package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/janbina/swm/geometry"
	"github.com/janbina/swm/util"
	"log"
)

type Window struct {
	win         *xwindow.Window
	protocols   []string
	moveState   *MoveState
	resizeState *ResizeState
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

	protocols, err := icccm.WmProtocolsGet(x, xWin)
	if err != nil {
		log.Println("Wm protocols not set")
	}

	win := xwindow.New(x, xWin)

	err = util.SetBorder(win, 3, 0xff00ff)
	if err != nil {
		log.Printf("Cannot set window border")
	}

	window := &Window{
		win:         win,
		protocols:   protocols,
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
	if w.HasProtocol("WM_DELETE_WINDOW") {
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

func (w *Window) Resize(d Directions) {
	g, _ := w.Geometry()
	x := g.X() + d.Left
	y := g.Y() + d.Top
	width := g.Width() + d.Right - d.Left
	height := g.Height() + d.Bottom - d.Top
	w.MoveResize(x, y, width, height)
}

func (w *Window) Move(x, y int) {
	w.win.Move(x, y)
}

func (w *Window) MoveResize(x, y, width, height int) {
	w.win.MoveResize(x, y, width, height)
}

func (w *Window) HasProtocol(x string) bool {
	for _, p := range w.protocols {
		if x == p {
			return true
		}
	}
	return false
}

func (w *Window) WasUnmapped() {
}
