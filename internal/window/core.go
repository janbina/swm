package window

import (
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/janbina/swm/internal/config"
	"github.com/janbina/swm/internal/decoration"
	"github.com/janbina/swm/internal/focus"
	"github.com/janbina/swm/internal/stack"
	"github.com/janbina/swm/internal/util"
)

type Window struct {
	win         *xwindow.Window
	parent      *xwindow.Window
	infoWin     *xwindow.Window
	infoTimer   *time.Timer
	decorations decoration.Decorations
	moveState   *MoveState
	resizeState *ResizeState
	savedStates map[state]windowState
	unmapIgnore int

	maxedVert        bool
	maxedHorz        bool
	iconified        bool
	tmpDeiconified   bool
	focused          bool
	mapped           bool
	layer            int
	demandsAttention bool
	fullscreen       bool
	skipTaskbar      bool
	skipPager        bool

	info         *WinInfo
	transientFor xproto.Window
}

type MoveState struct {
	rx, ry    int
	startGeom xrect.Rect
}

type ResizeState struct {
	rx, ry    int
	direction int
	startGeom xrect.Rect
}

func New(xWin *xwindow.Window, info *WinInfo, actions *WinActions) *Window {
	window := &Window{
		win:  xWin,
		info: info,
	}

	window.savedStates = make(map[state]windowState)

	_ = util.SetBorderWidth(window.win, 0)

	window.unmapIgnore++
	window.parent, _ = reparent(xWin)

	decorations := make(decoration.Decorations, 0)

	if actions.ShouldDecorate {
		decorations = append(decorations, createBorders(window.parent)...)
	}

	window.decorations = decorations

	window.MoveResizeWinSize(
		true,
		actions.Geometry.X(),
		actions.Geometry.Y(),
		actions.Geometry.Width(),
		actions.Geometry.Height(),
	)

	window.layer = actions.StackLayer
	window.iconified = actions.StartIconified

	window.updateFrameExtents()

	window.infoWin, _ = util.CreateTransparentWindow(xWin.X, window.parent.Id)

	return window
}

func reparent(xWin *xwindow.Window) (*xwindow.Window, error) {
	parent, err := util.CreateTransparentWindow(xWin.X, xWin.X.RootWin())
	if err != nil {
		return nil, err
	}

	var events uint32 = xproto.EventMaskSubstructureRedirect |
		xproto.EventMaskButtonPress |
		xproto.EventMaskButtonRelease |
		xproto.EventMaskFocusChange

	parent.Change(xproto.CwEventMask, events)

	err = xproto.ReparentWindowChecked(xWin.X.Conn(), xWin.Id, parent.Id, 0, 0).Check()
	if err != nil {
		return nil, err
	}

	return parent, nil
}

func createBorders(parent *xwindow.Window) decoration.Decorations {
	return decoration.Decorations{
		decoration.CreateBorder(parent, decoration.Top, config.BorderTop),
		decoration.CreateBorder(parent, decoration.Bottom, config.BorderBottom),
		decoration.CreateBorder(parent, decoration.Left, config.BorderLeft),
		decoration.CreateBorder(parent, decoration.Right, config.BorderRight),
	}
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
	w.mapped = true
	w.iconified = false
	_ = w.SetIcccmState(icccm.StateNormal)
}

func (w *Window) Unmap() {
	w.unmapIgnore++
	w.parent.Unmap()
	w.win.Unmap()
	w.mapped = false
	w.iconified = true
	_ = w.SetIcccmState(icccm.StateIconic)
}

func (w *Window) UnmapNotify() {
	w.unmapIgnore--
}

func (w *Window) UnmapNotifyShouldUnmanage() bool {
	return w.unmapIgnore == 0
}

func (w *Window) Iconify() {
	w.Unmap()
	w.AddStates("_NET_WM_STATE_HIDDEN")
	focus.FocusLast()
}

func (w *Window) DeIconify() {
	w.Map()
	w.RemoveStates("_NET_WM_STATE_HIDDEN")
	w.Focus()
	w.Raise()
}

func (w *Window) Destroy() {
	if w.info.Protocols["WM_DELETE_WINDOW"] {
		atoms, _ := util.Atoms(w.win.X, "WM_PROTOCOLS", "WM_DELETE_WINDOW")
		cm, _ := xevent.NewClientMessage(32, w.win.Id, atoms[0], int(atoms[1]))

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
	w.win.Detach()
	w.parent.Unmap()
}

func (w *Window) IsHidden() bool {
	return w.info.States["_NET_WM_STATE_HIDDEN"]
}

func (w *Window) IsIconified() bool {
	return w.iconified
}

func (w *Window) IsMouseMoveable() bool {
	return !w.fullscreen && !w.info.Types.Any("_NET_WM_WINDOW_TYPE_DESKTOP", "_NET_WM_WINDOW_TYPE_DOCK")
}

func (w *Window) IsMouseResizable() bool {
	return !w.fullscreen && !w.info.Types.Any("_NET_WM_WINDOW_TYPE_DESKTOP", "_NET_WM_WINDOW_TYPE_DOCK")
}
