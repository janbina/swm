package groupmanager

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xprop"
)

const SWM_VISIBLE_GROUPS_ATOM = "_SWM_VISIBLE_GROUPS"

func getGroupNames(groups []*group) []string {
	names := make([]string, len(groups))
	for i, group := range groups {
		names[i] = group.name
	}
	return names
}

func defaultDesktopName(pos int) string {
	return fmt.Sprintf("G.%d", pos+1)
}

func getDesktopNames(from, to int) []string {
	if from > to {
		return nil
	}
	names := make([]string, to-from+1)
	fromEwmh, _ := ewmh.DesktopNamesGet(X)
	for i := range names {
		i2 := i + from
		if i2 < len(fromEwmh) {
			names[i] = fromEwmh[i2]
		} else {
			names[i] = defaultDesktopName(i2)
		}
	}
	return names
}

func getDesktops() []string {
	num, _ := ewmh.NumberOfDesktopsGet(X)
	return getDesktopNames(0, int(num)-1)
}

func setDesktops() {
	_ = ewmh.NumberOfDesktopsSet(X, uint(len(groups)))
	fromEwmh, _ := ewmh.DesktopNamesGet(X)
	if len(fromEwmh) < len(groups) {
		// dont set names when shrinking
		setDesktopNames(getGroupNames(groups))
	}
}

func setDesktopNames(names []string) {
	_ = ewmh.DesktopNamesSet(X, names)
}

func setCurrentDesktop() {
	_ = ewmh.CurrentDesktopSet(X, uint(currentGroup))
}

func setVisibleGroups() {
	_ = xprop.ChangeProp32(X, X.RootWin(), SWM_VISIBLE_GROUPS_ATOM, "CARDINAL", GetVisibleGroups()...)
}

func setWinDesktop(win xproto.Window) {
	d := winToGroup[win]
	_ = xprop.ChangeProp32(X, win, "_NET_WM_DESKTOP", "CARDINAL", uint(d))
}