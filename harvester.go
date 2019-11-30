package main

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"sort"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/bah2830/webview"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type harvester struct {
	mainWindow     webview.WebView
	infoWindow     webview.WebView
	settingsWindow webview.WebView
	Settings       *Settings `json:"settings"`
	changeCh       chan bool
	db             *gorm.DB
	jiraClient     *jira.Client
	harvestClient  *HarvestClient
	harvestURL     *url.URL
	Timers         *Timers `json:"timers"`
	listener       net.Listener
	debug          bool
}

type Timers struct {
	Tasks TaskTimers `json:"tasks"`
}

func NewHarvester(db *gorm.DB, debug bool) (*harvester, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	go http.Serve(ln, http.FileServer(http.Dir("resources")))

	h := &harvester{
		db: db,
		Settings: &Settings{
			RefreshInterval: defaultRefreshInterval,
		},
		changeCh: make(chan bool),
		Timers: &Timers{
			Tasks: TaskTimers{},
		},
		listener: ln,
		debug:    debug,
	}

	if err := h.init(); err != nil {
		return nil, errors.WithMessage(err, "harvester init error")
	}

	h.renderMainWindow()

	return h, nil
}

func (h *harvester) Start() {
	// Hold onto the last copy of settings to check for diffs
	previousSettings := h.Settings

	// Start the purger to keep the database small
	go StartJiraPurger(h.db)

	tick := time.NewTicker(previousSettings.RefreshInterval)
	for {
		select {
		case <-tick.C:
			if err := h.Refresh(); err != nil {
				h.sendErr(err)
			}
		case <-h.changeCh:
			if err := h.Settings.Save(h.db); err != nil {
				h.sendErr(err)
				continue
			}

			// If refresh interval changed update the ticket
			if previousSettings.RefreshInterval != h.Settings.RefreshInterval {
				tick = time.NewTicker(h.Settings.RefreshInterval)
			}

			// If the jira credentials changed get a new client
			if h.Settings.Jira.URL != previousSettings.Jira.URL ||
				h.Settings.Jira.User != previousSettings.Jira.User ||
				h.Settings.Jira.Pass != previousSettings.Jira.Pass {
				if err := h.getNewJiraClient(); err != nil {
					h.sendErr(err)
					continue
				}
			}

			previousSettings = h.Settings
			if err := h.Refresh(); err != nil {
				h.sendErr(err)
			}
		}
	}
}

func (h *harvester) Refresh() error {
	// Get any current timers running in the local database
	timers, err := GetActiveTimers(h.db, h.jiraClient, h.harvestClient)
	if err != nil {
		return err
	}

	// Add jira details to any timers
	if h.jiraClient != nil {
		issues, err := h.getUsersActiveIssues()
		if err != nil {
			return err
		}

		// Add jiras to timers
		for _, jira := range issues {
			jiraIssue := jira
			timer, err := timers.GetByKey(jira.Key)
			if err != nil {
				timer = &TaskTimer{
					Key:  jira.Key,
					Jira: &jiraIssue,
				}
				timers = append(timers, timer)
				continue
			}
			timer.Jira = &jiraIssue
		}
	}

	// Add harvest details to timers
	if h.harvestClient != nil {
		if h.harvestURL == nil {
			company, err := h.harvestClient.getCompany()
			if err != nil {
				return err
			}
			u, _ := url.Parse(*company.BaseUri)
			h.harvestURL = u
		}

		tasks, err := h.harvestClient.getUserProjects()
		if err != nil {
			return err
		}

		// Add harvest projects to timers
		for _, task := range tasks {
			harvestTask := *task
			timer, err := timers.GetByKey(*task.Project.Code)
			if err != nil {
				timer = &TaskTimer{
					Key:     *task.Project.Code,
					Harvest: &harvestTask,
				}
				timers = append(timers, timer)
				continue
			}
			timer.Harvest = &harvestTask
		}
	}

	sort.SliceStable(timers, func(a, b int) bool {
		return sort.StringsAreSorted([]string{timers[a].Key, timers[b].Key})
	})
	h.Timers.Tasks = timers

	if h.mainWindow != nil {
		h.sendTimers()
	}

	return nil
}

func (h *harvester) init() error {
	settings, err := GetSettings(h.db)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return errors.WithMessage(err, "error getting settings")
	}
	if settings != nil {
		h.Settings = settings
	}

	// Setup the jira client
	if h.Settings.Jira.URL != "" && h.Settings.Jira.User != "" {
		if err := h.getNewJiraClient(); err != nil {
			return err
		}
	}

	// Setup the harvest client
	if h.Settings.Harvest.User != "" && h.Settings.Harvest.Pass != "" {
		if err := h.getNewHarvestClient(); err != nil {
			return err
		}
	}

	return nil
}

func (h *harvester) Stop() {
	if err := h.Settings.Save(h.db); err != nil {
		log.Fatal(err)
	}

	for _, timer := range h.Timers.Tasks {
		if err := timer.Stop(h.db, h.harvestClient); err != nil {
			log.Println(err)
		}
	}

	h.listener.Close()

	if h.infoWindow != nil {
		h.infoWindow.Terminate()
	}

	h.mainWindow.Terminate()
}
