package ui

import (
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// DoneScreen is the final screen shown after a successful CSV export.
type DoneScreen struct {
	App         fyne.App
	statusLabel *widget.Label
	fileLink    *widget.Hyperlink
}

func (s *DoneScreen) Content() fyne.CanvasObject {
	s.statusLabel = widget.NewLabel("")
	s.statusLabel.Wrapping = fyne.TextWrapWord
	s.statusLabel.Alignment = fyne.TextAlignCenter

	s.fileLink = widget.NewHyperlink("", nil)
	s.fileLink.Alignment = fyne.TextAlignCenter

	quitBtn := widget.NewButton("Quit", func() {
		if s.App != nil {
			s.App.Quit()
		}
	})

	buttons := container.NewHBox(layout.NewSpacer(), quitBtn, layout.NewSpacer())

	topLabel := widget.NewLabel("Export complete.")
	topLabel.Wrapping = fyne.TextWrapWord
	topCenter := container.NewVBox(layout.NewSpacer(), topLabel, layout.NewSpacer())
	top := container.NewBorder(nil, nil, screenImage("image_6.png"), nil, topCenter)

	body := container.NewVBox(
		layout.NewSpacer(),
		s.statusLabel,
		s.fileLink,
		layout.NewSpacer(),
	)
	return padded(container.NewBorder(top, buttons, nil, nil, body))
}

func (s *DoneScreen) OnEnter(state *WizardState) {
	if s.statusLabel == nil {
		return
	}

	s.statusLabel.SetText(fmt.Sprintf(
		"%d indicators saved to:",
		state.SavedCount))

	// Show the filename as a clickable link that opens the file.
	absPath := state.SavedPath
	displayName := filepath.Base(absPath)
	// On Windows absPath is like C:\Users\..., which needs file:///C:/...
	urlPath := filepath.ToSlash(absPath)
	if !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
	}
	fileURL, err := url.Parse("file://" + urlPath)
	if err != nil {
		log.Printf("[done] could not parse file URL: %v", err)
		s.fileLink.SetText(displayName)
		s.fileLink.SetURL(nil)
	} else {
		s.fileLink.SetText(displayName)
		s.fileLink.SetURL(fileURL)
	}
}

func (s *DoneScreen) OnLeave(state *WizardState) error { return nil }
