package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/motif"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/janbina/swm/focus"
	"github.com/janbina/swm/geometry"
	"github.com/janbina/swm/stack"
	"github.com/janbina/swm/util"
	"log"
)

const (
	borderColorActive   = 0x00BCD4
	borderColorInactive = 0xCDDC39
)

type Window struct {
	win         *xwindow.Window
	moveState   *MoveState
	resizeState *ResizeState
	savedStates map[string]windowState
	maxedVert   bool
	maxedHorz   bool
	iconified   bool
	focused     bool
	mapped      bool
	layer       int

	name         string
	protocols    util.StringSet
	hints        *icccm.Hints
	normalHints  *icccm.NormalHints
	states       util.StringSet
	types        util.StringSet
	transientFor xproto.Window
}

type MoveState struct {
	rx, ry    int
	startGeom geometry.Geometry
}

type ResizeState struct {
	rx, ry    int
	direction int
	startGeom geometry.Geometry
}

func New(x *xgbutil.XUtil, xWin xproto.Window) *Window {
	window := &Window{
		win: xwindow.New(x, xWin),
	}

	window.fetchXProperties()

	window.savedStates = make(map[string]windowState)

	if window.shouldDecorate() {
		if err := util.SetBorder(window.win, 2, borderColorInactive); err != nil {
			log.Printf("Cannot set window border")
		}
	}

	if window.states["_NET_WM_STATE_MAXIMIZED_VERT"] && window.states["_NET_WM_STATE_MAXIMIZED_HORZ"] {
		window.states["_NET_WM_STATE_MAXIMIZED_VERT"] = false
		window.states["_NET_WM_STATE_MAXIMIZED_HORZ"] = false
		window.states["MAXIMIZED"] = true
	}
	for _, s := range window.states.GetActive() {
		window.UpdateState(ewmh.StateAdd, s)
	}

	if !window.types.Any("_NET_WM_WINDOW_TYPE_DESKTOP", "_NET_WM_WINDOW_TYPE_DOCK") {
		focus.InitialAdd(window)
	}

	if window.types["_NET_WM_WINDOW_TYPE_DESKTOP"] {
		window.layer = stack.LayerDesktop
	} else if window.types["_NET_WM_WINDOW_TYPE_DOCK"] {
		window.layer = stack.LayerDock
	} else {
		window.layer = stack.LayerDefault
	}

	return window
}

func (w *Window) Id() xproto.Window {
	return w.win.Id
}

func (w *Window) Geometry() (*geometry.Geometry, error) {
	return geometry.Get(w.win)
}

func (w *Window) Listen(evMasks ...int) error {
	return w.win.Listen(evMasks...)
}

func (w *Window) Map() {
	w.win.Map()
	w.mapped = true
	_ = w.SetIcccmState(icccm.StateNormal)
}

func (w *Window) Unmap() {
	w.win.Unmap()
	w.mapped = false
	_ = w.SetIcccmState(icccm.StateIconic)
}

func (w *Window) Destroy() {
	if w.protocols["WM_DELETE_WINDOW"] {
		atoms, err := util.Atoms(w.win.X, "WM_PROTOCOLS", "WM_DELETE_WINDOW")

		cm, err := xevent.NewClientMessage(32, w.win.Id, atoms[0], int(atoms[1]))
		if err != nil {
			return
		}

		xproto.SendEvent(w.win.X.Conn(), false, w.win.Id, 0, string(cm.Bytes()))
	} else {
		w.win.Kill()
	}
}

func (w *Window) Destroyed() {
	_ = w.SetIcccmState(icccm.StateWithdrawn)
	focus.Remove(w)
	stack.Remove(w)
}

func (w *Window) IsHidden() bool {
	return w.states["_NET_WM_STATE_HIDDEN"]
}

func (w *Window) fetchXProperties() {
	var err error
	X := w.win.X
	id := w.win.Id

	w.hints = getHintsForWindow(X, id)

	w.protocols = make(util.StringSet)
	if protocols, err := icccm.WmProtocolsGet(X, id); err != nil {
		log.Printf("Wm protocols not set: %s", err)
	} else {
		w.protocols.SetAll(protocols)
	}

	w.normalHints, err = icccm.WmNormalHintsGet(X, id)
	if err != nil {
		log.Printf("Error getting normal hints: %s", err)
		w.normalHints = &icccm.NormalHints{}
	}

	w.states = make(util.StringSet)
	states, _ := ewmh.WmStateGet(X, id)
	w.states.SetAll(states)

	w.types = getTypesForWindow(X, id)

	w.name = w.loadName()
}

func (w *Window) shouldDecorate() bool {
	if w.types.Any("_NET_WM_WINDOW_TYPE_DESKTOP", "_NET_WM_WINDOW_TYPE_DOCK", "_NET_WM_WINDOW_TYPE_SPLASH") {
		return false
	}

	mh, err := motif.WmHintsGet(w.win.X, w.win.Id)
	if err == nil && !motif.Decor(mh) {
		return false
	}

	return true
}

func getHintsForWindow(X *xgbutil.XUtil, win xproto.Window) *icccm.Hints {
	hints, err := icccm.WmHintsGet(X, win)
	if err != nil {
		hints = &icccm.Hints{
			Flags:        icccm.HintInput | icccm.HintState,
			Input:        1,
			InitialState: icccm.StateNormal,
		}
	}
	return hints
}

func getTypesForWindow(X *xgbutil.XUtil, win xproto.Window) util.StringSet {
	typesSet := make(util.StringSet)
	if types, err := ewmh.WmWindowTypeGet(X, win); err != nil {
		typesSet["_NET_WM_WINDOW_TYPE_NORMAL"] = true
	} else {
		typesSet.SetAll(types)
	}
	return typesSet
}

func (w *Window) loadName() string {
	name, _ := ewmh.WmNameGet(w.win.X, w.win.Id)
	if len(name) > 0 {
		return name
	}

	name, _ = icccm.WmNameGet(w.win.X, w.win.Id)
	if len(name) > 0 {
		return name
	}

	return ""
}
