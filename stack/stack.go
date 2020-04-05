package stack

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"sort"
	"time"
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

var (
	X *xgbutil.XUtil
	// slice of all windows to be stacked
	windows []StackingWindow
	// maps win to timestamp it was raised last
	// also used as indicator if win is present in windows slice - those two should be kept in sync
	raiseTimestamp map[xproto.Window]int64
	// flag which tells if we are currently inside tmp stacking,
	// which is initiated by TmpRaise() and ended by Raise()
	// while active, ReStack() does nothing so tmp stack order is not messed up
	tmpStacking = false
)

func Initialize(x *xgbutil.XUtil) {
	X = x
	raiseTimestamp = make(map[xproto.Window]int64)
}

// Raise adds win to list of windows if its not there yet,
// and updates raise timestamps for win and its transients
func Raise(win StackingWindow) {
	if _, ok := raiseTimestamp[win.Id()]; !ok {
		windows = append(windows, win)
	}

	t := time.Now().UnixNano()
	raiseTimestamp[win.Id()] = t
	for _, w := range windows {
		if w.TransientFor(win.Id()) {
			raiseTimestamp[w.Id()] = t + 1
		}
	}

	tmpStacking = false

	ReStack()
}

// ReStack sorts windows by stacking order (layer, raise timestamp)
// and invokes realiseStacking(). Usually called from Raise() after raising window,
// but can be also called standalone, typically when we change some windows layer
// but we dont want to raise it - useful for fullscreen windows
func ReStack() {
	if tmpStacking {
		return
	}

	sort.SliceStable(windows, func(i, j int) bool {
		a := windows[i]
		b := windows[j]
		if a.Layer() == b.Layer() {
			return raiseTimestamp[a.Id()] < raiseTimestamp[b.Id()]
		}
		return a.Layer() < b.Layer()
	})

	realiseStacking(windows)

	updateEwmhStacking()
}

func Remove(win StackingWindow) {
	for i, w := range windows {
		if w.Id() == win.Id() {
			windows = append(windows[:i], windows[i+1:]...)
			return
		}
	}
	delete(raiseTimestamp, win.Id())
	updateEwmhStacking()
}

// temporarily raise win above all others, ignoring layers - used for cycling windows
func TmpRaise(win StackingWindow) {
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

	tmpStacking = true

	realiseStacking(tmpWins)
}

func realiseStacking(wins []StackingWindow) {
	if len(wins) <= 1 {
		return
	}
	wins[0].StackSibling(wins[1], xproto.StackModeBelow)
	for i := 1; i < len(wins); i++ {
		wins[i].StackSibling(wins[i-1], xproto.StackModeAbove)
	}
}

func updateEwmhStacking() {
	ids := make([]xproto.Window, len(windows))
	for i, win := range windows {
		ids[i] = win.Id()
	}
	_ = ewmh.ClientListStackingSet(X, ids)
}
