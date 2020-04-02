package stack

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"sort"
)

type StackingWindow interface {
	Id() xproto.Window
	Layer() int
	TransientFor(win xproto.Window) bool
	StackSibling(sibling StackingWindow, mode byte)
}

const (
	LayerDesktop = iota
	LayerBelow
	LayerDefault
	LayerAbove
	LayerDock
	LayerFullscreen
)

var X *xgbutil.XUtil
// windows in their stacking order, from lowest to highest
var windows []StackingWindow

func Initialize(x *xgbutil.XUtil) {
	X = x
}

// Raise adds win to list of windows if its not there yet,
// sorts windows by layer, putting win and its transients to the top of their layers,
// and issues appropriate StackSibling commands
func Raise(win StackingWindow) {
	// remove win and add it to the end to make sure it is in the slice exactly once
	remove(win)
	windows = append(windows, win)

	// sort windows by layer, win and its transients on top of their layers
	sort.SliceStable(windows, func(i, j int) bool {
		a := windows[i]
		b := windows[j]
		if a.Layer() < b.Layer() {
			return true
		}
		if a.Layer() > b.Layer() {
			return false
		}
		if b.TransientFor(win.Id()) {
			return true
		}
		if a.TransientFor(win.Id()) {
			return false
		}
		return b.Id() == win.Id()
	})

	// now find win and its transients in the slice and issue StackSibling for them
	// stack the last above the second last and all other below the next one
	if len(windows) <= 1 {
		return
	}
	last := len(windows) - 1
	for i, w := range windows {
		if w.Id() == win.Id() || w.TransientFor(win.Id()) {
			if i == last {
				w.StackSibling(windows[i - 1], xproto.StackModeAbove)
			} else {
				w.StackSibling(windows[i + 1], xproto.StackModeBelow)
			}
		}
	}

	updateEwmhStacking()
}

func Remove(win StackingWindow) {
	remove(win)
	updateEwmhStacking()
}

func remove(win StackingWindow) {
	for i, w := range windows {
		if w.Id() == win.Id() {
			windows = append(windows[:i], windows[i+1:]...)
			return
		}
	}
}

func updateEwmhStacking() {
	ids := make([]xproto.Window, len(windows))
	for i, win := range windows {
		ids[i] = win.Id()
	}
	_ = ewmh.ClientListStackingSet(X, ids)
}

func TempRaise(win StackingWindow) {
	if len(windows) <= 1 {
		return
	}
	tmpWins := make([]StackingWindow, 0, len(windows))
	for _, w := range windows {
		if w.Id() != win.Id() {
			tmpWins = append(tmpWins, w)
		}
	}
	tmpWins = append(tmpWins, win)

	tmpWins[0].StackSibling(tmpWins[1], xproto.StackModeBelow)
	for i := 1; i < len(tmpWins); i++ {
		tmpWins[i].StackSibling(tmpWins[i-1], xproto.StackModeAbove)
	}
}
