package window

import (
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"log"
)

func (w *Window) UpdateStates(action int, s1 string, s2 string) {
	if (s1 == "_NET_WM_STATE_MAXIMIZED_VERT" && s2 == "_NET_WM_STATE_MAXIMIZED_HORZ") ||
		(s2 == "_NET_WM_STATE_MAXIMIZED_VERT" && s1 == "_NET_WM_STATE_MAXIMIZED_HORZ") {
		w.UpdateState(action, "MAXIMIZED")
	} else {
		w.UpdateState(action, s1)
		if len(s2) > 0 {
			w.UpdateState(action, s2)
		}
	}
}

func (w *Window) UpdateState(action int, state string) {
	switch state {
	case "MAXIMIZED":
		switch action {
		case ewmh.StateRemove:
			w.UnMaximize()
		case ewmh.StateAdd:
			w.Maximize()
		case ewmh.StateToggle:
			w.MaximizeToggle()
		}
	case "_NET_WM_STATE_MAXIMIZED_VERT":
		switch action {
		case ewmh.StateRemove:
			w.UnMaximizeVert()
		case ewmh.StateAdd:
			w.MaximizeVert()
		case ewmh.StateToggle:
			w.MaximizeVertToggle()
		}
	case "_NET_WM_STATE_MAXIMIZED_HORZ":
		switch action {
		case ewmh.StateRemove:
			w.UnMaximizeHorz()
		case ewmh.StateAdd:
			w.MaximizeHorz()
		case ewmh.StateToggle:
			w.MaximizeHorzToggle()
		}
	default:
		log.Printf("Unsupported state: %s", state)
	}
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
