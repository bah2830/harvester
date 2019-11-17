package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultRefreshInterval = 5 * time.Minute
)

var (
	dbFile = flag.String("db.file", "./harvester.db", "Path to the local database")
)

func main() {
	flag.Parse()

	db, err := sql.Open("sqlite3", *dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := databaseMigrate(db); err != nil {
		log.Fatal(err)
	}

	h, err := newHarvester(db)
	if err != nil {
		log.Fatal(err)
	}
	go h.start()
	defer h.stop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("received kill signal")
		h.stop()
		os.Exit(0)
	}()

	h.app.Run()
}

func databaseMigrate(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "sqlite3", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
