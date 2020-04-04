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
	borderColorActive    = 0x00BCD4
	borderColorInactive  = 0xCDDC39
	borderColorAttention = 0xF44336
)

type Window struct {
	win         *xwindow.Window
	parent      *xwindow.Window
	moveState   *MoveState
	resizeState *ResizeState
	savedStates map[string]windowState

	maxedVert        bool
	maxedHorz        bool
	iconified        bool
	focused          bool
	mapped           bool
	layer            int
	demandsAttention bool
	fullscreen       bool

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

	_ = util.SetBorderWidth(window.win, 0)

	g, _ := window.win.Geometry()

	window.parent, _ = reparent(x, xWin)

	window.parent.MoveResize(g.Pieces())

	if window.shouldDecorate() {
		if err := util.SetBorder(window.parent, 2, borderColorInactive); err != nil {
			log.Printf("Cannot set window border")
		}
	}

	if window.states["_NET_WM_STATE_MAXIMIZED_VERT"] && window.states["_NET_WM_STATE_MAXIMIZED_HORZ"] {
		window.states["_NET_WM_STATE_MAXIMIZED_VERT"] = false
		window.states["_NET_WM_STATE_MAXIMIZED_HORZ"] = false
		window.states["MAXIMIZED"] = true
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

	window.iconified = window.normalHints.Flags&icccm.HintState > 0 && window.hints.InitialState == icccm.StateIconic

	return window
}

func reparent(X *xgbutil.XUtil, xWin xproto.Window) (*xwindow.Window, error) {
	parent, err := xwindow.Generate(X)
	if err != nil {
		return nil, err
	}

	err = parent.CreateChecked(X.RootWin(), 0, 0, 1, 1, xproto.CwEventMask,
		xproto.EventMaskSubstructureRedirect|
			xproto.EventMaskButtonPress|
			xproto.EventMaskButtonRelease|
			xproto.EventMaskFocusChange,
	)
	if err != nil {
		return nil, err
	}

	err = xproto.ReparentWindowChecked(X.Conn(), xWin, parent.Id, 0, 0).Check()
	if err != nil {
		return nil, err
	}

	return parent, nil
}

func (w *Window) Id() xproto.Window {
	return w.win.Id
}

func (w *Window) Geometry() (*geometry.Geometry, error) {
	return geometry.Get(w.parent)
}

func (w *Window) Listen(evMasks ...int) error {
	return w.win.Listen(evMasks...)
}

func (w *Window) Map() {
	w.parent.Map()
	w.win.Map()
	w.mapped = true
	w.iconified = false
	_ = w.SetIcccmState(icccm.StateNormal)
}

func (w *Window) Unmap() {
	w.parent.Unmap()
	w.win.Unmap()
	w.mapped = false
	w.iconified = true
	_ = w.SetIcccmState(icccm.StateIconic)
}

func (w *Window) Hide() {
	w.Unmap()
	w.addStates("_NET_WM_STATE_HIDDEN")
	focus.FocusLast() // Todo: how to do this better
}

func (w *Window) Show() {
	w.Map()
	w.removeStates("_NET_WM_STATE_HIDDEN")
	w.Focus()
	w.Raise()
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
	xproto.ReparentWindow(w.win.X.Conn(), w.win.Id, w.win.X.RootWin(), 0, 0)
	w.win.Destroy()
	w.parent.Destroy()
}

func (w *Window) IsHidden() bool {
	return w.states["_NET_WM_STATE_HIDDEN"]
}

func (w *Window) IsIconified() bool {
	return w.iconified
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
