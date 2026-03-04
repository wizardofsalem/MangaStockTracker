package main

import (
	"log"

	"salemcodex.com/mangastocktracker/internal/csvio"
	"salemcodex.com/mangastocktracker/internal/stock"
	"salemcodex.com/mangastocktracker/internal/viz"
)

func main() {
	stockMap, stockLots, _, _, err := csvio.ParseBought("bought.csv")
	if err != nil {
		log.Fatal(err)
	}

	_, err = csvio.ParseTargets("targets.csv")
	if err != nil {
		log.Fatal(err)
	}

	soldStock, _, totalSoldValue, err := csvio.ParseSold("sold.csv")
	if err != nil {
		log.Fatal(err)
	}

	weightedAverageSold, totalSold := stock.CalculateWeightedAverageSold(stockMap, soldStock)
	remaining := stock.ComputeRemainingFromLots(stockLots, soldStock)

	costOfSold := weightedAverageSold * float64(totalSold)
	profitOnSold := totalSoldValue - costOfSold
	if err := viz.GenerateCharts(remaining, soldStock, stockMap, totalSoldValue, costOfSold, profitOnSold, "charts.html"); err != nil {
		log.Fatal(err)
	}
	log.Println("Charts generated: charts.html")
}
