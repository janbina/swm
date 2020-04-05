package window

import (
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/janbina/swm/stack"
)

func (w *Window) Raise() {
	stack.Raise(w)
}

func (w *Window) Layer() int {
	if w.layer == stack.LayerFullscreen && !w.focused {
		// Effective layer of fullscreen window depends on its focus state
		return stack.LayerDefault
	}
	return w.layer
}

func (w *Window) TransientFor(_other stack.StackingWindow) bool {
	other, ok := _other.(*Window)
	if !ok {
		return false
	}
	if w.Id() == other.Id() {
		return false
	}
	if w.transientFor == 0 {
		w.transientFor, _ = icccm.WmTransientForGet(w.win.X, w.win.Id)
	}
	if w.transientFor == other.Id() {
		return true
	} else if w.transientFor != 0 {
		return false
	}

	if w.hints.Flags&icccm.HintWindowGroup > 0 &&
		other.hints.Flags&icccm.HintWindowGroup > 0 &&
		w.hints.WindowGroup == other.hints.WindowGroup {

		return w.hasTransientType() && !other.hasTransientType()
	}
	return false
}

func (w *Window) StackSibling(sibling stack.StackingWindow, mode byte) {
	if sW, ok := sibling.(*Window); ok {
		w.parent.StackSibling(sW.parent.Id, mode)
	}
}

func (w *Window) hasTransientType() bool {
	return w.types.Any(
		"_NET_WM_WINDOW_TYPE_TOOLBAR",
		"_NET_WM_WINDOW_TYPE_MENU",
		"_NET_WM_WINDOW_TYPE_UTILITY",
		"_NET_WM_WINDOW_TYPE_DIALOG",
	)
}
