package models

import "gorm.io/gorm"

type Portfolio struct {
	gorm.Model
	Ticker       string
	Rank         float64
	Close        float64
	Volatility   float64
	CloseAboveMa bool
	Gap          bool
	PctAlloc     float64
	Cost         float64
	NumShares    float64
}
