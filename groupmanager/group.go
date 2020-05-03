package groupmanager

import (
	"github.com/BurntSushi/xgb/xproto"
	"time"
)

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

func (g *group) makeVisible() {
	g.shownTimestamp = time.Now().UnixNano()
}

func (g *group) makeInvisible() {
	g.shownTimestamp = 0
}
