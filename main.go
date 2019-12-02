package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/bah2830/harvester/pkg/harvester"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

const (
	version = "alpha-1"
)

var (
	dbFile = flag.String("db.file", "", "Path to the local database")
	debug  = flag.Bool("debug", false, "Print debug information")
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
	if err := db.AutoMigrate(&harvester.Settings{}, &harvester.TaskTimer{}).Error; err != nil {
		log.Fatalln("Migration error", err)
	}

	h, err := harvester.NewHarvester(db, *debug)
	if err != nil {
		log.Fatalln("Unable to get new harvester", err)
	}
	go h.Start()
	defer h.Stop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		h.Stop()
	}()

	h.Run()
}
