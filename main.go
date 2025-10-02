package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

type QuantityValue struct {
	Quantity uint64
	Value    float64
}

func main() {

	//read in bought.csv
	boughtCsv, err := os.Open("bought.csv")
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(boughtCsv)

	stock := make(map[string]QuantityValue)

	var totalBoughtValue float64
	var totalBooksBought uint64
	for scanner.Scan() {
		splitLine := strings.Split(scanner.Text(), ",")

		lotValue, err := strconv.ParseFloat(splitLine[0], 64)
		if err != nil {
			log.Fatal(err)
		}

		mangaName := splitLine[1]

		quantity, err := strconv.ParseUint(splitLine[2], 10, 64)
		totalBooksBought += quantity

		if entry, ok := stock[mangaName]; ok {
			entry.Quantity += quantity
			entry.Value += lotValue

			stock[mangaName] = entry
		} else {
			stock[mangaName] = QuantityValue{Quantity: quantity, Value: lotValue}
		}

		totalBoughtValue += lotValue
	}

	err = boughtCsv.Close()
	if err != nil {
		log.Fatal(err)
	}

	soldCsv, err := os.Open("sold.csv")

	scanner = bufio.NewScanner(soldCsv)

	soldStock := make(map[string]QuantityValue)
	var totalSoldValue float64
	for scanner.Scan() {
		splitLine := strings.Split(scanner.Text(), ",")
		soldValue, err := strconv.ParseFloat(splitLine[0], 64)
		if err != nil {
			log.Fatal(err)
		}

		mangaName := splitLine[1]

		quantity, err := strconv.ParseUint(splitLine[2], 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		if entry, ok := soldStock[mangaName]; ok {
			entry.Quantity += quantity
			entry.Value += soldValue

			soldStock[mangaName] = entry
		} else {
			soldStock[mangaName] = QuantityValue{Quantity: quantity, Value: soldValue}
		}

		totalSoldValue += soldValue
	}

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

	//How to sensibily determine price of current stock?
	//have to calculate relative time periods of when stock bought and sold

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
	for manga, qv := range stock {
		log.Println("Average stock value of ", manga, " = ", qv.Value/float64(qv.Quantity))
		projectedEarnings := targetStockPrices[manga]*float64(qv.Quantity) - qv.Value
		log.Println("Projected earnings = ", projectedEarnings)
		totalProjectedEarnings += projectedEarnings
	}

	log.Println("Profit on current sold: ", totalAverageValue)
	log.Println("Current profit: ", totalSoldValue-totalBoughtValue)
	log.Println("Projected profit: ", totalProjectedEarnings)
	log.Println("Profit as percentage", (100*totalProjectedEarnings)/totalBoughtValue, "%")
	log.Println("Total books bought = ", totalBooksBought)
}
