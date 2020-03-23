package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xprop"
	"log"
)

type handlerFunc func(win *Window, data []uint32)

var handlers = map[string]handlerFunc{
	"_NET_WM_MOVERESIZE": handleMoveResizeMessage,
	"_NET_WM_STATE":      handleWmStateMessage,
}

func (w *Window) HandleClientMessage(name string, data []uint32) {
	if f, ok := handlers[name]; !ok {
		log.Printf("Unsupported client message: %s", name)
	} else {
		f(w, data)
	}
}

func handleMoveResizeMessage(win *Window, data []uint32) {
	xr := data[0]
	yr := data[1]
	dir := data[2]
	log.Printf("Move resize client message: %d, %d, %d", xr, yr, dir)
	if dir <= ewmh.SizeLeft {
		win.DragResizeBegin(int16(xr), int16(yr), int(dir))
	} else if dir == ewmh.Move {
		win.DragMoveBegin(int16(xr), int16(yr))
	} else {
		log.Printf("Unsupported direction: %d", dir)
	}
}

func handleWmStateMessage(win *Window, data []uint32) {
	action := data[0]
	p1, _ := xprop.AtomName(win.win.X, xproto.Atom(data[1]))
	p2, _ := xprop.AtomName(win.win.X, xproto.Atom(data[2]))
	log.Printf("Wm state client message: %d, %s, %s", action, p1, p2)
	win.UpdateState(int(action), p1)
	if len(p2) > 0 {
		win.UpdateState(int(action), p2)
	}
}