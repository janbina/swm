package main

import (
	"flag"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/janbina/swm/windowmanager"
	"log"
	"os/exec"
)

func main() {

	replace := flag.Bool("replace", false, "whether swm should replace current wm")
	flag.Parse()

	if err := windowmanager.Initialize(*replace); err != nil {
		log.Fatalf("Cannot initialize window manager: %s", err)
	}

	keybind.Initialize(windowmanager.X)

	if err := windowmanager.SetupRoot(); err != nil {
		log.Fatalf("Cannot setup root window: %s", err)
	}

	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			windowmanager.ShutDown()
		},
	).Connect(windowmanager.X, windowmanager.Root.Id, "Mod1-x", true)

	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			exec.Command("xterm").Start()
		},
	).Connect(windowmanager.X, windowmanager.Root.Id, "Mod1-return", true)

	if err := windowmanager.Run(); err != nil {
		log.Fatalf("Error running window manager: %s", err)
	}
}
