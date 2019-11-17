package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"
	jira "github.com/andygrunwald/go-jira"
)

type harvester struct {
	app            fyne.App
	mainWindow     fyne.Window
	settingsWindow fyne.Window
	settings       settings
	changeCh       chan bool
	db             *sql.DB
	jiraClient     *jira.Client
	activeJiras    []jira.Issue
}

type settings struct {
	RefreshInterval time.Duration
	DarkTheme       bool
	Jira, Harvest   settingsData
}

type settingsData struct {
	URL, User, Pass string
}

func newHarvester(db *sql.DB) (*harvester, error) {
	h := &harvester{
		app: app.New(),
		db:  db,
		settings: settings{
			RefreshInterval: defaultRefreshInterval,
			DarkTheme:       true,
		},
		changeCh: make(chan bool),
	}

	if err := h.init(); err != nil {
		return nil, err
	}

	h.app.SetIcon(harvestIcon)

	if h.settings.DarkTheme {
		h.app.Settings().SetTheme(theme.DarkTheme())
	} else {
		h.app.Settings().SetTheme(theme.LightTheme())
	}

	h.renderMainWindow()
	return h, nil
}

func (h *harvester) start() {
	// Hold onto the last copy of settings to check for diffs
	previousSettings := h.settings

	// Start the purger to keep the database small
	go h.jiraPurger()

	if err := h.refresh(); err != nil {
		log.Fatal(err)
	}

	tick := time.NewTicker(previousSettings.RefreshInterval)
	for {
		select {
		case <-tick.C:
			if err := h.refresh(); err != nil {
				log.Print(err)
			}
		case <-h.changeCh:
			if err := h.saveSettings(); err != nil {
				log.Printf("ERROR: %s", err.Error())
			}

			// If refresh interval changed update the ticket
			if previousSettings.RefreshInterval != h.settings.RefreshInterval {
				tick = time.NewTicker(h.settings.RefreshInterval)
			}

			// If the jira credentials changed get a new client
			if h.settings.Jira.URL != previousSettings.Jira.URL ||
				h.settings.Jira.User != previousSettings.Jira.User ||
				h.settings.Jira.Pass != previousSettings.Jira.Pass {
				if err := h.getNewJiraClient(); err != nil {
					log.Print(err)
				}
			}

			previousSettings = h.settings
			h.refresh()
		}
	}
}

func (h *harvester) refresh() error {
	if h.jiraClient != nil {
		issues, err := h.getUsersActiveIssues()
		if err != nil {
			return err
		}
		h.activeJiras = issues
	}

	h.redraw()
	return nil
}

func (h *harvester) init() error {
	var settings string
	err := h.db.QueryRow("select settings from settings").Scan(&settings)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if settings == "" {
		return nil
	}

	set, err := base64.StdEncoding.DecodeString(settings)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(set, &h.settings); err != nil {
		return err
	}

	if h.settings.Jira.URL != "" && h.settings.Jira.User != "" {
		if err := h.getNewJiraClient(); err != nil {
			return err
		}
	}

	return nil
}

func (h *harvester) getNewJiraClient() error {
	tp := jira.BasicAuthTransport{
		Username: h.settings.Jira.User,
		Password: h.settings.Jira.Pass,
	}

	jiraClient, err := jira.NewClient(tp.Client(), h.settings.Jira.URL)
	if err != nil {
		return err
	}

	h.jiraClient = jiraClient
	return nil
}

func (h *harvester) stop() {
	if err := h.saveSettings(); err != nil {
		log.Fatal(err)
	}

	for _, jiraTracker := range h.activeJiras {
		h.saveJiraTime(jiraTracker, "stop")
	}
}

func (h *harvester) saveSettings() error {
	if _, err := h.db.Exec("delete from settings where id > 0"); err != nil {
		return err
	}

	settings, err := json.Marshal(h.settings)
	if err != nil {
		return err
	}

	base64Settings := base64.StdEncoding.EncodeToString(settings)
	if _, err = h.db.Exec("insert into settings (settings) values (?)", base64Settings); err != nil {
		return err
	}

	return nil
}
