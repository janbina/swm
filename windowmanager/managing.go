package windowmanager

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/janbina/swm/groupmanager"
	"github.com/janbina/swm/focus"
	"github.com/janbina/swm/window"
	"log"
)

func manageWindow(w xproto.Window) {
	X.Grab()
	defer X.Ungrab()

	if _, ok := managedWindows[w]; ok {
		return
	}

	attrs, err := xproto.GetWindowAttributes(X.Conn(), w).Reply()
	if err == nil && attrs.OverrideRedirect {
		log.Printf("Ignoring window with override redirect")
		return
	}

	win := window.New(X, w)

	if win == nil {
		log.Printf("Cannot manage window id %d", w)
		return
	}

	managedWindows[w] = win
	groupmanager.AddWindow(w)

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

	setWmAllowedActions(w)

	if !win.IsIconified() && !win.IsHidden() && groupmanager.IsWinDesktopVisible(w) {
		win.Map()
		win.Focus()
		win.Raise()
	} else {
		win.Unmap()
	}
}

func unmanageWindow(w xproto.Window) {
	X.Grab()
	defer X.Ungrab()
	win := managedWindows[w]
	if win == nil {
		return
	}
	win.Destroyed()
	groupmanager.RemoveWindow(w)
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
	win.SetupMouseEvents()

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

	xevent.ConfigureRequestFun(func(x *xgbutil.XUtil, e xevent.ConfigureRequestEvent) {
		win.ConfigureRequest(e)
	}).Connect(X, w)
}
