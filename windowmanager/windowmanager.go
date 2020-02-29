package windowmanager

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/janbina/swm/window"
	"log"
)

var X *xgbutil.XUtil
var Root *xwindow.Window
var RootGeometry xrect.Rect
var Heads xinerama.Heads

var managedWindows []*window.Window

// Initialize connection to x server, take wm ownership and initialize variables
func Initialize(replace bool) error {
	var err error
	X, err = xgbutil.NewConn()
	if err != nil {
		return err
	}

	if err := takeWmOwnership(X, replace); err != nil {
		X.Conn().Close()
		return err
	}

	Root = xwindow.New(X, X.RootWin())

	RootGeometry, err = Root.Geometry()
	if err != nil {
		X.Conn().Close()
		return err
	}

	Heads, err = xinerama.PhysicalHeads(X)
	if err != nil || len(Heads) == 0 {
		Heads = xinerama.Heads{RootGeometry}
	}

	return nil
}

// Setup event listeners
func SetupRoot() error {
	if err := Root.Listen(
		xproto.EventMaskSubstructureRedirect,
		xproto.EventMaskSubstructureNotify,
	); err != nil {
		return err
	}

	xevent.ConfigureRequestFun(configureRequestFun).Connect(X, Root.Id)
	xevent.MapRequestFun(mapRequestFun).Connect(X, Root.Id)

	return nil
}

// Rut the window manager - main event loop
func Run() error {
	if X == nil {
		return fmt.Errorf("cannot run window manager, X is nil")
	}

	xevent.Main(X)
	return nil
}

func configureRequestFun(x *xgbutil.XUtil, e xevent.ConfigureRequestEvent) {
	log.Printf("Configure request: %s", e)
	xwindow.New(x, e.Window).Configure(
		int(e.ValueMask),
		int(e.X),
		int(e.Y),
		int(e.Width),
		int(e.Height),
		e.Sibling,
		e.StackMode,
	)
}

func mapRequestFun(x *xgbutil.XUtil, e xevent.MapRequestEvent) {
	log.Printf("Map request: %s", e)
	win := window.New(x, e.Window)
	managedWindows = append(managedWindows, win)
	win.Map()
	win.Focus()
	win.Listen(
		xproto.EventMaskStructureNotify,
		xproto.EventMaskEnterWindow,
	)
	xevent.DestroyNotifyFun(destroyNotifyFun).Connect(x, e.Window)
}

func destroyNotifyFun(x *xgbutil.XUtil, e xevent.DestroyNotifyEvent) {
	log.Printf("Destroy notify: %s", e)
	for i, win := range managedWindows {
		if win.Id() == e.Window {
			managedWindows = append(managedWindows[0:i], managedWindows[i+1:]...)
		}
	}
}
