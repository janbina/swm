package main

import (
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xinerama"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/janbina/swm/cmd/swm/keysym"
	"log"
	"os/exec"
)

var xc *xgb.Conn
var setupInfo *xproto.SetupInfo
var attachedScreens []xinerama.ScreenInfo
var keymap [256][]xproto.Keysym

var (
	atomWMProtocols    xproto.Atom
	atomWMDeleteWindow xproto.Atom
)

func main() {

	if err := initConnection(); err != nil {
		log.Fatalf("Cannot init connection: %s", err)
	}
	defer xc.Close()

	if err := queryAttachedScreens(); err != nil {
		log.Fatalf("Error querying attached screens: %s", err)
	}

	if err := initializeAtoms(); err != nil {
		log.Fatalf("Error initializing atoms: %s", err)
	}

	if err := takeWMOwnership(); err != nil {
		log.Fatalf("Cannot take WM ownership: %s", err)
	}

	if err := initKeyMap(); err != nil {
		log.Fatalf("Cannot init keymap: %s", err)
	}

	if errs := grabKeys(); len(errs) > 0 {
		log.Println("There were some errors grabbing keys:")
		for _, err := range errs {
			log.Println("\t", err)
		}
	}

	if err := initWorkspaces(); err != nil {
		log.Fatalf("Cannot init workspaces: %s", err)
	}

eventloop:
	for {
		event, err := xc.WaitForEvent()
		if err != nil {
			log.Println(err)
			continue
		}
		if event != nil {
			if err := handleEvent(event); err != nil {
				break eventloop
			}
		}
	}
}

func initConnection() error {
	xConn, err := xgb.NewConn()
	if err != nil {
		return err
	}

	if err := xinerama.Init(xConn); err != nil {
		xConn.Close()
		return err
	}

	connInfo := xproto.Setup(xConn)
	if connInfo == nil {
		xConn.Close()
		return fmt.Errorf("could not parse X connection info")
	} else if len(connInfo.Roots) != 1 {
		xConn.Close()
		return fmt.Errorf("inappropriate number of roots (%d), Xinerama probably didn't initialize correctly", len(connInfo.Roots))
	}

	xc = xConn
	setupInfo = connInfo
	return nil
}

func queryAttachedScreens() error {
	if r, err := xinerama.QueryScreens(xc).Reply(); err != nil {
		return err
	} else {
		if len(r.ScreenInfo) == 0 {
			attachedScreens = []xinerama.ScreenInfo{{
				Width:  setupInfo.Roots[0].WidthInPixels,
				Height: setupInfo.Roots[0].HeightInPixels,
			}}
		} else {
			attachedScreens = r.ScreenInfo
		}
		return nil
	}
}

func initializeAtoms() error {
	if a, err := getAtom("WM_PROTOCOLS"); err != nil {
		return err
	} else {
		atomWMProtocols = a
	}
	if a, err := getAtom("WM_DELETE_WINDOW"); err != nil {
		return err
	} else {
		atomWMDeleteWindow = a
	}
	return nil
}

func getAtom(name string) (xproto.Atom, error) {
	r, err := xproto.InternAtom(xc, false, uint16(len(name)), name).Reply()
	if err != nil {
		return xproto.AtomNone, err
	}
	if r == nil {
		// TODO: should we return error here or not?
		return xproto.AtomNone, nil
	}
	return r.Atom, nil
}

func takeWMOwnership() error {
	err := xproto.ChangeWindowAttributesChecked(
		xc,
		setupInfo.Roots[0].Root,
		xproto.CwEventMask,
		[]uint32{
			xproto.EventMaskKeyPress |
				xproto.EventMaskKeyRelease |
				xproto.EventMaskButtonPress |
				xproto.EventMaskButtonRelease |
				xproto.EventMaskStructureNotify |
				xproto.EventMaskSubstructureRedirect,
		}).Check()
	if _, ok := err.(xproto.AccessError); ok {
		return fmt.Errorf("could not become the WM, another WM is probably already running (%s)", err)
	}
	return err
}

func initKeyMap() error {
	// ASCII codes below 8 don't correspond to any keys on modern keyboards
	const loKey, hiKey = 8, 255

	mapping, err := xproto.GetKeyboardMapping(xc, loKey, hiKey-loKey+1).Reply()
	if err != nil {
		return err
	}
	if mapping == nil {
		return fmt.Errorf("cannot load keyboard map")
	}

	keysymsPerKeycode := int(mapping.KeysymsPerKeycode)

	for i := 0; i < hiKey-loKey+1; i++ {
		keymap[loKey+i] = mapping.Keysyms[i*keysymsPerKeycode : (i+1)*keysymsPerKeycode]
	}

	return nil
}

func grabKeys() []error {
	grabs := []struct {
		sym   xproto.Keysym
		mods  uint16
		codes []xproto.Keycode
	}{
		{sym: keysym.XK_BackSpace, mods: xproto.ModMaskControl | xproto.ModMask1},
		{sym: keysym.XK_e, mods: xproto.ModMask1},
		{sym: keysym.XK_q, mods: xproto.ModMask1},
		{sym: keysym.XK_q, mods: xproto.ModMask1 | xproto.ModMaskShift},
	}

	for i, syms := range keymap {
		for _, sym := range syms {
			for c := range grabs {
				if grabs[c].sym == sym {
					grabs[c].codes = append(grabs[c].codes, xproto.Keycode(i))
				}
			}
		}
	}

	var errs []error
	for _, grabbed := range grabs {
		for _, code := range grabbed.codes {
			if err := xproto.GrabKeyChecked(
				xc,
				false,
				setupInfo.Roots[0].Root,
				grabbed.mods,
				code,
				xproto.GrabModeAsync,
				xproto.GrabModeAsync,
			).Check(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}

func handleEvent(event xgb.Event) error {
	log.Println(event)
	switch e := event.(type) {
	case xproto.KeyPressEvent:
		if err := handleKeyPressEvent(KeyPressEvent(e)); err != nil {
			return err
		}
	case xproto.DestroyNotifyEvent:
		handleDestroyNotifyEvent(e)
	case xproto.ConfigureRequestEvent:
		handleConfigureRequestEvent(e)
	case xproto.MapRequestEvent:
		handleMapRequestEvent(e)
	case xproto.EnterNotifyEvent:
		handleEnterNotifyEvent(e)
	}
	return nil
}

type KeyPressEvent xproto.KeyPressEvent

func (e KeyPressEvent) hasModifiers(mods ...uint16) bool {
	for _, mod := range mods {
		if e.State&mod == 0 {
			return false
		}
	}
	return true
}

func handleKeyPressEvent(e KeyPressEvent) error {
	switch keymap[e.Detail][0] {
	case keysym.XK_BackSpace:
		if e.hasModifiers(xproto.ModMaskControl, xproto.ModMask1) {
			return fmt.Errorf("quit")
		}
	case keysym.XK_e:
		if e.hasModifiers(xproto.ModMask1) {
			cmd := exec.Command("xterm")
			err := cmd.Start()
			go func() {
				cmd.Wait()
			}()
			return err
		}
	case keysym.XK_q:
		if e.hasModifiers(xproto.ModMask1) {
			return destroyActiveWindow(e.hasModifiers(xproto.ModMaskShift))
		}
	}
	return nil
}

func handleDestroyNotifyEvent(e xproto.DestroyNotifyEvent) {
	for _, w := range workspaces {
		go func(w *Workspace) {
			if err := w.RemoveWindow(e.Window); err == nil {
				w.TileWindows()
			}
		}(w)
	}
	if activeWindow != nil && e.Window == *activeWindow {
		activeWindow = nil
	}
}

func handleConfigureRequestEvent(e xproto.ConfigureRequestEvent) {
	ev := xproto.ConfigureNotifyEvent{
		Event:            e.Window,
		Window:           e.Window,
		AboveSibling:     0,
		X:                e.X,
		Y:                e.Y,
		Width:            e.Width,
		Height:           e.Height,
		BorderWidth:      0,
		OverrideRedirect: false,
	}
	xproto.SendEventChecked(xc, false, e.Window, xproto.EventMaskStructureNotify, string(ev.Bytes()))
}

func handleMapRequestEvent(e xproto.MapRequestEvent) {
	w := workspaces["default"]
	xproto.MapWindowChecked(xc, e.Window)
	w.Add(e.Window)
	w.TileWindows()
}

func handleEnterNotifyEvent(e xproto.EnterNotifyEvent) {
	activeWindow = &e.Event
}
