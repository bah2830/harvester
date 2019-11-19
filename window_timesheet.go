package main

import (
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/jinzhu/now"
)

func (h *harvester) showJiraTimes(window fyne.Window) {
	var w fyne.Window
	if window != nil {
		w = window
	} else {
		w = h.app.NewWindow("Timers")
		w.SetFixedSize(true)
		w.SetPadded(false)
	}

	views := []struct {
		name string
		date time.Time
	}{
		{name: "today", date: now.BeginningOfDay()},
		{name: "week", date: now.BeginningOfWeek()},
		{name: "month", date: now.BeginningOfMonth()},
	}

	var hasErr error
	var tabs []*widget.TabItem
	for _, view := range views {
		times, err := h.drawTimeSummary(w, view.date, view.name)
		if err != nil {
			hasErr = err
		} else {
			tabs = append(tabs, widget.NewTabItem(view.name, times))
		}
	}

	if hasErr != nil {
		w.SetContent(widget.NewVBox(widget.NewLabel(hasErr.Error())))
	} else {
		w.SetContent(widget.NewVBox(
			widget.NewToolbar(
				widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
					h.showJiraTimes(w)
				}),
				widget.NewToolbarSpacer(),
				widget.NewToolbarAction(theme.CancelIcon(), func() {
					w.Close()
				}),
			),
			widget.NewTabContainer(tabs...),
		))
	}

	if window == nil {
		w.Show()
	}
}

// viewType can be today, week, month
func (h *harvester) getTimeSummary(createdAtTime time.Time, viewType string) ([]jiraTime, error) {
	var durationDays int
	switch viewType {
	case "today":
		durationDays = 1
	case "week", "month":
		durationDays = 7
	}

	var timers []TaskTimer
	if err := h.db.Where("started_at >= ?", createdAtTime).Order("started_at ASC").Find(&timers).Error; err != nil {
		return nil, err
	}

	times := make(map[string]jiraTime, 0)
	for _, timer := range timers {
		stoppedDate := time.Now()
		if timer.StoppedAt != nil {
			stoppedDate = *timer.StoppedAt
		}

		// Build the map key from the id and created week.
		// This will create one tracker per jira per week
		mapKey := timer.Key + "-" + strconv.Itoa(week(timer.StartedAt))

		// Find an existing tracker, if none exists create it.
		jiraTracker, ok := times[mapKey]
		if !ok {
			jiraTracker = jiraTime{
				JiraID:    timer.Key,
				Week:      week(timer.StartedAt),
				StartDay:  createdAtTime,
				Durations: make([]time.Duration, durationDays),
			}
		}

		day := day(timer.StartedAt, viewType)
		jiraTracker.Durations[day] = jiraTracker.Durations[day] + stoppedDate.Sub(timer.StartedAt)
		times[mapKey] = jiraTracker
	}

	// Turn map into slice and sort by the week number
	var timeSlice []jiraTime
	for _, j := range times {
		timeSlice = append(timeSlice, j)
	}
	sort.Slice(timeSlice, func(a, b int) bool {
		return timeSlice[a].Week < timeSlice[b].Week
	})

	return timeSlice, nil
}

func (h *harvester) drawTimeSummary(window fyne.Window, createdAtTime time.Time, viewType string) (*widget.Box, error) {
	times, err := h.getTimeSummary(createdAtTime, viewType)
	if err != nil {
		return nil, err
	}

	var objects []fyne.CanvasObject
	headerStyle := fyne.TextStyle{Bold: true}
	cellStyle := fyne.TextStyle{}

	var columns int
	switch viewType {
	case "today":
		columns = 2
		objects = append(objects, widget.NewLabelWithStyle("Jira", fyne.TextAlignLeading, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Today", fyne.TextAlignTrailing, headerStyle))
	case "week":
		objects = append(objects, widget.NewLabelWithStyle("Jira", fyne.TextAlignLeading, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Mon", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Tue", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Wed", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Thu", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Fri", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Sat", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Sun", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Total", fyne.TextAlignTrailing, headerStyle))
		columns = 9
	case "month":
		objects = append(objects, widget.NewLabelWithStyle("Week", fyne.TextAlignLeading, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Jira", fyne.TextAlignLeading, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Mon", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Tue", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Wed", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Thu", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Fri", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Sat", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Sun", fyne.TextAlignTrailing, headerStyle))
		objects = append(objects, widget.NewLabelWithStyle("Total", fyne.TextAlignTrailing, headerStyle))
		columns = 10
	}

	var total time.Duration
	for _, jiraTracker := range times {
		if viewType == "month" {
			objects = append(objects, widget.NewLabel(strconv.Itoa(jiraTracker.Week)))
		}

		var durationTotal time.Duration

		objects = append(objects, widget.NewHyperlink(jiraTracker.JiraID, h.getURL(jiraTracker.JiraID)))
		for _, dayTime := range jiraTracker.Durations {
			objects = append(objects, widget.NewLabelWithStyle(
				fmt.Sprintf("%.2f", dayTime.Hours()),
				fyne.TextAlignTrailing,
				cellStyle,
			))
			durationTotal += dayTime
			total += dayTime
		}

		if viewType != "today" {
			objects = append(objects, widget.NewLabelWithStyle(
				fmt.Sprintf("%.2f", durationTotal.Hours()),
				fyne.TextAlignTrailing,
				cellStyle,
			))
		}
	}

	if len(objects) == 0 {
		objects = append(
			objects,
			widget.NewLabelWithStyle("no times currently tracked", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
		)
		return widget.NewVBox(objects...), nil
	}

	// Add a copy button
	copyButton := widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		window.Clipboard().SetContent(h.getSummaryClipboard(times, viewType))
	})

	return widget.NewVBox(
		fyne.NewContainerWithLayout(layout.NewGridLayout(columns), objects...),
		canvas.NewLine(color.RGBA{R: 50, G: 50, B: 50, A: 0}),
		widget.NewHBox(
			widget.NewLabelWithStyle("Total", fyne.TextAlignLeading, headerStyle),
			layout.NewSpacer(),
			widget.NewLabelWithStyle(fmt.Sprintf("%.2f", total.Hours()), fyne.TextAlignTrailing, headerStyle),
		),
		layout.NewSpacer(),
		copyButton,
	), nil
}

func (h *harvester) getSummaryClipboard(times []jiraTime, viewType string) string {
	var clipboard string

	switch viewType {
	case "today":
		for _, t := range times {
			clipboard += fmt.Sprintf("%s\t%.2f\n", t.JiraID, t.Durations[0].Hours())
		}
	case "week":
		clipboard += "Jira\tMon\tTue\tWed\tThu\tFri\tSat\tSun\n"
		for _, t := range times {
			clipboard += t.JiraID
			for _, d := range t.Durations {
				clipboard += fmt.Sprintf("\t%.2f", d.Hours())
			}
			clipboard += "\n"
		}
	case "month":
		clipboard += "Week\tJira\tMon\tTue\tWed\tThu\tFri\tSat\tSun\n"
		for _, t := range times {
			clipboard += fmt.Sprintf("%d\t%s", t.Week, t.JiraID)
			for _, d := range t.Durations {
				clipboard += fmt.Sprintf("\t%.2f", d.Hours())
			}
			clipboard += "\n"
		}
	}

	return clipboard
}

func day(date time.Time, viewType string) int {
	var day int
	switch viewType {
	case "today":
		day = 0
	case "week", "month":
		// Shift the days back 1 so that the week starts on Monday
		day = int(date.Weekday()) - 1
		if day < 0 {
			day = 6
		}
	}

	return day
}

func week(date time.Time) int {
	beginningOfTheMonth := now.BeginningOfMonth()
	_, thisWeek := date.ISOWeek()
	_, beginningWeek := beginningOfTheMonth.ISOWeek()
	return 1 + thisWeek - beginningWeek
}
