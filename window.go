package main

import (
	"github.com/asticode/go-astilectron"
	astiptr "github.com/asticode/go-astitools/ptr"
)

func (h *harvester) renderMainWindow() error {
	var err error
	h.mainWindow, err = h.app.NewWindow("templates/index.html", &astilectron.WindowOptions{
		Title:           astiptr.Str("Harvester"),
		Height:          astiptr.Int(100),
		Width:           astiptr.Int(400),
		BackgroundColor: astiptr.Str("#1A1D21"),
	})
	if err != nil {
		return err
	}

	// Create windows
	if err := h.mainWindow.Create(); err != nil {
		return err
	}

	return nil
}
