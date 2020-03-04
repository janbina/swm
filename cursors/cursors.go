package cursors

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xcursor"
	"log"
)

var (
	LeftPtr xproto.Cursor
	Fleur   xproto.Cursor
)

func Initialize(X *xgbutil.XUtil) {
	LeftPtr = initCursor(X, xcursor.LeftPtr)
	Fleur = initCursor(X, xcursor.Fleur)
}

func initCursor(X *xgbutil.XUtil, cursor uint16) xproto.Cursor {
	cid, err := xcursor.CreateCursor(X, cursor)
	if err != nil {
		log.Printf("Cannot load cursor %d", cursor)
		return 0
	}
	return cid
}
