package heads

import (
	"fmt"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xrect"
)

var Heads xinerama.Heads
var HeadsStruts xinerama.Heads

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