package groupmanager

import "github.com/BurntSushi/xgb/xproto"

type group struct {
	name           string
	shownTimestamp int64
	windows        map[xproto.Window]bool
}

func createGroup(name string) *group {
	return &group{
		name:           name,
		shownTimestamp: 0,
		windows:        map[xproto.Window]bool{},
	}
}

func (g *group) isVisible() bool {
	return g.shownTimestamp > 0
}
