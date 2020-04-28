package groupmanager

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
)

type Changes struct {
	Invisible []xproto.Window
	Visible   []xproto.Window
}

type Mode int

const (
	// Id of group which is always visible
	// Taken from ewmh desktop specification: "0xFFFFFFFF indicates that the window should appear on all groups"
	alwaysVisibleGroup = 0xFFFFFFFF

	// Mode for initial window group:
	// * sticky - all windows are initially in group id 0xFFFFFFFF, which is always visible
	// * auto - window is in group which we get from _NET_WM_DESKTOP, or currentGroup
	ModeSticky Mode = iota
	ModeAuto
)

var (
	X *xgbutil.XUtil

	// names of all groups
	// used also to get the number of groups
	groups        []string
	groupToWins   map[int]map[xproto.Window]bool
	winToGroup    map[xproto.Window]int
	visibleGroups map[int]bool
	// group which was made visible *last*
	currentGroup  int
	GroupMode     Mode
)

func Initialize(x *xgbutil.XUtil) {
	X = x

	groupToWins = make(map[int]map[xproto.Window]bool)
	groupToWins[alwaysVisibleGroup] = make(map[xproto.Window]bool)
	winToGroup = make(map[xproto.Window]int)
	groups = getDesktops()
	currentGroup = alwaysVisibleGroup
	visibleGroups = make(map[int]bool)
	GroupMode = ModeAuto
	setDesktops()
	setCurrentDesktop()
}

func AddWindow(win xproto.Window) {
	g := getInitialGroupForWindow(win)
	winToGroup[win] = g
	ensureGroup(g)
	groupToWins[g][win] = true
	_ = ewmh.WmDesktopSet(X, win, uint(g))
}

func RemoveWindow(win xproto.Window) {
	d := winToGroup[win]
	delete(winToGroup, win)
	delete(groupToWins[d], win)
}

func GetNumGroups() int {
	return len(groups)
}

func IsGroupVisible(group int) bool {
	return group == alwaysVisibleGroup || visibleGroups[group]
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
			groups[i] = name
		}
	}
	if len(names) > len(groups) {
		setDesktopNames(names)
	} else {
		setDesktopNames(groups)
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
		groups = append(groups, getDesktopNames(currentNum, newLast)...)
		setDesktops()
	}

	return nil
}

func ToggleGroupVisibility(group int) *Changes {
	if group == alwaysVisibleGroup {
		return nil
	}
	wasVisible := visibleGroups[group]
	visibleGroups[group] = !wasVisible

	wins := winsOfGroup(group)

	updateCurrentGroup(group)

	if wasVisible {
		return createChanges(wins, nil)
	} else {
		return createChanges(nil, wins)
	}
}

func ShowGroupOnly(group int) *Changes {
	invisible := make([]xproto.Window, 0)
	var visible []xproto.Window

	for g, v := range visibleGroups {
		if v && g != group {
			visibleGroups[g] = false
			invisible = append(invisible, winsOfGroup(g)...)
		}
	}

	if !visibleGroups[group] {
		visibleGroups[group] = true
		visible = winsOfGroup(group)
	}

	updateCurrentGroup(group)

	return createChanges(invisible, visible)
}

func showGroupForce(group int, force bool) *Changes {
	if !IsGroupVisible(group) {
		return ToggleGroupVisibility(group)
	}
	if force {
		wins := winsOfGroup(group)

		updateCurrentGroup(group)

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
	delete(groupToWins[prev], win)
	ensureGroup(group)
	groupToWins[group][win] = true
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
		return alwaysVisibleGroup
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
	ensureGroup(to)
	for w := range groupToWins[from] {
		groupToWins[to][w] = true
		winToGroup[w] = to
		_ = ewmh.WmDesktopSet(X, w, uint(to))
	}
	delete(groupToWins, from)
}

func ensureGroup(d int) {
	if groupToWins[d] == nil {
		groupToWins[d] = make(map[xproto.Window]bool)
	}
}

func ensureEnoughGroups(group int) {
	if group == alwaysVisibleGroup || group < len(groups) {
		return
	}
	// we can safely ignore changes, cause we are adding new groups, so there are none
	_ = SetNumberOfGroups(group + 1)
}

func winsOfGroup(d int) []xproto.Window {
	ret := make([]xproto.Window, 0, len(groupToWins[d]))
	for w := range groupToWins[d] {
		ret = append(ret, w)
	}
	return ret
}

func updateCurrentGroup(group int) {
	if IsGroupVisible(group) {
		currentGroup = group
		setCurrentDesktop()
	}
}

func createChanges(invisible, visible []xproto.Window) *Changes {
	return &Changes{Invisible: invisible, Visible: visible}
}
