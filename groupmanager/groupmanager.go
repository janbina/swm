package groupmanager

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"time"
)

type Changes struct {
	Invisible []xproto.Window
	Visible   []xproto.Window
}

type Mode int

const (
	// Id of group which is always visible
	// Taken from ewmh desktop specification: "0xFFFFFFFF indicates that the window should appear on all groups"
	stickyGroupID = 0xFFFFFFFF

	// Mode for initial window group:
	// * sticky - all windows are initially in group id 0xFFFFFFFF, which is always visible
	// * auto - window is in group which we get from _NET_WM_DESKTOP, or currentGroup
	ModeSticky Mode = iota
	ModeAuto
)

var (
	X *xgbutil.XUtil

	groups       []*group
	stickyGroup  *group
	winToGroup   map[xproto.Window]int
	currentGroup int // group which was made visible *last*
	GroupMode    Mode
)

func Initialize(x *xgbutil.XUtil) {
	X = x

	desktops := getDesktops()
	groups := make([]*group, len(desktops))
	for i, name := range getDesktops() {
		groups[i] = createGroup(name)
	}

	stickyGroup = createGroup("sticky")
	winToGroup = map[xproto.Window]int{}
	currentGroup = stickyGroupID
	GroupMode = ModeAuto
	setDesktops()
	setCurrentDesktop()
}

func AddWindow(win xproto.Window) {
	g := getInitialGroupForWindow(win)
	winToGroup[win] = g
	getGroup(g).windows[win] = true
	_ = ewmh.WmDesktopSet(X, win, uint(g))
}

func RemoveWindow(win xproto.Window) {
	g := winToGroup[win]
	delete(winToGroup, win)
	delete(getGroup(g).windows, win)
}

func GetNumGroups() int {
	return len(groups)
}

func IsGroupVisible(group int) bool {
	if group == stickyGroupID {
		return true
	}
	return group >= 0 && group < len(groups) && getGroup(group).isVisible()
}

func IsWinGroupVisible(win xproto.Window) bool {
	return IsGroupVisible(winToGroup[win])
}

func GetWinGroup(win xproto.Window) int {
	return winToGroup[win]
}

func SetGroupNames(names []string) {
	for i, name := range names {
		if i < len(groups) {
			getGroup(i).name = name
		}
	}
	if len(names) > len(groups) {
		setDesktopNames(names)
	} else {
		setDesktopNames(getGroupNames(groups))
	}
}

// SetNumberOfGroups
// 1) If we are increasing number of groups, we just update internals and ewmh properties
// 2) If we are increasing number of groups, windows from removed groups are moved to group with highest index.
//    If current group is out of bounds after decrease, we show the group with highest index
func SetNumberOfGroups(num int) *Changes {
	if num < 1 {
		num = 1
	}
	currentNum := len(groups)
	newLast := num - 1

	if num < currentNum {
		for i := num; i < currentNum; i++ {
			moveWinsToGroup(i, newLast)
		}
		groups = groups[:num]
		setDesktops()
		if currentGroup >= newLast {
			return showGroupForce(newLast, true)
		}
	} else if num > currentNum {
		names := getDesktopNames(currentNum, newLast)
		for _, name := range names {
			groups = append(groups, createGroup(name))
		}
		setDesktops()
	}

	return nil
}

func ToggleGroupVisibility(group int) *Changes {
	if group == stickyGroupID {
		return nil
	}
	ensureEnoughGroups(group)
	wasVisible := IsGroupVisible(group)
	if wasVisible {
		getGroup(group).shownTimestamp = 0
	} else {
		getGroup(group).shownTimestamp = time.Now().UnixNano()
	}

	wins := winsOfGroup(group)

	updateCurrentGroup()

	if wasVisible {
		return createChanges(wins, nil)
	} else {
		return createChanges(nil, wins)
	}
}

func ShowGroupOnly(group int) *Changes {
	invisible := make([]xproto.Window, 0)
	var visible []xproto.Window

	ensureEnoughGroups(group)

	for i, g := range groups {
		if i != group && g.isVisible() {
			g.shownTimestamp = 0
			invisible = append(invisible, winsOfGroup(i)...)
		}
	}

	if !getGroup(group).isVisible() {
		getGroup(group).shownTimestamp = time.Now().UnixNano()
		visible = winsOfGroup(group)
	}

	updateCurrentGroup()

	return createChanges(invisible, visible)
}

func showGroupForce(group int, force bool) *Changes {
	if !IsGroupVisible(group) {
		return ToggleGroupVisibility(group)
	}
	if force {
		wins := winsOfGroup(group)

		ensureEnoughGroups(group)

		getGroup(group).shownTimestamp = time.Now().UnixNano()
		updateCurrentGroup()

		return createChanges(nil, wins)
	}
	return nil
}

func ShowGroup(group int) *Changes {
	return showGroupForce(group, false)
}

func HideGroup(group int) *Changes {
	if IsGroupVisible(group) {
		return ToggleGroupVisibility(group)
	}
	return nil
}

func SetGroupForWindow(win xproto.Window, group int) *Changes {
	prev := winToGroup[win]
	if prev == group {
		return nil
	}
	ensureEnoughGroups(group)
	delete(getGroup(prev).windows, win)
	getGroup(group).windows[win] = true
	winToGroup[win] = group
	_ = ewmh.WmDesktopSet(X, win, uint(group))

	if IsGroupVisible(prev) && !IsGroupVisible(group) {
		return createChanges([]xproto.Window{win}, nil)
	} else if !IsGroupVisible(prev) && IsGroupVisible(group) {
		return createChanges(nil, []xproto.Window{win})
	}
	return nil
}

func getInitialGroupForWindow(win xproto.Window) int {
	if GroupMode == ModeSticky {
		return stickyGroupID
	}
	g, err := ewmh.WmDesktopGet(X, win)
	if err != nil {
		// not specified
		return currentGroup
	}
	ensureEnoughGroups(int(g))
	return int(g)
}

func moveWinsToGroup(from, to int) {
	for w := range getGroup(from).windows {
		getGroup(to).windows[w] = true
		winToGroup[w] = to
		_ = ewmh.WmDesktopSet(X, w, uint(to))
	}
	getGroup(from).windows = map[xproto.Window]bool{}
}

func ensureEnoughGroups(group int) {
	if group == stickyGroupID || group < len(groups) {
		return
	}
	// we can safely ignore changes, cause we are adding new groups, so there are none
	_ = SetNumberOfGroups(group + 1)
}

func winsOfGroup(g int) []xproto.Window {
	ret := make([]xproto.Window, 0, len(getGroup(g).windows))
	for w := range getGroup(g).windows {
		ret = append(ret, w)
	}
	return ret
}

func updateCurrentGroup() {
	group := stickyGroupID
	max := int64(0)
	for i, g := range groups {
		if g.shownTimestamp > max {
			max = g.shownTimestamp
			group = i
		}
	}
	currentGroup = group
	setCurrentDesktop()
}

func getGroup(id int) *group {
	if id == stickyGroupID {
		return stickyGroup
	}
	return groups[id]
}

func createChanges(invisible, visible []xproto.Window) *Changes {
	return &Changes{Invisible: invisible, Visible: visible}
}
