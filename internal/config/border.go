package config

type BorderConfig struct {
	Size           int
	ColorNormal    uint32
	ColorActive    uint32
	ColorAttention uint32
}

const (
	borderColorActive    = 0xFF00BCD4
	borderColorInactive  = 0xFFB0BEC5
	borderColorAttention = 0xFFF44336
)

var BorderTop = &BorderConfig{
	Size:           1,
	ColorNormal:    borderColorInactive,
	ColorActive:    borderColorActive,
	ColorAttention: borderColorAttention,
}

var BorderBottom = &BorderConfig{
	Size:           1,
	ColorNormal:    borderColorInactive,
	ColorActive:    borderColorActive,
	ColorAttention: borderColorAttention,
}

var BorderLeft = &BorderConfig{
	Size:           1,
	ColorNormal:    borderColorInactive,
	ColorActive:    borderColorActive,
	ColorAttention: borderColorAttention,
}

var BorderRight = &BorderConfig{
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

func setBorder(border *BorderConfig, size int, colorNormal, colorActive, colorAttention uint32) {
	border.Size = size
	border.ColorNormal = colorNormal
	border.ColorActive = colorActive
	border.ColorAttention = colorAttention
}
