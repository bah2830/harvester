package harvester

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/asticode/go-astilectron"
	astiptr "github.com/asticode/go-astitools/ptr"
	"github.com/jinzhu/now"
	"github.com/skratchdot/open-golang/open"
)

type AppData struct {
	View     string    `json:"view"`
	Timers   *Timers   `json:"timers"`
	Settings *Settings `json:"settings"`
	Error    string    `json:"error"`
}

func (h *harvester) getWindow(opts *astilectron.WindowOptions) (*astilectron.Window, error) {

	displays := h.app.Displays()
	display := displays[0]
	for _, d := range displays {
		if d.Size().Height == 1080 && d.WorkArea().Position.X == -1920 {
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

	return w, err
}

func (h *harvester) renderMainWindow() error {
	w, err := h.getWindow(
		&astilectron.WindowOptions{
			Title:           astiptr.Str("Harvester"),
			Height:          astiptr.Int(200),
			MinHeight:       astiptr.Int(100),
			Width:           astiptr.Int(400),
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
		h.mainWindow.SendMessage(AppData{View: "main"})
	}()

	return nil
}

func (h *harvester) renderSettings() error {
	if h.settingsWindow == nil || h.settingsWindow.IsDestroyed() {
		w, err := h.getWindow(
			&astilectron.WindowOptions{
				Title:           astiptr.Str("Settings"),
				Height:          astiptr.Int(500),
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
		h.settingsWindow.SendMessage(AppData{View: "settings", Settings: h.Settings})
	}()

	return nil
}

func (h *harvester) renderTimesheet() error {
	if h.timesheetWindow == nil || h.timesheetWindow.IsDestroyed() {
		w, err := h.getWindow(
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
	h.timeoutListener(ready)
	go func() {
		<-ready
		h.timesheetWindow.SendMessage(AppData{View: "timesheet"})
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

func (h *harvester) timeoutListener(ready chan bool) {
	h.timesheetWindow.OnMessage(func(m *astilectron.EventMessage) interface{} {
		var data string
		m.Unmarshal(&data)

		var start, end time.Time
		switch data {
		case "ready":
			ready <- true
		case "day":
			start, end = now.BeginningOfDay(), now.EndOfDay()
		case "week":
			start, end = now.BeginningOfMonth(), now.EndOfMonth()
		}

		timesheet, err := h.getTimeSheet(start.UTC(), end.UTC())
		if err != nil {
			h.sendErr(h.timesheetWindow, err)
			return nil
		}
		return timesheet
	})
}

func (h *harvester) sendErr(w *astilectron.Window, err error) {
	fmt.Println(err)
	w.SendMessage(AppData{View: "error", Error: err.Error()})
}

func (h *harvester) sendTimers(auto bool) {
	h.mainWindow.SendMessage(AppData{View: "main", Timers: h.Timers})

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
