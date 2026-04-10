package ui

import "fyne.io/fyne/v2"

// WizardScreen defines the interface that each wizard step must implement.
type WizardScreen interface {
	// Content returns the Fyne canvas object tree for this screen.
	Content() fyne.CanvasObject
	// OnEnter is called when the wizard navigates to this screen.
	OnEnter(state *WizardState)
	// OnLeave is called when the wizard navigates away from this screen.
	// It returns an error if the screen cannot be left (e.g. validation failure).
	OnLeave(state *WizardState) error
}
