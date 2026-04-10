package ui

import (
	"errors"
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"threatintel-feed-wizard/api"
	"threatintel-feed-wizard/credential"
)

type APIKeyScreen struct {
	OnNext func()
	OnBack func()

	store        credential.Store
	apiKeyEntry  *widget.Entry
	regionSelect *widget.Select
	errorLabel   *widget.Label
}

func NewAPIKeyScreen(store credential.Store) *APIKeyScreen {
	return &APIKeyScreen{store: store}
}

func (s *APIKeyScreen) Content() fyne.CanvasObject {
	instructions := widget.NewLabel(
		"Enter your Trend Vision One API key.\n" +
			"Go to Administration > API Keys to generate one.\n" +
			"The key needs the \"Threat Intelligence\" role.",
	)
	instructions.Wrapping = fyne.TextWrapWord

	s.apiKeyEntry = widget.NewEntry()
	s.apiKeyEntry.SetPlaceHolder("Paste your API key here")

	regionNames := make([]string, len(api.Regions))
	for i, r := range api.Regions {
		regionNames[i] = r.Name
	}
	s.regionSelect = widget.NewSelect(regionNames, nil)
	s.regionSelect.SetSelected(api.Regions[0].Name)

	s.errorLabel = widget.NewLabel("")
	s.errorLabel.Hide()

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
	header := container.NewBorder(nil, nil, screenImage("image_2.png"), nil, instrCenter)

	form := container.NewVBox(
		header,
		widget.NewLabel("API Key"),
		s.apiKeyEntry,
		widget.NewLabel("Region"),
		s.regionSelect,
		s.errorLabel,
	)

	buttons := container.NewHBox(backBtn, layout.NewSpacer(), nextBtn)
	body := container.NewVBox(layout.NewSpacer(), form, layout.NewSpacer())
	return padded(container.NewBorder(nil, buttons, nil, nil, body))
}

func (s *APIKeyScreen) OnEnter(state *WizardState) {
	if s.apiKeyEntry == nil {
		return
	}
	if key, err := s.store.LoadAPIKey(); err == nil && key != "" {
		log.Printf("[apikey] loaded saved API key (len=%d)", len(key))
		s.apiKeyEntry.SetText(key)
	}
	if regionName, err := s.store.LoadRegion(); err == nil && regionName != "" {
		log.Printf("[apikey] loaded saved region: %s", regionName)
		s.regionSelect.SetSelected(regionName)
	}
	s.errorLabel.SetText("")
	s.errorLabel.Hide()
}

func (s *APIKeyScreen) OnLeave(state *WizardState) error {
	key := s.apiKeyEntry.Text
	if key == "" {
		s.errorLabel.SetText("API key is required.")
		s.errorLabel.Show()
		return errors.New("API key is required")
	}
	if err := s.store.SaveAPIKey(key); err != nil {
		msg := fmt.Sprintf("Could not save API key securely: %v", err)
		s.errorLabel.SetText(msg)
		s.errorLabel.Show()
		return fmt.Errorf("save api key: %w", err)
	}

	selectedName := s.regionSelect.Selected
	region := api.Regions[0]
	for _, r := range api.Regions {
		if r.Name == selectedName {
			region = r
			break
		}
	}
	if err := s.store.SaveRegion(selectedName); err != nil {
		msg := fmt.Sprintf("Could not save region: %v", err)
		s.errorLabel.SetText(msg)
		s.errorLabel.Show()
		return fmt.Errorf("save region: %w", err)
	}

	state.APIKey = key
	state.Region = region
	log.Printf("[apikey] saved key (len=%d) region=%s", len(key), region.Name)
	s.errorLabel.SetText("")
	s.errorLabel.Hide()
	return nil
}
