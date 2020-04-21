package util

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"
)

func SetBorderWidth(win *xwindow.Window, width uint32) error {
	return xproto.ConfigureWindowChecked(
		win.X.Conn(),
		win.Id,
		xproto.ConfigWindowBorderWidth,
		[]uint32{width},
	).Check()
}
