package util

import (
	"fmt"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/janbina/swm/config"
	"image"
)

func CreateTextBox(
	x *xgbutil.XUtil,
	text string,
	textSize float64,
	padding int,
	bg, fg uint32,
) (*xgraphics.Image, error) {
	font, err := GetFont(config.FontPath)
	if err != nil {
		return nil, fmt.Errorf("cannot get font: %s", err)
	}

	width, height := xgraphics.Extents(font, textSize, text)
	width += 2 * padding
	height += 2 * padding

	ximg := xgraphics.New(x, image.Rect(0, 0, width, height))
	ximg.For(func(x, y int) xgraphics.BGRA {
		return intColor2BGRA(bg)
	})

	_, _, err = ximg.Text(padding, padding, intColor2BGRA(fg), textSize, font, text)
	if err != nil {
		ximg.Destroy()
		return nil, fmt.Errorf("cannot draw text: %s", err)
	}

	return ximg, nil
}

func intColor2BGRA(color uint32) xgraphics.BGRA {
	B := color & 0xFF
	G := (color >> 8) & 0xFF
	R := (color >> 16) & 0xFF

	return xgraphics.BGRA{
		B: uint8(B),
		G: uint8(G),
		R: uint8(R),
		A: 0xFF,
	}
}
