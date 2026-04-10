package api

import "encoding/json"

// Region represents a Trend Vision One regional API server.
type Region struct {
	Name    string
	BaseURL string
}

// Regions lists all nine supported Trend Vision One regional servers.
var Regions = []Region{
	{Name: "United States", BaseURL: "https://api.xdr.trendmicro.com"},
	{Name: "Germany", BaseURL: "https://api.eu.xdr.trendmicro.com"},
	{Name: "Singapore", BaseURL: "https://api.sg.xdr.trendmicro.com"},
	{Name: "Japan", BaseURL: "https://api.xdr.trendmicro.co.jp"},
	{Name: "Australia", BaseURL: "https://api.au.xdr.trendmicro.com"},
	{Name: "India", BaseURL: "https://api.in.xdr.trendmicro.com"},
	{Name: "United Arab Emirates", BaseURL: "https://api.mea.xdr.trendmicro.com"},
	{Name: "United Kingdom", BaseURL: "https://api.uk.xdr.trendmicro.com"},
	{Name: "Canada", BaseURL: "https://api.ca.xdr.trendmicro.com"},
}

// Indicator represents a STIX 2.1 Indicator object.
type Indicator struct {
	ID              string           `json:"id"`
	Name            string           `json:"name,omitempty"`
	Description     string           `json:"description,omitempty"`
	IndicatorTypes  []string         `json:"indicator_types,omitempty"`
	Pattern         string           `json:"pattern"`
	PatternType     string           `json:"pattern_type"`
	PatternVersion  string           `json:"pattern_version,omitempty"`
	Created         string           `json:"created"`
	Modified        string           `json:"modified"`
	ValidFrom       string           `json:"valid_from"`
	ValidUntil      string           `json:"valid_until,omitempty"`
	Confidence      *int             `json:"confidence,omitempty"`
	Labels          []string         `json:"labels,omitempty"`
	KillChainPhases []KillChainPhase `json:"kill_chain_phases,omitempty"`
	CreatedByRef    string           `json:"created_by_ref,omitempty"`
}

// KillChainPhase represents a STIX 2.1 kill chain phase.
type KillChainPhase struct {
	KillChainName string `json:"kill_chain_name"`
	PhaseName     string `json:"phase_name"`
}

// STIXBundle represents a STIX 2.1 bundle from the API response.
type STIXBundle struct {
	Type    string            `json:"type"`
	ID      string            `json:"id"`
	Objects []json.RawMessage `json:"objects"`
}

// STIXObjectType is used for initial type discrimination of STIX objects.
type STIXObjectType struct {
	Type string `json:"type"`
}

// FeedResponse represents the top-level API response from the feeds endpoint.
type FeedResponse struct {
	Bundle   STIXBundle `json:"bundle"`
	NextLink string     `json:"nextLink,omitempty"`
}
