package main

import (
	"errors"
	"log"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/jinzhu/gorm"
)

var (
	ErrTimerNotExists = errors.New("timer not found")
)

type TaskTimer struct {
	ID        int        `gorm:"primary_key;AUTO_INCREMENT"`
	Key       string     `gorm:"index:key"`
	StartedAt time.Time  `gorm:"index:started_at"`
	StoppedAt *time.Time `gorm:"index:stopped_at;default:NULL"`

	Jira    *jira.Issue  `gorm:"-"`
	Harvest *harvestTask `gorm:"-"`
}

type TaskTimers []*TaskTimer

func (t *TaskTimer) Start(db *gorm.DB, harvestClient *HarvestClient) error {
	// Make sure an existing timer doesn't already exist
	existingTimers, err := GetActiveTimers(db, nil, harvestClient)
	if err != nil {
		return err
	}

	// Stop the current timers before starting the new one
	if len(existingTimers) > 0 {
		for _, timer := range existingTimers {
			if err := timer.Stop(db, harvestClient); err != nil {
				return err
			}
		}
	}

	// If this is not a new timer then create a new one
	if !t.New() {
		newTimer := &TaskTimer{
			Key:     t.Key,
			Jira:    t.Jira,
			Harvest: t.Harvest,
		}
		t = newTimer
	}

	t.StartedAt = time.Now().UTC()
	if err := db.Create(t).Error; err != nil {
		return err
	}

	// If a harvest task exists start the timer for it
	if t.Harvest != nil {
		return t.Harvest.startTimer(harvestClient)
	}

	return nil
}

func (t *TaskTimer) Stop(db *gorm.DB, harvestClient *HarvestClient) error {
	if t.New() {
		return nil
	}

	stoppedAt := time.Now().UTC()
	t.StoppedAt = &stoppedAt

	if err := db.Save(t).Error; err != nil {
		return err
	}

	if t.Harvest != nil {
		return t.Harvest.stopTimer(harvestClient)
	}

	return nil
}

func GetActiveTimers(db *gorm.DB, jiraClient *jira.Client, harvestClient *HarvestClient) (TaskTimers, error) {
	var timers TaskTimers
	if err := db.Where("stopped_at is null").Find(&timers).Error; err != nil {
		return nil, err
	}

	for _, timer := range timers {
		if jiraClient != nil {
			jira, _, err := jiraClient.Issue.Get(timer.Key, &jira.GetQueryOptions{})
			if err != nil {
				return nil, err
			}
			timer.Jira = jira
		}
		if harvestClient != nil {
			harvestTask, err := harvestClient.getUserProjectByKey(timer.Key)
			if err != nil {
				return nil, err
			}
			timer.Harvest = harvestTask
		}
	}

	return timers, nil
}

func (t *TaskTimer) IsRunning() bool {
	return !t.StartedAt.IsZero() && t.StoppedAt == nil
}

func (t *TaskTimer) New() bool {
	return t.ID == 0
}

// StartJiraPurger will check for old jiras every few hours and purge any that are more than 90 days old
func StartJiraPurger(db *gorm.DB) {
	purge := func() error {
		r := db.Where("started_at < ?", time.Now().UTC().Add(-90*24*time.Hour)).Delete(&TaskTimer{})
		return r.Error
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

func (timers TaskTimers) GetByKey(key string) (*TaskTimer, error) {
	for _, timer := range timers {
		if timer.Key == key {
			return timer, nil
		}
	}
	return nil, ErrTimerNotExists
}

func GetKeysWithTimes(db *gorm.DB) ([]string, error) {
	keyStructs := make([]struct{ Key string }, 0)
	if err := db.Table("task_timers").Select("key").Group("key").Scan(&keyStructs).Error; err != nil {
		return nil, err
	}

	keys := make([]string, len(keyStructs))
	for i, key := range keyStructs {
		keys[i] = key.Key
	}

	return keys, nil
}
