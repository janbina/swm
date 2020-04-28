package main

import (
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xwindow"
)

func testDesktopNames() int {
	errorCnt := 0
	numDesks := 10

	_ = ewmh.NumberOfDesktopsReq(X, numDesks)
	sleepMillis(10)

	before, _ := ewmh.DesktopNamesGet(X)
	names := []string{"adfg", "qrtqr", "xbnxn", "ghjgj"}
	swmctl(append([]string{"set-desktop-names"}, names...)...)
	sleepMillis(10)
	after, _ := ewmh.DesktopNamesGet(X)

	// new names has been set
	for i := range names {
		assert(names[i] == after[i], "Invalid desktop name", &errorCnt)
	}
	// names we didnt set remain the same
	for i := len(names); i < numDesks; i++ {
		assert(before[i] == after[i], "Invalid desktop name", &errorCnt)
	}

	// we can set names for desktops which are not present at the moment
	_ = ewmh.NumberOfDesktopsReq(X, 1)
	sleepMillis(10)
	names = []string{"a", "b", "c", "d"}
	swmctl(append([]string{"set-desktop-names"}, names...)...)
	sleepMillis(10)
	_ = ewmh.NumberOfDesktopsReq(X, len(names))
	sleepMillis(10)
	after, _ = ewmh.DesktopNamesGet(X)
	for i := range names {
		assert(names[i] == after[i], "Invalid desktop name", &errorCnt)
	}

	return errorCnt
}

func testGroupBasics() int {
	errorCnt := 0

	maxDesks := 10

	// valid number of desktops
	for i := 1; i <= maxDesks; i++ {
		_ = ewmh.NumberOfDesktopsReq(X, i)
		sleepMillis(10)
		assertEquals(i, numDesktops(), "Incorrect number of desktop", &errorCnt)
	}
	// zero is invalid, will be set to 1
	_ = ewmh.NumberOfDesktopsReq(X, 0)
	sleepMillis(10)
	assertEquals(1, numDesktops(), "Incorrect number of desktop", &errorCnt)
	_ = ewmh.NumberOfDesktopsReq(X, maxDesks)
	sleepMillis(10)
	assertEquals(maxDesks, numDesktops(), "Incorrect number of desktop", &errorCnt)

	// current desktop valid numbers
	for i := 0; i < maxDesks; i++ {
		_ = ewmh.CurrentDesktopReq(X, i)
		sleepMillis(10)
		assertEquals(i, activeDesktop(), "Incorrect active desktop", &errorCnt)
	}

	// shrinking is ok and updates current desktop
	_ = ewmh.NumberOfDesktopsReq(X, maxDesks)
	_ = ewmh.CurrentDesktopReq(X, maxDesks - 1)
	sleepMillis(10)
	for i := maxDesks; i > 0; i-- {
		_ = ewmh.NumberOfDesktopsReq(X, i)
		sleepMillis(10)
		assertEquals(i, numDesktops(), "Incorrect number of desktop", &errorCnt)
		assertEquals(i - 1, activeDesktop(), "Incorrect active desktop", &errorCnt)
	}

	return errorCnt
}

func testGroupWindowCreation() int {
	errorCnt := 0

	maxDesks := 10
	var wins []*xwindow.Window

	_ = ewmh.NumberOfDesktopsReq(X, maxDesks)

	for i := 0; i < maxDesks; i++ {
		_ = ewmh.CurrentDesktopReq(X, i)
		sleepMillis(10)
		w := createWindow()
		wins = append(wins, w)
		sleepMillis(10)
		d, _ := ewmh.WmDesktopGet(X, w.Id)
		assertEquals(i, int(d), "Incorrect desktop for window", &errorCnt)
	}

	destroyWindows(wins)

	return errorCnt
}

func testGroupWindowMovement() int {
	errorCnt := 0

	maxDesks := 10
	wins := createWindows(maxDesks)
	_ = ewmh.NumberOfDesktopsReq(X, maxDesks)

	for i, win := range wins {
		_ = ewmh.ClientEvent(X, win.Id, "_NET_WM_DESKTOP", i, 2)
		sleepMillis(10)
		d, _ := ewmh.WmDesktopGet(X, win.Id)
		assertEquals(i, int(d), "Incorrect desktop for window", &errorCnt)
	}

	// removing desktops moves windows from removed desktops to last desktop, others are left
	newDesks := maxDesks / 2
	_ = ewmh.NumberOfDesktopsReq(X, newDesks)
	sleepMillis(10)
	for i, win := range wins {
		d, _ := ewmh.WmDesktopGet(X, win.Id)
		expected := i
		if expected > newDesks - 1  {
			expected = newDesks - 1
		}
		assertEquals(expected, int(d), "Incorrect desktop for window", &errorCnt)
	}

	destroyWindows(wins)

	return errorCnt
}

func activeDesktop() int {
	d, _ := ewmh.CurrentDesktopGet(X)
	return int(d)
}

func numDesktops() int {
	n, _ := ewmh.NumberOfDesktopsGet(X)
	return int(n)
}
