package harvester

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
)

const (
	defaultRefreshInterval = 1 * time.Minute
)

type Settings struct {
	ID       int `gorm:"primary_key;AUTO_INCREMENT"`
	Settings string

	RefreshInterval time.Duration `gorm:"-"`
	Jira            SettingsData  `gorm:"-" json:"jira"`
	Harvest         SettingsData  `gorm:"-" json:"harvest"`
}

type SettingsData struct {
	URL  string `json:"url"`
	User string `json:"user"`
	Pass string `json:"pass"`
}

func (s *Settings) Save(db *gorm.DB) error {
	// Clear the encoded settings string
	s.Settings = ""

	settings, err := json.Marshal(s)
	if err != nil {
		return err
	}

	if s.ID == 0 {
		return db.Create(s).Error
	}

	base64Settings := base64.StdEncoding.EncodeToString(settings)
	s.Settings = base64Settings
	return db.Save(s).Error
}

func GetSettings(db *gorm.DB) (*Settings, error) {
	var settings Settings
	if err := db.Last(&settings).Error; err != nil {
		return nil, err
	}

	decodedSettings, err := base64.StdEncoding.DecodeString(settings.Settings)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(decodedSettings, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}
