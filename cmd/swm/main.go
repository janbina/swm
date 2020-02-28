package main

import (
	"flag"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xwindow"
	"log"
)

func main() {

	replace := flag.Bool("replace", false, "whether swm should replace current wm")
	flag.Parse()

	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	defer X.Conn().Close()

	if err := takeWmOwnership(X, *replace); err != nil {
		log.Fatalf("Cannot take wm ownership: %s", err)
	}

	keybind.Initialize(X)

	if err := initRoot(X); err != nil {
		log.Fatalf("Cannot initialize root window %s", err)
	}

	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			xevent.Quit(X)
		},
	).Connect(X, X.RootWin(), "Mod1-x", true)

	xevent.Main(X)
}

var root *xwindow.Window
var heads xinerama.Heads
func initRoot(X *xgbutil.XUtil) error {
	root = xwindow.New(X, X.RootWin())

	rootGeometry, err := root.Geometry()
	if err != nil {
		return err
	}

	heads, err = xinerama.PhysicalHeads(X)
	if err != nil || len(heads) == 0 {
		heads = xinerama.Heads{rootGeometry}
	}

	log.Println("Root geometry: ", rootGeometry)
	log.Println("Heads: ", heads)

	return nil
}
