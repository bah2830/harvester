package main

import (
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
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
		},
		changeCh: make(chan bool),
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
