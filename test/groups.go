package main

import (
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xwindow"
	"strconv"
	"strings"
)

func testDesktopNames() int {
	errorCnt := 0
	numDesks := 10

	_ = ewmh.NumberOfDesktopsReq(X, numDesks)
	waitForPropertyChange(X.RootWin(), "_NET_NUMBER_OF_DESKTOPS")

	before, _ := ewmh.DesktopNamesGet(X)
	names := []string{"adfg", "qrtqr", "xbnxn", "ghjgj"}
	swmctl(append([]string{"group", "names"}, names...)...)
	waitForPropertyChange(X.RootWin(), "_NET_DESKTOP_NAMES")
	after, _ := ewmh.DesktopNamesGet(X)

	assert(len(after) >= len(before), "Desktop names shouldnt shrink", &errorCnt)

	// new names has been set
	for i := range names {
		assert(names[i] == after[i], "Invalid desktop name", &errorCnt)
	}
	// names we didnt set remain the same
	for i := len(names); i < len(before); i++ {
		assert(before[i] == after[i], "Invalid desktop name", &errorCnt)
	}

	// we can set names for desktops which are not present at the moment
	_ = ewmh.NumberOfDesktopsReq(X, 1)
	waitForPropertyChange(X.RootWin(), "_NET_NUMBER_OF_DESKTOPS")
	names = []string{"a", "b", "c", "d"}
	swmctl(append([]string{"group", "names"}, names...)...)
	_ = ewmh.NumberOfDesktopsReq(X, len(names))
	waitForPropertyChange(X.RootWin(), "_NET_NUMBER_OF_DESKTOPS")
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
		waitForPropertyChange(X.RootWin(), "_NET_NUMBER_OF_DESKTOPS")
		assertEquals(i, numDesktops(), "Incorrect number of desktop", &errorCnt)
	}
	// zero is invalid, will be set to 1
	_ = ewmh.NumberOfDesktopsReq(X, 0)
	waitForPropertyChange(X.RootWin(), "_NET_NUMBER_OF_DESKTOPS")
	assertEquals(1, numDesktops(), "Incorrect number of desktop", &errorCnt)
	_ = ewmh.NumberOfDesktopsReq(X, maxDesks)
	waitForPropertyChange(X.RootWin(), "_NET_NUMBER_OF_DESKTOPS")
	assertEquals(maxDesks, numDesktops(), "Incorrect number of desktop", &errorCnt)

	// current desktop valid numbers
	for i := 0; i < maxDesks; i++ {
		_ = ewmh.CurrentDesktopReq(X, i)
		waitForPropertyChange(X.RootWin(), "_NET_CURRENT_DESKTOP")
		assertEquals(i, activeDesktop(), "Incorrect active desktop", &errorCnt)
	}

	// shrinking is ok and updates current desktop
	_ = ewmh.NumberOfDesktopsReq(X, maxDesks)
	_ = ewmh.CurrentDesktopReq(X, maxDesks-1)
	for i := maxDesks; i > 0; i-- {
		_ = ewmh.NumberOfDesktopsReq(X, i)
		waitForPropertyChange(X.RootWin(), "_NET_CURRENT_DESKTOP")
		assertEquals(i, numDesktops(), "Incorrect number of desktop", &errorCnt)
		assertEquals(i-1, activeDesktop(), "Incorrect active desktop", &errorCnt)
	}

	return errorCnt
}

func testGroupWindowCreation() int {
	errorCnt := 0

	maxDesks := 10
	var wins []*xwindow.Window

	_ = ewmh.NumberOfDesktopsReq(X, maxDesks)

	// automatic mode - inferring
	swmctl("group", "mode", "auto")
	for i := 0; i < maxDesks; i++ {
		_ = ewmh.CurrentDesktopReq(X, i)
		waitForPropertyChange(X.RootWin(), "_NET_CURRENT_DESKTOP")
		w := createWindow()
		wins = append(wins, w)
		d, _ := ewmh.WmDesktopGet(X, w.Id)
		assertEquals(i, int(d), "Incorrect desktop for window", &errorCnt)
	}

	// sticky mode
	swmctl("group", "mode", "sticky")
	for i := 0; i < maxDesks; i++ {
		_ = ewmh.CurrentDesktopReq(X, i)
		waitForPropertyChange(X.RootWin(), "_NET_CURRENT_DESKTOP")
		w := createWindow()
		wins = append(wins, w)
		d, _ := ewmh.WmDesktopGet(X, w.Id)
		assertEquals(0xFFFFFFFF, int(d), "Incorrect desktop for window", &errorCnt)
	}

	destroyWindows(wins)

	return errorCnt
}

func testGroupWindowMovement() int {
	errorCnt := 0

	maxDesks := 10
	wins := createWindows(maxDesks)
	_ = ewmh.NumberOfDesktopsReq(X, maxDesks)
	flushEvents()

	for i, win := range wins {
		_ = ewmh.ClientEvent(X, win.Id, "_NET_WM_DESKTOP", i, 2)
		waitForPropertyChange(win.Id, "_NET_WM_DESKTOP")
		d, _ := ewmh.WmDesktopGet(X, win.Id)
		assertEquals(i, int(d), "Incorrect desktop for window", &errorCnt)
	}

	// removing desktops moves windows from removed desktops to last desktop, others are left
	newDesks := maxDesks / 2
	_ = ewmh.NumberOfDesktopsReq(X, newDesks)
	waitForPropertyChange(wins[len(wins)-1].Id, "_NET_WM_DESKTOP")
	for i, win := range wins {
		d, _ := ewmh.WmDesktopGet(X, win.Id)
		expected := i
		if expected > newDesks-1 {
			expected = newDesks - 1
		}
		assertEquals(expected, int(d), "Incorrect desktop for window", &errorCnt)
	}

	destroyWindows(wins)

	return errorCnt
}

//noinspection GoNilness
func testGroupVisibility() int {
	errorCnt := 0

	maxDesks := 10
	_ = ewmh.NumberOfDesktopsReq(X, maxDesks)
	var wins []*xwindow.Window
	swmctl("group", "mode", "auto")

	// create one window on each desktop and one sticky
	for i := 0; i < maxDesks; i++ {
		_ = ewmh.CurrentDesktopReq(X, i)
		waitForPropertyChange(X.RootWin(), "_NET_CURRENT_DESKTOP")
		w := createWindow()
		wins = append(wins, w)
		d, _ := ewmh.WmDesktopGet(X, w.Id)
		assertEquals(i, int(d), "Incorrect desktop for window", &errorCnt)
	}
	sticky := createWindow()
	_ = ewmh.ClientEvent(X, sticky.Id, "_NET_WM_DESKTOP", 0xFFFFFFFF, 2)
	waitForPropertyChange(sticky.Id, "_NET_WM_DESKTOP")
	d, _ := ewmh.WmDesktopGet(X, sticky.Id)
	assertEquals(0xFFFFFFFF, int(d), "Incorrect desktop for window", &errorCnt)

	// set the only group to be the sticky one - no window should be mapped but the sticky one
	swmctl("group", "only", "-1")
	waitForPropertyChange(X.RootWin(), "_SWM_VISIBLE_GROUPS")
	assert(isWinMapped(sticky), "Window should be mapped", &errorCnt)
	assertSliceEquals([]int{}, getIntsFromSwm("group", "get-visible"), "Incorrect visible groups", &errorCnt)
	for _, win := range wins {
		assert(isWinIconified(win), "Window should be iconified", &errorCnt)
	}
	assert(isWinMapped(sticky), "Window should be mapped", &errorCnt)

	// set the only visible group and check that only windows from that group and sticky window is mapped
	for i := 0; i < maxDesks; i++ {
		swmctl("group", "only", intStr(i))
		waitForPropertyChange(X.RootWin(), "_SWM_VISIBLE_GROUPS")
		assertSliceEquals([]int{i}, getIntsFromSwm("group", "get-visible"), "Incorrect visible groups", &errorCnt)
		assert(isWinMapped(sticky), "Window should be mapped", &errorCnt)
		for j, win := range wins {
			if i == j {
				assert(isWinMapped(win), "Window should be mapped", &errorCnt)
			} else {
				assert(isWinIconified(win), "Window should be iconified", &errorCnt)
			}
		}
	}

	//multiple groups together, show, hide, toggle...
	swmctl("group", "only", "-1")
	waitForPropertyChange(X.RootWin(), "_SWM_VISIBLE_GROUPS")
	swmctl("group", "show", "1")
	waitForPropertyChange(X.RootWin(), "_SWM_VISIBLE_GROUPS")
	swmctl("group", "show", "3")
	waitForPropertyChange(X.RootWin(), "_SWM_VISIBLE_GROUPS")
	swmctl("group", "toggle", "5")
	waitForPropertyChange(X.RootWin(), "_SWM_VISIBLE_GROUPS")
	assertSliceEquals([]int{1, 3, 5}, getIntsFromSwm("group", "get-visible"), "Incorrect visible groups", &errorCnt)
	assert(isWinMapped(wins[1]), "Window should be mapped", &errorCnt)
	assert(isWinIconified(wins[2]), "Window should be iconified", &errorCnt)
	assert(isWinMapped(wins[3]), "Window should be mapped", &errorCnt)
	assert(isWinIconified(wins[4]), "Window should be iconified", &errorCnt)
	assert(isWinMapped(wins[5]), "Window should be mapped", &errorCnt)
	assert(isWinIconified(wins[6]), "Window should be iconified", &errorCnt)
	swmctl("group", "hide", "3")
	waitForPropertyChange(X.RootWin(), "_SWM_VISIBLE_GROUPS")
	swmctl("group", "show", "4")
	waitForPropertyChange(X.RootWin(), "_SWM_VISIBLE_GROUPS")
	swmctl("group", "toggle", "5")
	waitForPropertyChange(X.RootWin(), "_SWM_VISIBLE_GROUPS")
	swmctl("group", "toggle", "6")
	waitForPropertyChange(X.RootWin(), "_SWM_VISIBLE_GROUPS")
	assertSliceEquals([]int{1, 4, 6}, getIntsFromSwm("group", "get-visible"), "Incorrect visible groups", &errorCnt)
	assert(isWinMapped(wins[1]), "Window should be mapped", &errorCnt)
	assert(isWinIconified(wins[2]), "Window should be iconified", &errorCnt)
	assert(isWinIconified(wins[3]), "Window should be iconified", &errorCnt)
	assert(isWinMapped(wins[4]), "Window should be mapped", &errorCnt)
	assert(isWinIconified(wins[5]), "Window should be iconified", &errorCnt)
	assert(isWinMapped(wins[6]), "Window should be mapped", &errorCnt)

	destroyWindows(wins)
	sticky.Destroy()

	return errorCnt
}

func testGroupMembership() int {
	errorCnt := 0

	maxDesks := 10
	_ = ewmh.NumberOfDesktopsReq(X, maxDesks)
	swmctl("group", "mode", "sticky")

	win := createWindow()
	winId := intStr(int(win.Id))

	// Set group
	for i := 0; i < maxDesks; i++ {
		swmctl("group", "set", "-g", intStr(i), "-id", winId)
		waitForPropertyChange(win.Id, "_NET_WM_DESKTOP")
		assertSliceEquals([]int{i}, getIntsFromSwm("group", "get", "-id", winId), "Incorrect window groups", &errorCnt)
	}

	// Add group
	swmctl("group", "set", "-g", "0", "-id", winId)
	var groups []int
	for i := 0; i < maxDesks; i++ {
		swmctl("group", "add", "-g", intStr(i), "-id", winId)
		groups = append(groups, i)
		waitForPropertyChange(win.Id, "_NET_WM_DESKTOP")
		assertSliceEquals(groups, getIntsFromSwm("group", "get", "-id", winId), "Incorrect window groups", &errorCnt)
	}

	// Remove group
	swmctl("group", "set", "0")
	for i := 0; i < maxDesks-1; i++ {
		swmctl("group", "remove", "-g", intStr(i), "-id", winId)
		groups = groups[1:]
		waitForPropertyChange(win.Id, "_NET_WM_DESKTOP")
		assertSliceEquals(groups, getIntsFromSwm("group", "get", "-id", winId), "Incorrect window groups", &errorCnt)
	}

	// Removing the last group will add the window to sticky group
	assertSliceEquals([]int{maxDesks - 1}, getIntsFromSwm("group", "get", "-id", winId), "Incorrect window groups", &errorCnt)
	swmctl("group", "remove", "-g", intStr(maxDesks-1), "-id", winId)
	waitForPropertyChange(win.Id, "_NET_WM_DESKTOP")
	assertSliceEquals([]int{0xFFFFFFFF}, getIntsFromSwm("group", "get", "-id", winId), "Incorrect window groups", &errorCnt)

	win.Destroy()

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

func getIntsFromSwm(args ...string) []int {
	o, err := swmctlOut(args...)
	if err != nil {
		return nil
	}
	var ints []int
	for _, line := range strings.Split(o, "\n") {
		if len(line) == 0 {
			continue
		}
		i, err := strconv.Atoi(line)
		if err != nil {
			return nil
		}
		ints = append(ints, i)
	}
	return ints
}
