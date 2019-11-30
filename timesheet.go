package main

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/jinzhu/now"
)

type taskTimeInfo struct {
	Week      int
	JiraID    string
	Durations []time.Duration
	StartDay  time.Time
}

// viewType can be today, week, month
func (h *harvester) getTimeSummary(createdAtTime time.Time, viewType string) ([]taskTimeInfo, error) {
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

	times := make(map[string]taskTimeInfo, 0)
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
			jiraTracker = taskTimeInfo{
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
	var timeSlice []taskTimeInfo
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

func (h *harvester) getTimeSheetHTML() (string, error) {
	views := []struct {
		name string
		date time.Time
	}{
		{name: "today", date: now.BeginningOfDay()},
		{name: "week", date: now.BeginningOfWeek()},
		{name: "month", date: now.BeginningOfMonth()},
	}

	html := `<ul class="nav nav-tabs" role="tablist">`
	for i, view := range views {
		var active string
		if i == 0 {
			active = "active"
		}

		html += `
	<li class="nav-item">
		<a
			class="nav-link ` + active + `"
			id="` + view.name + `-tab"
			data-toggle="tab"
			href="#` + view.name + `"
			aria-controls="` + view.name + `"
			aria-selected="true"
		>
			` + view.name + `
		</a>
	</li>`
	}

	html += `
</ul>`

	html += `
<div class="tab-content" id="myTabContent">`

	for _, view := range views {
		times, err := h.getTimeSummary(view.date, view.name)
		if err != nil {
			return "", err
		}

		html += h.getTimeSheetViewHTML(times, view.name)
	}

	html += `
</div>
	`
	return html, nil
}

func (h *harvester) getTimeSheetViewHTML(times []taskTimeInfo, view string) string {
	var html string
	if view == "today" {
		html += `
	<div class="tab-pane fade show active" `
	} else {
		html += `
	<div class="tab-pane fade" `
	}
	html += `id="` + view + `" role="tabpanel" aria-labelledby="` + view + `-tab">`

	switch view {
	case "today":
		html += h.getTimeSheetTodayHTML(times)
	case "week":
		html += h.getTimeSheetWeekHTML(times)
	case "month":
		html += h.getTimeSheetMonthHTML(times)
	}

	html += `
</div>
`
	return html
}

func (h *harvester) getTimeSheetTodayHTML(times []taskTimeInfo) string {
	html := `
	<table class="table table-hover time-table">
		<thead>
			<tr>
				<th scope="col">Jira</th>
				<th scope="col">Hours</th>
			</tr>
		</thead>
		<tbody>`

	var total time.Duration
	for _, t := range times {
		var durationTotal time.Duration
		for _, time := range t.Durations {
			durationTotal += time
			total += time
		}

		html += `
			<tr>
				<td>` + t.JiraID + `</td>
				<td>` + fmt.Sprintf("%.2f", durationTotal.Hours()) + `</td>
			</tr>`
	}

	html += `
			<tr>
				<td>Total</td>
				<td>` + fmt.Sprintf("%.2f", total.Hours()) + `</td>
		</tbody>
	</table>`
	return html
}

func (h *harvester) getTimeSheetWeekHTML(times []taskTimeInfo) string {
	html := `
	<table class="table table-hover time-table">
		<thead>
			<tr>
				<th scope="col">Jira</th>
				<th scope="col">Mon</th>
				<th scope="col">Tue</th>
				<th scope="col">Wed</th>
				<th scope="col">Thu</th>
				<th scope="col">Fri</th>
				<th scope="col">Sat</th>
				<th scope="col">Sun</th>
				<th scope="col">Total</th>
			</tr>
		</thead>
		<tbody>`

	var total time.Duration
	for _, t := range times {
		html += `<tr><td>` + t.JiraID + `</td>`

		var durationTotal time.Duration
		for _, time := range t.Durations {
			html += `<td>` + fmt.Sprintf("%.2f", time.Hours()) + `</td>`
			durationTotal += time
			total += time
		}
		html += `<td>` + fmt.Sprintf("%.2f", durationTotal.Hours()) + `<td></tr>`
	}

	html += `
				<td colspan="8">Total</td><td>` + fmt.Sprintf("%.2f", total.Hours()) + `</td>
			</tr>
		</tbody>
	</table>`
	return html
}

func (h *harvester) getTimeSheetMonthHTML(times []taskTimeInfo) string {
	html := `
	<table class="table table-hover time-table">
		<thead>
			<tr>
				<th scope="col">Week</th>
				<th scope="col">Jira</th>
				<th scope="col">Mon</th>
				<th scope="col">Tue</th>
				<th scope="col">Wed</th>
				<th scope="col">Thu</th>
				<th scope="col">Fri</th>
				<th scope="col">Sat</th>
				<th scope="col">Sun</th>
				<th scope="col">Total</th>
			</tr>
		</thead>
		<tbody>`

	var total time.Duration
	for _, t := range times {
		html += `<tr>
			<td>` + fmt.Sprintf("%d", t.Week) + `</td>
			<td>` + t.JiraID + `</td>`

		var durationTotal time.Duration
		for _, time := range t.Durations {
			html += `<td>` + fmt.Sprintf("%.2f", time.Hours()) + `</td>`
			durationTotal += time
			total += time
		}
		html += `<td>` + fmt.Sprintf("%.2f", durationTotal.Hours()) + `<td></tr>`
	}

	html += `
					<td colspan="9">Total</td><td>` + fmt.Sprintf("%.2f", total.Hours()) + `</td>
				</tr>
			</tbody>
		</table>`
	return html
}
