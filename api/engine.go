package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const maxRetries = 3

// OnProgress is called after each page with the running total of indicators.
type OnProgress func(indicatorCount int)

// PartialError indicates the download completed partially. Indicators contains
// what was fetched before the error occurred.
type PartialError struct {
	Indicators []Indicator
	Cause      error
}

func (e *PartialError) Error() string {
	return fmt.Sprintf("partial download (%d indicators): %v", len(e.Indicators), e.Cause)
}

func (e *PartialError) Unwrap() error {
	return e.Cause
}

// DownloadEngine fetches threat intel feeds with pagination.
type DownloadEngine struct {
	Client  *http.Client
	APIKey  string
	BaseURL string
}

// NewDownloadEngine creates a DownloadEngine with a 30-second HTTP timeout.
func NewDownloadEngine(apiKey, baseURL string) *DownloadEngine {
	return &DownloadEngine{
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
}

// FetchIndicators retrieves all Indicator objects across paginated responses.
// On rate-limit (429) it retries with exponential backoff. If retries are
// exhausted it returns a *PartialError containing the indicators fetched so far.
func (e *DownloadEngine) FetchIndicators(ctx context.Context, startTime, endTime time.Time, onProgress OnProgress) ([]Indicator, error) {
	url := fmt.Sprintf("%s/v3.0/threatintel/feedIndicators?indicatorObjectFormat=stixBundle&top=10000&startDateTime=%s&endDateTime=%s",
		e.BaseURL,
		startTime.UTC().Format(time.RFC3339),
		endTime.UTC().Format(time.RFC3339),
	)

	var indicators []Indicator
	page := 0

	for url != "" {
		page++
		log.Printf("[engine] page %d: GET %s", page, url)

		feedResp, err := e.fetchPageWithRetry(ctx, url, page)
		if err != nil {
			// Return partial results if we already have some indicators.
			if len(indicators) > 0 {
				log.Printf("[engine] error on page %d with %d indicators collected, returning partial", page, len(indicators))
				return nil, &PartialError{Indicators: indicators, Cause: err}
			}
			return nil, err
		}

		pageIndicators := 0
		for _, raw := range feedResp.Bundle.Objects {
			var objType STIXObjectType
			if err := json.Unmarshal(raw, &objType); err != nil {
				continue
			}
			if objType.Type != "indicator" {
				continue
			}
			var ind Indicator
			if err := json.Unmarshal(raw, &ind); err != nil {
				return nil, fmt.Errorf("unmarshaling indicator: %w", err)
			}
			indicators = append(indicators, ind)
			pageIndicators++
		}

		log.Printf("[engine] page %d: %d indicators (total %d), hasNextLink=%v",
			page, pageIndicators, len(indicators), feedResp.NextLink != "")

		if onProgress != nil {
			onProgress(len(indicators))
		}

		url = feedResp.NextLink
	}

	log.Printf("[engine] done: %d pages, %d total indicators", page, len(indicators))
	return indicators, nil
}

// fetchPageWithRetry fetches a single page, retrying on HTTP 429 with
// exponential backoff. It respects the Retry-After header when present.
func (e *DownloadEngine) fetchPageWithRetry(ctx context.Context, url string, page int) (*FeedResponse, error) {
	backoff := 2 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+e.APIKey)

		resp, err := e.Client.Do(req)
		if err != nil {
			log.Printf("[engine] request error: %v", err)
			return nil, classifyNetworkError(err)
		}

		log.Printf("[engine] page %d attempt %d: HTTP %d", page, attempt+1, resp.StatusCode)

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()

			if attempt == maxRetries {
				return nil, classifyHTTPError(resp.StatusCode)
			}

			wait := backoff
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, err := time.ParseDuration(ra + "s"); err == nil {
					wait = secs
				}
			}

			log.Printf("[engine] rate limited, waiting %v before retry", wait)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
			backoff *= 2
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, classifyHTTPError(resp.StatusCode)
		}

		var feedResp FeedResponse
		if err := json.NewDecoder(resp.Body).Decode(&feedResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decoding response: %w", err)
		}
		resp.Body.Close()
		return &feedResp, nil
	}

	return nil, fmt.Errorf("exhausted retries for page %d", page)
}

// classifyHTTPError maps HTTP status codes to descriptive error messages.
func classifyHTTPError(statusCode int) error {
	switch statusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("invalid request: the server could not process the request (HTTP 400)")
	case http.StatusForbidden:
		return fmt.Errorf("access denied: verify that your API key is correct and has the required role assigned (HTTP 403)")
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limit exceeded: too many requests, please retry later (HTTP 429)")
	case http.StatusInternalServerError:
		return fmt.Errorf("server error: the Trend Vision One server encountered an internal error (HTTP 500)")
	default:
		return fmt.Errorf("unexpected status code: %d", statusCode)
	}
}

// classifyNetworkError inspects the error from http.Client.Do and returns
// a descriptive message for dial and timeout failures.
func classifyNetworkError(err error) error {
	if err == nil {
		return nil
	}

	// Check for timeout (net.Error with Timeout() == true).
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("connection timed out: the server did not respond in time, please check your network and try again")
	}

	// Check for dial errors (DNS resolution, connection refused, etc.).
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return fmt.Errorf("connection failed: unable to reach the server, please check your network connectivity")
	}

	// Fallback: check error message for common network keywords.
	msg := err.Error()
	if strings.Contains(msg, "dial") || strings.Contains(msg, "connect") {
		return fmt.Errorf("connection failed: unable to reach the server, please check your network connectivity")
	}
	if strings.Contains(msg, "timeout") {
		return fmt.Errorf("connection timed out: the server did not respond in time, please check your network and try again")
	}

	return fmt.Errorf("network error: %w", err)
}
