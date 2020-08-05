package window

import (
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/janbina/swm/internal/log"
)

var propertyHandlers = map[string]func(win *Window){
	"WM_NORMAL_HINTS": handleNormalHints,
}

func (w *Window) HandlePropertyNotify(e xevent.PropertyNotifyEvent) {
	name, err := xprop.AtomName(w.win.X, e.Atom)
	if err != nil {
		log.Infof("Cannot get property atom name for propertyNotify event: %s", err)
		return
	}
	log.Infof("Property notify event %s: %s", name, e)
	if f, ok := propertyHandlers[name]; !ok {
		log.Infof("Unsupported client message: %s", name)
	} else {
		f(w)
	}
}

func handleNormalHints(w *Window) {
	if h, err := icccm.WmNormalHintsGet(w.win.X, w.win.Id); err == nil {
		w.info.NormalHints = h
	}
}
