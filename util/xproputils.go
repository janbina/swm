package util

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xprop"
)

func Atoms(X *xgbutil.XUtil, names ...string) ([]xproto.Atom, error) {
	atoms := make([]xproto.Atom, len(names))
	var err error = nil

	for i, name := range names {
		atoms[i], err = xprop.Atm(X, name)
	}

	return atoms, err
}
