package windowmanager

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/janbina/swm/window"
	"log"
)

type win = window.Window

var windowCmHandlers = map[string]func(win *win, data []uint32){
	"_NET_WM_MOVERESIZE":     handleMoveResizeMessage,
	"_NET_WM_STATE":          handleWmStateMessage,
	"_NET_WM_DESKTOP":        handleWmDesktop,
	"_NET_ACTIVE_WINDOW":     handleActiveWindowMessage,
	"_NET_CLOSE_WINDOW":      handleCloseWindowMessage,
	"_NET_MOVERESIZE_WINDOW": handleMoveResizeMessage2,
	"WM_CHANGE_STATE":        handleWmChangeStateMessage,
}

var windowStateHandlers = map[string][3]func(window *win){
	//                                 {StateRemove, StateAdd, StateToggle}
	"MAXIMIZED":                       {(*win).UnMaximize, (*win).Maximize, (*win).MaximizeToggle},
	"_NET_WM_STATE_MAXIMIZED_VERT":    {(*win).UnMaximizeVert, (*win).MaximizeVert, (*win).MaximizeVertToggle},
	"_NET_WM_STATE_MAXIMIZED_HORZ":    {(*win).UnMaximizeHorz, (*win).MaximizeHorz, (*win).MaximizeHorzToggle},
	"_NET_WM_STATE_ABOVE":             {(*win).UnStackAbove, (*win).StackAbove, (*win).StackAboveToggle},
	"_NET_WM_STATE_BELOW":             {(*win).UnStackBelow, (*win).StackBelow, (*win).StackBelowToggle},
	"_NET_WM_STATE_DEMANDS_ATTENTION": {(*win).StopAttention, (*win).StartAttention, (*win).ToggleAttention},
	"_NET_WM_STATE_STICKY":            {UnstickWindow, StickWindow, ToggleWindowSticky},
	"_NET_WM_STATE_FULLSCREEN":        {(*win).UnFullscreen, (*win).Fullscreen, (*win).FullscreenToggle},
	"_NET_WM_STATE_MINIMIZE":          {(*win).DeIconify, (*win).Iconify, (*win).IconifyToggle},
}

func handleWindowClientMessage(X *xgbutil.XUtil, e xevent.ClientMessageEvent) {
	name, err := xprop.AtomName(X, e.Type)
	if err != nil {
		log.Printf("Cannot get property atom name for clientMessage event: %s", err)
		return
	}
	log.Printf("Client message %s: %s", name, e)
	if f, ok := windowCmHandlers[name]; !ok {
		log.Printf("Unsupported client message: %s", name)
	} else {
		if w, ok := managedWindows[e.Window].(*win); ok {
			f(w, e.Data.Data32)
		}
	}
}

func handleMoveResizeMessage(win *win, data []uint32) {
	xr := data[0]
	yr := data[1]
	dir := data[2]
	log.Printf("Move resize client message: %d, %d, %d", xr, yr, dir)
	if dir <= ewmh.SizeLeft {
		win.DragResizeBegin(int16(xr), int16(yr), int(dir))
	} else if dir == ewmh.Move {
		win.DragMoveBegin(int16(xr), int16(yr))
	} else {
		log.Printf("Unsupported direction: %d", dir)
	}
}

func handleWmChangeStateMessage(win *win, data []uint32) {
	if data[0] == icccm.StateIconic && !win.IsIconified() {
		win.IconifyToggle()
	}
}

func handleActiveWindowMessage(win *win, _ []uint32) {
	win.Focus()
	win.Raise()
}

func handleWmStateMessage(win *win, data []uint32) {
	action := data[0]
	p1, _ := xprop.AtomName(X, xproto.Atom(data[1]))
	p2, _ := xprop.AtomName(X, xproto.Atom(data[2]))
	log.Printf("Wm state client message: %d, %s, %s", action, p1, p2)

	updateWinStates(win, int(action), p1, p2)
}

func updateWinStates(win *win, action int, s1 string, s2 string) {
	if (s1 == "_NET_WM_STATE_MAXIMIZED_VERT" && s2 == "_NET_WM_STATE_MAXIMIZED_HORZ") ||
		(s2 == "_NET_WM_STATE_MAXIMIZED_VERT" && s1 == "_NET_WM_STATE_MAXIMIZED_HORZ") {
		updateWinState(win, action, "MAXIMIZED")
	} else {
		updateWinState(win, action, s1)
		if len(s2) > 0 {
			updateWinState(win, action, s2)
		}
	}
}

func updateWinState(win *win, action int, state string) {
	if fs, ok := windowStateHandlers[state]; !ok {
		log.Printf("Unsupported window state: %s", state)
	} else {
		fs[action](win)
	}
}

func handleWmDesktop(win *win, data []uint32) {
	MoveWindowToDesktop(win, int(data[0]))
}

func handleCloseWindowMessage(win *win, _ []uint32) {
	win.Destroy()
}

func handleMoveResizeMessage2(win *win, data []uint32) {
	//gravity := int(data[0] & 0xff)
	//flags := int((data[0] >> 8) & 0xf)
	x, y, w, h := int(data[1]), int(data[2]), int(data[3]), int(data[4])
	win.MoveResize(x, y, w, h)
}
