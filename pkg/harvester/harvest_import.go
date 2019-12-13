package harvester

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/becoded/go-harvest/harvest"
	"github.com/dgraph-io/badger"
)

type harvestEntries []*harvest.TimeEntry

func (h harvestEntries) getByKeyAndTime(key string, date time.Time) (*harvest.TimeEntry, error) {
	for _, e := range h {
		if *e.Project.Code == key && date.YearDay() == e.SpentDate.YearDay() {
			return e, nil
		}
	}

	return nil, fmt.Errorf("No entry with key %s and time %s found", key, date.String())
}

func (h *harvester) backfillHarvest() error {
	if h.harvestClient == nil {
		return nil
	}

	log.Println("harvest backfill start")

	ctx, c := context.WithTimeout(context.Background(), time.Minute)
	defer c()

	timesheet, _, err := h.harvestClient.Timesheet.List(ctx, &harvest.TimeEntryListOptions{
		From: &harvest.Date{Time: time.Now().Add(-35 * 24 * time.Hour)},
	})
	if err != nil {
		return err
	}

	var harvestEntries harvestEntries
	for _, e := range timesheet.TimeEntries {
		if *e.UserAssignment.IsProjectManager && *e.Hours > 0 {
			harvestEntries = append(harvestEntries, e)
		}
	}

	storedTimers, err := getTimersByOpts(h.db, badger.DefaultIteratorOptions)
	if err != nil {
		return err
	}

	harvestTasks, err := h.harvestClient.getUserProjects()
	if err != nil {
		return err
	}

	for _, task := range storedTimers {
		if math.Round(task.Duration.Hours()*100)/100 == 0 {
			continue
		}

		_, err := harvestTasks.getByKey(task.Key)
		if err != nil {
			continue
		}

		entry, _ := harvestEntries.getByKeyAndTime(task.Key, task.Day)
		if entry != nil {
			if *entry.IsRunning {
				continue
			}

			if hoursMatch(*entry.Hours, task.Duration.Hours()) {
				continue
			}

			hourAdjust := task.Duration.Hours() - *entry.Hours
			if *entry.Hours > task.Duration.Hours() {
				hourAdjust = -1 * hourAdjust
			}
			newTime := *entry.Hours + hourAdjust

			log.Printf(
				"Updating from %.2f to %.2f (%.2f) for key %s on %s\n",
				*entry.Hours,
				task.Duration.Hours(),
				hourAdjust,
				task.Key,
				task.Day.Format("2006-01-02"),
			)

			ctx, c := context.WithTimeout(context.Background(), time.Minute)
			defer c()
			_, _, err := h.harvestClient.Timesheet.UpdateTimeEntry(ctx, *entry.Id, &harvest.TimeEntryUpdate{
				Hours: &newTime,
			})
			if err != nil {
				return err
			}
		} else {
			log.Printf(
				"Adding %.2f for key %s on %s\n",
				task.Duration.Hours(),
				task.Key,
				task.Day.Format("2006-01-02"),
			)

			hours := task.Duration.Hours()

			ctx, c := context.WithTimeout(context.Background(), time.Minute)
			defer c()
			_, _, err := h.harvestClient.Timesheet.CreateTimeEntryViaDuration(ctx, &harvest.TimeEntryCreateViaDuration{
				ProjectId: entry.Project.Id,
				TaskId:    entry.Task.Id,
				Hours:     &hours,
				SpentDate: &harvest.Date{Time: task.Day},
			})
			if err != nil {
				return err
			}
		}
	}

	log.Println("harvest backfill complete")
	return nil
}

// will return true if the two floats are within a certian percent of each other
func hoursMatch(a, b float64) bool {
	return (a/b) < 1.05 && (a/b) > .95
}
