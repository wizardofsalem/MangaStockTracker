package stock

func CalculateAverageValue(value float64, quantity uint64) float64 {
	return value / float64(quantity)
}

// AverageSoldPrices returns average sale price per unit for each manga.
func AverageSoldPrices(sold map[string]QuantityValue) map[string]float64 {
	avg := make(map[string]float64)
	for name, qv := range sold {
		if qv.Quantity > 0 {
			avg[name] = qv.Value / float64(qv.Quantity)
		}
	}
	return avg
}

func CalculateWeightedAverageSold(stock map[string]QuantityValue, sold map[string]QuantityValue) (float64, uint64) {
	var weightedAverage float64
	var totalSold int64
	for soldStockName, qv := range sold {
		stockAvgValue := stock[soldStockName].AverageValue
		weightedAverage += float64(qv.Quantity) * stockAvgValue
		totalSold += int64(qv.Quantity)
	}

	return weightedAverage / float64(totalSold), uint64(totalSold)
}

// ComputeRemainingFromLots computes remaining quantity/value per title
// by consuming sold quantities from lots in FIFO order. stockLots maps
// a manga name to a slice of lots (each with Quantity and Value).
func ComputeRemainingFromLots(stockLots map[string][]Lot, sold map[string]QuantityValue) map[string]QuantityValue {
	remaining := make(map[string]QuantityValue)

	for name, lots := range stockLots {
		soldQty := uint64(0)
		if s, ok := sold[name]; ok {
			soldQty = s.Quantity
		}

		var remQty uint64
		var remValue float64

		for _, lot := range lots {
			if soldQty == 0 {
				remQty += lot.Quantity
				remValue += lot.Value
				continue
			}

			if soldQty >= lot.Quantity {
				// entire lot sold
				soldQty -= lot.Quantity
				continue
			}

			// partial consumption of this lot
			remainingInLot := lot.Quantity - soldQty
			avg := 0.0
			if lot.Quantity > 0 {
				avg = lot.Value / float64(lot.Quantity)
			}
			remQty += remainingInLot
			remValue += float64(remainingInLot) * avg
			soldQty = 0
		}

		if remQty > 0 {
			remaining[name] = QuantityValue{Quantity: remQty, Value: remValue, AverageValue: remValue / float64(remQty)}
		}
	}

	return remaining
}
