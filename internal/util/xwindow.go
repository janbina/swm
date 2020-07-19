package util

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/janbina/swm/internal/heads"
)

func SetBorderWidth(win *xwindow.Window, width uint32) error {
	return xproto.ConfigureWindowChecked(
		win.X.Conn(),
		win.Id,
		xproto.ConfigWindowBorderWidth,
		[]uint32{width},
	).Check()
}

// Like xwindow.Create(), but uses max allowed depth, so it can be transparent
func CreateTransparentWindow(xu *xgbutil.XUtil, parent xproto.Window) (*xwindow.Window, error) {
	win, err := xwindow.Generate(xu)
	if err != nil {
		return nil, err
	}

	screen := heads.Screen()

	err = xproto.CreateWindowChecked(
		xu.Conn(),
		screen.Depth,
		win.Id,
		parent,
		0, 0, 1, 1, 0,
		xproto.WindowClassInputOutput,
		screen.Visual,
		xproto.CwBackPixel|xproto.CwBorderPixel|xproto.CwColormap,
		[]uint32{0, 0, uint32(screen.Colormap)},
	).Check()

	if err != nil {
		return nil, err
	}

	return win, nil
}
