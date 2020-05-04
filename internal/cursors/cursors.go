package cursors

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xcursor"
	"log"
)

var (
	LeftPtr           xproto.Cursor
	Fleur             xproto.Cursor
	TopSide           xproto.Cursor
	TopRightCorner    xproto.Cursor
	RightSide         xproto.Cursor
	BottomRightCorner xproto.Cursor
	BottomSide        xproto.Cursor
	BottomLeftCorner  xproto.Cursor
	LeftSide          xproto.Cursor
	TopLeftCorner     xproto.Cursor
)

func Initialize(X *xgbutil.XUtil) {
	LeftPtr = initCursor(X, xcursor.LeftPtr)
	Fleur = initCursor(X, xcursor.Fleur)
	TopSide = initCursor(X, xcursor.TopSide)
	TopRightCorner = initCursor(X, xcursor.TopRightCorner)
	RightSide = initCursor(X, xcursor.RightSide)
	BottomRightCorner = initCursor(X, xcursor.BottomRightCorner)
	BottomSide = initCursor(X, xcursor.BottomSide)
	BottomLeftCorner = initCursor(X, xcursor.BottomLeftCorner)
	LeftSide = initCursor(X, xcursor.LeftSide)
	TopLeftCorner = initCursor(X, xcursor.TopLeftCorner)
}

func initCursor(X *xgbutil.XUtil, cursor uint16) xproto.Cursor {
	cid, err := xcursor.CreateCursor(X, cursor)
	if err != nil {
		log.Printf("Cannot load cursor %d", cursor)
		return 0
	}
	return cid
}
