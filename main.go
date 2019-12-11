package main

import (
	"flag"
	"log"
	"os"

	"github.com/bah2830/harvester/pkg/harvester"
	"github.com/dgraph-io/badger"
)

const (
	version = "alpha-1"
)

var (
	dbDir = flag.String("db.dir", "", "Path to the local database directory")
)

func main() {
	flag.Parse()

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Unable to get user home directory", err)
	}

	if *dbDir == "" {
		*dbDir = home + "/.harvester/db"
	}

	log.Printf("using database dir at %s", *dbDir)

	db, err := badger.Open(badger.DefaultOptions(*dbDir))
	if err != nil {
		log.Fatal("Unable to open database", err)
	}
	defer db.Close()

	h, err := harvester.NewHarvester(db)
	if err != nil {
		log.Fatalln("Unable to get new harvester", err)
	}
	go h.Start()
	defer h.Stop()

	if err := h.Run(); err != nil {
		log.Fatalln(err)
	}
}
