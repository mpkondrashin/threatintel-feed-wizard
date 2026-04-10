package ui

import (
	"log"

	"fyne.io/fyne/v2"
)

// WizardController manages screen transitions and holds shared wizard state.
type WizardController struct {
	state   *WizardState
	window  fyne.Window
	size    fyne.Size
	screens []WizardScreen
	current int
}

// NewWizardController creates a WizardController with the given window, shared
// state, and ordered screen slice. The caller is responsible for wiring each
// screen's OnNext/OnBack callbacks to call controller.Next() / controller.Back().
func NewWizardController(window fyne.Window, state *WizardState, screens []WizardScreen, size fyne.Size) *WizardController {
	return &WizardController{
		state:   state,
		window:  window,
		size:    size,
		screens: screens,
		current: 0,
	}
}

// Next validates the current screen via OnLeave, advances to the next screen,
// sets the window content (which creates widgets), then calls OnEnter.
// If OnLeave returns an error the navigation is aborted.
func (c *WizardController) Next() {
	if c.current >= len(c.screens)-1 {
		return
	}

	if err := c.screens[c.current].OnLeave(c.state); err != nil {
		log.Printf("[wizard] Next blocked by OnLeave error on screen %d: %v", c.current, err)
		return
	}

	c.current++
	log.Printf("[wizard] Next → screen %d", c.current)
	c.window.SetContent(c.screens[c.current].Content())
	c.window.Resize(c.size)
	c.screens[c.current].OnEnter(c.state)
}

// Back navigates to the previous screen. OnLeave is called on the current
// screen but any error is ignored so the user can always go back.
func (c *WizardController) Back() {
	if c.current <= 0 {
		return
	}

	_ = c.screens[c.current].OnLeave(c.state)

	c.current--
	log.Printf("[wizard] Back → screen %d", c.current)
	c.window.SetContent(c.screens[c.current].Content())
	c.window.Resize(c.size)
	c.screens[c.current].OnEnter(c.state)
}

// Show initialises the wizard by setting the window content for the first
// screen and then calling OnEnter. Call this once after creating the controller.
func (c *WizardController) Show() {
	if len(c.screens) == 0 {
		return
	}
	log.Printf("[wizard] Show → screen %d", c.current)
	c.window.SetContent(c.screens[c.current].Content())
	c.window.Resize(c.size)
	c.screens[c.current].OnEnter(c.state)
}
