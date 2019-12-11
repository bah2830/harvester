package harvester

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/asticode/go-astilectron"
	astiptr "github.com/asticode/go-astitools/ptr"
	"github.com/bah2830/harvester/pkg/assets"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type harvester struct {
	app           *astilectron.Astilectron
	menu          *astilectron.Menu
	mainWindow    *Window
	Settings      *Settings `json:"settings"`
	changeCh      chan bool
	db            *gorm.DB
	jiraClient    *jira.Client
	harvestClient *HarvestClient
	harvestURL    *url.URL
	Timers        *Timers `json:"timers"`
	CurrentTimer  *TaskTimer
	listener      net.Listener
	debug         bool
}

type Timers struct {
	Tasks TaskTimers `json:"tasks"`
}

func NewHarvester(db *gorm.DB) (*harvester, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	go http.Serve(ln, http.FileServer(assets.AssetFile()))

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Unable to get user home directory", err)
	}

	harvesterDir := home + "/.harvester"
	prepareIcons(harvesterDir, ln)

	app, err := astilectron.New(astilectron.Options{
		AppName:            "Harvester",
		AppIconDefaultPath: harvesterDir + "/icon.png",
		AppIconDarwinPath:  harvesterDir + "/icon.icns",
		DataDirectoryPath:  harvesterDir + "/electron",
	})
	if err != nil {
		return nil, err
	}

	h := &harvester{
		app:      app,
		db:       db,
		Settings: &Settings{},
		changeCh: make(chan bool),
		Timers: &Timers{
			Tasks: TaskTimers{},
		},
		listener: ln,
	}

	if err := h.init(); err != nil {
		return nil, errors.WithMessage(err, "harvester init error")
	}

	if err := h.app.Start(); err != nil {
		return nil, err
	}

	h.createMenu(harvesterDir + "/timer.png")

	if err := h.renderMainWindow(); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *harvester) Start() {
	// Hold onto the last copy of settings to check for diffs
	previousSettings := *h.Settings

	// Start the purger to keep the database small
	go StartJiraPurger(h.db)

	if err := h.Refresh(); err != nil {
		h.sendErr(err)
	}

	tick := time.NewTicker(defaultRefreshInterval)
	for {
		select {
		case <-time.After(10 * time.Second):
			if h.CurrentTimer != nil {
				h.updateTimers(h.CurrentTimer.Key)
				h.sendTimers(false)
			}
		case <-tick.C:
			if err := h.Refresh(); err != nil {
				h.sendErr(err)
			}
		case <-h.changeCh:
			if err := h.Settings.Save(h.db); err != nil {
				h.sendErr(err)
				continue
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

			// If the harvest credentials changed get a new client
			if h.Settings.Harvest.User != previousSettings.Harvest.User ||
				h.Settings.Harvest.Pass != previousSettings.Harvest.Pass {
				if err := h.getNewHarvestClient(); err != nil {
					h.sendErr(err)
					continue
				}
			}

			previousSettings = *h.Settings
			if err := h.Refresh(); err != nil {
				h.sendErr(err)
			}
		}
	}
}

func (h *harvester) Refresh() error {
	// Get any current timers running in the local database
	timers, err := h.GetActiveTimers()
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
			var newTimer bool

			harvestTask := *task
			timer, err := timers.GetByKey(*task.Project.Code)
			if err != nil {
				newTimer = true
				timer = &TaskTimer{
					Key:     *task.Project.Code,
					Harvest: &harvestTask,
				}
			}

			if h.jiraClient != nil && timer.Jira == nil {
				jira, err := h.getJiraByKey(*task.Project.Code)
				if err != nil || jira.Fields.Status.Name == "Done" {
					continue
				}
				timer.Jira = jira
			}

			timer.Harvest = &harvestTask

			if newTimer {
				timers = append(timers, timer)
			}
		}
	}

	sort.SliceStable(timers, func(a, b int) bool {
		return sort.StringsAreSorted([]string{timers[a].Key, timers[b].Key})
	})
	h.Timers.Tasks = timers

	if h.mainWindow != nil {
		h.sendTimers(true)
	}

	return nil
}

func (h *harvester) init() error {
	settings, err := GetSettings(h.db)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		log.Println(err)
	}

	if settings == nil {
		return nil
	}

	h.Settings = settings

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

func (h *harvester) stopAllTimers() error {
	for _, timer := range h.Timers.Tasks {
		if err := h.StopTimer(timer); err != nil {
			return err
		}
	}
	return nil
}

func (h *harvester) Stop() {
	if err := h.Settings.Save(h.db); err != nil {
		log.Fatal(err)
	}

	if err := h.stopAllTimers(); err != nil {
		log.Fatal(err)
	}

	h.listener.Close()

	h.mainWindow.Close()
	h.app.Close()
}

func (h *harvester) Run() error {
	h.app.HandleSignals()
	h.app.Wait()
	return nil
}

func prepareIcons(harvesterDir string, ln net.Listener) {
	icons := []string{
		"icon.png",
		"icon.icns",
		"timer.png",
	}

	for _, icon := range icons {
		path := harvesterDir + "/" + icon
		if _, err := os.Stat(path); err == nil {
			continue
		}

		r, err := http.Get("http://" + ln.Addr().String() + "/img/icons/" + icon)
		if err != nil {
			continue
		}
		defer r.Body.Close()

		o, err := os.Create(path)
		if err != nil {
			continue
		}
		defer o.Close()

		io.Copy(o, r.Body)
	}

	return
}

func (h *harvester) createMenu(trayIcon string) {
	t := h.app.NewTray(&astilectron.TrayOptions{
		Image: astiptr.Str(trayIcon),
	})
	t.Create()

	h.menu = t.NewMenu([]*astilectron.MenuItemOptions{
		{
			Label: astiptr.Str("Stop Timers"),
			OnClick: func(e astilectron.Event) (deleteListener bool) {
				h.stopAllTimers()
				h.sendTimers(false)
				return
			},
		},
		{Type: astilectron.MenuItemTypeSeparator},
		{
			Label: astiptr.Str("Open Dev Tools"),
			Type:  astilectron.MenuItemTypeCheckbox,
			OnClick: func(e astilectron.Event) (deleteListener bool) {
				if *e.MenuItemOptions.Checked {
					h.debug = true
					if h.mainWindow != nil && h.mainWindow.IsShown() {
						h.mainWindow.OpenDevTools()
					}
				} else {
					h.debug = true
					if h.mainWindow != nil && h.mainWindow.IsShown() {
						h.mainWindow.CloseDevTools()
					}
				}
				return
			},
		},
		{Type: astilectron.MenuItemTypeSeparator},
		{
			Label: astiptr.Str("Quit"),
			OnClick: func(e astilectron.Event) (deleteListener bool) {
				h.app.Close()
				return
			},
		},
	})

	h.menu.Create()
}
