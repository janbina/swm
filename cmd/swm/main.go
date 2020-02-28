package main

import (
	"flag"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xwindow"
	"log"
	"os/exec"
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

	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			exec.Command("xterm").Start()
		},
	).Connect(X, X.RootWin(), "Mod1-return", true)

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

	// as a wm, we have to listen for some events on root window...
	if err := root.Listen(
		xproto.EventMaskSubstructureRedirect,
		xproto.EventMaskSubstructureNotify,
	); err != nil {
		return err
	}

	// Create notify just informs us that new window has been created, we don't have to respond,
	// in fact we can do nothing about that
	xevent.CreateNotifyFun(
		func(xu *xgbutil.XUtil, event xevent.CreateNotifyEvent) {
			log.Printf("Create notify: %s", event)
		},
	).Connect(X, root.Id)

	// Now app sends configuration request and we should respond somehow
	xevent.ConfigureRequestFun(
		func(xu *xgbutil.XUtil, e xevent.ConfigureRequestEvent) {
			log.Printf("Configure request: %s", e)
			xwindow.New(X, e.Window).Configure(
				int(e.ValueMask),
				int(e.X),
				int(e.Y),
				int(e.Width),
				int(e.Height),
				e.Sibling,
				e.StackMode,
			)
		},
	).Connect(X, root.Id)

	// Now map request is the one where we would show window on screen
	xevent.MapRequestFun(
		func(xu *xgbutil.XUtil, e xevent.MapRequestEvent) {
			log.Printf("Map request: %s", e)
			xwindow.New(X, e.Window).Map()
		},
	).Connect(X, root.Id)

	// Map notify is again just notification that window was mapped
	xevent.MapNotifyFun(
		func(xu *xgbutil.XUtil, event xevent.MapNotifyEvent) {
			log.Printf("Map notify: %s", event)
		},
	).Connect(X, root.Id)

	return nil
}
