package windowmanager

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/mousebind"
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
	win.Map()
	win.Focus()
	win.Raise()
	setupListeners(w, win)
}

func unmanageWindow(w xproto.Window) {
	win := managedWindows[w]
	if win == nil {
		return
	}
	mousebind.Detach(X, w)
	xevent.Detach(X, w)
	xproto.ChangeSaveSet(X.Conn(), xproto.SetModeDelete, w)
	win.Destroyed()
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

	xevent.ClientMessageFun(func(x *xgbutil.XUtil, e xevent.ClientMessageEvent) {
		win.HandleClientMessage(e)
	}).Connect(X, w)

	xevent.DestroyNotifyFun(func(x *xgbutil.XUtil, e xevent.DestroyNotifyEvent) {
		log.Printf("Destroy notify: %s", e)
		unmanageWindow(e.Window)
	}).Connect(X, w)

	win.HandleFocusIn().Connect(X, w)
	win.HandleFocusOut().Connect(X, w)
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
