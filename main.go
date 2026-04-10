package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"threatintel-feed-wizard/api"
	"threatintel-feed-wizard/credential"
	csvpkg "threatintel-feed-wizard/csv"
	"threatintel-feed-wizard/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

//go:embed icon.png
var iconData []byte

func main() {
	cliMode := flag.Bool("cli", false, "Run in CLI mode (no GUI)")
	apiKey := flag.String("key", "", "API key (CLI mode)")
	region := flag.String("region", "", "Region name (CLI mode, default: saved region or United States)")
	startFlag := flag.String("start", "", "Start date-time in ISO 8601 UTC (CLI mode, default: last sync or 7 days ago)")
	output := flag.String("out", "", "Output CSV file path (CLI mode)")
	flag.Parse()

	if *cliMode {
		runCLI(*apiKey, *region, *startFlag, *output)
		return
	}

	runGUI()
}

func runCLI(apiKey, regionName, startStr, output string) {
	store := &credential.KeyringStore{}

	if apiKey == "" {
		if saved, err := store.LoadAPIKey(); err == nil && saved != "" {
			apiKey = saved
			log.Println("[cli] loaded API key from keyring")
		} else {
			fmt.Fprintln(os.Stderr, "Error: --key is required (or save one via the GUI first)")
			os.Exit(1)
		}
	}

	if regionName == "" {
		if saved, err := store.LoadRegion(); err == nil && saved != "" {
			regionName = saved
			log.Printf("[cli] loaded region from keyring: %s", regionName)
		} else {
			regionName = "United States"
		}
	}

	var region api.Region
	found := false
	for _, r := range api.Regions {
		if r.Name == regionName {
			region = r
			found = true
			break
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "Error: unknown region %q\nAvailable regions:\n", regionName)
		for _, r := range api.Regions {
			fmt.Fprintf(os.Stderr, "  - %s\n", r.Name)
		}
		os.Exit(1)
	}

	var startTime time.Time
	if startStr != "" {
		var err error
		startTime, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid --start format: %v\nUse ISO 8601 UTC, e.g. 2024-01-01T00:00:00Z\n", err)
			os.Exit(1)
		}
	} else if ts, err := store.LoadLastSync(); err == nil && ts != "" {
		startTime, _ = time.Parse(time.RFC3339, ts)
		log.Printf("[cli] using last sync as start: %s", ts)
	} else {
		startTime = time.Now().UTC().Add(-7 * 24 * time.Hour)
		log.Println("[cli] no last sync found, defaulting to 7 days ago")
	}
	endTime := time.Now().UTC()

	log.Printf("[cli] fetching from %s (%s) range %s → %s", region.Name, region.BaseURL,
		startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))
	engine := api.NewDownloadEngine(apiKey, region.BaseURL)
	indicators, err := engine.FetchIndicators(context.Background(), startTime, endTime, func(count int) {
		log.Printf("[cli] progress: %d indicators", count)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Download error: %v\n", err)
		os.Exit(1)
	}
	log.Printf("[cli] downloaded %d indicators", len(indicators))

	if output == "" {
		now := time.Now()
		base := fmt.Sprintf("TAITI_%02d%02d%02d", now.Year()%100, now.Month(), now.Day())
		output = dedupFilename(base, ".csv")
	}
	f, err := os.Create(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := csvpkg.WriteCSV(f, indicators); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Saved %d indicators to %s\n", len(indicators), output)

	// Persist end time for next run.
	if err := store.SaveLastSync(endTime.Format(time.RFC3339)); err != nil {
		log.Printf("[cli] warning: could not save last sync: %v", err)
	}
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

func runGUI() {
	a := app.NewWithID("com.threatintel.feedwizard")
	a.SetIcon(fyne.NewStaticResource("icon.png", iconData))
	w := a.NewWindow("ThreatIntel Feed Wizard")
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(560, 380))

	store := &credential.KeyringStore{}
	state := &ui.WizardState{}

	introScreen := &ui.IntroScreen{}
	apiKeyScreen := ui.NewAPIKeyScreen(store)
	dateTimeScreen := ui.NewDateTimeScreen(store)
	downloadScreen := &ui.DownloadScreen{}
	saveScreen := ui.NewSaveScreen(w, store)
	doneScreen := &ui.DoneScreen{App: a}

	screens := []ui.WizardScreen{introScreen, apiKeyScreen, dateTimeScreen, downloadScreen, saveScreen, doneScreen}
	windowSize := fyne.NewSize(560, 380)
	ctrl := ui.NewWizardController(w, state, screens, windowSize)

	introScreen.OnNext = ctrl.Next
	apiKeyScreen.OnNext = ctrl.Next
	apiKeyScreen.OnBack = ctrl.Back
	dateTimeScreen.OnNext = ctrl.Next
	dateTimeScreen.OnBack = ctrl.Back
	downloadScreen.OnNext = ctrl.Next
	downloadScreen.OnBack = ctrl.Back
	saveScreen.OnBack = ctrl.Back
	saveScreen.OnDone = ctrl.Next

	ctrl.Show()
	w.ShowAndRun()
}
