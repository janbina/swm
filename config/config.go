package config

import "github.com/janbina/swm/decoration"

const (
	borderColorActive    = 0x00BCD4
	borderColorInactive  = 0xB0BEC5
	borderColorAttention = 0xF44336
)

var BorderTop = &decoration.BorderConfig{
	Size:           3,
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
