package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"

	"fyne.io/fyne"
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

	lastTypeQuery = "select track_type from time_tracking where jira_id = ? order by created_at DESC"
	timingQuery   = "select jira_id, track_type, created_at from time_tracking where created_at >= ? order by created_at ASC"
)

type jiraTime struct {
	JiraID    string
	Duration  time.Duration
	LastStart time.Time
}

func (h *harvester) getUsersActiveIssues() ([]jira.Issue, error) {
	issues, _, err := h.jiraClient.Issue.Search(issueQuery, nil)
	if err != nil {
		return nil, err
	}
	return issues, err
}

func (h *harvester) saveJiraTime(jiraID, trackType string) error {
	// Check the last value for this jira and only update if it's changed
	var lastType string
	if err := h.db.QueryRow(lastTypeQuery, jiraID).Scan(&lastType); err != nil && err != sql.ErrNoRows {
		return err
	}
	if lastType == trackType {
		return nil
	}

	if _, err := h.db.Exec("insert into time_tracking (jira_id, track_type) values (?, ?)", jiraID, trackType); err != nil {
		return err
	}
	return nil
}

func (h *harvester) drawJiraObjects() []fyne.CanvasObject {
	jiraRows := make([]fyne.CanvasObject, len(h.activeJiras))

	for i, jiraIssue := range h.activeJiras {
		// Trim the summary to 27 chars max
		summary := jiraIssue.Fields.Summary
		if len(summary) > 27 {
			summary = summary[0:24] + "..."
		}

		// Build the user friendly url
		jiraURL, _ := url.Parse(h.settings.Jira.URL + "/browse/" + jiraIssue.Key)

		// Get the latest status for this jira if it exists
		var trackType string

		h.db.QueryRow(lastTypeQuery, jiraIssue.Key).Scan(&trackType)
		if trackType == "" || trackType == "stop" {
			trackType = "start"
		} else {
			trackType = "stop"
		}

		jiraRows[i] = widget.NewHBox(
			widget.NewHyperlinkWithStyle(fmt.Sprintf("%s: %s", jiraIssue.Key, summary), jiraURL, fyne.TextAlignLeading, fyne.TextStyle{Monospace: true}),
			widget.NewToolbarSpacer().ToolbarObject(),
			h.addButton(jiraIssue.Key, trackType),
		)
	}

	return jiraRows
}

func (h *harvester) addButton(jiraID, trackType string) *widget.Button {
	return widget.NewButton(trackType, func() {
		// If the save is successful refresh
		if err := h.saveJiraTime(jiraID, trackType); err == nil {
			h.refresh()
		}
	})
}

func (h *harvester) showJiraTimes() {
	todayTimes, err := h.getTimeSummary(now.BeginningOfDay())
	if err != nil {
		log.Print(err)
		return
	}

	weekTimes, err := h.getTimeSummary(now.BeginningOfWeek())
	if err != nil {
		log.Print(err)
		return
	}

	monthTimes, err := h.getTimeSummary(now.BeginningOfMonth())
	if err != nil {
		log.Print(err)
		return
	}

	w := h.app.NewWindow("Timers")
	w.SetFixedSize(true)
	w.SetContent(
		widget.NewTabContainer(
			widget.NewTabItem("today", widget.NewVBox(todayTimes...)),
			widget.NewTabItem("week", widget.NewVBox(weekTimes...)),
			widget.NewTabItem("month", widget.NewVBox(monthTimes...)),
		),
	)

	w.Show()
}

func (h *harvester) getTimeSummary(createdAtTime time.Time) ([]fyne.CanvasObject, error) {
	times := make(map[string]jiraTime, 0)
	var jiraID, trackType, createdAt string
	rows, err := h.db.Query(timingQuery, createdAtTime.Format(dateTimeFormat))
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err := rows.Scan(&jiraID, &trackType, &createdAt); err != nil {
			return nil, err
		}

		createDate, err := time.Parse(sqliteTimeFormat, createdAt)
		if err != nil {
			return nil, err
		}

		jiraTracker, ok := times[jiraID]
		if !ok {
			jiraTracker = jiraTime{
				JiraID: jiraID,
			}
		}

		if trackType == "start" {
			jiraTracker.LastStart = createDate
		} else {
			jiraTracker.Duration = jiraTracker.Duration + createDate.Sub(jiraTracker.LastStart)
			jiraTracker.LastStart = time.Time{}
		}
		times[jiraID] = jiraTracker
	}

	var timeObjects []fyne.CanvasObject
	for jiraID, jiraTracker := range times {
		// If any timers are not stopped track them until the current time
		if !jiraTracker.LastStart.IsZero() {
			jiraTracker.Duration = jiraTracker.Duration + time.Now().Sub(jiraTracker.LastStart)
			jiraTracker.LastStart = time.Time{}
			times[jiraID] = jiraTracker
		}

		timeObjects = append(timeObjects, widget.NewHBox(
			widget.NewLabel(jiraID),
			widget.NewToolbarSpacer().ToolbarObject(),
			widget.NewLabel(fmt.Sprintf("%.3f", jiraTracker.Duration.Hours())),
		))
	}

	return timeObjects, nil
}
