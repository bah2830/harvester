package harvester

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/dgraph-io/badger"
)

var (
	ErrTimerNotExists = errors.New("timer not found")
)

type TaskTimer struct {
	DBKey     []byte       `json:"-"`
	Key       string       `json:"key"`
	StartedAt *time.Time   `json:"startedAt"`
	StoppedAt *time.Time   `json:"stoppedAt"`
	Running   bool         `json:"running"`
	Runtime   string       `json:"runtime"`
	Jira      *jira.Issue  `json:"jira"`
	Harvest   *harvestTask `json:"harvest"`
}

type TaskTimers []*TaskTimer

func (h *harvester) StartTimer(t *TaskTimer) error {
	if err := h.stopAllTimers(); err != nil {
		return err
	}

	startedAt := time.Now().UTC()
	newTimer := &TaskTimer{
		Key:       t.Key,
		Jira:      t.Jira,
		Harvest:   t.Harvest,
		StartedAt: &startedAt,
		Running:   true,
	}
	newTimer.Runtime = newTimer.CurrentRuntime()

	if err := h.saveTimer(newTimer); err != nil {
		return err
	}

	// If a harvest task exists start the timer for it
	if newTimer.Harvest != nil {
		if err := newTimer.Harvest.startTimer(h.harvestClient); err != nil {
			return err
		}
	}

	h.replaceTask(newTimer)
	return nil
}

func (h *harvester) StopTimer(t *TaskTimer) error {
	if !t.Running {
		return nil
	}

	stoppedAt := time.Now().UTC()
	t.StoppedAt = &stoppedAt
	if err := h.saveTimer(t); err != nil {
		return err
	}

	if t.Harvest != nil {
		return t.Harvest.stopTimer(h.harvestClient)
	}

	// Rest the in memory task
	newTimer := &TaskTimer{
		Key:     t.Key,
		Jira:    t.Jira,
		Harvest: t.Harvest,
	}
	h.replaceTask(newTimer)
	return nil
}

func (t *TaskTimer) getDBKey() []byte {
	return []byte(fmt.Sprintf("timer.%s.%s", t.StartedAt.Format("20060102_150405.99"), t.Key))
}

func (h *harvester) saveTimer(t *TaskTimer) error {
	taskCopy := *t
	taskCopy.Jira = nil
	taskCopy.Harvest = nil
	taskData, err := json.Marshal(taskCopy)
	if err != nil {
		return err
	}

	return h.db.Update(func(txn *badger.Txn) error {
		return txn.Set(t.getDBKey(), taskData)
	})
}

func (h *harvester) replaceTask(t *TaskTimer) {
	for i, task := range h.Timers.Tasks {
		if task.Key == t.Key {
			h.Timers.Tasks[i] = t
			return
		}
	}

	h.Timers.Tasks = append(h.Timers.Tasks, t)
}

func (h *harvester) GetActiveTimers() (TaskTimers, error) {
	allTimers, err := getTimersByOpts(h.db, badger.DefaultIteratorOptions)
	if err != nil {
		return nil, err
	}

	var timers TaskTimers
	for _, timer := range allTimers {
		if timer.StoppedAt != nil {
			continue
		}

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

		timer.Runtime = timer.CurrentRuntime()
		timers = append(timers, timer)
	}

	return timers, nil
}

func (t *TaskTimer) CurrentRuntime() string {
	if t.StartedAt == nil {
		return ""
	}

	runTime := time.Since(*t.StartedAt)
	return fmt.Sprintf("%02d:%02.0f", int(runTime.Hours()), runTime.Minutes()-float64(int(runTime.Hours())*60))
}

// StartJiraPurger will check for old jiras every few hours and purge any that are more than 90 days old
func StartJiraPurger(db *badger.DB) {
	purge := func() error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		timers, err := getTimersByOpts(db, opts)
		if err != nil {
			return err
		}

		return db.Update(func(txn *badger.Txn) error {
			for _, timer := range timers {
				if time.Since(*timer.StartedAt).Hours() < (90 * 24) {
					return nil
				}

				if err := txn.Delete(timer.DBKey); err != nil {
					return err
				}
			}
			return nil
		})
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

func GetKeysWithTimes(db *badger.DB, start, end time.Time) ([]string, error) {
	timers, err := getTimersByOpts(db, badger.DefaultIteratorOptions)
	if err != nil {
		return nil, err
	}

	var keys []string
	for _, timer := range timers {
		if timer.StartedAt.Before(start) {
			continue
		}
		if timer.StoppedAt == nil || timer.StoppedAt.After(end) {
			break
		}
		keys = append(keys, timer.Key)
	}

	return keys, nil
}

func getTimersByOpts(db *badger.DB, opts badger.IteratorOptions) (TaskTimers, error) {
	if opts.Prefix == nil {
		opts.Prefix = []byte("timer.")
	}

	var timers TaskTimers
	err := db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(opts)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()

			var taskTimer TaskTimer
			err := item.Value(func(v []byte) error {
				return json.Unmarshal(v, &taskTimer)
			})
			if err != nil {
				return err
			}

			taskTimer.DBKey = item.KeyCopy(nil)
			timers = append(timers, &taskTimer)
		}
		return nil
	})

	return timers, err
}
