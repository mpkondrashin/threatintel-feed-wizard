package ui

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"threatintel-feed-wizard/api"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type DownloadScreen struct {
	OnNext func()
	OnBack func()

	headerLabel *widget.Label
	statusLabel *widget.Label
	nextBtn     *widget.Button
	cancel      context.CancelFunc
}

func (s *DownloadScreen) Content() fyne.CanvasObject {
	s.headerLabel = widget.NewLabel("")
	s.headerLabel.Wrapping = fyne.TextWrapWord

	s.statusLabel = widget.NewLabel("Preparing download...")
	s.statusLabel.Wrapping = fyne.TextWrapWord
	s.statusLabel.Alignment = fyne.TextAlignCenter

	s.nextBtn = widget.NewButton("Next", func() {
		if s.OnNext != nil {
			s.OnNext()
		}
	})
	s.nextBtn.Disable()

	backBtn := widget.NewButton("Back", func() {
		if s.OnBack != nil {
			s.OnBack()
		}
	})

	headerCenter := container.NewVBox(layout.NewSpacer(), s.headerLabel, layout.NewSpacer())
	header := container.NewBorder(nil, nil, screenImage("image_4.png"), nil, headerCenter)

	buttons := container.NewHBox(backBtn, layout.NewSpacer(), s.nextBtn)
	body := container.NewVBox(header, layout.NewSpacer(), s.statusLabel, layout.NewSpacer())
	return padded(container.NewBorder(nil, buttons, nil, nil, body))
}

func (s *DownloadScreen) OnEnter(state *WizardState) {
	if s.statusLabel == nil {
		log.Println("[download] OnEnter: statusLabel is nil, skipping")
		return
	}

	days := int(math.Ceil(time.Since(state.StartTime).Hours() / 24))
	s.headerLabel.SetText(fmt.Sprintf("Downloading indicators for the last %d days.", days))

	s.nextBtn.Disable()
	s.statusLabel.SetText("Downloading... 0 indicators so far")

	log.Printf("[download] starting fetch: region=%s baseURL=%s start=%s end=%s",
		state.Region.Name, state.Region.BaseURL,
		state.StartTime.Format("2006-01-02T15:04:05Z"),
		state.EndTime.Format("2006-01-02T15:04:05Z"))
	engine := api.NewDownloadEngine(state.APIKey, state.Region.BaseURL)

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go func() {
		indicators, err := engine.FetchIndicators(ctx, state.StartTime, state.EndTime, func(count int) {
			log.Printf("[download] progress: %d indicators", count)
			fyne.Do(func() {
				s.statusLabel.SetText(fmt.Sprintf("Downloading... %d indicators so far", count))
			})
		})

		if err != nil {
			var partial *api.PartialError
			if errors.As(err, &partial) && len(partial.Indicators) > 0 {
				log.Printf("[download] partial: %d indicators, cause: %v", len(partial.Indicators), partial.Cause)
				state.Indicators = partial.Indicators
				fyne.Do(func() {
					s.statusLabel.SetText(fmt.Sprintf(
						"⚠ Partial download: %d indicators.\n%s\nYou can save what was downloaded or go back and retry.",
						len(partial.Indicators), partial.Cause))
					s.nextBtn.Enable()
				})
				return
			}
			log.Printf("[download] error: %v", err)
			fyne.Do(func() {
				s.statusLabel.SetText(fmt.Sprintf("Download failed: %s", err.Error()))
			})
			return
		}

		log.Printf("[download] complete: %d indicators", len(indicators))
		state.Indicators = indicators
		fyne.Do(func() {
			s.statusLabel.SetText(fmt.Sprintf("Download complete: %d indicators", len(indicators)))
			s.nextBtn.Enable()
		})
	}()
}

func (s *DownloadScreen) OnLeave(state *WizardState) error {
	if s.cancel != nil {
		log.Println("[download] cancelling in-flight download")
		s.cancel()
		s.cancel = nil
	}
	return nil
}
