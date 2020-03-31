package window

import (
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"log"
)

var stateHandlers = map[string][3]func(window *Window){
	//                              {StateRemove, StateAdd, StateToggle}
	"MAXIMIZED":                    {(*Window).UnMaximize, (*Window).Maximize, (*Window).MaximizeToggle},
	"_NET_WM_STATE_MAXIMIZED_VERT": {(*Window).UnMaximizeVert, (*Window).MaximizeVert, (*Window).MaximizeVertToggle},
	"_NET_WM_STATE_MAXIMIZED_HORZ": {(*Window).UnMaximizeHorz, (*Window).MaximizeHorz, (*Window).MaximizeHorzToggle},
}

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
	if fs, ok := stateHandlers[state]; !ok {
		log.Printf("Unsupported window state: %s", state)
	} else {
		fs[action](w)
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
