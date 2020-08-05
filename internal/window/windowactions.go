package window

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/motif"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/janbina/swm/internal/groupmanager"
	"github.com/janbina/swm/internal/heads"
	"github.com/janbina/swm/internal/stack"
	"github.com/janbina/swm/internal/util"
)

type WinActions struct {
	ShouldManage   bool
	ShouldDecorate bool
	IsFocusable    bool
	StartIconified bool
	Groups         []int
	StackLayer     int
	Geometry       xrect.Rect
}

func GetWinActions(x *xgbutil.XUtil, info *WinInfo) *WinActions {
	return &WinActions{
		ShouldManage:   shouldManage(info),
		ShouldDecorate: shouldDecorate(info),
		IsFocusable:    isFocusable(info),
		StartIconified: startIconified(info),
		Groups:         getInitialGroups(info),
		StackLayer:     getStackLayer(info),
		Geometry:       getGeometry(x, info),
	}
}

func shouldManage(info *WinInfo) bool {
	if info.Attributes != nil && info.Attributes.OverrideRedirect {
		return false
	}
	return true
}

func shouldDecorate(info *WinInfo) bool {
	if info.Types.Any("_NET_WM_WINDOW_TYPE_DESKTOP", "_NET_WM_WINDOW_TYPE_DOCK", "_NET_WM_WINDOW_TYPE_SPLASH") {
		return false
	}
	return motif.Decor(info.MotifHints)
}

func isFocusable(info *WinInfo) bool {
	if info.Types.Any("_NET_WM_WINDOW_TYPE_DESKTOP", "_NET_WM_WINDOW_TYPE_DOCK") {
		return false
	}
	return true
}

func startIconified(info *WinInfo) bool {
	return info.NormalHints.Flags&icccm.HintState > 0 && info.Hints.InitialState == icccm.StateIconic
}

func getInitialGroups(info *WinInfo) []int {
	return []int{getInitialGroup(info)}
}

func getInitialGroup(info *WinInfo) int {
	if groupmanager.GroupMode == groupmanager.ModeSticky {
		return groupmanager.StickyGroupID
	}

	if info.Desktop != nil {
		return int(*info.Desktop)
	}

	return groupmanager.GetCurrentGroup()
}

func getStackLayer(info *WinInfo) int {
	if info.Types["_NET_WM_WINDOW_TYPE_DESKTOP"] {
		return stack.LayerDesktop
	} else if info.Types["_NET_WM_WINDOW_TYPE_DOCK"] {
		return stack.LayerDock
	}
	return stack.LayerDefault
}

func getGeometry(x *xgbutil.XUtil, info *WinInfo) xrect.Rect {
	g := xrect.New(info.Geometry.Pieces())

	if info.NormalHints.Flags&icccm.SizeHintUSPosition == 0 &&
		info.NormalHints.Flags&icccm.SizeHintPPosition == 0 {
		if pointer, err := util.QueryPointer(x); err == nil {
			if head, err := heads.GetHeadForPointerStruts(pointer.X, pointer.Y); err == nil {
				xGap := head.Width() - g.Width()
				yGap := head.Height() - g.Height()
				g.XSet(head.X() + xGap/2)
				g.YSet(head.Y() + yGap/2)
			}
		}
	}

	return g
}
