package windowmanager

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/janbina/swm/cursors"
	"github.com/janbina/swm/focus"
	"github.com/janbina/swm/geometry"
	"github.com/janbina/swm/heads"
	"github.com/janbina/swm/stack"
	"github.com/janbina/swm/window"
)

type ManagedWindow interface {
	Destroy()
	Resize(directions window.Directions)
	Move(x int, y int)
	MoveResize(x int, y int, width int, height int)
	Geometry() (*geometry.Geometry, error)
	Map()
	Unmap()
	IsHidden() bool
	SetupMouseEvents(moveShortcut string, resizeShortcut string)
	Destroyed()
}

const (
	minDesktops   = 1 //minimum number of desktops created at startup
	stickyDesktop = 0xFFFFFFFF
)

var (
	X                  *xgbutil.XUtil
	Root               *xwindow.Window
	RootGeometry       xrect.Rect
	RootGeometryStruts xrect.Rect

	desktops       []string
	desktopToWins  map[int][]xproto.Window
	currentDesktop int
	managedWindows map[xproto.Window]ManagedWindow
	strutWindows   map[xproto.Window]bool

	cycleState int
)

// Take wm ownership and initialize variables
func Initialize(x *xgbutil.XUtil, replace bool) error {
	var err error
	X = x

	keybind.Initialize(X)
	mousebind.Initialize(X)
	cursors.Initialize(X)
	focus.Initialize(X)
	stack.Initialize(X)

	if err = takeWmOwnership(X, replace); err != nil {
		return err
	}

	Root = xwindow.New(X, X.RootWin())

	Root.Change(xproto.CwCursor, uint32(cursors.LeftPtr))

	RootGeometry, err = Root.Geometry()
	if err != nil {
		return err
	}
	setDesktopGeometry()
	setDesktopViewport()

	heads.Heads, err = xinerama.PhysicalHeads(X)
	if err != nil || len(heads.Heads) == 0 {
		heads.Heads = xinerama.Heads{RootGeometry}
	}

	managedWindows = make(map[xproto.Window]ManagedWindow)
	desktopToWins = make(map[int][]xproto.Window)
	strutWindows = make(map[xproto.Window]bool)

	desktops = getDesktops() // init desktops
	currentDesktop = getCurrentDesktop()
	applyStruts()
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
	xevent.ClientMessageFun(handleRootClientMessage).Connect(X, Root.Id)

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

func Run() {
	xevent.Main(X)
}

func Shutdown() {
	xevent.Quit(X)
}
