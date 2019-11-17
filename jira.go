package main

import (
	"database/sql"
	"errors"
	"fmt"
	"image/color"
	"log"
	"net/url"
	"sort"
	"strconv"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	jira "github.com/andygrunwald/go-jira"
	"github.com/jinzhu/now"
)

const (
	dateTimeFormat   = "2006-01-02 15:04:05"
	sqliteTimeFormat = "2006-01-02T15:04:05Z"

	issueQuery = `assignee = currentUser()
		AND Resolution = Unresolved
		AND status not in ("To Do", "Selected")
		ORDER BY updated DESC, status DESC`

	lastStartQuery = `select
			id,
			started_at,
			stopped_at
		from
			jira_time_tracking
		where
			jira_id = ?
		order by started_at DESC`

	timingQuery = `select
			jira_id,
			jira_description,
			started_at,
			stopped_at
		from
			jira_time_tracking
		where
			started_at >= ?
		order by started_at ASC`
)

type jiraTime struct {
	Week            int
	JiraID          string
	JiraDescription string
	Durations       []time.Duration
	StartDay        time.Time
}

func (h *harvester) getUsersActiveIssues() ([]jira.Issue, error) {
	issues, _, err := h.jiraClient.Issue.Search(issueQuery, nil)
	if err != nil {
		log.Print(err)
		return nil, errors.New("error getting active jira issues")
	}
	return issues, err
}

func (h *harvester) saveJiraTime(jira jira.Issue, trackType string) error {
	// Check the last value for this jira and only update if it's changed
	var id int
	var startedAt string
	var stoppedAt *string
	err := h.db.QueryRow(lastStartQuery, jira.Key).Scan(&id, &startedAt, &stoppedAt)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// Make sure the database is in a state to take this action
	if id == 0 && trackType == "stop" {
		return nil
	} else if startedAt != "" && stoppedAt != nil && trackType == "stop" {
		return nil
	} else if startedAt != "" && stoppedAt == nil && trackType == "start" {
		return nil
	}

	if trackType == "stop" {
		_, err := h.db.Exec(
			"update jira_time_tracking set stopped_at = ? where id = ?",
			time.Now().UTC().Format(dateTimeFormat),
			id,
		)
		return err
	}

	_, err = h.db.Exec(
		"insert into jira_time_tracking (jira_id, jira_description) values (?, ?)",
		jira.Key,
		jira.Fields.Summary,
	)
	return err
}

func (h *harvester) drawJiraObjects() []fyne.CanvasObject {
	jiraRows := make([]fyne.CanvasObject, 0, len(h.activeJiras))

	for _, jiraIssue := range h.activeJiras {
		// Trim the summary to 27 chars max
		summary := jiraIssue.Fields.Summary
		if len(summary) > 27 {
			summary = summary[0:24] + "..."
		}

		// Get the latest status for this jira if it exists
		var id int
		var startedAt string
		var stoppedAt *string
		err := h.db.QueryRow(lastStartQuery, jiraIssue.Key).Scan(&id, &startedAt, &stoppedAt)
		if err != nil && err != sql.ErrNoRows {
			return []fyne.CanvasObject{widget.NewLabel(err.Error())}
		}

		var status string
		if id == 0 {
			status = "start"
		} else if startedAt != "" && stoppedAt == nil {
			status = "stop"
		} else if startedAt != "" && stoppedAt != nil {
			status = "start"
		}

		jiraRows = append(jiraRows,
			widget.NewHBox(
				widget.NewHyperlinkWithStyle(
					fmt.Sprintf("%s: %s", jiraIssue.Key, summary),
					h.getURL(jiraIssue.Key),
					fyne.TextAlignLeading,
					fyne.TextStyle{Monospace: true},
				),
				widget.NewToolbarSpacer().ToolbarObject(),
				h.addButton(jiraIssue, status),
			),
			canvas.NewLine(color.RGBA{R: 23, G: 26, B: 30, A: 255}),
		)
	}

	return jiraRows
}

func (h *harvester) getURL(jiraID string) *url.URL {
	jiraURL, _ := url.Parse(h.settings.Jira.URL + "/browse/" + jiraID)
	return jiraURL
}

func (h *harvester) addButton(jira jira.Issue, trackType string) *widget.Button {
	return widget.NewButton(trackType, func() {
		if err := h.saveJiraTime(jira, trackType); err == nil {
			h.redraw()
		}
	})
}

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

	rows, err := h.db.Query(timingQuery, createdAtTime.Format(dateTimeFormat))
	if err != nil {
		return nil, err
	}

	times := make(map[string]jiraTime, 0)
	for rows.Next() {
		var jiraID, jiraDescription, startedAt string
		var stoppedAt *string
		if err := rows.Scan(&jiraID, &jiraDescription, &startedAt, &stoppedAt); err != nil {
			return nil, err
		}

		createDate, err := time.Parse(sqliteTimeFormat, startedAt)
		if err != nil {
			return nil, err
		}

		stoppedDate := time.Now()
		if stoppedAt != nil {
			stoppedDate, err = time.Parse(sqliteTimeFormat, *stoppedAt)
			if err != nil {
				return nil, err
			}
		}

		// Build the map key from the id and created week.
		// This will create one tracker per jira per week
		mapKey := jiraID + "-" + strconv.Itoa(week(createDate))

		// Find an existing tracker, if none exists create it.
		jiraTracker, ok := times[mapKey]
		if !ok {
			jiraTracker = jiraTime{
				JiraID:          jiraID,
				JiraDescription: jiraDescription,
				Week:            week(createDate),
				StartDay:        createdAtTime,
				Durations:       make([]time.Duration, durationDays),
			}
		}

		day := day(createDate, viewType)
		jiraTracker.Durations[day] = jiraTracker.Durations[day] + stoppedDate.Sub(createDate)
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

// jiraPurger will check for old jiras every few hours and purge any that are more than 90 days old
func (h *harvester) jiraPurger() {
	purge := func() error {
		_, err := h.db.Exec("delete from jira_time_tracking where started_at < ?", time.Now().UTC().Add(-90*24*time.Hour))
		return err
	}

	if err := purge(); err != nil {
		log.Print(err)
	}

	tick := time.NewTicker(3 * time.Hour)
	for range tick.C {
		if err := purge(); err != nil {
			log.Print(err)
		}
	}
}
