package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/janbina/swm/communication"
	"github.com/janbina/swm/cursors"
	"github.com/janbina/swm/windowmanager"
	"log"
	"os"
	"os/exec"
)

func main() {

	replace := flag.Bool("replace", false, "whether swm should replace current wm")
	showSocket := flag.Bool("show-socket", false, "show path to swm server socket")
	flag.Parse()

	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatalf("Cannot initialize x connection: %s", err)
	}
	defer X.Conn().Close()

	if *showSocket {
		socket := communication.GetSocketFilePath(X.Conn())
		fmt.Println(socket)
		os.Exit(0)
	}

	keybind.Initialize(X)
	mousebind.Initialize(X)
	cursors.Initialize(X)

	if err := windowmanager.Initialize(X, *replace); err != nil {
		log.Fatalf("Cannot initialize window manager: %s", err)
	}

	if err := windowmanager.SetupRoot(); err != nil {
		log.Fatalf("Cannot setup root window: %s", err)
	}

	windowmanager.ManageExistingClients()

	go communication.Listen(X.Conn())

	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			exec.Command("xterm").Start()
		},
	).Connect(windowmanager.X, windowmanager.Root.Id, "Mod1-return", true)

	if err := windowmanager.Run(); err != nil {
		log.Fatalf("Error running window manager: %s", err)
	}
}
