package main

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xinerama"
	"github.com/BurntSushi/xgb/xproto"
	"log"
)

var keymap [256][]xproto.Keysym

func main() {

	xConn, xRootWindow, err := initConnection()
	if err != nil {
		log.Fatalf("Cannot init connection: %s", err)
	}
	defer xConn.Close()

	if err := takeWMOwnership(xConn, xRootWindow.Root); err != nil {
		log.Fatalf("Cannot take WM ownership: %s", err)
	}

	if err := initKeyMap(xConn); err != nil {
		log.Fatalf("Cannot get keymap: %s", err)
	}

eventloop:
	for {
		event, err := xConn.WaitForEvent()
		if err != nil {
			log.Println(err)
		}
		if event != nil {
			if err := handleEvent(event); err != nil {
				break eventloop
			}
		}
	}
}

func initConnection() (*xgb.Conn, *xproto.ScreenInfo, error) {
	xConn, err := xgb.NewConn()
	if err != nil {
		return nil, nil, err
	}

	if err := xinerama.Init(xConn); err != nil {
		xConn.Close()
		return nil, nil, err
	}

	connInfo := xproto.Setup(xConn)
	if connInfo == nil {
		xConn.Close()
		return nil, nil, errors.New("could not parse X connection info")
	} else if len(connInfo.Roots) != 1 {
		xConn.Close()
		return nil, nil, fmt.Errorf("inappropriate number of roots (%d), Xinerama probably didn't initialize correctly", len(connInfo.Roots))
	}

	return xConn, &connInfo.Roots[0], nil
}

func takeWMOwnership(xConn *xgb.Conn, xRootWindow xproto.Window) error {
	err := xproto.ChangeWindowAttributesChecked(
		xConn,
		xRootWindow,
		xproto.CwEventMask,
		[]uint32{
			xproto.EventMaskKeyPress |
				xproto.EventMaskKeyRelease |
				xproto.EventMaskButtonPress |
				xproto.EventMaskButtonRelease,
		}).Check()
	if _, ok := err.(xproto.AccessError); ok {
		return fmt.Errorf("could not become the WM, another WM is probably already running")
	}
	return err
}

func initKeyMap(xConn *xgb.Conn) error {
	// ASCII codes below 8 don't correspond to any keys on modern keyboards
	const loKey, hiKey = 8, 255

	mapping, err := xproto.GetKeyboardMapping(xConn, loKey, hiKey-loKey+1).Reply()
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

func handleEvent(event xgb.Event) error {
	log.Println(event)
	switch e := event.(type) {
	case xproto.KeyPressEvent:
		if err := handleKeyPressEvent(e); err != nil {
			return err
		}
	}
	return nil
}

const XK_BackSpace = 0xff08

func handleKeyPressEvent(key xproto.KeyPressEvent) error {
	switch keymap[key.Detail][0] {
	case XK_BackSpace:
		if (key.State&xproto.ModMaskControl != 0) && (key.State&xproto.ModMask1 != 0) {
			return fmt.Errorf("quit signal")
		}
		return nil
	default:
		return nil
	}
}
