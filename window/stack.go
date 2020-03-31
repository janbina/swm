package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/janbina/swm/stack"
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
	return w.transientFor == otherId
}

func (w *Window) StackSibling(sibling stack.StackingWindow, mode byte) {
	if sW, ok := sibling.(*Window); ok {
		w.win.StackSibling(sW.win.Id, mode)
	}
}
