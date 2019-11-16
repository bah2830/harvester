package main

import (
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/theme"
)

type harvester struct {
	app            fyne.App
	mainWindow     fyne.Window
	settingsWindow fyne.Window
	settings       settings
	changeCh       chan bool
}

type settings struct {
	refreshInterval time.Duration
	darkTheme       bool
	jira, harvest   settingsData
}

type settingsData struct {
	url, user, pass string
}

func newHarvester() *harvester {
	h := &harvester{
		app: app.New(),
		settings: settings{
			refreshInterval: defaultRefreshInterval,
			darkTheme:       true,
		},
		changeCh: make(chan bool),
	}

	if h.settings.darkTheme {
		h.app.Settings().SetTheme(theme.DarkTheme())
	} else {
		h.app.Settings().SetTheme(theme.LightTheme())
	}

	h.renderMainWindow()
	return h
}

func (h *harvester) start() {
	// @TODO: get data from database

	interval := h.settings.refreshInterval

	tick := time.NewTicker(interval)
	for {
		select {
		case <-tick.C:
			h.refresh()
		case <-h.changeCh:
			if interval != h.settings.refreshInterval {
				interval = h.settings.refreshInterval
				tick = time.NewTicker(interval)
			}
		}
	}
}

func (h *harvester) stop() {
	// @TODO: added shutdown code to save all currently data to the database before quiting
}
