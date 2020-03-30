package window

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"log"
)

type handlerFunc func(win *Window, data []uint32)

var handlers = map[string]handlerFunc{
	"_NET_WM_MOVERESIZE": handleMoveResizeMessage,
	"_NET_WM_STATE":      handleWmStateMessage,
	"_NET_ACTIVE_WINDOW": handleActiveWindowMessage,
	"WM_CHANGE_STATE":    handleWmChangeStateMessage,
}

func (w *Window) HandleClientMessage(e xevent.ClientMessageEvent) {
	name, err := xprop.AtomName(w.win.X, e.Type)
	if err != nil {
		log.Printf("Cannot get property atom name for clientMessage event: %s", err)
		return
	}
	log.Printf("Client message %s: %s", name, e)
	if f, ok := handlers[name]; !ok {
		log.Printf("Unsupported client message: %s", name)
	} else {
		f(w, e.Data.Data32)
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

	win.UpdateStates(int(action), p1, p2)
}

func handleWmChangeStateMessage(win *Window, data []uint32) {
	if data[0] == icccm.StateIconic && !win.iconified {
		win.IconifyToggle()
	}
}

func handleActiveWindowMessage(win *Window, _ []uint32) {
	win.Focus()
}
