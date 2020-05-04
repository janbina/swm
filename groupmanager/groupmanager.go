package groupmanager

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"sort"
)

type Changes struct {
	Invisible []xproto.Window
	Visible   []xproto.Window
	Raise     []xproto.Window
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
	winToGroups  map[xproto.Window]map[int]bool
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
	winToGroups = map[xproto.Window]map[int]bool{}
	currentGroup = stickyGroupID
	GroupMode = ModeAuto
	setDesktops()
	setCurrentDesktop()
	setVisibleGroups()
}

func AddWindow(win xproto.Window) {
	g := getInitialGroupForWindow(win)
	winToGroups[win] = map[int]bool{g: true}
	getGroup(g).windows[win] = true
	setWinDesktop(win)
}

func RemoveWindow(win xproto.Window) {
	for g := range winToGroups[win] {
		delete(getGroup(g).windows, win)
	}
	delete(winToGroups, win)
}

func GetNumGroups() int {
	return len(groups)
}

func GetCurrentGroup() int {
	return currentGroup
}

func IsGroupVisible(group int) bool {
	if group == stickyGroupID {
		return true
	}
	return group >= 0 && group < len(groups) && getGroup(group).isVisible()
}

func IsWinGroupVisible(win xproto.Window) bool {
	for g := range winToGroups[win] {
		if IsGroupVisible(g) {
			return true
		}
	}
	return false
}

func GetWinGroups(win xproto.Window) []uint {
	groups := make([]uint, 0, len(winToGroups[win]))
	for g := range winToGroups[win] {
		groups = append(groups, uint(g))
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i] < groups[j]
	})
	return groups
}

func GetWinGroupNames(win xproto.Window) []string {
	groups := GetWinGroups(win)
	names := make([]string, len(groups))
	for i, g := range groups {
		if g == stickyGroupID {
			names[i] = "S"
		} else {
			names[i] = getGroup(int(g)).name
			if len(names[i]) == 0 {
				names[i] = fmt.Sprintf("%d", g)
			}
		}
	}
	return names
}

func IsWinInGroup(win xproto.Window, group int) bool {
	return winToGroups[win][group]
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
		setVisibleGroups()
		if currentGroup >= newLast {
			return ShowGroup(newLast)
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
	if IsGroupVisible(group) {
		return HideGroup(group)
	} else {
		return ShowGroup(group)
	}
}

func ShowGroupOnly(group int) *Changes {
	if group < 0 {
		group = stickyGroupID
	}

	ensureEnoughGroups(group)

	for i, g := range groups {
		if i != group {
			g.makeInvisible()
		}
	}
	getGroup(group).makeVisible()

	updateCurrentGroup()
	setVisibleGroups()

	return createChangesWithRaise(group)
}

func ShowGroup(group int) *Changes {
	if group < 0 || group == stickyGroupID {
		return nil
	}
	ensureEnoughGroups(group)
	getGroup(group).makeVisible()

	updateCurrentGroup()
	setVisibleGroups()

	return createChangesWithRaise(group)

}

func HideGroup(group int) *Changes {
	if group < 0 || group == stickyGroupID {
		return nil
	}
	ensureEnoughGroups(group)
	getGroup(group).makeInvisible()

	updateCurrentGroup()
	setVisibleGroups()

	return createChanges()
}

func SetGroupForWindow(win xproto.Window, group int) *Changes {
	if group < 0 {
		group = stickyGroupID
	}

	ensureEnoughGroups(group)
	for g := range winToGroups[win] {
		delete(getGroup(g).windows, win)
	}
	getGroup(group).windows[win] = true

	winToGroups[win] = map[int]bool{group: true}
	setWinDesktop(win)

	return createChanges()
}

func AddWindowToGroup(win xproto.Window, group int) *Changes {
	if group < 0 {
		group = stickyGroupID
	}

	ensureEnoughGroups(group)
	getGroup(group).windows[win] = true
	winToGroups[win][group] = true
	setWinDesktop(win)

	return createChanges()
}

func RemoveWindowFromGroup(win xproto.Window, group int) *Changes {
	if group < 0 {
		group = stickyGroupID
	}
	if group != stickyGroupID && group >= len(groups) {
		return nil
	}

	delete(getGroup(group).windows, win)
	delete(winToGroups[win], group)

	if len(winToGroups[win]) == 0 {
		// window has to be in some group...
		g := getInitialGroupForWindow(win)
		winToGroups[win] = map[int]bool{g: true}
		getGroup(g).windows[win] = true
	}

	setWinDesktop(win)

	return createChanges()
}

func GetVisibleGroups() []uint {
	ids := make([]uint, 0)
	for i, group := range groups {
		if group.isVisible() {
			ids = append(ids, uint(i))
		}
	}
	return ids
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
	// only if from is windows only group
	for w := range getGroup(from).windows {
		if len(winToGroups[w]) == 1 {
			getGroup(to).windows[w] = true
			winToGroups[w] = map[int]bool{to: true}
			setWinDesktop(w)
		}
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

func createChanges() *Changes {
	return createChangesWithRaise(-1)
}

func createChangesWithRaise(raiseGroup int) *Changes {
	invisible := make([]xproto.Window, 0)
	visible := make([]xproto.Window, 0)
	raise := make([]xproto.Window, 0)

	for win := range winToGroups {
		if IsWinGroupVisible(win) {
			visible = append(visible, win)
		} else {
			invisible = append(invisible, win)
		}
		if winToGroups[win][raiseGroup] {
			raise = append(raise, win)
		}
	}

	return &Changes{Invisible: invisible, Visible: visible, Raise: raise}
}
