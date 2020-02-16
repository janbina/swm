package main

import (
	"fmt"
	"github.com/BurntSushi/xgb/xinerama"
	"github.com/BurntSushi/xgb/xproto"
	"log"
	"sync"
	"time"
)

type ManagedWindow xproto.Window

type Workspace struct {
	Screen  *xinerama.ScreenInfo
	windows []ManagedWindow

	mu *sync.Mutex
}

var workspaces []*Workspace
var activeWindow *xproto.Window

func initWorkspaces() error {
	tree, err := xproto.QueryTree(xc, setupInfo.Roots[0].Root).Reply()
	if err != nil {
		return err
	}
	if tree != nil {
		defaultw := &Workspace{mu: &sync.Mutex{}}
		for _, c := range tree.Children {
			if err := defaultw.Add(c); err != nil {
				log.Println(err)
			}
		}

		if len(attachedScreens) > 0 {
			defaultw.Screen = &attachedScreens[0]
		}

		workspaces = append(workspaces, defaultw)

		if err := defaultw.TileWindows(); err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (w *Workspace) Add(win xproto.Window) error {
	// Ensure that we can manage this window.
	if err := xproto.ConfigureWindowChecked(
		xc,
		win,
		xproto.ConfigWindowBorderWidth,
		[]uint32{
			2,
		}).Check(); err != nil {
		return err
	}

	// Get notifications when this window is deleted.
	if err := xproto.ChangeWindowAttributesChecked(
		xc,
		win,
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

	w.windows = append(w.windows, ManagedWindow(win))

	return nil
}

// TileWindows tiles all the windows of the workspace into the screen that
// the workspace is attached to.
func (w *Workspace) TileWindows() error {
	if w.Screen == nil {
		return fmt.Errorf("workspace not attached to a screen")
	}

	if len(w.windows) == 0 {
		return nil
	}
	width := uint32(w.Screen.Width) / uint32(len(w.windows))
	height := uint32(w.Screen.Height)
	var err error
	for i, window := range w.windows {
		err2 := xproto.ConfigureWindowChecked(
			xc,
			xproto.Window(window),
			xproto.ConfigWindowX|
				xproto.ConfigWindowY|
				xproto.ConfigWindowWidth|
				xproto.ConfigWindowHeight,
			[]uint32{
				uint32(i) * width,
				0,
				width,
				height,
			},
		).Check()
		if err == nil && err2 != nil {
			err = err2
		}
	}
	return err
}

// RemoveWindow removes a window from the workspace. It returns
// an error if the window is not being managed by w.
func (w *Workspace) RemoveWindow(win xproto.Window) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i, window := range w.windows {
		if win == xproto.Window(window) {
			w.windows = append(w.windows[0:i], w.windows[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("Window not managed by workspace")
}

func destroyActiveWindow(aggressive bool) error {
	if aggressive {
		if activeWindow != nil {
			return xproto.DestroyWindowChecked(xc, *activeWindow).Check()
		}
		return nil
	} else {
		prop, err := xproto.GetProperty(
			xc,
			false,
			*activeWindow,
			atomWMProtocols,
			xproto.GetPropertyTypeAny,
			0,
			64,
		).Reply()
		if err != nil {
			return err
		}
		if prop == nil {
			// There were no properties, so the window doesn't follow ICCCM.
			// Just destroy it.
			return destroyActiveWindow(true)
		}
		for v := prop.Value; len(v) >= 4; v = v[4:] {
			switch xproto.Atom(uint32(v[0]) | uint32(v[1])<<8 | uint32(v[2])<<16 | uint32(v[3])<<24) {
			case atomWMDeleteWindow:
				t := time.Now().Unix()
				return xproto.SendEventChecked(
					xc,
					false,
					*activeWindow,
					xproto.EventMaskNoEvent,
					string(xproto.ClientMessageEvent{
						Format: 32,
						Window: *activeWindow,
						Type:   atomWMProtocols,
						Data: xproto.ClientMessageDataUnionData32New([]uint32{
							uint32(atomWMDeleteWindow),
							uint32(t),
							0,
							0,
							0,
						}),
					}.Bytes()),
				).Check()
			}
		}
		// No WM_DELETE_WINDOW protocol, so destroy.
		return destroyActiveWindow(true)
	}
}
