package main

import (
	"log"

	"salemcodex.com/mangastocktracker/internal/csvio"
	"salemcodex.com/mangastocktracker/internal/stock"
)

func main() {
	stockMap, stockLots, totalBoughtValue, totalBooksBought, err := csvio.ParseBought("bought.csv")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Total bought value = %.2f", totalBoughtValue)

	targetStockPrices, err := csvio.ParseTargets("targets.csv")
	if err != nil {
		log.Fatal(err)
	}

	soldStock, _, totalSoldValue, err := csvio.ParseSold("sold.csv")
	if err != nil {
		log.Fatal(err)
	}

	weightedAverageSold, totalSold := stock.CalculateWeightedAverageSold(stockMap, soldStock)
	soldAvg := totalSoldValue / float64(totalSold)
	profitMarginFactor := (soldAvg - weightedAverageSold) / weightedAverageSold

	log.Printf("Average sale price (overall) = %.2f", soldAvg)
	log.Printf("Weighted average cost of sold units = %.2f", weightedAverageSold)
	avgMap := stock.AverageSoldPrices(soldStock)
	for manga, avg := range avgMap {
		log.Printf("Sold - %s avg price = %.2f qty=%d", manga, avg, soldStock[manga].Quantity)
	}

	remaining := stock.ComputeRemainingFromLots(stockLots, soldStock)

	var totalRemainingValue float64
	var totalRemainingQty uint64
	for manga, qv := range remaining {
		log.Println("Remaining -", manga, "qty=", qv.Quantity, "avg=", qv.AverageValue)
		totalRemainingQty += qv.Quantity
		totalRemainingValue += qv.Value
	}

	var totalAverageValue float64
	for manga, qv := range stockMap {
		averageValue := qv.Value / float64(qv.Quantity)

		var averageLotValue float64
		if entry, ok := soldStock[manga]; ok {
			averageLotValue = averageValue * float64(entry.Quantity)
		}

		totalAverageValue += soldStock[manga].Value - averageLotValue
	}

	var totalProjectedEarnings float64
	for manga, qv := range remaining {
		projectedEarnings := targetStockPrices[manga]*float64(qv.Quantity) - qv.Value
		totalProjectedEarnings += projectedEarnings
	}

	log.Println("\nProfit on current sold: ", totalAverageValue)
	log.Println("Projected profit: ", totalProjectedEarnings)
	log.Println("Total books bought = ", totalBooksBought)
	log.Println("Total remaining qty = ", totalRemainingQty)
	log.Println("Total remaining value = ", totalRemainingValue)

	totalCurrentValue := totalRemainingValue + totalSoldValue
	log.Println("Starting total bought value = ", totalBoughtValue)
	log.Println("Total current value (remaining cost + sold revenue) = ", totalCurrentValue)

	absoluteGrowth := totalCurrentValue - totalBoughtValue
	if totalBoughtValue > 0 {
		growthPercent := (absoluteGrowth) / totalBoughtValue * 100
		log.Printf("Growth vs start = %.2f (%.2f%%)", absoluteGrowth, growthPercent)
	} else {
		log.Println("Growth vs start = N/A (no initial bought value)")
	}
	log.Println("Revenue = ", totalSoldValue)
	log.Println("Current profit: ", totalSoldValue-totalBoughtValue)
	log.Println("Current profit percentage = %", profitMarginFactor*100)
}
