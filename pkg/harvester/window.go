package harvester

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/asticode/go-astilectron"
	astiptr "github.com/asticode/go-astitools/ptr"
	"github.com/atotto/clipboard"
	"github.com/jinzhu/now"
	"github.com/skratchdot/open-golang/open"
)

type Window struct {
	*astilectron.Window
	View        string
	CurrentData *AppData
}

type AppData struct {
	View     string    `json:"view"`
	Timers   *Timers   `json:"timers"`
	Settings *Settings `json:"settings"`
	Error    string    `json:"error"`
}

func (h *harvester) getWindow(view string, opts *astilectron.WindowOptions) (*Window, error) {
	displays := h.app.Displays()
	display := displays[0]
	for _, d := range displays {
		if d.Size().Height == 1080 {
			display = d
			break
		}
	}

	w, err := h.app.NewWindowInDisplay(
		display,
		"http://"+h.listener.Addr().String()+"/templates/main.html",
		opts,
	)
	if err != nil {
		return nil, err
	}

	return &Window{Window: w, View: view}, err
}

func (h *harvester) renderMainWindow() error {
	w, err := h.getWindow(
		"main",
		&astilectron.WindowOptions{
			Title:           astiptr.Str("Harvester"),
			Height:          astiptr.Int(200),
			MinHeight:       astiptr.Int(100),
			Width:           astiptr.Int(350),
			MinWidth:        astiptr.Int(300),
			BackgroundColor: astiptr.Str("#1A1D21"),
		},
	)
	if err != nil {
		return err
	}

	h.mainWindow = w
	if err := h.mainWindow.Create(); err != nil {
		return err
	}

	if h.debug {
		if err := h.mainWindow.OpenDevTools(); err != nil {
			return err
		}
	}

	ready := make(chan bool)
	h.mainListener(ready)
	go func() {
		<-ready
		h.mainWindow.sendMessage(&AppData{})
	}()

	return nil
}

func (h *harvester) renderSettings() error {
	if h.settingsWindow == nil || h.settingsWindow.IsDestroyed() {
		w, err := h.getWindow(
			"settings",
			&astilectron.WindowOptions{
				Title:           astiptr.Str("Settings"),
				Height:          astiptr.Int(565),
				Width:           astiptr.Int(400),
				BackgroundColor: astiptr.Str("#1A1D21"),
				Resizable:       astiptr.Bool(false),
			},
		)
		if err != nil {
			return err
		}

		h.settingsWindow = w
	}

	if err := h.settingsWindow.Create(); err != nil {
		return err
	}

	if h.debug {
		if err := h.settingsWindow.OpenDevTools(); err != nil {
			return err
		}
	}

	ready := make(chan bool)
	h.settingsListener(ready)
	go func() {
		<-ready
		h.settingsWindow.sendMessage(&AppData{Settings: h.Settings})
	}()

	return nil
}

func (h *harvester) renderTimesheet() error {
	if h.timesheetWindow == nil || h.timesheetWindow.IsDestroyed() {
		w, err := h.getWindow(
			"timesheet",
			&astilectron.WindowOptions{
				Title:           astiptr.Str("TimeSheet"),
				Height:          astiptr.Int(500),
				Width:           astiptr.Int(600),
				BackgroundColor: astiptr.Str("#1A1D21"),
				Resizable:       astiptr.Bool(false),
			},
		)
		if err != nil {
			return err
		}

		h.timesheetWindow = w
	}

	if err := h.timesheetWindow.Create(); err != nil {
		return err
	}

	if h.debug {
		if err := h.timesheetWindow.OpenDevTools(); err != nil {
			return err
		}
	}

	ready := make(chan bool)
	h.timesheetListener(ready)
	go func() {
		<-ready
		h.timesheetWindow.sendMessage(&AppData{})
	}()

	return nil
}

func (h *harvester) mainListener(ready chan bool) {
	h.mainWindow.OnMessage(func(m *astilectron.EventMessage) interface{} {
		var data string
		if err := m.Unmarshal(&data); err != nil {
			h.sendErr(h.mainWindow, err)
			return err
		}

		switch {
		case data == "ready":
			ready <- true
		case data == "copy":
		case data == "refresh":
			if err := h.Refresh(); err != nil {
				h.sendErr(h.mainWindow, err)
				return err
			}
		case data == "timesheet":
			if err := h.renderTimesheet(); err != nil {
				h.sendErr(h.mainWindow, err)
				return err
			}
		case data == "settings":
			if err := h.renderSettings(); err != nil {
				h.sendErr(h.mainWindow, err)
				return err
			}
		case data == "harvest":
			if h.harvestURL != nil {
				if err := open.Run(h.harvestURL.String()); err != nil {
					h.sendErr(h.mainWindow, err)
					return err
				}
			}
		case strings.Contains(data, "|"):
			parts := strings.Split(data, "|")
			task, err := h.Timers.Tasks.GetByKey(parts[1])
			if err != nil {
				h.sendErr(h.mainWindow, err)
				return err
			}

			var currentRunning string
			switch parts[0] {
			case "start":
				currentRunning = task.Key
				err = task.Start(h.db, h.harvestClient)
			case "stop":
				err = task.Stop(h.db, h.harvestClient)
			case "open":
				err = open.Run(h.Settings.Jira.URL + "/browse/" + parts[1])
			}
			if err != nil {
				h.sendErr(h.mainWindow, err)
				return err
			}

			h.updateTimers(currentRunning)
			h.sendTimers(false)
		default:
			log.Println("unknown rpc handler " + data)
		}

		return nil
	})
}

func (h *harvester) settingsListener(ready chan bool) {
	h.settingsWindow.OnMessage(func(m *astilectron.EventMessage) interface{} {
		var raw string
		m.Unmarshal(&raw)
		if raw == "ready" {
			ready <- true
			return nil
		}

		var settings Settings
		if err := json.Unmarshal([]byte(raw), &settings); err != nil {
			h.sendErr(h.settingsWindow, err)
			return err
		}

		h.Settings.Jira.URL = settings.Jira.URL
		h.Settings.Jira.User = settings.Jira.User
		h.Settings.Jira.Pass = settings.Jira.Pass
		h.Settings.Harvest.User = settings.Harvest.User
		h.Settings.Harvest.Pass = settings.Harvest.Pass
		h.changeCh <- true

		return nil
	})
}

func (h *harvester) timesheetListener(ready chan bool) {
	h.timesheetWindow.OnMessage(func(m *astilectron.EventMessage) interface{} {
		var data string
		m.Unmarshal(&data)

		var start, end time.Time
		switch {
		case data == "ready":
			ready <- true
		default:
			now.WeekStartDay = time.Monday

			parts := strings.Split(data, "|")
			if len(parts) < 3 {
				return nil
			}

			var startTime time.Time
			var err error
			if parts[2] != "=" {
				startTime, err = time.Parse("2006-01-02T15:04:05Z", parts[1])
				if err != nil {
					h.sendErr(h.timesheetWindow, err)
					return nil
				}
			}

			addDuration := 24 * time.Hour
			if parts[0] == "week" {
				addDuration = 24 * 7 * time.Hour
			}

			switch parts[2] {
			case "=":
				start = now.BeginningOfDay()
				if parts[0] == "week" {
					start = now.BeginningOfWeek()
				}
			case "-":
				start = startTime.Add(-addDuration)
			case "+":
				start = startTime.Add(addDuration)
			case "copy":
				start = now.With(startTime).BeginningOfMonth()
				end = now.With(startTime).EndOfMonth()
				keys, err := GetKeysWithTimes(h.db, start, end)
				if err != nil {
					h.sendErr(h.timesheetWindow, err)
					return nil
				}

				clipboard.WriteAll(strings.Join(keys, "\n"))
				return nil
			}

			end = start.Add(addDuration - 1*time.Minute)
		}

		timesheet, err := h.getTimeSheet(start.UTC(), end.UTC())
		if err != nil {
			h.sendErr(h.timesheetWindow, err)
			return nil
		}
		return timesheet
	})
}

func (h *harvester) sendErr(w *Window, err error) {
	fmt.Println(err)
	current := *w.CurrentData
	current.Error = err.Error()
	w.SendMessage(current)
}

func (h *harvester) sendTimers(auto bool) {
	h.mainWindow.sendMessage(&AppData{Timers: h.Timers})

	// Change the height of the window to match the number of timers
	if auto {
		height := 64 + len(h.Timers.Tasks)*40
		h.mainWindow.SetBounds(astilectron.RectangleOptions{
			SizeOptions: astilectron.SizeOptions{
				Height: astiptr.Int(height),
			},
		})
	}
}

func (w *Window) sendMessage(message *AppData) {
	message.View = w.View
	w.CurrentData = message
	w.SendMessage(message)
}
