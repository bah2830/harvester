package harvester

import (
	"sort"
	"strconv"
	"time"

	"github.com/jinzhu/now"
)

type TimeSheet struct {
	Today []TaskTimeInfo `json:"today"`
	Week  []TaskTimeInfo `json:"week"`
	Month []TaskTimeInfo `json:"month"`
}

type TaskTimeInfo struct {
	Week      int             `json:"week"`
	JiraID    string          `json:"jiraId"`
	Durations []time.Duration `json:"durations"`
	StartDay  time.Time       `json:"startDay"`
}

// viewType can be today, week, month
func (h *harvester) getTimeSummary(createdAtTime time.Time, viewType string) ([]TaskTimeInfo, error) {
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

	times := make(map[string]TaskTimeInfo, 0)
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
			jiraTracker = TaskTimeInfo{
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
	var timeSlice []TaskTimeInfo
	for _, j := range times {
		timeSlice = append(timeSlice, j)
	}
	sort.Slice(timeSlice, func(a, b int) bool {
		return timeSlice[a].Week < timeSlice[b].Week
	})

	return timeSlice, nil
}

func week(date time.Time) int {
	beginningOfTheMonth := now.BeginningOfMonth()
	_, thisWeek := date.ISOWeek()
	_, beginningWeek := beginningOfTheMonth.ISOWeek()
	return 1 + thisWeek - beginningWeek
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

func (h *harvester) getTimeSheet() (*TimeSheet, error) {
	views := []struct {
		name string
		date time.Time
	}{
		{name: "today", date: now.BeginningOfDay()},
		{name: "week", date: now.BeginningOfWeek()},
		{name: "month", date: now.BeginningOfMonth()},
	}

	timesheet := &TimeSheet{}
	for _, view := range views {
		times, err := h.getTimeSummary(view.date, view.name)
		if err != nil {
			return nil, err
		}

		switch view.name {
		case "today":
			timesheet.Today = times
		case "week":
			timesheet.Week = times
		case "month":
			timesheet.Month = times
		}

	}
	return timesheet, nil
}
