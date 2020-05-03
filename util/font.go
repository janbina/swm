package util

import (
	"github.com/BurntSushi/freetype-go/freetype/truetype"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"os"
)

var cachedFonts = map[string]*truetype.Font{}

func GetFont(path string) (*truetype.Font, error) {
	if cachedFonts[path] != nil {
		return cachedFonts[path], nil
	}

	fontReader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	font, err := xgraphics.ParseFont(fontReader)
	if err != nil {
		return nil, err
	}

	cachedFonts[path] = font
	return font, nil
}
