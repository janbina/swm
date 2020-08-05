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
	"github.com/janbina/swm/internal/cursors"
	"github.com/janbina/swm/internal/focus"
	"github.com/janbina/swm/internal/groupmanager"
	"github.com/janbina/swm/internal/heads"
	"github.com/janbina/swm/internal/log"
	"github.com/janbina/swm/internal/stack"
	"github.com/janbina/swm/internal/window"
)

var (
	X                  *xgbutil.XUtil
	Root               *xwindow.Window
	RootGeometry       xrect.Rect
	RootGeometryStruts xrect.Rect

	managedWindows map[xproto.Window]*window.Window
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
	groupmanager.Initialize(X)

	registerCommands()

	if err = takeWmOwnership(X, replace); err != nil {
		return err
	}

	Root = xwindow.New(X, X.RootWin())

	Root.Change(xproto.CwCursor, uint32(cursors.LeftPtr))

	managedWindows = make(map[xproto.Window]*window.Window)
	strutWindows = make(map[xproto.Window]bool)

	if err = loadGeometriesAndHeads(); err != nil {
		return err
	}

	heads.InitScreen(X)

	setEwmhSupported(X)

	return nil
}

// Setup event listeners
func SetupRoot() error {
	if err := Root.Listen(
		xproto.EventMaskStructureNotify,
		xproto.EventMaskSubstructureRedirect,
		xproto.EventMaskSubstructureNotify,
	); err != nil {
		return err
	}

	xevent.ConfigureRequestFun(configureRequestFun).Connect(X, Root.Id)
	xevent.MapRequestFun(mapRequestFun).Connect(X, Root.Id)
	xevent.ClientMessageFun(handleRootClientMessage).Connect(X, Root.Id)
	xevent.ConfigureNotifyFun(func(X *xgbutil.XUtil, e xevent.ConfigureNotifyEvent) {
		log.Infof("Root geometry changed: %s", e)
		_ = loadGeometriesAndHeads()
	}).Connect(X, Root.Id)

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

func loadGeometriesAndHeads() error {
	var err error
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

	applyStruts()

	return nil
}
