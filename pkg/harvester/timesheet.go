package harvester

import (
	"math"
	"sort"
	"time"

	"github.com/dgraph-io/badger"
)

type TimeSheet struct {
	TimeStart time.Time      `json:"timeStart"`
	TimeEnd   time.Time      `json:"timeEnd"`
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

	timers, err := getTimersByOpts(h.db, badger.DefaultIteratorOptions)
	if err != nil {
		return nil, err
	}

	// Add any currently running timers that are not already stored
CURRENT_TIMER:
	for _, currentTimer := range h.Timers {
		if currentTimer.StartedAt == nil {
			continue
		}

		for _, timer := range timers {
			if timer.Key == currentTimer.Key && timer.Day.YearDay() == currentTimer.StartedAt.YearDay() {
				continue CURRENT_TIMER
			}
		}

		timers = append(
			timers,
			StoredTimer{
				Key:      currentTimer.Key,
				Day:      *currentTimer.StartedAt,
				Duration: time.Since(*currentTimer.StartedAt),
			},
		)
	}

	var total float64
	daysTotal := make([]float64, days)

	times := make(map[string]TaskTimeInfo, 0)
	for _, timer := range timers {
		// If a timer is currently running for this key then add the duration to it
		if runningTimer, err := h.Timers.GetByKey(timer.Key); err == nil {
			if runningTimer.StartedAt != nil && runningTimer.StartedAt.Format("20060102") == timer.Day.Format("20060102") {
				timer.Duration = timer.Duration + time.Since(*runningTimer.StartedAt)
			}
		}

		if timer.Day.Before(startTime) {
			continue
		}
		if timer.Day.After(endTime) {
			break
		}

		// Find an existing tracker, if none exists create it.
		jiraTracker, ok := times[timer.Key]
		if !ok {
			jiraTracker = TaskTimeInfo{
				Key:       timer.Key,
				Durations: make([]float64, days),
			}
		}

		day := day(viewType, timer.Day.Local())
		runTime := timer.Duration.Hours()

		jiraTracker.Durations[day] = jiraTracker.Durations[day] + runTime
		jiraTracker.TotalTime = jiraTracker.TotalTime + runTime

		daysTotal[day] = daysTotal[day] + runTime
		total = total + runTime

		times[timer.Key] = jiraTracker
	}

	// Turn map into slice and sort by the week number
	trackedTasks := make([]TaskTimeInfo, 0)
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
		TimeStart: startTime,
		TimeEnd:   endTime,
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
