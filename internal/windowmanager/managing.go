package windowmanager

import (
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/davecgh/go-spew/spew"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/janbina/swm/internal/focus"
	"github.com/janbina/swm/internal/groupmanager"
	"github.com/janbina/swm/internal/log"
	"github.com/janbina/swm/internal/window"
)

func manageWindow(w xproto.Window) {
	X.Grab()
	defer X.Ungrab()

	if _, ok := managedWindows[w]; ok {
		return
	}

	xWin := xwindow.New(X, w)

	winInfo := window.GetWindowInfo(xWin)
	winActions := window.GetWinActions(X, winInfo)

	log.Info("Win INFO: %s", spew.Sdump(winInfo))
	log.Info("Win Actions: %s", spew.Sdump(winActions))

	if !winActions.ShouldManage {
		return
	}

	win := window.New(xWin, winInfo, winActions)

	if win == nil {
		log.Info("Cannot manage window id %d", w)
		return
	}

	if winActions.IsFocusable {
		focus.InitialAdd(win)
	}

	managedWindows[w] = win
	groupmanager.AddWindow(w, winActions.Groups)

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

	if !win.IsIconified() && !win.IsHidden() && groupmanager.IsWinGroupVisible(w) {
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
		log.Info("Destroy notify: %s", e)
		unmanageWindow(e.Window)
	}).Connect(X, w)

	xevent.UnmapNotifyFun(func(x *xgbutil.XUtil, e xevent.UnmapNotifyEvent) {
		if win.UnmapNotifyShouldUnmanage() {
			unmanageWindow(e.Window)
		} else {
			win.UnmapNotify()
		}
	}).Connect(X, w)

	xevent.PropertyNotifyFun(func(x *xgbutil.XUtil, e xevent.PropertyNotifyEvent) {
		win.HandlePropertyNotify(e)
	}).Connect(X, w)

	xevent.ConfigureRequestFun(func(x *xgbutil.XUtil, e xevent.ConfigureRequestEvent) {
		win.ConfigureRequest(e)
	}).Connect(X, w)
}
