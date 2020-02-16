package main

import (
	"fmt"
	"github.com/BurntSushi/xgb/xinerama"
	"github.com/BurntSushi/xgb/xproto"
	"sync"
)

type ManagedWindow xproto.Window

type Column []ManagedWindow

type Workspace struct{
	Screen *xinerama.ScreenInfo
	columns []Column

	mu *sync.Mutex
}

var workspaces map[string]*Workspace
var activeWindow *xproto.Window

func (w *Workspace) Add(win xproto.Window) error {
	// Ensure that we can manage this window.
	if err := xproto.ConfigureWindowChecked(
		xc,
		xproto.Window(win),
		xproto.ConfigWindowBorderWidth,
		[]uint32{
			2,
		}).Check(); err != nil {
		return err
	}

	// Get notifications when this window is deleted.
	if err := xproto.ChangeWindowAttributesChecked(
		xc,
		xproto.Window(win),
		xproto.CwEventMask,
		[]uint32{
			xproto.EventMaskStructureNotify |
				xproto.EventMaskEnterWindow,
		},
	).Check(); err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	switch len(w.columns) {
	case 0:
		w.columns = []Column{
			{ ManagedWindow(win) },
		}
	case 1:
		if len(w.columns[0]) == 0 {
			// No active window in first column, so use it.
			w.columns[0] = append(w.columns[0], ManagedWindow(win))
		} else {
			// There's something in the primary column, so create a new one.
			w.columns = append(w.columns, Column{ ManagedWindow(win) })
		}
	default:
		// Add to the first empty column we can find, and shortcircuit out
		// if applicable.
		for i, c := range w.columns {
			if len(c) == 0 {
				w.columns[i] = append(w.columns[i], ManagedWindow(win))
				return nil
			}
		}

		// No empty columns, add to the last one.
		i := len(w.columns)-1
		w.columns[i] = append(w.columns[i], ManagedWindow(win))
	}
	return nil
}

// TileWindows tiles all the windows of the workspace into the screen that
// the workspace is attached to.
func (w *Workspace) TileWindows() error {
	if w.Screen == nil {
		return fmt.Errorf("Workspace not attached to a screen.")
	}

	n := uint32(len(w.columns))
	if n == 0 {
		return nil
	}
	size := uint32(w.Screen.Width) / n
	var err error
	for i, c := range w.columns {
		if err != nil {
			// Don't overwrite err if there's an error, but still
			// tile the rest of the columns instead of returning.
			c.TileColumn(uint32(i)*size, size, uint32(w.Screen.Height))
		} else {
			err = c.TileColumn(uint32(i)*size, size, uint32(w.Screen.Height))
		}
	}
	return err
}

// TileColumn sends ConfigureWindow messages to tile the ManagedWindows
// Using the geometry of the parameters passed
func (c Column) TileColumn(xstart, colwidth, colheight uint32) error {
	n := uint32(len(c))
	if n == 0 {
		return nil
	}

	height := colheight / n
	var err error
	for i, win := range c {
		if werr := xproto.ConfigureWindowChecked(
			xc,
			xproto.Window(win),
			xproto.ConfigWindowX|
				xproto.ConfigWindowY|
				xproto.ConfigWindowWidth|
				xproto.ConfigWindowHeight,
			[]uint32{
				xstart,
				uint32(i) * height,
				colwidth,
				height,
			}).Check(); werr != nil {
			err = werr
		}
	}
	return err
}

// RemoveWindow removes a window from the workspace. It returns
// an error if the window is not being managed by w.
func (wp *Workspace) RemoveWindow(w xproto.Window) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	for colnum, column := range wp.columns {
		idx := -1
		for i, candwin := range column {
			if w == xproto.Window(candwin) {
				idx = i
				break
			}
		}
		if idx != -1 {
			// Found the window at at idx, so delete it and return.
			// (I wish Go made it easier to delete from a slice.)
			wp.columns[colnum] = append(column[0:idx], column[idx+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Window not managed by workspace")
}
