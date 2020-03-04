package windowmanager

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xcursor"
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

var moveDragShortcut = ""

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

		manageWindow(child)
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

func Shutdown() {
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

func ResizeActiveWindow(directions window.Directions) {
	if activeWindow != nil {
		ResizeWindow(activeWindow.Id(), directions)
	}
}

func ResizeWindow(xWin xproto.Window, directions window.Directions) {
	win := FindWindowById(xWin)
	if win == nil {
		return
	}
	win.Resize(directions)
}

func MoveActiveWindow(x, y int) {
	if activeWindow != nil {
		MoveWindow(activeWindow.Id(), x, y)
	}
}

func MoveWindow(xWin xproto.Window, x, y int) {
	win := FindWindowById(xWin)
	if win == nil {
		return
	}
	win.Move(x, y)
}

func MoveResizeActiveWindow(x, y, width, height int) {
	if activeWindow != nil {
		MoveResizeWindow(activeWindow.Id(), x, y, width, height)
	}
}

func MoveResizeWindow(xWin xproto.Window, x, y, width, height int) {
	win := FindWindowById(xWin)
	if win == nil {
		return
	}
	win.MoveResize(x, y, width, height)
}

func SetMoveDragShortcut(s string) error {
	if _, _, err := mousebind.ParseString(X, s); err != nil {
		return err
	}
	moveDragShortcut = s
	moveDragShortcutChanged()
	return nil
}

func GetCurrentScreenGeometry() xrect.Rect {
	return Heads[0]
}

func GetActiveWindowGeometry() (xrect.Rect, error) {
	if activeWindow != nil {
		return GetWindowGeometry(activeWindow.Id())
	}
	return nil, fmt.Errorf("no active window")
}

func GetWindowGeometry(xWin xproto.Window) (xrect.Rect, error) {
	win := FindWindowById(xWin)
	if win == nil {
		return nil, fmt.Errorf("cannot find window with id %d", xWin)
	}
	return win.Geometry()
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
	manageWindow(e.Window)
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

func manageWindow(w xproto.Window) {
	win := window.New(X, w)
	managedWindows = append(managedWindows, win)
	win.Map()
	win.Listen(
		xproto.EventMaskStructureNotify,
		xproto.EventMaskEnterWindow,
		xproto.EventMaskFocusChange,
	)
	xevent.DestroyNotifyFun(destroyNotifyFun).Connect(X, w)
	xevent.FocusInFun(func(x *xgbutil.XUtil, e xevent.FocusInEvent) {
		log.Printf("Focus in event: %s", e)
		activeWindow = win
	}).Connect(X, w)
	win.Focus()
	setupMoveDrag(win)
}

func moveDragShortcutChanged() {
	for _, win := range managedWindows {
		mousebind.DetachPress(X, win.Id())
		setupMoveDrag(win)
	}
}

func setupMoveDrag(win *window.Window) {
	if _, _, err := mousebind.ParseString(X, moveDragShortcut); err != nil {
		return
	}
	dStart := xgbutil.MouseDragBeginFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) (bool, xproto.Cursor) {
			cursor, _ := xcursor.CreateCursor(X, xcursor.Fleur)
			return win.DragMoveBegin(rx, ry, ex, ey), cursor
		})
	dStep := xgbutil.MouseDragFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
			win.DragMoveStep(rx, ry, ex, ey)
		})
	dEnd := xgbutil.MouseDragFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
			win.DragMoveEnd(rx, ry, ex, ey)
		})
	mousebind.Drag(X, X.Dummy(), win.Id(), moveDragShortcut, true, dStart, dStep, dEnd)
}
