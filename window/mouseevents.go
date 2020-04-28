package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/janbina/swm/config"
	"github.com/janbina/swm/cursors"
	"log"
)

func (w *Window) SetupMouseEvents() {
	X := w.win.X

	// Detach old events
	mousebind.Detach(X, w.win.Id)

	if _, _, err := mousebind.ParseString(X, config.MoveDragShortcut); err == nil {
		mousebind.Drag(
			X, X.Dummy(), w.win.Id, config.MoveDragShortcut, true,
			dragMoveBegin(w), dragMoveStep(w), dragMoveEnd(w),
		)
	}

	if _, _, err := mousebind.ParseString(X, config.ResizeDragShortcut); err == nil {
		mousebind.Drag(
			X, X.Dummy(), w.win.Id, config.ResizeDragShortcut, true,
			dragResizeBegin(w, ewmh.Infer), dragResizeStep(w), dragResizeEnd(w),
		)
	}

	_ = mousebind.ButtonPressFun(func(X *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
		w.Focus()
		w.Raise()
		xevent.ReplayPointer(X)
	}).Connect(X, w.win.Id, "1", true, true)
}

func (w *Window) DragMoveBegin(xr, yr int16) {
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

func (w *Window) DragResizeBegin(xr, yr int16, dir int) {
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
		dragResizeBegin(w, dir), dragResizeStep(w), dragResizeEnd(w),
	)
}

func (w *Window) DragResizeBeginEvent(xr, yr, xe, ye int16) {
	X := w.win.X
	mousebind.DragBegin(
		X,
		xevent.ButtonPressEvent{
			ButtonPressEvent: &xproto.ButtonPressEvent{
				RootX:  xr,
				RootY:  yr,
				EventX: xe,
				EventY: ye,
			},
		},
		X.Dummy(),
		w.win.Id,
		dragResizeBegin(w, ewmh.Infer), dragResizeStep(w), dragResizeEnd(w),
	)
}

func dragMoveBegin(w *Window) xgbutil.MouseDragBeginFun {
	return func(X *xgbutil.XUtil, rx, ry, ex, ey int) (bool, xproto.Cursor) {
		if !w.IsMouseMoveable() {
			return false, 0
		}
		log.Printf("Drag move begin: %d, %d, %d, %d", rx, ry, ex, ey)

		g, _ := w.Geometry()
		w.moveState = &MoveState{
			rx:        rx,
			ry:        ry,
			startGeom: g,
		}

		w.Focus()
		w.Raise()

		return true, cursors.Fleur
	}
}

func dragMoveStep(w *Window) xgbutil.MouseDragFun {
	return func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
		log.Printf("Drag move step: %d, %d, %d, %d", rx, ry, ex, ey)

		g := w.moveState.startGeom
		x := g.X() + rx - w.moveState.rx
		y := g.Y() + ry - w.moveState.ry

		w.Move(x, y)
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

func dragResizeBegin(w *Window, direction int) xgbutil.MouseDragBeginFun {
	return func(X *xgbutil.XUtil, rx, ry, ex, ey int) (bool, xproto.Cursor) {
		if !w.IsMouseResizable() {
			return false, 0
		}
		log.Printf("Drag resize begin: %d, %d, %d, %d", rx, ry, ex, ey)

		dir := direction
		if dir == ewmh.Infer {
			dir = getDragDirection(w, ex, ey)
		}
		cursor := getCursorForDirection(dir)

		g, _ := w.Geometry()

		w.resizeState = &ResizeState{
			rx:        rx,
			ry:        ry,
			direction: dir,
			startGeom: g,
		}

		w.Focus()
		w.Raise()

		return true, cursor
	}
}

func dragResizeStep(win *Window) xgbutil.MouseDragFun {
	return func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
		log.Printf("Drag resize step: %d, %d, %d, %d", rx, ry, ex, ey)

		d := win.resizeState.direction
		changeX := d == ewmh.SizeLeft || d == ewmh.SizeTopLeft || d == ewmh.SizeBottomLeft
		changeY := d == ewmh.SizeTop || d == ewmh.SizeTopLeft || d == ewmh.SizeTopRight
		changeW := d != ewmh.SizeTop && d != ewmh.SizeBottom
		changeH := d != ewmh.SizeLeft && d != ewmh.SizeRight

		xDiff := rx - win.resizeState.rx
		yDiff := ry - win.resizeState.ry

		g := win.resizeState.startGeom
		x, y, w, h := g.Pieces()
		if changeX {
			x += xDiff
		}
		if changeY {
			y += yDiff
		}
		if changeW {
			if changeX {
				w -= xDiff
			} else {
				w += xDiff
			}
		}
		if changeH {
			if changeY {
				h -= yDiff
			} else {
				h += yDiff
			}
		}

		flags := ConfigAll
		if w < int(win.normalHints.MinWidth) {
			w = int(win.normalHints.MinWidth)
			flags &= ^ConfigX
		}
		if h < int(win.normalHints.MinHeight) {
			h = int(win.normalHints.MinHeight)
			flags &= ^ConfigY
		}
		win.MoveResize(true, x, y, w, h, flags)
	}
}

func dragResizeEnd(w *Window) xgbutil.MouseDragFun {
	return func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
		log.Printf("Drag resize end: %d, %d, %d, %d", rx, ry, ex, ey)
		w.resizeState = nil
	}
}
