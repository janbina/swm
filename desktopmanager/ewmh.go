package desktopmanager

import (
	"fmt"
	"github.com/BurntSushi/xgbutil/ewmh"
)

func defaultDesktopName(pos int) string {
	return fmt.Sprintf("D.%d", pos+1)
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
	if num < minDesktops {
		num = minDesktops
	}
	return getDesktopNames(0, int(num)-1)
}

func SetDesktops() {
	_ = ewmh.NumberOfDesktopsSet(X, uint(len(desktops)))
	fromEwmh, _ := ewmh.DesktopNamesGet(X)
	if len(fromEwmh) < len(desktops) {
		// dont set names when shrinking
		setDesktopNames(desktops)
	}
}

func setDesktopNames(names []string) {
	_ = ewmh.DesktopNamesSet(X, names)
}

func getCurrentDesktopEwmh() int {
	d, _ := ewmh.CurrentDesktopGet(X)
	return int(d)
}

func SetCurrentDesktop() {
	_ = ewmh.CurrentDesktopSet(X, uint(currentDesktop))
}
