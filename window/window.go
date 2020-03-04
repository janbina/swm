package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/janbina/swm/util"
	"log"
)

type Window struct {
	win *xwindow.Window
	protocols []string
	moveState *MoveState
}

type MoveState struct {
	Moving bool
	dX, dY int
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
	win := &Window{
		win: xwindow.New(x, xWin),
		protocols: protocols,
		moveState: &MoveState{
			Moving: false,
			dX:     0,
			dY:     0,
		},
	}

	return win
}

func (w *Window) Id() xproto.Window {
	return w.win.Id
}

func (w *Window) Geometry() (xrect.Rect, error) {
	return w.win.Geometry()
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

		cm, err := xevent.NewClientMessage(32, w.Id(), atoms[0], int(atoms[1]))
		if err != nil {
			return
		}

		xproto.SendEvent(w.win.X.Conn(), false, w.Id(), 0, string(cm.Bytes()))
	} else {
		w.win.Kill()
	}
}

func (w *Window) Resize(d Directions) {
	g, _ := w.win.Geometry()
	x := g.X() + d.Left
	y := g.Y() + d.Top
	width := g.Width() + d.Right - d.Left
	height := g.Height() + d.Bottom - d.Top
	w.win.MoveResize(x, y, width, height)
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

func (w *Window) DragMoveBegin(rx, ry, ex, ey int) bool {
	log.Printf("Drag move begin: %d, %d, %d, %d", rx, ry, ex, ey)

	g, _ := w.win.Geometry()
	w.moveState = &MoveState{
		Moving: true,
		dX:     g.X() - rx,
		dY:     g.Y() - ry,
	}

	return true
}

func (w *Window) DragMoveStep(rx, ry, ex, ey int) {
	log.Printf("Drag move step: %d, %d, %d, %d", rx, ry, ex, ey)

	w.win.Move(w.moveState.dX + rx, w.moveState.dY + ry)
}

func (w *Window) DragMoveEnd(rx, ry, ex, ey int) {
	log.Printf("Drag move end: %d, %d, %d, %d", rx, ry, ex, ey)

	w.moveState = &MoveState{
		Moving: false,
		dX:     0,
		dY:     0,
	}
}
