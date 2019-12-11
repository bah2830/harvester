package harvester

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/dgraph-io/badger"
	"github.com/jinzhu/now"
)

var (
	ErrTimerNotExists = errors.New("timer not found")
)

type TaskTimer struct {
	Key       string       `json:"key"`
	StartedAt *time.Time   `json:"startedAt"`
	Running   bool         `json:"running"`
	Runtime   string       `json:"runtime"`
	Jira      *jira.Issue  `json:"jira"`
	Harvest   *harvestTask `json:"harvest"`
}
type TaskTimers []*TaskTimer

type StoredTimer struct {
	dbKey    []byte
	Key      string        `json:"key"`
	Day      time.Time     `json:"day"`
	Duration time.Duration `json:"duration"`
}
type StoredTimers []StoredTimer

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

	// Get any existing times in the database
	var timer *StoredTimer
	err := h.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(t.getDBKey())
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &timer)
		})
	})
	if err != nil {
		return err
	}

	if timer == nil {
		timer = &StoredTimer{
			Key:      t.Key,
			Day:      now.BeginningOfDay(),
			Duration: time.Since(*t.StartedAt),
		}
	} else {
		timer.Duration = timer.Duration + time.Since(*t.StartedAt)
	}

	err = h.db.Update(func(txn *badger.Txn) error {
		data, err := json.Marshal(timer)
		if err != nil {
			return err
		}

		return txn.Set(t.getDBKey(), data)
	})
	if err != nil {
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
	return []byte(fmt.Sprintf("timer.%s.%s", t.Key, t.StartedAt.Format("20060102")))
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
	for i, task := range h.Timers {
		if task.Key == t.Key {
			h.Timers[i] = t
			return
		}
	}

	h.Timers = append(h.Timers, t)
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
		timers, err := getTimersByOpts(db, badger.DefaultIteratorOptions)
		if err != nil {
			return err
		}

		return db.Update(func(txn *badger.Txn) error {
			for _, timer := range timers {
				if time.Since(timer.Day).Hours() < (90 * 24) {
					return nil
				}

				if err := txn.Delete(timer.dbKey); err != nil {
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
		if timer.Day.Before(start) {
			continue
		}
		if timer.Day.After(end) {
			break
		}
		keys = append(keys, timer.Key)
	}

	return keys, nil
}

func getTimersByOpts(db *badger.DB, opts badger.IteratorOptions) (StoredTimers, error) {
	if opts.Prefix == nil {
		opts.Prefix = []byte("timer.")
	}

	var timers StoredTimers
	err := db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(opts)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()

			var timer StoredTimer
			err := item.Value(func(v []byte) error {
				return json.Unmarshal(v, &timer)
			})
			if err != nil {
				return err
			}

			timer.dbKey = item.KeyCopy(nil)
			timers = append(timers, timer)
		}
		return nil
	})

	return timers, err
}
