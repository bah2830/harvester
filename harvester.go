package main

import (
	"log"
	"net/url"
	"sort"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/theme"
	jira "github.com/andygrunwald/go-jira"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type harvester struct {
	app           fyne.App
	mainWindow    fyne.Window
	settings      Settings
	changeCh      chan bool
	db            *gorm.DB
	jiraClient    *jira.Client
	harvestClient *HarvestClient
	harvestURL    *url.URL
	timers        TaskTimers
}

func NewHarvester(db *gorm.DB) (*harvester, error) {
	h := &harvester{
		app: app.New(),
		db:  db,
		settings: Settings{
			RefreshInterval: defaultRefreshInterval,
			DarkTheme:       true,
		},
		changeCh: make(chan bool),
		timers:   TaskTimers{},
	}

	if err := h.init(); err != nil {
		return nil, errors.WithMessage(err, "harvester init error")
	}

	h.renderMainWindow()
	return h, nil
}

func (h *harvester) start() {
	// Hold onto the last copy of settings to check for diffs
	previousSettings := h.settings

	// Start the purger to keep the database small
	go StartJiraPurger(h.db)

	tick := time.NewTicker(previousSettings.RefreshInterval)
	for {
		select {
		case <-tick.C:
			if err := h.refresh(); err != nil {
				log.Print(err)
			}
		case <-h.changeCh:
			if err := h.settings.Save(h.db); err != nil {
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
	go func() {
		defer h.redraw()

		// Get any current timers running in the local database
		timers, err := GetActiveTimers(h.db, h.jiraClient, h.harvestClient)
		if err != nil {
			dialog.ShowError(err, h.mainWindow)
			return
		}

		// Add jira details to any timers
		if h.jiraClient != nil {
			issues, err := h.getUsersActiveIssues()
			if err != nil {
				dialog.ShowError(err, h.mainWindow)
				return
			}

			// Add jiras to timers
			for _, jira := range issues {
				jiraIssue := jira
				timer, err := timers.GetByKey(jira.Key)
				if err != nil {
					timer = &TaskTimer{
						Key:  jira.Key,
						jira: &jiraIssue,
					}
					timers = append(timers, timer)
					continue
				}
				timer.jira = &jiraIssue
			}
		}

		// Add harvest details to timers
		if h.harvestClient != nil {
			if h.harvestURL == nil {
				company, err := h.harvestClient.getCompany()
				if err != nil {
					dialog.ShowError(err, h.mainWindow)
					return
				}
				u, _ := url.Parse(*company.BaseUri)
				h.harvestURL = u
			}

			tasks, err := h.harvestClient.getUserProjects()
			if err != nil {
				dialog.ShowError(err, h.mainWindow)
				return
			}

			// Add harvest projects to timers
			for _, task := range tasks {
				harvestTask := *task
				timer, err := timers.GetByKey(*task.Project.Code)
				if err != nil {
					timer = &TaskTimer{
						Key:     *task.Project.Code,
						harvest: &harvestTask,
					}
					timers = append(timers, timer)
					continue
				}
				timer.harvest = &harvestTask
			}
		}

		sort.SliceStable(timers, func(a, b int) bool {
			return sort.StringsAreSorted([]string{timers[a].Key, timers[b].Key})
		})
		h.timers = timers
	}()

	return nil
}

func (h *harvester) init() error {
	settings, err := GetSettings(h.db)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return errors.WithMessage(err, "error getting settings")
	}
	if settings != nil {
		h.settings = *settings
	}

	// Setup the jira client
	if h.settings.Jira.URL != "" && h.settings.Jira.User != "" {
		if err := h.getNewJiraClient(); err != nil {
			return err
		}
	}

	// Setup the harvest client
	if h.settings.Harvest.User != "" && h.settings.Harvest.Pass != "" {
		if err := h.getNewHarvestClient(); err != nil {
			return err
		}
	}

	// Setup the application look and feel
	if h.settings.DarkTheme {
		h.app.Settings().SetTheme(defaultTheme())
	} else {
		h.app.Settings().SetTheme(theme.LightTheme())
	}

	return nil
}

func (h *harvester) stop() {
	if err := h.settings.Save(h.db); err != nil {
		log.Fatal(err)
	}

	for _, timer := range h.timers {
		if err := timer.Stop(h.db, h.harvestClient); err != nil {
			log.Println(err)
		}
	}
}
