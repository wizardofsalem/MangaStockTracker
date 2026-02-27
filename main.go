package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type QuantityValue struct {
	Quantity     uint64
	Value        float64
	AverageValue float64
}

type Lot struct {
	QuantityValue
	Date *time.Time
}

func calculateAverageValue(value float64, quantity uint64) float64 {
	return value / float64(quantity)
}

func parseOptionalDate(fields []string, idx int) (*time.Time, error) {
	if idx < 0 || idx >= len(fields) {
		return nil, nil
	}
	raw := strings.TrimSpace(fields[idx])
	if raw == "" {
		return nil, nil
	}

	layouts := []string{
		"2006-01-02",
		"1/2/2006",
		"01/02/2006",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("invalid date %q", raw)
}

func parseSold() (map[string]QuantityValue, map[string][]Lot, float64) {
	soldCsv, err := os.Open("sold.csv")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(soldCsv)

	soldStock := make(map[string]QuantityValue)
	soldLots := make(map[string][]Lot)
	var totalSoldValue float64

	for scanner.Scan() {
		splitLine := strings.Split(scanner.Text(), ",")
		if len(splitLine) < 3 {
			continue
		}
		soldValue, err := strconv.ParseFloat(splitLine[0], 64)
		if err != nil {
			panic(err)
		}

		mangaName := splitLine[1]

		quantity, err := strconv.ParseUint(splitLine[2], 10, 64)
		if err != nil {
			panic(err)
		}

		date, err := parseOptionalDate(splitLine, 3)
		if err != nil {
			panic(err)
		}

		soldLot := Lot{
			QuantityValue: QuantityValue{Quantity: quantity, Value: soldValue, AverageValue: calculateAverageValue(soldValue, quantity)},
			Date:          date,
		}
		soldLots[mangaName] = append(soldLots[mangaName], soldLot)

		if entry, ok := soldStock[mangaName]; ok {
			entry.Quantity += quantity
			entry.Value += soldValue

			soldStock[mangaName] = entry
		} else {
			soldStock[mangaName] = QuantityValue{Quantity: quantity, Value: soldValue}
		}

		totalSoldValue += soldValue
	}

	return soldStock, soldLots, totalSoldValue
}

// averageSoldPrices returns average sale price per unit for each manga
func averageSoldPrices(sold map[string]QuantityValue) map[string]float64 {
	avg := make(map[string]float64)
	for name, qv := range sold {
		if qv.Quantity > 0 {
			avg[name] = qv.Value / float64(qv.Quantity)
		}
	}
	return avg
}

func calcualteWeightedAverageSold(stock map[string]QuantityValue, sold map[string]QuantityValue) (float64, uint64) {
	var weightedAverage float64
	var totalSold int64
	for soldStockName, qv := range sold {
		stockAvgValue := stock[soldStockName].AverageValue
		weightedAverage += float64(qv.Quantity) * stockAvgValue
		totalSold += int64(qv.Quantity)
	}

	return weightedAverage / float64(totalSold), uint64(totalSold)
}

func computeRemaining(stock map[string]QuantityValue, sold map[string]QuantityValue) map[string]QuantityValue {
	remaining := make(map[string]QuantityValue)

	for name, qv := range stock {
		soldQty := uint64(0)
		if s, ok := sold[name]; ok {
			soldQty = s.Quantity
		}

		if soldQty >= qv.Quantity {
			// fully sold or oversold -> no remaining
			continue
		}

		remQty := qv.Quantity - soldQty
		// use weighted average cost per unit for remaining value
		avgPerUnit := 0.0
		if qv.Quantity > 0 {
			avgPerUnit = qv.Value / float64(qv.Quantity)
		}
		remValue := float64(remQty) * avgPerUnit

		remaining[name] = QuantityValue{Quantity: remQty, Value: remValue, AverageValue: avgPerUnit}
	}

	return remaining
}

// computeRemainingFromLots computes remaining quantity/value per title
// by consuming sold quantities from lots in FIFO order. stockLots maps
// a manga name to a slice of lots (each with Quantity and Value).
func computeRemainingFromLots(stockLots map[string][]Lot, sold map[string]QuantityValue) map[string]QuantityValue {
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

func main() {
	// read in bought.csv
	boughtCsv, err := os.Open("bought.csv")
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(boughtCsv)

	stock := make(map[string]QuantityValue)
	stockLots := make(map[string][]Lot)

	var totalBoughtValue float64
	var totalBooksBought uint64
	for scanner.Scan() {
		splitLine := strings.Split(scanner.Text(), ",")
		if len(splitLine) < 3 {
			continue
		}

		lotValue, err := strconv.ParseFloat(splitLine[0], 64)
		if err != nil {
			log.Fatal(err)
		}

		mangaName := splitLine[1]

		quantity, err := strconv.ParseUint(splitLine[2], 10, 64)
		totalBooksBought += quantity

		date, err := parseOptionalDate(splitLine, 3)
		if err != nil {
			log.Fatal(err)
		}

		// record the lot (preserve order for FIFO consumption)
		lot := Lot{
			QuantityValue: QuantityValue{Quantity: quantity, Value: lotValue, AverageValue: calculateAverageValue(lotValue, quantity)},
			Date:          date,
		}
		stockLots[mangaName] = append(stockLots[mangaName], lot)

		if entry, ok := stock[mangaName]; ok {
			entry.Quantity += quantity
			entry.Value += lotValue

			stock[mangaName] = entry
		} else {
			stock[mangaName] = QuantityValue{Quantity: quantity, Value: lotValue}
		}
		qv := stock[mangaName]
		qv.AverageValue = calculateAverageValue(qv.Value, qv.Quantity)
		stock[mangaName] = qv
		totalBoughtValue += lotValue
	}

	err = boughtCsv.Close()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Total bought value = %.2f", totalBoughtValue)

	targetsCsv, err := os.Open("targets.csv")

	scanner = bufio.NewScanner(targetsCsv)

	targetStockPrices := make(map[string]float64)
	for scanner.Scan() {
		splitLine := strings.Split(scanner.Text(), ",")

		mangaName := splitLine[1]

		targetPrice, err := strconv.ParseFloat(splitLine[0], 64)
		if err != nil {
			log.Fatal(err)
		}

		targetStockPrices[mangaName] = targetPrice
	}

	// How to sensibily determine price of current stock?
	// have to calculate relative time periods of when stock bought and sold
	soldStock, _, totalSoldValue := parseSold()
	weightedAverageSold, totalSold := calcualteWeightedAverageSold(stock, soldStock)
	soldAvg := totalSoldValue / float64(totalSold)
	profitMarginFactor := (soldAvg - weightedAverageSold) / weightedAverageSold

	// display average sale prices
	log.Printf("Average sale price (overall) = %.2f", soldAvg)
	log.Printf("Weighted average cost of sold units = %.2f", weightedAverageSold)
	avgMap := averageSoldPrices(soldStock)
	for manga, avg := range avgMap {
		log.Printf("Sold - %s avg price = %.2f qty=%d", manga, avg, soldStock[manga].Quantity)
	}
	// compute remaining using FIFO across lots so adding new lots updates averages correctly
	remaining := computeRemainingFromLots(stockLots, soldStock)

	var totalRemainingValue float64
	var totalRemainingQty uint64
	for manga, qv := range remaining {
		log.Println("Remaining -", manga, "qty=", qv.Quantity, "avg=", qv.AverageValue)
		totalRemainingQty += qv.Quantity
		totalRemainingValue += qv.Value
	}
	var totalAverageValue float64
	for manga, qv := range stock {
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
	// total current value = value of remaining stock (at purchase cost) + revenue from sold stock
	totalCurrentValue := totalRemainingValue + totalSoldValue
	log.Println("Starting total bought value = ", totalBoughtValue)
	log.Println("Total current value (remaining cost + sold revenue) = ", totalCurrentValue)
	// absolute growth and percentage vs start
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
