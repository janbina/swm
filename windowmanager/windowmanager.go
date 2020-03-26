package windowmanager

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/janbina/swm/cursors"
	"github.com/janbina/swm/geometry"
	"github.com/janbina/swm/window"
	"log"
)

var X *xgbutil.XUtil
var Root *xwindow.Window
var RootGeometry xrect.Rect
var Heads xinerama.Heads

var moveDragShortcut = "Mod1-1"
var resizeDragShortcut = "Mod1-3"

var desktops []string
var desktopToWins map[int][]xproto.Window
var currentDesktop int
var managedWindows map[xproto.Window]*window.Window
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

	Root.Change(xproto.CwCursor, uint32(cursors.LeftPtr))

	RootGeometry, err = Root.Geometry()
	if err != nil {
		X.Conn().Close()
		return err
	}

	Heads, err = xinerama.PhysicalHeads(X)
	if err != nil || len(Heads) == 0 {
		Heads = xinerama.Heads{RootGeometry}
	}

	managedWindows = make(map[xproto.Window]*window.Window)
	desktopToWins = make(map[int][]xproto.Window)

	desktops = getDesktops() // init desktops
	currentDesktop = getCurrentDesktop()
	setDesktops()
	setCurrentDesktop()

	setEwmhSupported(X)

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
	xevent.ClientMessageFun(handleClientMessage).Connect(X, Root.Id)

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

func FindWindowById(id uint32) *window.Window {
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

func DestroyWindow(id uint32) {
	win := FindWindowById(id)
	if win == nil {
		return
	}
	log.Printf("Destroy win %d", id)
	win.Destroy()
}

func ResizeActiveWindow(directions window.Directions) {
	if activeWindow != nil {
		ResizeWindow(activeWindow.Id(), directions)
	}
}

func ResizeWindow(id uint32, directions window.Directions) {
	win := FindWindowById(id)
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

func MoveWindow(id uint32, x, y int) {
	win := FindWindowById(id)
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

func MoveResizeWindow(id uint32, x, y, width, height int) {
	win := FindWindowById(id)
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
	dragShortcutChanged()
	return nil
}

func SetResizeDragShortcut(s string) error {
	if _, _, err := mousebind.ParseString(X, s); err != nil {
		return err
	}
	resizeDragShortcut = s
	dragShortcutChanged()
	return nil
}

func GetCurrentScreenGeometry() xrect.Rect {
	return Heads[0]
}

func GetActiveWindowGeometry() (*geometry.Geometry, error) {
	if activeWindow != nil {
		return GetWindowGeometry(activeWindow.Id())
	}
	return nil, fmt.Errorf("no active window")
}

func GetWindowGeometry(id uint32) (*geometry.Geometry, error) {
	win := FindWindowById(id)
	if win == nil {
		return nil, fmt.Errorf("cannot find window with id %d", id)
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

func destroyNotify(w xproto.Window) {
	delete(managedWindows, w)
	updateClientList()
}

func manageWindow(w xproto.Window) {
	win := window.New(X, w)
	managedWindows[w] = win
	d := getDesktopForWindow(w)
	desktopToWins[d] = append(desktopToWins[d], w)
	ewmh.WmDesktopSet(X, w, uint(d))

	updateClientList()
	win.Map()
	win.Listen(
		xproto.EventMaskStructureNotify,
		xproto.EventMaskEnterWindow,
		xproto.EventMaskFocusChange,
	)
	xevent.FocusInFun(func(x *xgbutil.XUtil, e xevent.FocusInEvent) {
		log.Printf("Focus in event: %s", e)
		activeWindow = win
	}).Connect(X, w)
	xevent.UnmapNotifyFun(func(x *xgbutil.XUtil, e xevent.UnmapNotifyEvent) {
		log.Printf("UNMAP notify: %s", e)
		//destroyNotify(win)
	}).Connect(X, w)
	xevent.ClientMessageFun(func(x *xgbutil.XUtil, e xevent.ClientMessageEvent) {
		name, err := xprop.AtomName(x, e.Type)
		if err != nil {
			log.Printf("Cannot get property atom name for clientMessage event: %s", err)
			return
		}
		win.HandleClientMessage(name, e.Data.Data32)
	}).Connect(X, w)
	xevent.DestroyNotifyFun(func(x *xgbutil.XUtil, e xevent.DestroyNotifyEvent) {
		log.Printf("Destroy notify: %s", e)
		mousebind.Detach(x, w)
		xevent.Detach(x, w)
		win.Destroyed()
		destroyNotify(w)
	}).Connect(X, w)
	win.Focus()
	win.SetupMouseEvents(moveDragShortcut, resizeDragShortcut)
}

func setNumberOfDesktops(num int) {
	if num < 1 {
		num = 1
	}
	currentNum := len(desktops)
	newLast := num -1
	if num < currentNum {
		for i := num; i < currentNum; i++ {
			for _, x := range desktopToWins[i] {
				ewmh.WmDesktopSet(X, x, uint(newLast))
			}
			desktopToWins[newLast] = append(desktopToWins[newLast], desktopToWins[i]...)
		}
		desktops = desktops[:num]
		setDesktops()
		if currentDesktop > newLast {
			switchToDesktop(newLast)
		}
	} else if num > currentNum {
		desktops = append(desktops, getDesktopNames(currentNum, newLast)...)
		setDesktops()
	}
}

func switchToDesktop(index int) {
	if currentDesktop != index && index < len(desktops) {
		for _, w := range desktopToWins[currentDesktop] {
			managedWindows[w].Unmap()
		}
		for _, w := range desktopToWins[index] {
			managedWindows[w].Map()
		}
		currentDesktop = index
		setCurrentDesktop()
	}
}

func SetDesktopNames(names []string) {
	for i, name := range names {
		if i < len(desktops) {
			desktops[i] = name
		}
	}
	if len(names) > len(desktops) {
		setDesktopNames(names)
	} else {
		setDesktopNames(desktops)
	}
}

func getDesktopForWindow(win xproto.Window) int {
	_d, err := ewmh.WmDesktopGet(X, win)
	d := int(_d)
	if err != nil {
		// not specified
		return currentDesktop
	}
	if d == 0xFFFFFFFF || d < len(desktops) {
		return d
	}
	// TODO: Current, last, create additional desktops, or what?
	return len(desktops) - 1
}

func dragShortcutChanged() {
	for _, win := range managedWindows {
		win.SetupMouseEvents(moveDragShortcut, resizeDragShortcut)
	}
}
