package util

import (
	"fmt"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/icccm"
)

type QueryPointerResponse struct {
	X, Y int
	Win xproto.Window
	WinX, WinY int
}

func QueryPointer(x *xgbutil.XUtil) (*QueryPointerResponse, error) {
	r, err := xproto.QueryPointer(x.Conn(), x.RootWin()).Reply()
	if err != nil {
		return nil, err
	}
	return &QueryPointerResponse{
		int(r.RootX), int(r.RootY),
		r.Child, int(r.WinX), int(r.WinY),
	}, nil
}

func QueryPointerClient(x *xgbutil.XUtil) (*QueryPointerResponse, error) {
	r, err := xproto.QueryPointer(x.Conn(), x.RootWin()).Reply()
	if err != nil {
		return nil, err
	}
	client, err := getClientChild(x, r.Child)
	if err != nil {
		return nil, err
	}
	r, err = xproto.QueryPointer(x.Conn(), client).Reply()
	if err != nil {
		return nil, err
	}
	return &QueryPointerResponse{
		int(r.RootX), int(r.RootY),
		client, int(r.WinX), int(r.WinY),
	}, nil
}

func getClientChild(x *xgbutil.XUtil, win xproto.Window) (xproto.Window, error) {
	if isClientWindow(x, win) {
		return win, nil
	}
	t, err := xproto.QueryTree(x.Conn(), win).Reply()
	if err != nil {
		return 0, err
	}
	for _, child := range t.Children {
		if isClientWindow(x, child) {
			return child, nil
		}
		if c, err := getClientChild(x, child); err != nil {
			return c, nil
		}
	}
	return 0, fmt.Errorf("cannot find client window")
}

func isClientWindow(x *xgbutil.XUtil, win xproto.Window) bool {
	s, e := icccm.WmStateGet(x, win)
	return s != nil && e == nil
}
