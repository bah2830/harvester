package main

import (
	"github.com/asticode/go-astilectron"
	astiptr "github.com/asticode/go-astitools/ptr"
)

func (h *harvester) renderMainWindow() error {
	d := h.app.Displays()
	w, err := h.app.NewWindowInDisplay(
		d[len(d)-1],
		"http://localhost:46557/resources/templates/index.html",
		&astilectron.WindowOptions{
			Title:           astiptr.Str("Harvester"),
			MinHeight:       astiptr.Int(100),
			Width:           astiptr.Int(400),
			MinWidth:        astiptr.Int(400),
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

	h.addListeners()
	h.mainWindow.OpenDevTools()
	return nil
}

func (h *harvester) sendErr(err error) {
	errS := struct {
		Type    string
		Message string
	}{
		Type:    "error",
		Message: err.Error(),
	}

	h.mainWindow.SendMessage(errS)
}

func (h *harvester) sendTimers(timers TaskTimers) {
	timerS := struct {
		Type   string
		Timers TaskTimers
	}{
		Type:   "renderTimers",
		Timers: timers,
	}

	h.mainWindow.SendMessage(timerS)
}

func (h *harvester) addListeners() {
	h.mainWindow.OnMessage(func(m *astilectron.EventMessage) interface{} {
		var message string
		m.Unmarshal(&message)

		switch message {
		case "refresh":
			if err := h.refresh(); err != nil {
				h.sendErr(err)
				return nil
			}
		}

		return nil
	})
}
