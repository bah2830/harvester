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
	app             *astilectron.Astilectron
	menu            *astilectron.Menu
	mainWindow      *Window
	timesheetWindow *Window
	settingsWindow  *Window
	Settings        *Settings `json:"settings"`
	changeCh        chan bool
	db              *gorm.DB
	jiraClient      *jira.Client
	harvestClient   *HarvestClient
	harvestURL      *url.URL
	Timers          *Timers `json:"timers"`
	listener        net.Listener
	debug           bool
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

	defaultIcon, darwinIcon, trayIcon := prepareIcons(ln)

	app, err := astilectron.New(astilectron.Options{
		AppName:            "Harvester",
		AppIconDefaultPath: defaultIcon,
		AppIconDarwinPath:  darwinIcon,
		DataDirectoryPath:  "./electron",
	})
	if err != nil {
		return nil, err
	}

	h := &harvester{
		app: app,
		db:  db,
		Settings: &Settings{
			RefreshInterval: defaultRefreshInterval,
		},
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

	h.createMenu(trayIcon)

	if err := h.renderMainWindow(); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *harvester) Start() {
	// Hold onto the last copy of settings to check for diffs
	previousSettings := h.Settings

	// Start the purger to keep the database small
	go StartJiraPurger(h.db)

	if err := h.Refresh(); err != nil {
		h.sendErr(h.mainWindow, err)
	}

	tick := time.NewTicker(previousSettings.RefreshInterval)
	for {
		select {
		case <-time.After(30 * time.Second):
			h.sendTimers(false)
		case <-tick.C:
			if err := h.Refresh(); err != nil {
				h.sendErr(h.mainWindow, err)
			}
		case <-h.changeCh:
			if err := h.Settings.Save(h.db); err != nil {
				h.sendErr(h.mainWindow, err)
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
					h.sendErr(h.mainWindow, err)
					continue
				}
			}

			previousSettings = h.Settings
			if err := h.Refresh(); err != nil {
				h.sendErr(h.mainWindow, err)
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
		h.sendTimers(true)
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

func (h *harvester) stopAllTimers() error {
	for _, timer := range h.Timers.Tasks {
		if err := timer.Stop(h.db, h.harvestClient); err != nil {
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

	if h.timesheetWindow != nil {
		h.timesheetWindow.Close()
	}
	if h.settingsWindow != nil {
		h.settingsWindow.Close()
	}

	h.mainWindow.Close()
	h.app.Close()
}

func (h *harvester) Run() error {
	h.app.HandleSignals()
	h.app.Wait()
	return nil
}

func prepareIcons(ln net.Listener) (defaultIcon string, darwinIcon string, trayIcon string) {
	// Download the icons into the temp dir
	r, _ := http.Get("http://" + ln.Addr().String() + "/img/icons/icon.png")
	if r != nil {
		path := os.TempDir() + "/harvester-icon.png"
		o, _ := os.Create(path)
		if o != nil {
			if _, err := io.Copy(o, r.Body); err == nil {
				defaultIcon = path
			}
			o.Close()
		}
		r.Body.Close()
	}

	r, _ = http.Get("http://" + ln.Addr().String() + "/img/icons/icon.icns")
	if r != nil {
		path := os.TempDir() + "/harvester-icon.icns"
		o, _ := os.Create(path)
		if o != nil {
			if _, err := io.Copy(o, r.Body); err == nil {
				darwinIcon = path
			}
			o.Close()
		}
		r.Body.Close()
	}

	r, _ = http.Get("http://" + ln.Addr().String() + "/img/icons/timer.png")
	if r != nil {
		path := os.TempDir() + "/tray-icon.png"
		o, _ := os.Create(path)
		if o != nil {
			if _, err := io.Copy(o, r.Body); err == nil {
				trayIcon = path
			}
			o.Close()
		}
		r.Body.Close()
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
					if h.settingsWindow != nil && h.settingsWindow.IsShown() {
						h.settingsWindow.OpenDevTools()
					}
					if h.timesheetWindow != nil && h.timesheetWindow.IsShown() {
						h.timesheetWindow.OpenDevTools()
					}
				} else {
					h.debug = true
					if h.mainWindow != nil && h.mainWindow.IsShown() {
						h.mainWindow.CloseDevTools()
					}
					if h.settingsWindow != nil && h.settingsWindow.IsShown() {
						h.settingsWindow.CloseDevTools()
					}
					if h.timesheetWindow != nil && h.timesheetWindow.IsShown() {
						h.timesheetWindow.CloseDevTools()
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
