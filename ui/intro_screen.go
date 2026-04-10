package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// IntroScreen is the first wizard step.
type IntroScreen struct {
	OnNext func()
}

func (s *IntroScreen) Content() fyne.CanvasObject {
	description := widget.NewLabel(
		"ThreatIntel Feed Wizard downloads threat intelligence data " +
			"from Trend Vision One and saves Indicators of Compromise (IoCs) as a CSV file.",
	)
	description.Wrapping = fyne.TextWrapWord

	clickNext := widget.NewLabel("Click Next to get started.")
	clickNext.Alignment = fyne.TextAlignCenter

	nextBtn := widget.NewButton("Next", func() {
		if s.OnNext != nil {
			s.OnNext()
		}
	})

	buttons := container.NewHBox(layout.NewSpacer(), nextBtn)
	descCenter := container.NewVBox(layout.NewSpacer(), description, layout.NewSpacer())
	header := container.NewBorder(nil, nil, screenImage("image_1.png"), nil, descCenter)
	body := container.NewVBox(header, layout.NewSpacer(), clickNext, layout.NewSpacer())
	return padded(container.NewBorder(nil, buttons, nil, nil, body))
}

func (s *IntroScreen) OnEnter(state *WizardState) {}

func (s *IntroScreen) OnLeave(state *WizardState) error { return nil }
