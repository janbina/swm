package util

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xwindow"
)

func SetBorder(win *xwindow.Window, width uint32, color uint32) error {
	if err := SetBorderWidth(win, width); err != nil {
		return err
	}
	if err := SetBorderColor(win, color); err != nil {
		return err
	}
	return nil
}

func SetBorderWidth(win *xwindow.Window, width uint32) error {
	return xproto.ConfigureWindowChecked(
		win.X.Conn(),
		win.Id,
		xproto.ConfigWindowBorderWidth,
		[]uint32{width},
	).Check()
}

func SetBorderColor(win *xwindow.Window, color uint32) error {
	return xproto.ChangeWindowAttributesChecked(
		win.X.Conn(),
		win.Id,
		xproto.CwBorderPixel,
		[]uint32{color},
	).Check()
}

func GetBorderWidth(win *xwindow.Window) uint16 {
	g, err := xproto.GetGeometry(win.X.Conn(), xproto.Drawable(win.Id)).Reply()
	if err != nil {
		return 0
	}
	return g.BorderWidth
}
