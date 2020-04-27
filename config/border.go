package config

import "github.com/janbina/swm/decoration"

const (
	borderColorActive    = 0x00BCD4
	borderColorInactive  = 0xB0BEC5
	borderColorAttention = 0xF44336
)

var BorderTop = &decoration.BorderConfig{
	Size:           1,
	ColorNormal:    borderColorInactive,
	ColorActive:    borderColorActive,
	ColorAttention: borderColorAttention,
}

var BorderBottom = &decoration.BorderConfig{
	Size:           1,
	ColorNormal:    borderColorInactive,
	ColorActive:    borderColorActive,
	ColorAttention: borderColorAttention,
}

var BorderLeft = &decoration.BorderConfig{
	Size:           1,
	ColorNormal:    borderColorInactive,
	ColorActive:    borderColorActive,
	ColorAttention: borderColorAttention,
}

var BorderRight = &decoration.BorderConfig{
	Size:           1,
	ColorNormal:    borderColorInactive,
	ColorActive:    borderColorActive,
	ColorAttention: borderColorAttention,
}

func SetTopBorder(size int, colorNormal, colorActive, colorAttention uint32) {
	setBorder(BorderTop, size, colorNormal, colorActive, colorAttention)
}

func SetBottomBorder(size int, colorNormal, colorActive, colorAttention uint32) {
	setBorder(BorderBottom, size, colorNormal, colorActive, colorAttention)
}

func SetLeftBorder(size int, colorNormal, colorActive, colorAttention uint32) {
	setBorder(BorderLeft, size, colorNormal, colorActive, colorAttention)
}

func SetRightBorder(size int, colorNormal, colorActive, colorAttention uint32) {
	setBorder(BorderRight, size, colorNormal, colorActive, colorAttention)
}

func SetAllBorders(size int, colorNormal, colorActive, colorAttention uint32) {
	SetTopBorder(size, colorNormal, colorActive, colorAttention)
	SetBottomBorder(size, colorNormal, colorActive, colorAttention)
	SetLeftBorder(size, colorNormal, colorActive, colorAttention)
	SetRightBorder(size, colorNormal, colorActive, colorAttention)
}

func setBorder(border *decoration.BorderConfig, size int, colorNormal, colorActive, colorAttention uint32) {
	border.Size = size
	border.ColorNormal = colorNormal
	border.ColorActive = colorActive
	border.ColorAttention = colorAttention
}
