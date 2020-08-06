package groupmanager

import (
	"time"

	"github.com/BurntSushi/xgb/xproto"
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

func (g *group) addWindow(window xproto.Window) {
	g.windows[window] = true
}

func (g *group) removeWindow(window xproto.Window) {
	delete(g.windows, window)
}

func (g *group) removeAllWindows() {
	g.windows = map[xproto.Window]bool{}
}

func (g *group) getWindows() []xproto.Window {
	wins := make([]xproto.Window, 0, len(g.windows))
	for win := range g.windows {
		wins = append(wins, win)
	}
	return wins
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
