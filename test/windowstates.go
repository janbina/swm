package main

import (
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
)

func testWindowStates() int {
	errorCnt := 0

	win := createWindow()
	otherWin := createWindow()

	initGeom := geom(win)
	screenGeom, _ := xwindow.New(X, X.RootWin()).Geometry()
	var newGeom xrect.Rect

	// horizontal maximization
	flushEvents()
	_ = ewmh.WmStateReqExtra(X, win.Id, ewmh.StateAdd, "_NET_WM_STATE_MAXIMIZED_HORZ", "", 2)
	repeat(2, waitForConfigureNotify)
	newGeom = geom(win)
	assertGeomEquals(
		xrect.New(screenGeom.X(), initGeom.Y(), screenGeom.Width(), initGeom.Height()),
		newGeom,
		"invalid geometry",
		&errorCnt,
	)
	// upon unmaximization, restores previous geometry
	flushEvents()
	_ = ewmh.WmStateReqExtra(X, win.Id, ewmh.StateRemove, "_NET_WM_STATE_MAXIMIZED_HORZ", "", 2)
	repeat(2, waitForConfigureNotify)
	newGeom = geom(win)
	assertGeomEquals(
		initGeom,
		newGeom,
		"invalid geometry",
		&errorCnt,
	)

	// vertical maximization
	flushEvents()
	_ = ewmh.WmStateReqExtra(X, win.Id, ewmh.StateAdd, "_NET_WM_STATE_MAXIMIZED_VERT", "", 2)
	repeat(2, waitForConfigureNotify)
	newGeom = geom(win)
	assertGeomEquals(
		xrect.New(initGeom.X(), screenGeom.Y(), initGeom.Width(), screenGeom.Height()),
		newGeom,
		"invalid geometry",
		&errorCnt,
	)
	// upon unmaximization, restores previous geometry
	flushEvents()
	_ = ewmh.WmStateReqExtra(X, win.Id, ewmh.StateRemove, "_NET_WM_STATE_MAXIMIZED_VERT", "", 2)
	repeat(2, waitForConfigureNotify)
	newGeom = geom(win)
	assertGeomEquals(
		initGeom,
		newGeom,
		"invalid geometry",
		&errorCnt,
	)

	// vertical & horizontal maximization
	flushEvents()
	_ = ewmh.WmStateReqExtra(X, win.Id, ewmh.StateAdd, "_NET_WM_STATE_MAXIMIZED_VERT", "_NET_WM_STATE_MAXIMIZED_HORZ", 2)
	repeat(4, waitForConfigureNotify)
	newGeom = geom(win)
	assertGeomEquals(
		xrect.New(screenGeom.X(), screenGeom.Y(), screenGeom.Width(), screenGeom.Height()),
		newGeom,
		"invalid geometry",
		&errorCnt,
	)
	// upon unmaximization, restores previous geometry
	flushEvents()
	_ = ewmh.WmStateReqExtra(X, win.Id, ewmh.StateRemove, "_NET_WM_STATE_MAXIMIZED_VERT", "_NET_WM_STATE_MAXIMIZED_HORZ", 2)
	repeat(4, waitForConfigureNotify)
	newGeom = geom(win)
	assertGeomEquals(
		initGeom,
		newGeom,
		"invalid geometry",
		&errorCnt,
	)

	// fullscreen
	flushEvents()
	_ = ewmh.WmStateReqExtra(X, win.Id, ewmh.StateAdd, "_NET_WM_STATE_FULLSCREEN", "", 2)
	repeat(2, waitForConfigureNotify)
	decorGeom := geom(win) // testing that borders are removed
	newGeom, _ = win.Geometry()
	assertGeomEquals(
		xrect.New(screenGeom.X(), screenGeom.Y(), screenGeom.Width(), screenGeom.Height()),
		xrect.New(decorGeom.X(), decorGeom.Y(), newGeom.Width(), newGeom.Height()),
		"invalid geometry",
		&errorCnt,
	)
	// upon unfullscreen, restores previous geometry
	flushEvents()
	_ = ewmh.WmStateReqExtra(X, win.Id, ewmh.StateRemove, "_NET_WM_STATE_FULLSCREEN", "", 2)
	repeat(2, waitForConfigureNotify)
	newGeom = geom(win)
	assertGeomEquals(
		initGeom,
		newGeom,
		"invalid geometry",
		&errorCnt,
	)

	// hiding and unhiding
	flushEvents()
	_ = ewmh.WmStateReqExtra(X, win.Id, ewmh.StateAdd, "_NET_WM_STATE_HIDDEN", "", 2)
	waitForUnmapNotify()
	assert(isWinIconified(win), "Window should be iconified", &errorCnt)
	flushEvents()
	_ = ewmh.WmStateReqExtra(X, win.Id, ewmh.StateRemove, "_NET_WM_STATE_HIDDEN", "", 2)
	waitForMapNotify()
	assert(isWinMapped(win), "Window should be mapped", &errorCnt)

	// focusing
	activeId := getActiveWindow()
	active := otherWin
	inactive := win
	if activeId == win.Id {
		active = win
		inactive = otherWin
	}
	flushEvents()
	_ = ewmh.WmStateReqExtra(X, inactive.Id, ewmh.StateAdd, "_NET_WM_STATE_FOCUSED", "", 2)
	assertActive(inactive, &errorCnt)
	flushEvents()
	_ = ewmh.WmStateReqExtra(X, active.Id, ewmh.StateAdd, "_NET_WM_STATE_FOCUSED", "", 2)
	assertActive(active, &errorCnt)

	// ABOVE, BELOW INTRO - testing that activation works as expected on normal windows
	// _NET_CLIENT_LIST_STACKING has bottom-to-top stacking order
	initStacking, _ := ewmh.ClientListStackingGet(X)
	// activating bottom window should move it to top of stack
	flushEvents()
	_ = ewmh.ActiveWindowReq(X, initStacking[0])
	waitForConfigureNotify()
	newStacking, _ := ewmh.ClientListStackingGet(X)
	assertEquals(int(initStacking[0]), int(newStacking[1]), "incorrect stacking order", &errorCnt)

	// ABOVE
	// now adding ABOVE state to bottom window should automatically move it to top
	initStacking = newStacking
	aboveWin := initStacking[0]
	normalWin := initStacking[1]
	_ = ewmh.WmStateReqExtra(X, aboveWin, ewmh.StateAdd, "_NET_WM_STATE_ABOVE", "", 2)
	waitForConfigureNotify()
	newStacking, _ = ewmh.ClientListStackingGet(X)
	assertEquals(int(aboveWin), int(newStacking[1]), "incorrect stacking order", &errorCnt)
	// and activating normal window should not change this...
	_ = ewmh.ActiveWindowReq(X, normalWin)
	waitForConfigureNotify()
	newStacking, _ = ewmh.ClientListStackingGet(X)
	assertEquals(int(normalWin), int(newStacking[0]), "incorrect stacking order", &errorCnt)
	// removing above state makes the other win able to be on top again
	_ = ewmh.WmStateReqExtra(X, aboveWin, ewmh.StateRemove, "_NET_WM_STATE_ABOVE", "", 2)
	_ = ewmh.ActiveWindowReq(X, normalWin)
	waitForConfigureNotify()
	newStacking, _ = ewmh.ClientListStackingGet(X)
	assertEquals(int(normalWin), int(newStacking[1]), "incorrect stacking order", &errorCnt)

	// BELOW
	// now adding BELOW state to top window should automatically move it to bottom
	initStacking, _ = ewmh.ClientListStackingGet(X)
	belowWin := initStacking[1]
	_ = ewmh.WmStateReqExtra(X, belowWin, ewmh.StateAdd, "_NET_WM_STATE_BELOW", "", 2)
	waitForConfigureNotify()
	newStacking, _ = ewmh.ClientListStackingGet(X)
	assertEquals(int(belowWin), int(newStacking[0]), "incorrect stacking order", &errorCnt)
	// and activating that window should not change this...
	_ = ewmh.ActiveWindowReq(X, belowWin)
	waitForConfigureNotify()
	newStacking, _ = ewmh.ClientListStackingGet(X)
	assertEquals(int(belowWin), int(newStacking[0]), "incorrect stacking order", &errorCnt)
	// removing below state makes it possible to be on top again
	_ = ewmh.WmStateReqExtra(X, belowWin, ewmh.StateRemove, "_NET_WM_STATE_BELOW", "", 2)
	_ = ewmh.ActiveWindowReq(X, belowWin)
	waitForConfigureNotify()
	newStacking, _ = ewmh.ClientListStackingGet(X)
	assertEquals(int(belowWin), int(newStacking[1]), "incorrect stacking order", &errorCnt)

	win.Destroy()
	otherWin.Destroy()

	return errorCnt
}
