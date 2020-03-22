package window

import (
	"log"
)

type handlerFunc func(win *Window, data []uint32)

var handlers = map[string]handlerFunc{
	"_NET_WM_MOVERESIZE": handleMoveResizeMessage,
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
	if dir == 8 { //_NET_WM_MOVERESIZE_MOVE - movement only
		win.DragBegin(int16(xr), int16(yr))
	}
}
