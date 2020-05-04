package main

import (
	"flag"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/janbina/swm/internal/communication"
	"github.com/janbina/swm/internal/windowmanager"
	"github.com/shibukawa/configdir"
	"log"
	"os/exec"
	"path/filepath"
)

func main() {

	replace := flag.Bool("replace", false, "whether swm should replace current wm")
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

	runConfig()

	windowmanager.ManageExistingClients()

	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			exec.Command("xterm").Start()
		},
	).Connect(windowmanager.X, windowmanager.Root.Id, "control-Mod1-return", true)

	windowmanager.Run()
}

func runConfig() {
	dirName := "swm"
	fileName := "swmrc"

	log.Printf("Trying to execute config")
	dir := configdir.New("", dirName).QueryFolderContainsFile(fileName)

	if dir == nil {
		log.Printf("No config file to execute")
		return
	}

	file := filepath.Join(dir.Path, fileName)

	log.Printf("Found config file at \"%s\"", file)

	err := exec.Command(file).Run()

	if err != nil {
		log.Printf("Error executing config file: %s", err)
	}
}
