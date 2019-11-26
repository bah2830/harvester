package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultRefreshInterval = 1 * time.Minute
	version                = "alpha-1"
)

var (
	dbFile = flag.String("db.file", "", "Path to the local database")
)

func main() {
	flag.Parse()

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Unable to get user home directory", err)
	}

	if *dbFile == "" {
		*dbFile = home + "/.harvester.db"
	}

	log.Printf("using database file at %s", *dbFile)

	db, err := gorm.Open("sqlite3", *dbFile)
	if err != nil {
		log.Fatalln("Unable to open sql database", err)
	}
	defer db.Close()

	// Build the database schema if needed
	if err := db.AutoMigrate(&Settings{}, &TaskTimer{}).Error; err != nil {
		log.Fatalln("Migration error", err)
	}

	h, err := NewHarvester(db)
	if err != nil {
		log.Fatalln("Unable to get new harvester", err)
	}
	go h.start()
	defer h.stop()

	h.app.Wait()
}
