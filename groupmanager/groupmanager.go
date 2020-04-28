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

const (
	// Id of group which is always visible
	// Taken from ewmh desktop specification: "0xFFFFFFFF indicates that the window should appear on all groups"
	alwaysVisibleGroup = 0xFFFFFFFF
)

var (
	X *xgbutil.XUtil

	groups        []string
	groupToWins   map[int]map[xproto.Window]bool
	winToGroup    map[xproto.Window]int
	currentGroup  int
	visibleGroups map[int]bool
)

func Initialize(x *xgbutil.XUtil) {
	X = x

	groupToWins = make(map[int]map[xproto.Window]bool)
	groupToWins[alwaysVisibleGroup] = make(map[xproto.Window]bool)
	winToGroup = make(map[xproto.Window]int)
	groups = getDesktops()
	currentGroup = alwaysVisibleGroup
	visibleGroups = make(map[int]bool)
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

func SetNumberOfGroups(num int) *Changes {
	//if num < 1 {
	//	num = 1
	//}
	//currentNum := len(groups)
	//newLast := num - 1
	//
	//if num < currentNum {
	//	for i := num; i < currentNum; i++ {
	//		moveWinsToGroup(i, newLast)
	//	}
	//	groups = groups[:num]
	//	setDesktops()
	//	if currentGroup > newLast {
	//		return SwitchToDesktop(newLast)
	//	} else if currentGroup == newLast {
	//		return createChanges(nil, winsOfGroup(currentGroup))
	//	}
	//} else if num > currentNum {
	//	groups = append(groups, getDesktopNames(currentNum, newLast)...)
	//	setDesktops()
	//	return createChanges(nil, nil)
	//}

	currentNum := len(groups)
	newLast := num - 1
	if num > currentNum {
		groups = append(groups, getDesktopNames(currentNum, newLast)...)
		setDesktops()
		return createChanges(nil, nil)
	}

	return createChanges(nil, nil)
}

func ToggleGroupVisibility(group int) *Changes {
	if group == alwaysVisibleGroup {
		return createChanges(nil, nil)
	}
	wasVisible := visibleGroups[group]
	visibleGroups[group] = !wasVisible

	wins := make([]xproto.Window, 0, len(groupToWins[group]))
	for w := range groupToWins[group] {
		wins = append(wins, w)
	}

	currentGroup = group
	setCurrentDesktop()

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

	currentGroup = group
	setCurrentDesktop()

	return createChanges(invisible, visible)
}

func SetGroupForWindow(win xproto.Window, group int) *Changes {
	prev := winToGroup[win]
	if prev == group {
		return createChanges(nil, nil)
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
	return createChanges(nil, nil)
}

func getInitialGroupForWindow(win xproto.Window) int {
	_g, err := ewmh.WmDesktopGet(X, win)
	g := int(_g)
	if err != nil {
		// not specified
		return currentGroup
	}
	if g == alwaysVisibleGroup || g < len(groups) {
		return g
	}
	// TODO: Current, last, create additional groups, or what?
	return len(groups) - 1
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

func createChanges(invisible, visible []xproto.Window) *Changes {
	return &Changes{Invisible: invisible, Visible: visible}
}
