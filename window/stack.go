package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/janbina/swm/stack"
	"github.com/janbina/swm/util"
)

func (w *Window) Raise() {
	stack.Raise(w)
}

func (w *Window) Layer() int {
	return w.layer
}

func (w *Window) TransientFor(otherId xproto.Window) bool {
	if w.Id() == otherId {
		return false
	}
	if w.transientFor == 0 {
		w.transientFor, _ = icccm.WmTransientForGet(w.win.X, w.win.Id)
	}
	if w.transientFor == otherId {
		return true
	} else if w.transientFor != 0 {
		return false
	}

	otherHints := getHintsForWindow(w.win.X, otherId)
	if w.hints.Flags&icccm.HintWindowGroup > 0 &&
		otherHints.Flags&icccm.HintWindowGroup > 0 &&
		w.hints.WindowGroup == otherHints.WindowGroup &&
		hasTransientType(w.types) {

		return !hasTransientType(getTypesForWindow(w.win.X, otherId))
	}
	return false
}

func (w *Window) StackSibling(sibling stack.StackingWindow, mode byte) {
	if sW, ok := sibling.(*Window); ok {
		w.win.StackSibling(sW.win.Id, mode)
	}
}

func hasTransientType(set util.StringSet) bool {
	return set.Any(
		"_NET_WM_WINDOW_TYPE_TOOLBAR",
		"_NET_WM_WINDOW_TYPE_MENU",
		"_NET_WM_WINDOW_TYPE_UTILITY",
		"_NET_WM_WINDOW_TYPE_DIALOG",
	)
}
