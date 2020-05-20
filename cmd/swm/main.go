package main

import (
	"flag"
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/janbina/swm/internal/communication"
	"github.com/janbina/swm/internal/config"
	"github.com/janbina/swm/internal/windowmanager"
)

func main() {

	replace := flag.Bool("replace", false, "whether swm should replace current wm")
	customConfig := flag.String("c", "", "path to swmrc file")
	flag.Parse()

	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatalf("Cannot initialize x connection: %s", err)
	}
	defer X.Conn().Close()

	if err := windowmanager.Initialize(X, *replace); err != nil {
		log.Fatalf("Cannot initialize window manager: %s", err)
	}

	if err := windowmanager.SetupRoot(); err != nil {
		log.Fatalf("Cannot setup root window: %s", err)
	}

	go communication.Listen(X.Conn())

	config.FindAndRunSwmrc(*customConfig)

	windowmanager.ManageExistingClients()

	windowmanager.Run()
}
