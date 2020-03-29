package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/motif"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/janbina/swm/geometry"
	"github.com/janbina/swm/util"
	"log"
)

type Window struct {
	win         *xwindow.Window
	moveState   *MoveState
	resizeState *ResizeState
	savedStates map[string]windowState
	maxedVert   bool
	maxedHorz   bool
	iconified   bool

	protocols   util.StringSet
	normalHints *icccm.NormalHints
	states      util.StringSet
	types       util.StringSet
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
		if err := util.SetBorder(window.win, 3, 0xff00ff); err != nil {
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

	return window
}

// Unique id of this window
// It is backed by window Id, but we don't want to expose that
// so all manipulations which needs xproto.Window must happen here
func (w *Window) Id() uint32 {
	return uint32(w.win.Id)
}

func (w *Window) Geometry() (*geometry.Geometry, error) {
	return geometry.Get(w.win)
}

func (w *Window) Listen(evMasks ...int) error {
	return w.win.Listen(evMasks...)
}

func (w *Window) Map() {
	w.win.Map()
	_ = w.SetIcccmState(icccm.StateNormal)
}

func (w *Window) Unmap() {
	w.win.Unmap()
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
}

func (w *Window) fetchXProperties() {
	var err error
	X := w.win.X
	id := w.win.Id

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

	w.types = make(util.StringSet)
	if types, err := ewmh.WmWindowTypeGet(X, id); err != nil {
		w.types["_NET_WM_WINDOW_TYPE_NORMAL"] = true
	} else {
		w.types.SetAll(types)
	}
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
