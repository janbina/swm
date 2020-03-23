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
	maxedVert   bool
	maxedHorz   bool

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

type Directions struct {
	Left   int
	Right  int
	Bottom int
	Top    int
}

func New(x *xgbutil.XUtil, xWin xproto.Window) *Window {
	window := &Window{
		win: xwindow.New(x, xWin),
	}

	window.FetchXProperties()

	if window.shouldDecorate() {
		if err := util.SetBorder(window.win, 3, 0xff00ff); err != nil {
			log.Printf("Cannot set window border")
		}
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
}

func (w *Window) Focus() {
	w.win.Focus()
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
}

func (w *Window) Resize(d Directions) {
	g, _ := w.Geometry()
	x := g.X() + d.Left
	y := g.Y() + d.Top

	width := g.TotalWidth() + d.Right - d.Left
	height := g.TotalHeight() + d.Bottom - d.Top
	w.MoveResize(x, y, width, height)
}

func (w *Window) Move(x, y int) {
	w.win.Move(x, y)
}

func (w *Window) MoveResize(x, y, width, height int) {
	g, _ := w.Geometry()
	realWidth := width - 2 * g.BorderWidth()
	realHeight := height - 2 * g.BorderWidth()

	if realWidth < int(w.normalHints.MinWidth) {
		realWidth = int(w.normalHints.MinWidth)
	}
	if realHeight < int(w.normalHints.MinHeight) {
		realHeight = int(w.normalHints.MinHeight)
	}
	w.win.MoveResize(x, y, realWidth, realHeight)
}

func (w *Window) Configure(flags, x, y, width, height int) {
	g, _ := w.Geometry()
	realWidth := width - 2 * g.BorderWidth()
	realHeight := height - 2 * g.BorderWidth()

	if realWidth < int(w.normalHints.MinWidth) {
		realWidth = int(w.normalHints.MinWidth)
	}
	if realHeight < int(w.normalHints.MinHeight) {
		realHeight = int(w.normalHints.MinHeight)
	}
	w.win.Configure(flags, x, y, realWidth, realHeight, 0, 0)
}

func (w *Window) WasUnmapped() {
}

func (w *Window) FetchXProperties() {
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

func (w *Window) UpdateState(action int, state string) {
	switch state {
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

func (w *Window) MaximizeVert() {
	if w.maxedVert {
		return
	}
	w.maxedVert = true
	w.addStates("_NET_WM_STATE_MAXIMIZED_VERT")

	g, _ := xwindow.New(w.win.X, w.win.X.RootWin()).Geometry() // TODO: get real geometry
	log.Printf("GEOM: %s", g)
	w.Configure(xproto.ConfigWindowY|xproto.ConfigWindowHeight, 0, g.Y(), 0, g.Height())
}

func (w *Window) UnMaximizeVert() {
	if !w.maxedVert {
		return
	}
	w.maxedVert = false
	w.removeStates("_NET_WM_STATE_MAXIMIZED_VERT")

	// TODO: alter the geometry
}

func (w *Window) MaximizeVertToggle() {
	if w.maxedVert {
		w.UnMaximizeVert()
	} else {
		w.MaximizeVert()
	}
}

func (w *Window) MaximizeHorz() {
	if w.maxedHorz {
		return
	}
	w.maxedHorz = true
	w.addStates("_NET_WM_STATE_MAXIMIZED_HORZ")

	g, _ := xwindow.New(w.win.X, w.win.X.RootWin()).Geometry() // TODO: get real geometry
	w.Configure(xproto.ConfigWindowX|xproto.ConfigWindowWidth, g.X(), 0, g.Width(), 0)
}

func (w *Window) UnMaximizeHorz() {
	if !w.maxedHorz {
		return
	}
	w.maxedHorz = false
	w.removeStates("_NET_WM_STATE_MAXIMIZED_HORZ")

	// TODO: alter the geometry
}

func (w *Window) MaximizeHorzToggle() {
	if w.maxedHorz {
		w.UnMaximizeHorz()
	} else {
		w.MaximizeHorz()
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
