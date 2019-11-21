package main

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/bah2830/harvester/icons"
)

func (h *harvester) renderMainWindow() {
	h.mainWindow = h.app.NewWindow("Harvester")
	h.mainWindow.Resize(fyne.Size{Width: 400, Height: 100})
	h.mainWindow.SetPadded(false)
	h.refresh()
	h.mainWindow.Show()

	go h.startResizeWatch()
	go h.startRedrawIntetrval()
}

func (h *harvester) startResizeWatch() {
	var lastSize int
	tick := time.NewTicker(100 * time.Millisecond)
	for range tick.C {
		if h.mainWindow.Canvas().Size().Width != lastSize {
			lastSize = h.mainWindow.Canvas().Size().Width
			h.redraw()
		}
	}
}

func (h *harvester) startRedrawIntetrval() {
	tick := time.NewTicker(5 * time.Second)
	for range tick.C {
		h.redraw()
	}
}

func (h *harvester) redraw() {
	var mainObjects []fyne.CanvasObject
	mainObjects = append(mainObjects, h.drawTimers()...)
	mainContent := widget.NewVBox(mainObjects...)

	h.mainWindow.SetContent(widget.NewVBox(
		widget.NewToolbar(
			widget.NewToolbarAction(icons.ResourceHarvestPng, func() {
				h.app.OpenURL(h.harvestURL)
			}),
			widget.NewToolbarAction(theme.ContentCopyIcon(), func() {
				h.getJiraListClipboard()
			}),
			widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
				h.refresh()
			}),
			widget.NewToolbarSpacer(),
			widget.NewToolbarAction(theme.InfoIcon(), func() {
				h.renderAboutWindow()
			}),
			widget.NewToolbarAction(theme.SettingsIcon(), func() {
				h.renderSettingsWindow()
			}),
		),
		mainContent,
		widget.NewHBox(
			widget.NewButtonWithIcon("Show Times", theme.InfoIcon(), func() {
				h.showJiraTimes(nil)
			}),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("Stop All", theme.CancelIcon(), func() {
				for _, timer := range h.timers {
					timer.Stop(h.db, h.harvestClient)
				}
				h.redraw()
			}),
		),
	))
}

func (h *harvester) getJiraListClipboard() {
	keys, err := GetKeysWithTimes(h.db)
	if err != nil {
		dialog.ShowError(err, h.mainWindow)
		return
	}

	var clipboard string

	for _, key := range keys {
		for _, timer := range h.timers {
			if timer.Key == key && timer.harvest == nil {
				clipboard += fmt.Sprintf("%s: %s\n", key, timer.jira.Fields.Summary)
				break
			}
		}
	}

	if len(clipboard) == 0 {
		dialog.ShowInformation("", h.wrapString("No jiras with tracked times are currently missing from harvest"), h.mainWindow)
		return
	}

	h.mainWindow.Clipboard().SetContent(clipboard)
}

func (h *harvester) renderAboutWindow() {
	w := h.app.NewWindow("About")
	w.SetContent(widget.NewHBox(
		widget.NewLabel("Version"),
		widget.NewLabel(version),
	))
	w.Show()
}

func breakString(msg string, size int) string {
	for i := size; i < len(msg); i += (size + 1) {
		msg = msg[:i] + "\n" + msg[i:]
	}
	return msg
}

func (h *harvester) wrapString(input string) string {
	sizeLimit := (h.mainWindow.Canvas().Size().Width - 100) / 9
	words := strings.Fields(strings.TrimSpace(input))
	if len(words) == 0 {
		return input

	}
	wrapped := words[0]
	spaceLeft := sizeLimit - len(wrapped)
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			wrapped += "\n" + word
			spaceLeft = sizeLimit - len(word)
		} else {
			wrapped += " " + word
			spaceLeft -= 1 + len(word)
		}
	}

	return wrapped
}
