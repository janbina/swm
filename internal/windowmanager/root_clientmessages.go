package windowmanager

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/janbina/swm/internal/log"
)

var rootCmHandlers = map[string]func(data []uint32){
	"_NET_NUMBER_OF_DESKTOPS": handleNumberOfDesktops,
	"_NET_CURRENT_DESKTOP":    handleCurrentDesktop,
}

func handleRootClientMessage(X *xgbutil.XUtil, e xevent.ClientMessageEvent) {
	name, err := xprop.AtomName(X, e.Type)
	if err != nil {
		log.Warn("Error getting atom name for client message %s: %s", e, err)
		return
	}
	log.Debug("Handle root client message: %s (%s)", name, e)
	if f, ok := rootCmHandlers[name]; !ok {
		log.Info("Unsupported root client message: %s", name)
	} else {
		f(e.Data.Data32)
	}
}

func handleNumberOfDesktops(data []uint32) {
	setNumberOfDesktops(int(data[0]))
}

func handleCurrentDesktop(data []uint32) {
	switchToDesktop(int(data[0]))
}
