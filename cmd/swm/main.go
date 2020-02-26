package main

import (
	"flag"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
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

	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			xevent.Quit(X)
		},
	).Connect(X, X.RootWin(), "Mod1-x", true)

	xevent.Main(X)
}
