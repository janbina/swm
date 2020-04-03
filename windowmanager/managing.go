package windowmanager

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/janbina/swm/focus"
	"github.com/janbina/swm/window"
	"log"
)

func manageWindow(w xproto.Window) {
	win := window.New(X, w)
	managedWindows[w] = win
	d := getDesktopForWindow(w)
	desktopToWins[d] = append(desktopToWins[d], w)
	_ = ewmh.WmDesktopSet(X, w, uint(d))

	xproto.ChangeSaveSet(X.Conn(), xproto.SetModeInsert, w)

	if s, _ := ewmh.WmStrutPartialGet(X, w); s != nil {
		strutWindows[w] = true
		applyStruts()
	}

	updateClientList()
	setupListeners(w, win)

	for _, s := range win.GetActiveStates() {
		updateWinState(win, ewmh.StateAdd, s)
	}

	if !win.IsHidden() && (d == currentDesktop || d == stickyDesktop) {
		win.Map()
		win.Focus()
		win.Raise()
	} else {
		win.Unmap()
	}
}

func unmanageWindow(w xproto.Window) {
	win := managedWindows[w]
	if win == nil {
		return
	}
	win.Destroyed()
	xproto.ChangeSaveSet(X.Conn(), xproto.SetModeDelete, w)
	focus.FocusLast()
	delete(managedWindows, w)
	updateClientList()
	if strutWindows[w] {
		delete(strutWindows, w)
		applyStruts()
	}
}

func setupListeners(w xproto.Window, win *window.Window) {
	win.SetupMouseEvents(moveDragShortcut, resizeDragShortcut)

	_ = win.Listen(
		xproto.EventMaskStructureNotify,
		xproto.EventMaskEnterWindow,
		xproto.EventMaskFocusChange,
	)

	win.SetupFocusListeners()

	xevent.ClientMessageFun(handleWindowClientMessage).Connect(X, w)

	xevent.DestroyNotifyFun(func(x *xgbutil.XUtil, e xevent.DestroyNotifyEvent) {
		log.Printf("Destroy notify: %s", e)
		unmanageWindow(e.Window)
	}).Connect(X, w)

	xevent.PropertyNotifyFun(func(x *xgbutil.XUtil, e xevent.PropertyNotifyEvent) {
		win.HandlePropertyNotify(e)
	}).Connect(X, w)
}

func getDesktopForWindow(win xproto.Window) int {
	_d, err := ewmh.WmDesktopGet(X, win)
	d := int(_d)
	if err != nil {
		// not specified
		return currentDesktop
	}
	if d == stickyDesktop || d < len(desktops) {
		return d
	}
	// TODO: Current, last, create additional desktops, or what?
	return len(desktops) - 1
}
