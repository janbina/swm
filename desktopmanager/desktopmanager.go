package desktopmanager

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
	minDesktops   = 1 //minimum number of desktops created at startup
	stickyDesktop = 0xFFFFFFFF
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
	desktopToWins[stickyDesktop] = make(map[xproto.Window]bool)
	winToDesktop = make(map[xproto.Window]int)
	desktops = getDesktops()
	currentDesktop = getCurrentDesktop()
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
	return desktop == stickyDesktop || desktop == currentDesktop
}

func IsWinDesktopVisible(win xproto.Window) bool {
	return IsDesktopVisible(winToDesktop[win])
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
			return &Changes{
				Invisible: nil,
				Visible:   winsOfDesktop(currentDesktop),
			}
		}
	} else if num > currentNum {
		desktops = append(desktops, getDesktopNames(currentNum, newLast)...)
		SetDesktops()
	}

	return &Changes{}
}

func SwitchToDesktop(index int) *Changes {
	if currentDesktop == index || index >= len(desktops) {
		return &Changes{}
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

	return &Changes{
		Invisible: invisible,
		Visible:   visible,
	}
}

func getInitialDesktopForWindow(win xproto.Window) int {
	_d, err := ewmh.WmDesktopGet(X, win)
	d := int(_d)
	if err != nil {
		// not specified
		return currentDesktop
	}
	if d == stickyDesktop || d < len(desktops) {
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
