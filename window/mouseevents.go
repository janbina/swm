package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/janbina/swm/cursors"
	"log"
)

func dragMoveBegin(w *Window) xgbutil.MouseDragBeginFun {
	return func(X *xgbutil.XUtil, rx, ry, ex, ey int) (bool, xproto.Cursor) {
		log.Printf("Drag move begin: %d, %d, %d, %d", rx, ry, ex, ey)

		g, _ := w.Geometry()
		w.moveState = &MoveState{
			rx:        rx,
			ry:        ry,
			startGeom: *g,
		}

		return true, cursors.Fleur
	}
}

func dragMoveStep(w *Window) xgbutil.MouseDragFun {
	return func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
		log.Printf("Drag move step: %d, %d, %d, %d", rx, ry, ex, ey)

		g := w.moveState.startGeom
		g.AddX(rx - w.moveState.rx)
		g.AddY(ry - w.moveState.ry)

		w.Move(g.X(), g.Y())
	}
}

func dragMoveEnd(w *Window) xgbutil.MouseDragFun {
	return func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
		log.Printf("Drag move end: %d, %d, %d, %d", rx, ry, ex, ey)
		w.moveState = nil
	}
}

func getDragDirection(win *Window, ex, ey int) int {
	direction := ewmh.SizeRight

	g, err := win.Geometry()
	if err != nil {
		return direction
	}
	w := g.Width()
	h := g.Height()
	topThird := ey < h/3
	bottomThird := ey > h/3*2
	leftThird := ex < w/3
	rightThird := ex > w/3*2

	if topThird {
		if leftThird {
			direction = ewmh.SizeTopLeft
		} else if rightThird {
			direction = ewmh.SizeTopRight
		} else {
			direction = ewmh.SizeTop
		}
	} else if bottomThird {
		if leftThird {
			direction = ewmh.SizeBottomLeft
		} else if rightThird {
			direction = ewmh.SizeBottomRight
		} else {
			direction = ewmh.SizeBottom
		}
	} else {
		if ex < w/2 {
			direction = ewmh.SizeLeft
		} else {
			direction = ewmh.SizeRight
		}
	}

	return direction
}

func getCursorForDirection(d int) xproto.Cursor {
	switch d {
	case ewmh.SizeTop:
		return cursors.TopSide
	case ewmh.SizeTopRight:
		return cursors.TopRightCorner
	case ewmh.SizeBottomRight:
		return cursors.BottomRightCorner
	case ewmh.SizeBottom:
		return cursors.BottomSide
	case ewmh.SizeBottomLeft:
		return cursors.BottomLeftCorner
	case ewmh.SizeLeft:
		return cursors.LeftSide
	case ewmh.SizeTopLeft:
		return cursors.TopLeftCorner
	case ewmh.SizeRight:
		return cursors.RightSide
	default:
		return cursors.RightSide
	}
}

func (w *Window) DragResizeBegin(rx, ry, ex, ey int) (bool, xproto.Cursor) {
	log.Printf("Drag resize begin: %d, %d, %d, %d", rx, ry, ex, ey)

	direction := getDragDirection(w, ex, ey)
	cursor := getCursorForDirection(direction)

	g, _ := w.Geometry()

	w.resizeState = &ResizeState{
		rx:        rx,
		ry:        ry,
		direction: direction,
		startGeom: *g,
	}

	return true, cursor
}

func (w *Window) DragResizeStep(rx, ry, ex, ey int) {
	log.Printf("Drag resize step: %d, %d, %d, %d", rx, ry, ex, ey)

	d := w.resizeState.direction
	changeX := d == ewmh.SizeLeft || d == ewmh.SizeTopLeft || d == ewmh.SizeBottomLeft
	changeY := d == ewmh.SizeTop || d == ewmh.SizeTopLeft || d == ewmh.SizeTopRight
	changeW := d != ewmh.SizeTop && d != ewmh.SizeBottom
	changeH := d != ewmh.SizeLeft && d != ewmh.SizeRight

	xDiff := rx - w.resizeState.rx
	yDiff := ry - w.resizeState.ry

	g := w.resizeState.startGeom
	if changeX {
		g.AddX(xDiff)
	}
	if changeY {
		g.AddY(yDiff)
	}
	if changeW {
		if changeX {
			g.AddWidth(-xDiff)
		} else {
			g.AddWidth(xDiff)
		}
	}
	if changeH {
		if changeY {
			g.AddHeight(-yDiff)
		} else {
			g.AddHeight(yDiff)
		}
	}

	flags := xproto.ConfigWindowX | xproto.ConfigWindowY | xproto.ConfigWindowWidth | xproto.ConfigWindowHeight
	if g.Width() < int(w.normalHints.MinWidth) {
		g.SetWidth(int(w.normalHints.MinWidth))
		flags &= ^xproto.ConfigWindowX
	}
	if g.Height() < int(w.normalHints.MinHeight) {
		g.SetHeight(int(w.normalHints.MinHeight))
		flags &= ^xproto.ConfigWindowY
	}
	w.Configure(flags, g.X(), g.Y(), g.Width(), g.Height())
}

func (w *Window) DragResizeEnd(rx, ry, ex, ey int) {
	log.Printf("Drag resize end: %d, %d, %d, %d", rx, ry, ex, ey)
	w.resizeState = nil
}

func (w *Window) SetupMouseEvents(moveShortcut string, resizeShortcut string) {
	X := w.win.X

	// Detach old events
	mousebind.Detach(X, w.win.Id)

	if _, _, err := mousebind.ParseString(X, moveShortcut); err == nil {
		mousebind.Drag(
			X, X.Dummy(), w.win.Id, moveShortcut, true,
			dragMoveBegin(w), dragMoveStep(w), dragMoveEnd(w),
		)
	}

	if _, _, err := mousebind.ParseString(X, resizeShortcut); err == nil {
		dStart := xgbutil.MouseDragBeginFun(
			func(X *xgbutil.XUtil, rx, ry, ex, ey int) (bool, xproto.Cursor) {
				return w.DragResizeBegin(rx, ry, ex, ey)
			})
		dStep := xgbutil.MouseDragFun(
			func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
				w.DragResizeStep(rx, ry, ex, ey)
			})
		dEnd := xgbutil.MouseDragFun(
			func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
				w.DragResizeEnd(rx, ry, ex, ey)
			})
		mousebind.Drag(X, X.Dummy(), w.win.Id, resizeShortcut, true, dStart, dStep, dEnd)
	}
}

func (w *Window) DragBegin(xr, yr int16) {
	X := w.win.X
	mousebind.DragBegin(
		X,
		xevent.ButtonPressEvent{
			ButtonPressEvent: &xproto.ButtonPressEvent{
				RootX: xr,
				RootY: yr,
			},
		},
		X.Dummy(),
		w.win.Id,
		dragMoveBegin(w), dragMoveStep(w), dragMoveEnd(w),
	)
}
