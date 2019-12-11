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

func (h *harvester) createWindow() error {
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

	h.mainWindow = &Window{Window: w}
	if err := h.mainWindow.Create(); err != nil {
		return err
	}

	if h.debug {
		if err := h.mainWindow.OpenDevTools(); err != nil {
			return err
		}
	}

	return err
}

func (h *harvester) renderMainWindow() error {
	if h.mainWindow == nil {
		if err := h.createWindow(); err != nil {
			return err
		}

		ready := make(chan bool)
		h.mainListener(ready)
		go func() {
			<-ready
			h.mainWindow.sendMessage(&AppData{View: "main"})
		}()
	} else {
		h.mainWindow.SetBounds(astilectron.RectangleOptions{
			SizeOptions: astilectron.SizeOptions{
				Height: astiptr.Int(200),
				Width:  astiptr.Int(350),
			},
		})
		h.mainWindow.sendMessage(&AppData{View: "main"})
		h.Refresh()
	}

	return nil
}

func (h *harvester) renderSettings() error {
	h.mainWindow.SetBounds(astilectron.RectangleOptions{
		SizeOptions: astilectron.SizeOptions{
			Height: astiptr.Int(610),
			Width:  astiptr.Int(430),
		},
	})
	return h.mainWindow.sendMessage(&AppData{View: "settings", Settings: h.Settings})
}

func (h *harvester) renderTimesheet() error {
	h.mainWindow.SetBounds(astilectron.RectangleOptions{
		SizeOptions: astilectron.SizeOptions{
			Height: astiptr.Int(500),
			Width:  astiptr.Int(630),
		},
	})

	return h.mainWindow.sendMessage(&AppData{View: "timesheet"})

}

func (h *harvester) mainListener(ready chan bool) {
	h.mainWindow.OnMessage(func(m *astilectron.EventMessage) interface{} {
		var data string
		if err := m.Unmarshal(&data); err != nil {
			h.sendErr(err)
			return err
		}

		switch {
		case data == "ready":
			ready <- true
		case data == "copy":
		case data == "refresh":
			if err := h.Refresh(); err != nil {
				h.sendErr(err)
				return err
			}
		case data == "timesheet":
			var err error
			if h.mainWindow.View == "timesheet" {
				err = h.renderMainWindow()
			} else {
				err = h.renderTimesheet()
			}
			if err != nil {
				h.sendErr(err)
				return err
			}
		case data == "settings":
			var err error
			if h.mainWindow.View == "settings" {
				err = h.renderMainWindow()
			} else {
				err = h.renderSettings()
			}
			if err != nil {
				h.sendErr(err)
				return err
			}
		case data == "harvest":
			if h.harvestURL != nil {
				if err := open.Run(h.harvestURL.String()); err != nil {
					h.sendErr(err)
					return err
				}
			}
		case strings.HasPrefix(data, "settings|"):
			data = strings.TrimPrefix(data, "settings|")

			var settings Settings
			if err := json.Unmarshal([]byte(data), &settings); err != nil {
				h.sendErr(err)
				return err
			}

			if h.Settings == nil {
				h.Settings = &Settings{}
			}

			h.Settings.Jira = SettingsData{
				URL:  settings.Jira.URL,
				User: settings.Jira.User,
				Pass: settings.Jira.Pass,
			}

			h.Settings.Harvest = SettingsData{
				User: settings.Harvest.User,
				Pass: settings.Harvest.Pass,
			}

			h.changeCh <- true

			h.renderMainWindow()
		case strings.HasPrefix(data, "day|") || strings.HasPrefix(data, "week|"):
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
					h.sendErr(err)
					return nil
				}
			}

			addDuration := 24 * time.Hour
			if parts[0] == "week" {
				addDuration = 24 * 7 * time.Hour
			}

			var start, end time.Time

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
					h.sendErr(err)
					return nil
				}

				var noProjects string
				for _, k := range keys {
					tracker, err := h.Timers.Tasks.GetByKey(k)
					if err == nil {
						if tracker.Harvest == nil {
							noProjects += fmt.Sprintf("\n%s; %s", tracker.Key, tracker.Jira.Fields.Summary)
						}
					}
				}

				clipboard.WriteAll(noProjects)
				return nil
			}

			end = start.Add(addDuration - 1*time.Minute)

			timesheet, err := h.getTimeSheet(start.UTC(), end.UTC())
			if err != nil {
				h.sendErr(err)
				return nil
			}
			return timesheet
		case strings.Contains(data, "|"):
			parts := strings.Split(data, "|")
			task, err := h.Timers.Tasks.GetByKey(parts[1])
			if err != nil {
				h.sendErr(err)
				return err
			}

			var currentRunning string
			switch parts[0] {
			case "start":
				currentRunning = task.Key
				err = h.StartTimer(task)
			case "stop":
				err = h.StopTimer(task)
			case "open":
				err = open.Run(h.Settings.Jira.URL + "/browse/" + parts[1])
			}
			if err != nil {
				h.sendErr(err)
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

func (h *harvester) sendErr(err error) {
	fmt.Println(err)
	current := *h.mainWindow.CurrentData
	current.Error = err.Error()
	h.mainWindow.SendMessage(current)
}

func (h *harvester) sendTimers(auto bool) {
	if h.mainWindow.View != "main" {
		return
	}

	h.mainWindow.sendMessage(&AppData{View: "main", Timers: h.Timers})

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

func (w *Window) sendMessage(message *AppData) error {
	w.View = message.View
	w.CurrentData = message
	return w.SendMessage(message)
}
