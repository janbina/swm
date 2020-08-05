package window

import (
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
)

func (w *Window) GetActiveStates() []string {
	return w.info.States.GetActive()
}

func (w *Window) AddStates(states ...string) {
	w.info.States.SetAll(states)
	ewmh.WmStateSet(w.win.X, w.win.Id, w.info.States.GetActive())
}

func (w *Window) RemoveStates(states ...string) {
	w.info.States.UnSetAll(states)
	ewmh.WmStateSet(w.win.X, w.win.Id, w.info.States.GetActive())
}

func (w *Window) SetIcccmState(state uint) error {
	return icccm.WmStateSet(w.win.X, w.win.Id, &icccm.WmState{State: state})
}
