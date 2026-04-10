package ui

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"threatintel-feed-wizard/credential"
)

type DateTimeScreen struct {
	OnNext func()
	OnBack func()

	store      credential.Store
	datePicker *widget.DateEntry
	timeEntry  *widget.Entry
	errorLabel *widget.Label
}

func NewDateTimeScreen(store credential.Store) *DateTimeScreen {
	return &DateTimeScreen{store: store}
}

func (s *DateTimeScreen) Content() fyne.CanvasObject {
	instructions := widget.NewLabel(
		"Select the start date and time for the indicator query.\n" +
			"Only indicators created after this time will be downloaded.",
	)
	instructions.Wrapping = fyne.TextWrapWord

	s.datePicker = widget.NewDateEntry()
	s.timeEntry = widget.NewEntry()
	s.timeEntry.SetPlaceHolder("HH:MM (UTC, 24h)")

	s.errorLabel = widget.NewLabel("")

	nextBtn := widget.NewButton("Next", func() {
		if s.OnNext != nil {
			s.OnNext()
		}
	})
	backBtn := widget.NewButton("Back", func() {
		if s.OnBack != nil {
			s.OnBack()
		}
	})

	instrCenter := container.NewVBox(layout.NewSpacer(), instructions, layout.NewSpacer())
	header := container.NewBorder(nil, nil, screenImage("image_3.png"), nil, instrCenter)

	form := container.NewVBox(
		header,
		widget.NewLabel("Start Date"),
		s.datePicker,
		widget.NewLabel("Start Time (UTC)"),
		s.timeEntry,
		s.errorLabel,
	)

	buttons := container.NewHBox(backBtn, layout.NewSpacer(), nextBtn)
	body := container.NewVBox(layout.NewSpacer(), form, layout.NewSpacer())
	return padded(container.NewBorder(nil, buttons, nil, nil, body))
}

func (s *DateTimeScreen) OnEnter(state *WizardState) {
	if s.datePicker == nil {
		return
	}

	var defaultTime time.Time
	if ts, err := s.store.LoadLastSync(); err == nil && ts != "" {
		log.Printf("[datetime] loaded last sync: %s", ts)
		defaultTime, _ = time.Parse(time.RFC3339, ts)
	}
	if defaultTime.IsZero() {
		defaultTime = time.Now().UTC().Add(-7 * 24 * time.Hour)
	}

	s.datePicker.SetDate(&defaultTime)
	s.timeEntry.SetText(fmt.Sprintf("%02d:%02d", defaultTime.Hour(), defaultTime.Minute()))
	s.errorLabel.SetText("")
}

func (s *DateTimeScreen) OnLeave(state *WizardState) error {
	if s.datePicker.Date == nil {
		s.errorLabel.SetText("Please select a date.")
		return errors.New("date is required")
	}

	timeText := strings.TrimSpace(s.timeEntry.Text)
	hour, minute := 0, 0
	if timeText != "" {
		parts := strings.Split(timeText, ":")
		if len(parts) != 2 {
			s.errorLabel.SetText("Time format must be HH:MM")
			return errors.New("invalid time format")
		}
		var err error
		hour, err = strconv.Atoi(parts[0])
		if err != nil || hour < 0 || hour > 23 {
			s.errorLabel.SetText("Hour must be 00-23")
			return errors.New("invalid hour")
		}
		minute, err = strconv.Atoi(parts[1])
		if err != nil || minute < 0 || minute > 59 {
			s.errorLabel.SetText("Minute must be 00-59")
			return errors.New("invalid minute")
		}
	}

	d := *s.datePicker.Date
	state.StartTime = time.Date(d.Year(), d.Month(), d.Day(), hour, minute, 0, 0, time.UTC)
	state.EndTime = time.Now().UTC()
	log.Printf("[datetime] range: %s → %s", state.StartTime.Format(time.RFC3339), state.EndTime.Format(time.RFC3339))

	s.errorLabel.SetText("")
	return nil
}
