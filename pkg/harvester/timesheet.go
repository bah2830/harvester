package harvester

import (
	"math"
	"sort"
	"time"
)

type TimeSheet struct {
	Tasks     []TaskTimeInfo `json:"tasks"`
	DaysTotal []float64      `json:"daysTotal"`
	Total     float64        `json:"total"`
}
type TaskTimeInfo struct {
	Key       string    `json:"key"`
	Durations []float64 `json:"durations"`
	TotalTime float64   `json:"totalTime"`
}

// viewType can be today, week, month
func (h *harvester) getTimeSheet(startTime, endTime time.Time) (*TimeSheet, error) {
	viewType := "day"
	days := 1
	if endTime.Sub(startTime).Hours() > 24 {
		viewType = "week"
		days = 7
	}

	var timers []TaskTimer
	r := h.db.Where("started_at BETWEEN ? AND ?", startTime, endTime).Find(&timers)
	if r.Error != nil {
		return nil, r.Error
	}

	var total float64
	daysTotal := make([]float64, days)

	times := make(map[string]TaskTimeInfo, 0)
	for _, timer := range timers {
		stoppedDate := time.Now()
		if timer.StoppedAt != nil {
			stoppedDate = (*timer.StoppedAt).Local()
		}

		// Find an existing tracker, if none exists create it.
		jiraTracker, ok := times[timer.Key]
		if !ok {
			jiraTracker = TaskTimeInfo{
				Key:       timer.Key,
				Durations: make([]float64, days),
			}
		}

		day := day(viewType, timer.StartedAt.Local())
		runTime := stoppedDate.Sub(timer.StartedAt.Local()).Hours()

		jiraTracker.Durations[day] = jiraTracker.Durations[day] + runTime
		jiraTracker.TotalTime = jiraTracker.TotalTime + runTime

		daysTotal[day] = daysTotal[day] + runTime
		total = total + runTime

		times[timer.Key] = jiraTracker
	}

	// Turn map into slice and sort by the week number
	var trackedTasks []TaskTimeInfo
	for _, j := range times {

		// Round the durations down to the nearest 100ths
		j.TotalTime = math.Round(j.TotalTime*100) / 100
		for i := range j.Durations {
			j.Durations[i] = math.Round(j.Durations[i]*100) / 100
		}

		trackedTasks = append(trackedTasks, j)
	}
	sort.Slice(trackedTasks, func(a, b int) bool {
		return trackedTasks[a].Key < trackedTasks[b].Key
	})

	for i := range daysTotal {
		daysTotal[i] = math.Round(daysTotal[i]*100) / 100
	}

	return &TimeSheet{
		Tasks:     trackedTasks,
		DaysTotal: daysTotal,
		Total:     math.Round(total*100) / 100,
	}, nil
}

func day(view string, date time.Time) int {
	if view == "day" {
		return 0
	}

	// Shift the days back 1 so that the week starts on Monday
	day := int(date.Weekday()) - 1
	if day < 0 {
		day = 6
	}

	return day
}
