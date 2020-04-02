package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/janbina/swm/focus"
	"github.com/janbina/swm/util"
)

func (w *Window) IsFocusable() bool {
	return w.mapped
}

func (w *Window) IsFocused() bool {
	return w.focused
}

func (w *Window) CanFocus() bool {
	if w.hints.Flags&icccm.HintInput > 0 {
		return w.hints.Input == 1
	}
	return true
}

func (w *Window) ShouldSendFocusNotify() bool {
	return w.protocols["WM_TAKE_FOCUS"]
}

func (w *Window) PrepareForFocus() {
	if w.iconified {
		w.IconifyToggle()
	}
}

func (w *Window) Focused() {
	w.focused = true
	focus.SetFocus(w)
	_ = util.SetBorderColor(w.parent, borderColorActive)
	_ = ewmh.ActiveWindowSet(w.win.X, w.win.Id)
	w.addStates("_NET_WM_STATE_FOCUSED")

}

func (w *Window) Unfocused() {
	w.focused = false
	_ = util.SetBorderColor(w.parent, borderColorInactive)
	_ = ewmh.ActiveWindowSet(w.win.X, 0)
	w.removeStates("_NET_WM_STATE_FOCUSED")
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
