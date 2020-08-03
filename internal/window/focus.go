package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/janbina/swm/internal/focus"
	"github.com/janbina/swm/internal/stack"
	"github.com/janbina/swm/internal/util"
)

func (w *Window) IsFocusable() bool {
	return w.mapped
}

func (w *Window) IsFocused() bool {
	return w.focused
}

func (w *Window) CanFocus() bool {
	if w.info.Hints.Flags&icccm.HintInput > 0 {
		return w.info.Hints.Input == 1
	}
	return true
}

func (w *Window) ShouldSendFocusNotify() bool {
	return w.info.Protocols["WM_TAKE_FOCUS"]
}

func (w *Window) PrepareForFocus(tmp bool) {
	if w.iconified {
		w.tmpDeiconified = tmp
		w.IconifyToggle()
	}
}

func (w *Window) Focused() {
	w.StopAttention()
	w.focused = true
	focus.SetFocus(w)
	w.decorations.Active()
	_ = ewmh.ActiveWindowSet(w.win.X, w.win.Id)
	w.AddStates("_NET_WM_STATE_FOCUSED")
	if w.layer == stack.LayerFullscreen {
		// Effective layer of fullscreen window depends on its focus state (see Window.Layer()),
		// so we have to restack after changing its focus state
		stack.ReStack()
	}
}

func (w *Window) Unfocused() {
	w.focused = false
	if w.tmpDeiconified {
		w.tmpDeiconified = false
		w.IconifyToggle()
	}
	w.decorations.InActive()
	_ = ewmh.ActiveWindowSet(w.win.X, 0)
	w.RemoveStates("_NET_WM_STATE_FOCUSED")
	if w.layer == stack.LayerFullscreen {
		// Effective layer of fullscreen window depends on its focus state (see Window.Layer()),
		// so we have to restack after changing its focus state
		stack.ReStack()
	}
}

func (w *Window) RemoveTmpDeiconified() {
	w.tmpDeiconified = false
}

func (w *Window) FocusToggle() {
	if w.focused {
		w.Unfocused()
	} else {
		w.Focused()
	}
}

func (w *Window) ApplyFocus() {
	if w.CanFocus() {
		w.win.Focus()
	}
	if w.ShouldSendFocusNotify() {
		atoms, err := util.Atoms(w.win.X, "WM_PROTOCOLS", "WM_TAKE_FOCUS")
		if err != nil {
			return
		}

		cm, err := xevent.NewClientMessage(32, w.win.Id, atoms[0], int(atoms[1]), int(w.win.X.TimeGet()))
		if err != nil {
			return
		}

		xproto.SendEvent(w.win.X.Conn(), false, w.win.Id, 0, string(cm.Bytes()))
	}
}

func (w *Window) Focus() {
	focus.Focus(w)
}

func (w *Window) SetupFocusListeners() {
	w.handleFocusIn().Connect(w.win.X, w.parent.Id)
	w.handleFocusOut().Connect(w.win.X, w.parent.Id)
}

func (w *Window) handleFocusIn() xevent.FocusInFun {
	return func(X *xgbutil.XUtil, e xevent.FocusInEvent) {
		if focus.AcceptClientFocus(e.Mode, e.Detail) {
			w.Focused()
		}
	}
}

func (w *Window) handleFocusOut() xevent.FocusOutFun {
	return func(X *xgbutil.XUtil, e xevent.FocusOutEvent) {
		if focus.AcceptClientFocus(e.Mode, e.Detail) {
			w.Unfocused()
		}
	}
}
