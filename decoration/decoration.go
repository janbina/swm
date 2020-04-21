package decoration

import "github.com/BurntSushi/xgbutil/xrect"

type Decoration interface {
	ApplyRect(config *WinConfig, rect xrect.Rect) xrect.Rect
	WidthNeeded(config *WinConfig) int
	HeightNeeded(config *WinConfig) int
	Left(config *WinConfig) int
	Right(config *WinConfig) int
	Top(config *WinConfig) int
	Bottom(config *WinConfig) int
	Active()
	InActive()
	Attention()
	Destroy()
}

type WinConfig struct {
	Fullscreen bool
}

type Decorations []Decoration

func (d *Decorations) ApplyRect(config *WinConfig, rect xrect.Rect) xrect.Rect {
	for _, decoration := range *d {
		rect = decoration.ApplyRect(config, rect)
	}
	return rect
}

func (d *Decorations) WidthNeeded(config *WinConfig) int {
	sum := 0
	for _, decoration := range *d {
		sum += decoration.WidthNeeded(config)
	}
	return sum
}

func (d *Decorations) HeightNeeded(config *WinConfig) int {
	sum := 0
	for _, decoration := range *d {
		sum += decoration.HeightNeeded(config)
	}
	return sum
}

func (d *Decorations) Left(config *WinConfig) int {
	sum := 0
	for _, decoration := range *d {
		sum += decoration.Left(config)
	}
	return sum
}

func (d *Decorations) Right(config *WinConfig) int {
	sum := 0
	for _, decoration := range *d {
		sum += decoration.Right(config)
	}
	return sum
}

func (d *Decorations) Top(config *WinConfig) int {
	sum := 0
	for _, decoration := range *d {
		sum += decoration.Top(config)
	}
	return sum
}

func (d *Decorations) Bottom(config *WinConfig) int {
	sum := 0
	for _, decoration := range *d {
		sum += decoration.Bottom(config)
	}
	return sum
}

func (d *Decorations) Active() {
	d.forAll(Decoration.Active)
}

func (d *Decorations) InActive() {
	d.forAll(Decoration.InActive)
}

func (d *Decorations) Attention() {
	d.forAll(Decoration.Attention)
}

func (d *Decorations) Destroy() {
	d.forAll(Decoration.Destroy)
}

func (d *Decorations) forAll(f func(d Decoration)) {
	for _, decoration := range *d {
		f(decoration)
	}
}
