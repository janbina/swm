package heads

import (
	"fmt"
	"log"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xrect"
)

type ScreenInfo struct {
	Depth    byte
	Visual   xproto.Visualid
	Colormap xproto.Colormap
}

var screen *ScreenInfo
var Heads xinerama.Heads
var HeadsStruts xinerama.Heads

// Find max allowed depth, its visual and colormap
func InitScreen(X *xgbutil.XUtil) {
	depth := X.Screen().RootDepth
	visual := X.Screen().RootVisual
	colormap := X.Screen().DefaultColormap

	for _, d := range X.Screen().AllowedDepths {
		if d.Depth > depth {
			if d.VisualsLen > 0 {
				depth = d.Depth
				visual = d.Visuals[0].VisualId
			}
		}
	}

	if visual != X.Screen().RootVisual {
		colormap, _ = xproto.NewColormapId(X.Conn())
		err := xproto.CreateColormapChecked(
			X.Conn(),
			xproto.ColormapAllocNone,
			colormap,
			X.RootWin(),
			visual,
		).Check()

		if err != nil {
			log.Printf("Error creating colormap for visual %d: %s", visual, err)

			screen = &ScreenInfo{
				Depth:    X.Screen().RootDepth,
				Visual:   X.Screen().RootVisual,
				Colormap: X.Screen().DefaultColormap,
			}
			return
		}
	}

	screen = &ScreenInfo{
		Depth:    depth,
		Visual:   visual,
		Colormap: colormap,
	}
}

func Screen() *ScreenInfo {
	return screen
}

func GetHeadForRect(rect xrect.Rect) (xrect.Rect, error) {
	if len(Heads) == 0 {
		return nil, fmt.Errorf("no heads")
	}
	if i := xrect.LargestOverlap(rect, Heads); i < 0 {
		return Heads[0], nil
	} else {
		return Heads[i], nil
	}
}

func GetHeadForRectStruts(rect xrect.Rect) (xrect.Rect, error) {
	if len(HeadsStruts) == 0 {
		return nil, fmt.Errorf("no heads")
	}
	if i := xrect.LargestOverlap(rect, HeadsStruts); i < 0 {
		return Heads[0], nil
	} else {
		return HeadsStruts[i], nil
	}
}

func GetHeadForPointerStruts(x, y int) (xrect.Rect, error) {
	if len(HeadsStruts) == 0 {
		return nil, fmt.Errorf("no heads")
	}
	for _, head := range HeadsStruts {
		if xInRect(x, head) && yInRect(y, head) {
			return head, nil
		}
	}
	return HeadsStruts[0], nil
}

func xInRect(xT int, rect xrect.Rect) bool {
	return xT >= rect.X() && xT < (rect.X()+rect.Width())
}

func yInRect(yT int, rect xrect.Rect) bool {
	return yT >= rect.Y() && yT < (rect.Y()+rect.Height())
}
