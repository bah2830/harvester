package main

import (
	"flag"
	"time"
)

const (
	defaultRefreshInterval = 30 * time.Second
)

var (
	dbFile = flag.String("db.file", "./harvester.db", "Path to the local database")
)

func main() {
	flag.Parse()
	h := newHarvester()
	go h.start()
	defer h.stop()
	h.app.Run()
}
