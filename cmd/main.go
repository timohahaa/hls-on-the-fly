package main

import (
	"flag"

	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/hls-on-the-fly/internal/origin"
)

func main() {
	addr := flag.String("addr", ":8001", "http port to listen on")
	domain := flag.String("domain", "127.0.0.1:8001", "domain to use in manifest links")
	filesBasePath := flag.String("basePath", "", "base path of a folder with files")

	flag.Parse()

	app, err := origin.New(*addr, *domain, *filesBasePath)
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
