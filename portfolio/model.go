package portfolio

import (
	"gorm.io/gorm"
)

type Holding struct {
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

type Instrument struct {
	gorm.Model
	Ticker      string
	Description string
	Metrics     []Metric
}
type Metric struct {
	gorm.Model
	InstrumentID uint // automagic foreign key from instrument
	Date         string
	Value        float64
}
