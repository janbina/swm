package windowmanager

import (
	"fmt"
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/janbina/swm/internal/log"
	"github.com/janbina/swm/internal/util"
)

func takeWmOwnership(X *xgbutil.XUtil, replace bool) error {

	xTime, err := currentTime(X)
	if err != nil {
		return err
	}

	selectionAtom, err := xprop.Atm(X, getSelectionAtomName(X))
	if err != nil {
		return err
	}

	// check if another wm is running
	otherWm := xproto.Window(xproto.WindowNone)
	if reply, err := xproto.GetSelectionOwner(X.Conn(), selectionAtom).Reply(); err != nil {
		return err
	} else if reply.Owner != xproto.WindowNone {
		if !replace {
			return fmt.Errorf("another window manager is already running, use '--replace' flag")
		}

		otherWm = reply.Owner

		// listen to DestroyNotify events
		if err = xwindow.New(X, reply.Owner).Listen(xproto.EventMaskStructureNotify); err != nil {
			return err
		}

		log.Info("Waiting for other wm to shutdown and transfer ownership to us.")
	} else {
		if err2 := xwindow.New(X, X.RootWin()).Listen(xproto.EventMaskSubstructureRedirect); err2 != nil {
			// cannot listen to substructure redirect - there must be another wm, but it doesnt support icccm selection

			// try ewmh
			win, _ := ewmh.SupportingWmCheckGet(X, X.RootWin())
			if win != xproto.WindowNone {
				if !replace {
					return fmt.Errorf("another window manager is already running, use '--replace' flag")
				} else {
					xwindow.New(X, win).Kill()
				}
			} else {
				return fmt.Errorf("another window manager is already running and it doesn't support icccm or ewmh, so we cannot destroy it")
			}
		}
	}

	log.Info("Setting selection owner.")
	if err := xproto.SetSelectionOwnerChecked(X.Conn(), X.Dummy(), selectionAtom, xTime).Check(); err != nil {
		return err
	}

	// check we actually got ownership
	log.Info("Getting selection owner.")
	if reply, err := xproto.GetSelectionOwner(X.Conn(), selectionAtom).Reply(); err != nil {
		return err
	} else if reply.Owner != X.Dummy() {
		return fmt.Errorf("cannot get ownership, owner is '%d', we need it to be '%d'", reply.Owner, X.Dummy())
	}

	// wait for other wm to shut down - ICCCM 2.8
	if otherWm != xproto.Window(xproto.WindowNone) {
		if err := waitForWmShutdown(X, otherWm); err != nil {
			log.Info("Other wm didnt destroy its window in reasonable time, will kill it.")
			xwindow.New(X, otherWm).Kill()
		}
	}

	if err := announce(X); err != nil {
		return err
	}

	// listen for SelectionClear events - another wm wants to take control
	xevent.SelectionClearFun(disown).Connect(X, X.Dummy())

	log.Info("Swm is now your wm.")

	return nil
}

func waitForWmShutdown(X *xgbutil.XUtil, otherWm xproto.Window) error {
	timeout := 3 * time.Second
	delay := 100 * time.Millisecond

	for t := time.Duration(0); t <= timeout; t += delay {
		log.Info("Polling for event...")
		if ev, err := X.Conn().PollForEvent(); ev != nil && err == nil {
			log.Info("Got event: %s", ev)
			if destNotify, ok := ev.(xproto.DestroyNotifyEvent); ok {
				if destNotify.Window == otherWm {
					return nil
				}
			}
		} else if err != nil {
			log.Info("Got error: %s", err)
		}
		time.Sleep(delay)
	}

	return fmt.Errorf("timeout waiting for other wm to shut down")
}

// letting others know that we are taking control now - ICCCM 2.8
func announce(X *xgbutil.XUtil) error {
	atoms, err := util.Atoms(X, "MANAGER", getSelectionAtomName(X))
	if err != nil {
		return err
	}
	if clientMessage, err := xevent.NewClientMessage(
		32, X.RootWin(), atoms[0], int(X.TimeGet()), int(atoms[1]), int(X.Dummy()),
	); err != nil {
		return err
	} else if clientMessage == nil {
		return fmt.Errorf("client message was nil")
	} else {
		xproto.SendEvent(
			X.Conn(), false, X.RootWin(), xproto.EventMaskStructureNotify, string(clientMessage.Bytes()),
		)
		return nil
	}
}

func disown(X *xgbutil.XUtil, _ xevent.SelectionClearEvent) {
	log.Warn("Exiting, will be replaced by another wm.")
	xevent.Quit(X)
}

func currentTime(X *xgbutil.XUtil) (xproto.Timestamp, error) {
	if err := xwindow.New(X, X.RootWin()).Listen(xproto.EventMaskPropertyChange); err != nil {
		return 0, err
	}

	atoms, err := util.Atoms(X, "WM_CLASS", "STRING")
	if err != nil {
		return 0, err
	}

	// ICCCM 2.1
	if err := xproto.ChangePropertyChecked(
		X.Conn(), xproto.PropModeAppend, X.RootWin(), atoms[0], atoms[1], 8, 0, nil,
	).Check(); err != nil {
		return 0, err
	}

	timeout := 3 * time.Second
	delay := 100 * time.Millisecond
	for t := time.Duration(0); t <= timeout; t += delay {
		if event, err := X.Conn().PollForEvent(); err == nil {
			if propertyNotify, ok := event.(xproto.PropertyNotifyEvent); ok {
				X.TimeSet(propertyNotify.Time)
				return propertyNotify.Time, nil
			}
		}
		time.Sleep(delay)
	}
	return 0, fmt.Errorf("cannot get valid timestamp")
}

func getSelectionAtomName(X *xgbutil.XUtil) string {
	return fmt.Sprintf("WM_S%d", X.Conn().DefaultScreen)
}
