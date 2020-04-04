package windowmanager

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"log"
)

func updateClientList() {
	ids := make([]xproto.Window, 0, len(managedWindows))
	for window := range managedWindows {
		ids = append(ids, window)
	}
	_ = ewmh.ClientListSet(X, ids)
}

func setDesktopGeometry() {
	_ = ewmh.DesktopGeometrySet(
		X,
		&ewmh.DesktopGeometry{
			Width:  RootGeometry.Width(),
			Height: RootGeometry.Height(),
		},
	)
}

func setDesktopViewport() {
	_ = ewmh.DesktopViewportSet(X, []ewmh.DesktopViewport{{0, 0}})
}

func setWorkArea(numDesktops int) {
	areas := make([]ewmh.Workarea, numDesktops)
	for i := range areas {
		areas[i] = ewmh.Workarea{
			X:      RootGeometryStruts.X(),
			Y:      RootGeometryStruts.Y(),
			Width:  uint(RootGeometryStruts.Width()),
			Height: uint(RootGeometryStruts.Height()),
		}
	}
	_ = ewmh.WorkareaSet(X, areas)
}

func setWmAllowedActions(win xproto.Window) {
	_ = ewmh.WmAllowedActionsSet(X, win, windowAllowedActions)
}

func setEwmhSupported(X *xgbutil.XUtil) {
	// Set supported atoms
	if err := ewmh.SupportedSet(X, ewmhSupported); err != nil {
		log.Println(err)
	}

	// While we're at it, set the supporting wm hint too.
	if err := ewmh.SupportingWmCheckSet(X, X.RootWin(), X.Dummy()); err != nil {
		log.Println(err)
	}
	if err := ewmh.SupportingWmCheckSet(X, X.Dummy(), X.Dummy()); err != nil {
		log.Println(err)
	}
	if err := ewmh.WmNameSet(X, X.Dummy(), "Swm"); err != nil {
		log.Println(err)
	}
}

var ewmhSupported = []string{
	"_NET_SUPPORTED",
	"_NET_CLIENT_LIST",
	"_NET_NUMBER_OF_DESKTOPS",
	"_NET_DESKTOP_GEOMETRY",
	"_NET_CURRENT_DESKTOP",
	"_NET_VISIBLE_DESKTOPS",
	"_NET_DESKTOP_NAMES",
	"_NET_ACTIVE_WINDOW",
	"_NET_SUPPORTING_WM_CHECK",
	"_NET_CLOSE_WINDOW",
	"_NET_MOVERESIZE_WINDOW",
	"_NET_RESTACK_WINDOW",
	"_NET_WM_NAME",
	"_NET_WM_DESKTOP",
	"_NET_WM_WINDOW_TYPE",
	"_NET_WM_WINDOW_TYPE_DESKTOP",
	"_NET_WM_WINDOW_TYPE_DOCK",
	"_NET_WM_WINDOW_TYPE_TOOLBAR",
	"_NET_WM_WINDOW_TYPE_MENU",
	"_NET_WM_WINDOW_TYPE_UTILITY",
	"_NET_WM_WINDOW_TYPE_SPLASH",
	"_NET_WM_WINDOW_TYPE_DIALOG",
	"_NET_WM_WINDOW_TYPE_DROPDOWN_MENU",
	"_NET_WM_WINDOW_TYPE_POPUP_MENU",
	"_NET_WM_WINDOW_TYPE_TOOLTIP",
	"_NET_WM_WINDOW_TYPE_NOTIFICATION",
	"_NET_WM_WINDOW_TYPE_COMBO",
	"_NET_WM_WINDOW_TYPE_DND",
	"_NET_WM_WINDOW_TYPE_NORMAL",
	"_NET_WM_STATE",
	"_NET_WM_STATE_STICKY",
	"_NET_WM_STATE_MAXIMIZED_VERT",
	"_NET_WM_STATE_MAXIMIZED_HORZ",
	"_NET_WM_STATE_SKIP_TASKBAR",
	"_NET_WM_STATE_SKIP_PAGER",
	"_NET_WM_STATE_HIDDEN",
	"_NET_WM_STATE_FULLSCREEN",
	"_NET_WM_STATE_ABOVE",
	"_NET_WM_STATE_BELOW",
	"_NET_WM_STATE_DEMANDS_ATTENTION",
	"_NET_WM_STATE_FOCUSED",
	"_NET_WM_ALLOWED_ACTIONS",
	"_NET_WM_ACTION_MOVE",
	"_NET_WM_ACTION_RESIZE",
	"_NET_WM_ACTION_MINIMIZE",
	"_NET_WM_ACTION_STICK",
	"_NET_WM_ACTION_MAXIMIZE_HORZ",
	"_NET_WM_ACTION_MAXIMIZE_VERT",
	"_NET_WM_ACTION_FULLSCREEN",
	"_NET_WM_ACTION_CHANGE_DESKTOP",
	"_NET_WM_ACTION_CLOSE",
	"_NET_WM_ACTION_ABOVE",
	"_NET_AM_ACTION_BELOW",
	"_NET_WM_STRUT_PARTIAL",
	"_NET_WM_ICON",
	"_NET_FRAME_EXTENTS",
	"WM_TRANSIENT_FOR",
}

var windowAllowedActions = []string{
	"_NET_WM_ACTION_MOVE",
	"_NET_WM_ACTION_RESIZE",
	"_NET_WM_ACTION_MINIMIZE",
	"_NET_WM_ACTION_SHADE",
	"_NET_WM_ACTION_STICK",
	"_NET_WM_ACTION_MAXIMIZE_HORZ",
	"_NET_WM_ACTION_MAXIMIZE_VERT",
	"_NET_WM_ACTION_FULLSCREEN",
	"_NET_WM_ACTION_CHANGE_DESKTOP",
	"_NET_WM_ACTION_CLOSE",
	"_NET_WM_ACTION_ABOVE",
	"_NET_WM_ACTION_BELOW",
}
