package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/xgbutil"
	"github.com/janbina/swm/internal/buildconfig"
	"github.com/janbina/swm/internal/communication"
	"github.com/janbina/swm/internal/config"
	"github.com/janbina/swm/internal/log"
	"github.com/janbina/swm/internal/windowmanager"
)

func main() {

	replace := flag.Bool("replace", false, "whether swm should replace current wm")
	customConfig := flag.String("c", "", "path to swmrc file")
	version := flag.Bool("v", false, "print swm version")
	debugLog := flag.String("d", "", "path to debug log file")
	flag.Parse()

	log.Init(*debugLog)
	defer log.Sync()

	if *version {
		fmt.Println(buildconfig.Version)
		os.Exit(0)
	}

	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal("Cannot initialize x connection: %s", err)
	}
	defer X.Conn().Close()

	if err := windowmanager.Initialize(X, *replace); err != nil {
		log.Fatal("Cannot initialize window manager: %s", err)
	}

	if err := windowmanager.SetupRoot(); err != nil {
		log.Fatal("Cannot setup root window: %s", err)
	}

	go communication.Listen(X.Conn())

	config.FindAndRunSwmrc(*customConfig)

	windowmanager.ManageExistingClients()

	windowmanager.Run()
}
