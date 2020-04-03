package window

import (
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
)

func (w *Window) GetActiveStates() []string {
	return w.states.GetActive()
}

func (w *Window) addStates(states ...string) {
	w.states.SetAll(states)
	ewmh.WmStateSet(w.win.X, w.win.Id, w.states.GetActive())
}

func (w *Window) removeStates(states ...string) {
	w.states.UnSetAll(states)
	ewmh.WmStateSet(w.win.X, w.win.Id, w.states.GetActive())
}

func (w *Window) SetIcccmState(state uint) error {
	return icccm.WmStateSet(w.win.X, w.win.Id, &icccm.WmState{State: state})
}
