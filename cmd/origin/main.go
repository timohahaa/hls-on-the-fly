package main

import (
	"flag"

	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/hls-on-the-fly/internal/origin"
)

func main() {
	addr := flag.String("addr", ":8001", "http port to listen on")

	flag.Parse()

	app, err := origin.New(*addr)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
