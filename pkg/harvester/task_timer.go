package harvester

import (
	"errors"
	"fmt"
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
	Key       string     `gorm:"index:key" json:"key"`
	StartedAt time.Time  `gorm:"index:started_at" json:"startedAt"`
	StoppedAt *time.Time `gorm:"index:stopped_at;default:NULL" json:"stoppedAt"`

	Running bool   `gorm:"-" json:"running"`
	Runtime string `gorm:"-" json:"runtime"`

	Jira    *jira.Issue  `gorm:"-" json:"jira"`
	Harvest *harvestTask `gorm:"-" json:"harvest"`
}

type TaskTimers []*TaskTimer

func (h *harvester) StartTimer(t *TaskTimer) error {
	// Make sure an existing timer doesn't already exist
	if h.CurrentTimer != nil {
		if err := h.StopTimer(h.CurrentTimer); err != nil {
			return err
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
	if err := h.db.Create(t).Error; err != nil {
		return err
	}

	// If a harvest task exists start the timer for it
	if t.Harvest != nil {
		return t.Harvest.startTimer(h.harvestClient)
	}

	t.Running = true
	t.Runtime = t.CurrentRuntime()
	h.CurrentTimer = t
	return nil
}

func (h *harvester) StopTimer(t *TaskTimer) error {
	if t.New() {
		return nil
	}

	stoppedAt := time.Now().UTC()
	t.StoppedAt = &stoppedAt

	if err := h.db.Save(t).Error; err != nil {
		return err
	}

	if t.Harvest != nil {
		return t.Harvest.stopTimer(h.harvestClient)
	}

	t.Running = false
	h.CurrentTimer = nil
	return nil
}

func (h *harvester) updateTimers(currentRunning string) {
	for i, task := range h.Timers.Tasks {
		if task.Key != currentRunning {
			task.ID = 0
			task.StartedAt = time.Time{}
			task.StoppedAt = nil
		}

		task.Running = task.IsRunning()
		task.Runtime = task.CurrentRuntime()
		h.Timers.Tasks[i] = task
	}
}

func (h *harvester) GetActiveTimers() (TaskTimers, error) {
	var timers TaskTimers
	if err := h.db.Where("stopped_at is null").Find(&timers).Error; err != nil {
		return nil, err
	}

	for _, timer := range timers {
		if h.jiraClient != nil {
			jira, err := h.getJiraByKey(timer.Key)
			if err != nil {
				return nil, err
			}
			timer.Jira = jira
		}
		if h.harvestClient != nil {
			harvestTask, err := h.harvestClient.getUserProjectByKey(timer.Key)
			if err != nil {
				return nil, err
			}
			timer.Harvest = harvestTask
		}

		if timer.IsRunning() {
			timer.Running = true
			timer.Runtime = timer.CurrentRuntime()
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

func (t *TaskTimer) CurrentRuntime() string {
	runTime := time.Since(t.StartedAt)
	return fmt.Sprintf("%02d:%02.0f", int(runTime.Hours()), runTime.Minutes()-float64(int(runTime.Hours())*60))
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

func GetKeysWithTimes(db *gorm.DB, start, end time.Time) ([]string, error) {
	keyStructs := make([]struct{ Key string }, 0)
	if err := db.Table("task_timers").Select("key").Where("started_at between ? and ?", start, end).Group("key").Scan(&keyStructs).Error; err != nil {
		return nil, err
	}

	keys := make([]string, len(keyStructs))
	for i, key := range keyStructs {
		keys[i] = key.Key
	}

	return keys, nil
}
