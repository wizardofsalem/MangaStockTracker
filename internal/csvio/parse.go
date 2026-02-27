package csvio

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"salemcodex.com/mangastocktracker/internal/stock"
)

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

func ParseBought(path string) (map[string]stock.QuantityValue, map[string][]stock.Lot, float64, uint64, error) {
	boughtCsv, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	defer func() { _ = boughtCsv.Close() }()

	scanner := bufio.NewScanner(boughtCsv)

	stockMap := make(map[string]stock.QuantityValue)
	stockLots := make(map[string][]stock.Lot)

	var totalBoughtValue float64
	var totalBooksBought uint64

	for scanner.Scan() {
		splitLine := strings.Split(scanner.Text(), ",")
		if len(splitLine) < 3 {
			continue
		}

		lotValue, err := strconv.ParseFloat(splitLine[0], 64)
		if err != nil {
			return nil, nil, 0, 0, err
		}

		mangaName := splitLine[1]

		quantity, err := strconv.ParseUint(splitLine[2], 10, 64)
		if err != nil {
			return nil, nil, 0, 0, err
		}
		totalBooksBought += quantity

		date, err := parseOptionalDate(splitLine, 3)
		if err != nil {
			return nil, nil, 0, 0, err
		}

		lot := stock.Lot{
			QuantityValue: stock.QuantityValue{Quantity: quantity, Value: lotValue, AverageValue: stock.CalculateAverageValue(lotValue, quantity)},
			Date:          date,
		}
		stockLots[mangaName] = append(stockLots[mangaName], lot)

		if entry, ok := stockMap[mangaName]; ok {
			entry.Quantity += quantity
			entry.Value += lotValue
			stockMap[mangaName] = entry
		} else {
			stockMap[mangaName] = stock.QuantityValue{Quantity: quantity, Value: lotValue}
		}

		qv := stockMap[mangaName]
		qv.AverageValue = stock.CalculateAverageValue(qv.Value, qv.Quantity)
		stockMap[mangaName] = qv
		totalBoughtValue += lotValue
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, 0, 0, err
	}

	return stockMap, stockLots, totalBoughtValue, totalBooksBought, nil
}

func ParseSold(path string) (map[string]stock.QuantityValue, map[string][]stock.Lot, float64, error) {
	soldCsv, err := os.Open(path)
	if err != nil {
		return nil, nil, 0, err
	}
	defer func() { _ = soldCsv.Close() }()

	scanner := bufio.NewScanner(soldCsv)

	soldStock := make(map[string]stock.QuantityValue)
	soldLots := make(map[string][]stock.Lot)
	var totalSoldValue float64

	for scanner.Scan() {
		splitLine := strings.Split(scanner.Text(), ",")
		if len(splitLine) < 3 {
			continue
		}
		soldValue, err := strconv.ParseFloat(splitLine[0], 64)
		if err != nil {
			return nil, nil, 0, err
		}

		mangaName := splitLine[1]

		quantity, err := strconv.ParseUint(splitLine[2], 10, 64)
		if err != nil {
			return nil, nil, 0, err
		}

		date, err := parseOptionalDate(splitLine, 3)
		if err != nil {
			return nil, nil, 0, err
		}

		soldLot := stock.Lot{
			QuantityValue: stock.QuantityValue{Quantity: quantity, Value: soldValue, AverageValue: stock.CalculateAverageValue(soldValue, quantity)},
			Date:          date,
		}
		soldLots[mangaName] = append(soldLots[mangaName], soldLot)

		if entry, ok := soldStock[mangaName]; ok {
			entry.Quantity += quantity
			entry.Value += soldValue
			soldStock[mangaName] = entry
		} else {
			soldStock[mangaName] = stock.QuantityValue{Quantity: quantity, Value: soldValue}
		}

		totalSoldValue += soldValue
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, 0, err
	}

	return soldStock, soldLots, totalSoldValue, nil
}

func ParseTargets(path string) (map[string]float64, error) {
	targetsCsv, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = targetsCsv.Close() }()

	scanner := bufio.NewScanner(targetsCsv)

	targetStockPrices := make(map[string]float64)
	for scanner.Scan() {
		splitLine := strings.Split(scanner.Text(), ",")
		if len(splitLine) < 2 {
			continue
		}

		mangaName := splitLine[1]

		targetPrice, err := strconv.ParseFloat(splitLine[0], 64)
		if err != nil {
			return nil, err
		}

		targetStockPrices[mangaName] = targetPrice
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return targetStockPrices, nil
}
