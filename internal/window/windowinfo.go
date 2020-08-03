package window

import (
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/motif"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/janbina/swm/internal/util"
)

type WinInfo struct {
	Attributes  *xproto.GetWindowAttributesReply
	Hints       *icccm.Hints
	NormalHints *icccm.NormalHints
	MotifHints  *motif.Hints
	Protocols   util.StringSet
	States      util.StringSet
	Types       util.StringSet
	Name        string
	Class       *icccm.WmClass
	Desktop     *uint
	Geometry    xrect.Rect
}

func GetWindowInfo(win *xwindow.Window) *WinInfo {
	x := win.X
	id := win.Id

	return &WinInfo{
		Attributes:  getAttributes(x, id),
		Hints:       getHints(x, id),
		NormalHints: getNormalHints(x, id),
		MotifHints:  getMotifHints(x, id),
		Protocols:   getProtocols(x, id),
		States:      getStates(x, id),
		Types:       getTypes(x, id),
		Name:        getName(x, id),
		Class:       getClass(x, id),
		Desktop:     getDesktop(x, id),
		Geometry:    getDefaultGeometry(win),
	}
}

func getAttributes(x *xgbutil.XUtil, id xproto.Window) *xproto.GetWindowAttributesReply {
	attrs, err := xproto.GetWindowAttributes(x.Conn(), id).Reply()
	if err != nil {
		log.Printf("Error getting window attributes: %s", err)
	}
	return attrs
}

func getHints(x *xgbutil.XUtil, id xproto.Window) *icccm.Hints {
	hints, err := icccm.WmHintsGet(x, id)
	if err != nil {
		log.Printf("Error getting wm hints: %s", err)
		hints = &icccm.Hints{
			Flags:        icccm.HintInput | icccm.HintState,
			Input:        1,
			InitialState: icccm.StateNormal,
		}
	}
	return hints
}

func getProtocols(x *xgbutil.XUtil, id xproto.Window) util.StringSet {
	protocols := make(util.StringSet)
	p, err := icccm.WmProtocolsGet(x, id)
	if err != nil {
		log.Printf("Error getting wm protocols: %s", err)
	} else {
		protocols.SetAll(p)
	}
	return protocols
}

func getNormalHints(x *xgbutil.XUtil, id xproto.Window) *icccm.NormalHints {
	normalHints, err := icccm.WmNormalHintsGet(x, id)
	if err != nil {
		log.Printf("Error getting wm normal hints: %s", err)
		normalHints = &icccm.NormalHints{}
	}
	return normalHints
}

func getMotifHints(x *xgbutil.XUtil, id xproto.Window) *motif.Hints {
	motifHints, err := motif.WmHintsGet(x, id)
	if err != nil {
		log.Printf("Error getting motif hints: %s", err)
		motifHints = &motif.Hints{}
	}
	return motifHints
}

func getStates(x *xgbutil.XUtil, id xproto.Window) util.StringSet {
	states := make(util.StringSet)
	s, err := ewmh.WmStateGet(x, id)
	if err != nil {
		log.Printf("Error getting wm state: %s", err)
	}
	states.SetAll(s)
	return states
}

func getClass(x *xgbutil.XUtil, id xproto.Window) *icccm.WmClass {
	class, err := icccm.WmClassGet(x, id)
	if err != nil {
		log.Printf("Error getting wm class: %s", err)
		class = &icccm.WmClass{
			Instance: "",
			Class:    "",
		}
	}
	return class
}

func getTypes(x *xgbutil.XUtil, id xproto.Window) util.StringSet {
	typesSet := make(util.StringSet)
	types, err := ewmh.WmWindowTypeGet(x, id)
	if err != nil {
		log.Printf("Error getting window types: %s", err)
		typesSet["_NET_WM_WINDOW_TYPE_NORMAL"] = true
	} else {
		typesSet.SetAll(types)
	}
	return typesSet
}

func getName(x *xgbutil.XUtil, id xproto.Window) string {
	name, _ := ewmh.WmNameGet(x, id)
	if len(name) > 0 {
		return name
	}

	name, _ = icccm.WmNameGet(x, id)
	if len(name) > 0 {
		return name
	}

	return ""
}

func getDesktop(x *xgbutil.XUtil, id xproto.Window) *uint {
	d, err := ewmh.WmDesktopGet(x, id)
	if err != nil {
		return nil
	}
	return &d
}

func getDefaultGeometry(win *xwindow.Window) xrect.Rect {
	g, err := win.Geometry()
	if err != nil {
		return xrect.New(0, 0, 1, 1)
	}
	return g
}
