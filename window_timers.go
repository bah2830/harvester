package main

import (
	"fmt"
	"image/color"
	"net/url"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/widget"
	"github.com/bah2830/harvester/icons"
)

func (h *harvester) drawTimers() []fyne.CanvasObject {
	rows := make([]fyne.CanvasObject, 0, len(h.timers))

	var getSummary = func(summary string) string {
		sizeLimit := (h.mainWindow.Canvas().Size().Width - 100) / 9
		if len(summary) > sizeLimit {
			summary = summary[0:sizeLimit-3] + "..."
		}
		return summary
	}

	for _, timer := range h.timers {
		var summary string
		var icon *widget.Icon
		if timer.harvest != nil {
			icon = widget.NewIcon(icons.ResourceHarvestPng)
			summary = getSummary(fmt.Sprintf("%s: %s", timer.Key, *timer.harvest.Project.Name))
		} else {
			icon = widget.NewIcon(icons.ResourceJiraPng)
			summary = getSummary(fmt.Sprintf("%s: %s", timer.Key, timer.jira.Fields.Summary))
		}

		rows = append(
			rows,
			widget.NewHBox(
				widget.NewLabel(" "),
				icon,
				widget.NewHyperlinkWithStyle(
					summary,
					h.getURL(timer.Key),
					fyne.TextAlignLeading,
					fyne.TextStyle{Monospace: true},
				),
				widget.NewToolbarSpacer().ToolbarObject(),
				h.addButton(timer),
				widget.NewLabel(" "),
			),
			canvas.NewLine(color.RGBA{R: 35, G: 38, B: 42, A: 255}),
		)
	}

	return rows
}

func (h *harvester) getURL(jiraID string) *url.URL {
	jiraURL, _ := url.Parse(h.settings.Jira.URL + "/browse/" + jiraID)
	return jiraURL
}

func (h *harvester) addButton(timer *TaskTimer) *widget.Button {
	status := "start"
	if timer.IsRunning() {
		status = "stop"
	}

	return widget.NewButton(status, func() {
		defer h.redraw()

		if !timer.IsRunning() {
			if err := timer.Start(h.db, h.harvestClient); err != nil {
				dialog.ShowError(err, h.mainWindow)
				return
			}

			// Update all current timers to stopped
			for i, t := range h.timers {
				if t.Key == timer.Key {
					continue
				}
				h.timers[i].ID = 0
				h.timers[i].StartedAt = time.Time{}
				h.timers[i].StoppedAt = nil
			}
		} else {
			if err := timer.Stop(h.db, h.harvestClient); err != nil {
				dialog.ShowError(err, h.mainWindow)
				return
			}
		}
	})
}
