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
var activeWindow *window.Window

// Initialize connection to x server, take wm ownership and initialize variables
func Initialize(x *xgbutil.XUtil, replace bool) error {
	var err error
	X = x

	if err = takeWmOwnership(X, replace); err != nil {
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


func ManageExistingClients() error {
	tree, err := xproto.QueryTree(X.Conn(), Root.Id).Reply()
	if err != nil {
		return err
	}
	for _, child := range tree.Children {
		if child == X.Dummy() {
			continue
		}

		attrs, err := xproto.GetWindowAttributes(X.Conn(), child).Reply()
		if err != nil {
			continue
		}
		if attrs.MapState == xproto.MapStateUnmapped {
			continue
		}

		win := window.New(X, child)
		managedWindows = append(managedWindows, win)
		win.Map()
		win.Listen(
			xproto.EventMaskStructureNotify,
			xproto.EventMaskEnterWindow,
			xproto.EventMaskFocusChange,
		)
		xevent.DestroyNotifyFun(destroyNotifyFun).Connect(X, child)
		xevent.FocusInFun(func(x *xgbutil.XUtil, e xevent.FocusInEvent) {
			log.Printf("Focus in event: %s", e)
			activeWindow = win
		}).Connect(X, child)
		win.Focus()
	}
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

func ShutDown() {
	xevent.Quit(X)
}

func FindWindowById(id xproto.Window) *window.Window {
	for _, win := range managedWindows {
		if win.Id() == id {
			return win
		}
	}
	return nil
}

func DestroyActiveWindow() {
	if activeWindow != nil {
		DestroyWindow(activeWindow.Id())
	}
}

func DestroyWindow(xWin xproto.Window) {
	win := FindWindowById(xWin)
	if win == nil {
		return
	}
	log.Printf("Destroy win %d", xWin)
	win.Destroy()
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
	win.Listen(
		xproto.EventMaskStructureNotify,
		xproto.EventMaskEnterWindow,
		xproto.EventMaskFocusChange,
	)
	xevent.DestroyNotifyFun(destroyNotifyFun).Connect(x, e.Window)
	xevent.FocusInFun(func(x *xgbutil.XUtil, e xevent.FocusInEvent) {
		log.Printf("Focus in event: %s", e)
		activeWindow = win
	}).Connect(x, e.Window)
	win.Focus()
}

func destroyNotifyFun(x *xgbutil.XUtil, e xevent.DestroyNotifyEvent) {
	log.Printf("Destroy notify: %s", e)
	for i, win := range managedWindows {
		if win.Id() == e.Window {
			managedWindows = append(managedWindows[0:i], managedWindows[i+1:]...)
			if activeWindow != nil && activeWindow.Id() == e.Window {
				activeWindow = nil
			}
			return
		}
	}
}
