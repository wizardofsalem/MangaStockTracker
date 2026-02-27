package stock

import "time"

type QuantityValue struct {
	Quantity     uint64
	Value        float64
	AverageValue float64
}

type Lot struct {
	QuantityValue
	Date *time.Time
}
