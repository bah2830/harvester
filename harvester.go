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

	if h.settings.DarkTheme {
		h.app.Settings().SetTheme(theme.DarkTheme())
	} else {
		h.app.Settings().SetTheme(theme.LightTheme())
	}

	h.renderMainWindow()
	return h, nil
}

func (h *harvester) start() {
	interval := h.settings.RefreshInterval

	if err := h.refresh(); err != nil {
		log.Fatal(err)
	}

	tick := time.NewTicker(interval)
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

			if interval != h.settings.RefreshInterval {
				interval = h.settings.RefreshInterval
				tick = time.NewTicker(interval)
			}
		}
	}
}

func (h *harvester) refresh() error {
	issues, err := h.getUsersActiveIssues()
	if err != nil {
		return err
	}
	h.activeJiras = issues

	h.redraw()
	return nil
}

func (h *harvester) init() error {
	var settings string
	if err := h.db.QueryRow("select settings from settings").Scan(&settings); err != nil {
		return err
	}

	set, err := base64.StdEncoding.DecodeString(settings)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(set, &h.settings); err != nil {
		return err
	}

	if h.settings.Jira.URL != "" && h.settings.Jira.User != "" {
		tp := jira.BasicAuthTransport{
			Username: h.settings.Jira.User,
			Password: h.settings.Jira.Pass,
		}

		jiraClient, err := jira.NewClient(tp.Client(), h.settings.Jira.URL)
		if err != nil {
			return err
		}
		h.jiraClient = jiraClient
	}

	return nil
}

func (h *harvester) stop() {
	if err := h.saveSettings(); err != nil {
		log.Fatal(err)
	}

	for _, jiraTracker := range h.activeJiras {
		h.saveJiraTime(jiraTracker.Key, "stop")
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
