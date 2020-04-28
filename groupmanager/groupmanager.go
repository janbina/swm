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
	minDesktops = 1 //minimum number of desktops created at startup
	allDesktops = 0xFFFFFFFF
)

var (
	X *xgbutil.XUtil

	desktops       []string
	desktopToWins  map[int]map[xproto.Window]bool
	winToDesktop   map[xproto.Window]int
	currentDesktop int
)

func Initialize(x *xgbutil.XUtil) {
	X = x

	desktopToWins = make(map[int]map[xproto.Window]bool)
	desktopToWins[allDesktops] = make(map[xproto.Window]bool)
	winToDesktop = make(map[xproto.Window]int)
	desktops = getDesktops()
	currentDesktop = getCurrentDesktopEwmh()
}

func AddWindow(win xproto.Window) {
	d := getInitialDesktopForWindow(win)
	winToDesktop[win] = d
	ensureDesktop(d)
	desktopToWins[d][win] = true
	_ = ewmh.WmDesktopSet(X, win, uint(d))
}

func RemoveWindow(win xproto.Window) {
	d := winToDesktop[win]
	delete(winToDesktop, win)
	delete(desktopToWins[d], win)
}

func GetNumDesktops() int {
	return len(desktops)
}

func IsDesktopVisible(desktop int) bool {
	return desktop == allDesktops || desktop == currentDesktop
}

func IsWinDesktopVisible(win xproto.Window) bool {
	return IsDesktopVisible(winToDesktop[win])
}

func GetCurrentDesktop() int {
	return currentDesktop
}

func GetWinDesktop(win xproto.Window) int {
	return winToDesktop[win]
}

func SetDesktopNames(names []string) {
	for i, name := range names {
		if i < len(desktops) {
			desktops[i] = name
		}
	}
	if len(names) > len(desktops) {
		setDesktopNames(names)
	} else {
		setDesktopNames(desktops)
	}
}

func SetNumberOfDesktops(num int) *Changes {
	if num < 1 {
		num = 1
	}
	currentNum := len(desktops)
	newLast := num - 1
	if num < currentNum {
		for i := num; i < currentNum; i++ {
			moveWinsToDesktop(i, newLast)
		}
		desktops = desktops[:num]
		SetDesktops()
		if currentDesktop > newLast {
			return SwitchToDesktop(newLast)
		} else if currentDesktop == newLast {
			return createChanges(nil, winsOfDesktop(currentDesktop))
		}
	} else if num > currentNum {
		desktops = append(desktops, getDesktopNames(currentNum, newLast)...)
		SetDesktops()
	}

	return createChanges(nil, nil)
}

func SwitchToDesktop(index int) *Changes {
	if currentDesktop == index || index >= len(desktops) {
		return createChanges(nil, nil)
	}

	invisible := make([]xproto.Window, 0, len(desktopToWins[currentDesktop]))
	visible := make([]xproto.Window, 0, len(desktopToWins[index]))

	for w := range desktopToWins[currentDesktop] {
		invisible = append(invisible, w)
	}
	for w := range desktopToWins[index] {
		visible = append(visible, w)
	}
	currentDesktop = index
	SetCurrentDesktop()

	return createChanges(invisible, visible)
}

func MoveWindowToDesktop(win xproto.Window, desktop int) *Changes {
	prev := winToDesktop[win]
	if prev == desktop || (desktop >= len(desktops) && desktop != allDesktops) {
		return createChanges(nil, nil)
	}
	delete(desktopToWins[prev], win)
	ensureDesktop(desktop)
	desktopToWins[desktop][win] = true
	winToDesktop[win] = desktop
	_ = ewmh.WmDesktopSet(X, win, uint(desktop))

	if IsDesktopVisible(prev) && !IsDesktopVisible(desktop) {
		return createChanges([]xproto.Window{win}, nil)
	} else if !IsDesktopVisible(prev) && IsDesktopVisible(desktop) {
		return createChanges(nil, []xproto.Window{win})
	}
	return createChanges(nil, nil)
}

func getInitialDesktopForWindow(win xproto.Window) int {
	_d, err := ewmh.WmDesktopGet(X, win)
	d := int(_d)
	if err != nil {
		// not specified
		return currentDesktop
	}
	if d == allDesktops || d < len(desktops) {
		return d
	}
	// TODO: Current, last, create additional desktops, or what?
	return len(desktops) - 1
}

func moveWinsToDesktop(from, to int) {
	ensureDesktop(to)
	for w := range desktopToWins[from] {
		desktopToWins[to][w] = true
		winToDesktop[w] = to
		_ = ewmh.WmDesktopSet(X, w, uint(to))
	}
	delete(desktopToWins, from)
}

func ensureDesktop(d int) {
	if desktopToWins[d] == nil {
		desktopToWins[d] = make(map[xproto.Window]bool)
	}
}

func winsOfDesktop(d int) []xproto.Window {
	ret := make([]xproto.Window, 0, len(desktopToWins[d]))
	for w := range desktopToWins[d] {
		ret = append(ret, w)
	}
	return ret
}

func createChanges(invisible, visible []xproto.Window) *Changes {
	return &Changes{Invisible: invisible, Visible: visible}
}
