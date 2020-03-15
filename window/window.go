package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/janbina/swm/cursors"
	"github.com/janbina/swm/util"
	"log"
)

type Window struct {
	parent    *xwindow.Window
	win       *xwindow.Window
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

	w := xwindow.New(x, xWin)
	p, err := createParent(w)

	win := &Window{
		parent:    p,
		win:       xwindow.New(x, xWin),
		protocols: protocols,
		moveState: &MoveState{
			Moving: false,
			dX:     0,
			dY:     0,
		},
	}

	return win
}

// Create parent window for [win]
func createParent(win *xwindow.Window) (*xwindow.Window, error) {
	X := win.X

	parent, err := xwindow.Generate(X)
	if err != nil {
		return nil, err
	}

	g, err := win.Geometry()
	if err != nil {
		return nil, err
	}

	err = parent.CreateChecked(
		X.RootWin(),
		g.X(),
		g.Y(),
		g.Width(),
		g.Height(),
		xproto.CwBackPixel,
		0xff0000, // TODO: red bg so we can easily see if we are off a bit, maybe set to something else later
	)
	if err != nil {
		return nil, err
	}

	// Set window border to 0, as we will either use our own borders or we don't want any
	err = xproto.ConfigureWindowChecked(X.Conn(), win.Id, xproto.ConfigWindowBorderWidth, []uint32{0}, ).Check()
	if err != nil {
		return nil, err
	}

	err = xproto.ReparentWindowChecked(X.Conn(), win.Id, parent.Id, 0, 0).Check()
	if err != nil {
		return nil, err
	}

	return parent, nil
}

func (w *Window) Id() xproto.Window {
	return w.win.Id
}

func (w *Window) Geometry() (xrect.Rect, error) {
	return w.parent.Geometry()
}

func (w *Window) Listen(evMasks ...int) error {
	return w.win.Listen(evMasks...)
}

func (w *Window) Map() {
	w.parent.Map()
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
	g, _ := w.Geometry()
	x := g.X() + d.Left
	y := g.Y() + d.Top
	width := g.Width() + d.Right - d.Left
	height := g.Height() + d.Bottom - d.Top
	w.MoveResize(x, y, width, height)
}

func (w *Window) Move(x, y int) {
	w.parent.Move(x, y)
}

func (w *Window) MoveResize(x, y, width, height int) {
	w.parent.MoveResize(x, y, width, height)
	w.win.Resize(width, height)
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

	g, _ := w.Geometry()
	w.moveState = &MoveState{
		Moving: true,
		dX:     g.X() - rx,
		dY:     g.Y() - ry,
	}

	return true
}

func (w *Window) DragMoveStep(rx, ry, ex, ey int) {
	log.Printf("Drag move step: %d, %d, %d, %d", rx, ry, ex, ey)

	w.Move(w.moveState.dX+rx, w.moveState.dY+ry)
}

func (w *Window) DragMoveEnd(rx, ry, ex, ey int) {
	log.Printf("Drag move end: %d, %d, %d, %d", rx, ry, ex, ey)

	w.moveState = &MoveState{
		Moving: false,
		dX:     0,
		dY:     0,
	}
}

func (w *Window) SetupMouseEvents(moveShortcut string) {
	X := w.win.X

	// Detach old events
	mousebind.Detach(X, w.parent.Id)

	if _, _, err := mousebind.ParseString(X, moveShortcut); err != nil {
		return
	}
	dStart := xgbutil.MouseDragBeginFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) (bool, xproto.Cursor) {
			return w.DragMoveBegin(rx, ry, ex, ey), cursors.Fleur
		})
	dStep := xgbutil.MouseDragFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
			w.DragMoveStep(rx, ry, ex, ey)
		})
	dEnd := xgbutil.MouseDragFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
			w.DragMoveEnd(rx, ry, ex, ey)
		})
	mousebind.Drag(X, X.Dummy(), w.parent.Id, moveShortcut, true, dStart, dStep, dEnd)
}
