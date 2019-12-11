package harvester

import (
	"encoding/json"
	"time"

	"github.com/dgraph-io/badger"
)

const (
	defaultRefreshInterval = 5 * time.Minute
)

type Settings struct {
	Jira    SettingsData `json:"jira"`
	Harvest SettingsData `json:"harvest"`
}

type SettingsData struct {
	URL  string `json:"url"`
	User string `json:"user"`
	Pass string `json:"pass"`
}

func (s *Settings) Save(db *badger.DB) error {
	settings, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("settings"), settings)
	})
}

func GetSettings(db *badger.DB) (*Settings, error) {
	var settingsValue []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("settings"))
		if err != nil {
			return err
		}

		settingsValue, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	var settings Settings
	if err := json.Unmarshal(settingsValue, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}
