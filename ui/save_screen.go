package ui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"threatintel-feed-wizard/credential"
	csvpkg "threatintel-feed-wizard/csv"
)

type SaveScreen struct {
	OnBack func()
	OnDone func() // called after successful save to navigate to done screen

	window      fyne.Window
	store       credential.Store
	pathEntry   *widget.Entry
	statusLabel *widget.Label
	state       *WizardState
}

func NewSaveScreen(window fyne.Window, store credential.Store) *SaveScreen {
	return &SaveScreen{window: window, store: store}
}

func (s *SaveScreen) Content() fyne.CanvasObject {
	s.pathEntry = widget.NewEntry()
	s.pathEntry.SetPlaceHolder("File path")

	s.statusLabel = widget.NewLabel("")
	s.statusLabel.Wrapping = fyne.TextWrapWord
	s.statusLabel.Alignment = fyne.TextAlignCenter

	browseBtn := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil {
				s.statusLabel.SetText(fmt.Sprintf("Error: %s", err.Error()))
				return
			}
			if dir == nil {
				return
			}
			// Combine selected folder with the current filename.
			filename := filepath.Base(s.pathEntry.Text)
			if filename == "" || filename == "." {
				filename = "indicators.csv"
			}
			s.pathEntry.SetText(filepath.Join(dir.Path(), filename))
		}, s.window)
	})

	saveBtn := widget.NewButton("Save", func() { s.save() })
	backBtn := widget.NewButton("Back", func() {
		if s.OnBack != nil {
			s.OnBack()
		}
	})

	pathRow := container.NewBorder(nil, nil, nil, browseBtn, s.pathEntry)

	saveLabel := widget.NewLabel("Save indicators to CSV")
	saveLabel.Wrapping = fyne.TextWrapWord
	labelCenter := container.NewVBox(layout.NewSpacer(), saveLabel, layout.NewSpacer())
	top := container.NewBorder(nil, nil, screenImage("image_5.png"), nil, labelCenter)

	form := container.NewVBox(
		pathRow,
		s.statusLabel,
	)

	buttons := container.NewHBox(backBtn, layout.NewSpacer(), saveBtn)
	body := container.NewVBox(layout.NewSpacer(), form, layout.NewSpacer())
	return padded(container.NewBorder(top, buttons, nil, nil, body))
}

func (s *SaveScreen) OnEnter(state *WizardState) {
	s.state = state
	if s.pathEntry == nil {
		return
	}
	now := time.Now()
	base := fmt.Sprintf("TAITI_%02d%02d%02d", now.Year()%100, now.Month(), now.Day())
	s.pathEntry.SetText(dedupFilename(base, ".csv"))
	s.statusLabel.SetText("")
}

// dedupFilename returns "name.ext" if it doesn't exist, otherwise
// "name (1).ext", "name (2).ext", etc.
func dedupFilename(name, ext string) string {
	candidate := name + ext
	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		return candidate
	}
	for i := 1; ; i++ {
		candidate = fmt.Sprintf("%s (%d)%s", name, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

func (s *SaveScreen) OnLeave(state *WizardState) error { return nil }

func (s *SaveScreen) save() {
	path := s.pathEntry.Text
	if path == "" {
		s.statusLabel.SetText("Please enter a file path.")
		return
	}

	// Resolve to absolute path for the done screen link.
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	log.Printf("[save] writing to %s", absPath)
	file, err := os.Create(absPath)
	if err != nil {
		s.statusLabel.SetText(fmt.Sprintf("Error: could not create file: %s", err.Error()))
		return
	}
	defer file.Close()

	if err := csvpkg.WriteCSV(file, s.state.Indicators); err != nil {
		s.statusLabel.SetText(fmt.Sprintf("Error: could not write CSV: %s", err.Error()))
		return
	}

	count := len(s.state.Indicators)
	log.Printf("[save] wrote %d indicators to %s", count, absPath)

	endTS := s.state.EndTime.UTC().Format(time.RFC3339)
	if err := s.store.SaveLastSync(endTS); err != nil {
		log.Printf("[save] warning: could not save last sync time: %v", err)
	} else {
		log.Printf("[save] saved last sync time: %s", endTS)
	}

	s.state.SavedPath = absPath
	s.state.SavedCount = count
	if s.OnDone != nil {
		s.OnDone()
	}
}
