package ui

import (
	"time"

	"threatintel-feed-wizard/api"
)

// WizardState holds shared state across all wizard screens.
type WizardState struct {
	APIKey     string
	Region     api.Region
	StartTime  time.Time
	EndTime    time.Time
	Indicators []api.Indicator
	SavedPath  string
	SavedCount int
}
