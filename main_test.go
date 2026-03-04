package main

import (
	"testing"

	"salemcodex.com/mangastocktracker/internal/csvio"
	"salemcodex.com/mangastocktracker/internal/stock"
)

func TestStockTracking(t *testing.T) {
	stockMap, stockLots, totalBoughtValue, totalBooksBought, err := csvio.ParseBought("test/bought_test.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Test total bought value: 30 + 20 + 50 + 40 = 140
	if totalBoughtValue != 140.0 {
		t.Errorf("Expected totalBoughtValue=140.0, got %.2f", totalBoughtValue)
	}

	// Test total books bought: 10 + 5 + 10 + 8 = 33
	if totalBooksBought != 33 {
		t.Errorf("Expected totalBooksBought=33, got %d", totalBooksBought)
	}

	// Test manga_a: qty=23, value=90, avg=3.91
	if stockMap["manga_a"].Quantity != 23 {
		t.Errorf("Expected manga_a quantity=23, got %d", stockMap["manga_a"].Quantity)
	}
	if stockMap["manga_a"].Value != 90.0 {
		t.Errorf("Expected manga_a value=90.0, got %.2f", stockMap["manga_a"].Value)
	}
	expectedAvg := 90.0 / 23.0
	if stockMap["manga_a"].AverageValue != expectedAvg {
		t.Errorf("Expected manga_a avg=%.4f, got %.4f", expectedAvg, stockMap["manga_a"].AverageValue)
	}

	// Test manga_b: qty=10, value=50, avg=5.0
	if stockMap["manga_b"].Quantity != 10 {
		t.Errorf("Expected manga_b quantity=10, got %d", stockMap["manga_b"].Quantity)
	}
	if stockMap["manga_b"].Value != 50.0 {
		t.Errorf("Expected manga_b value=50.0, got %.2f", stockMap["manga_b"].Value)
	}
	if stockMap["manga_b"].AverageValue != 5.0 {
		t.Errorf("Expected manga_b avg=5.0, got %.2f", stockMap["manga_b"].AverageValue)
	}

	soldStock, _, totalSoldValue, err := csvio.ParseSold("test/sold_test.csv")
	if err != nil {
		t.Fatal(err)
	}

	// Test total sold value: 24 + 30 = 54
	if totalSoldValue != 54.0 {
		t.Errorf("Expected totalSoldValue=54.0, got %.2f", totalSoldValue)
	}

	// Test sold manga_a: qty=6, value=24, avg=4.0
	if soldStock["manga_a"].Quantity != 6 {
		t.Errorf("Expected sold manga_a quantity=6, got %d", soldStock["manga_a"].Quantity)
	}
	if soldStock["manga_a"].Value != 24.0 {
		t.Errorf("Expected sold manga_a value=24.0, got %.2f", soldStock["manga_a"].Value)
	}

	// Test weighted average sold
	weightedAvgSold, totalSold := stock.CalculateWeightedAverageSold(stockMap, soldStock)
	if totalSold != 11 {
		t.Errorf("Expected totalSold=11, got %d", totalSold)
	}
	// manga_a: 6 units @ 3.91, manga_b: 5 units @ 5.0
	// weighted avg = (6*3.91 + 5*5.0) / 11 = 4.27
	expectedWeightedAvg := (6*expectedAvg + 5*5.0) / 11.0
	if weightedAvgSold != expectedWeightedAvg {
		t.Errorf("Expected weightedAvgSold=%.4f, got %.4f", expectedWeightedAvg, weightedAvgSold)
	}

	// Test remaining stock (FIFO)
	remaining := stock.ComputeRemainingFromLots(stockLots, soldStock)

	// manga_a: bought 10, 5, 8, sold 6 -> remaining 17 (4 from first lot @ 3.0, 5 from second @ 4.0, 8 from third @ 5.0)
	// remaining value = 4*3.0 + 5*4.0 + 8*5.0 = 12 + 20 + 40 = 72
	if remaining["manga_a"].Quantity != 17 {
		t.Errorf("Expected remaining manga_a quantity=17, got %d", remaining["manga_a"].Quantity)
	}
	expectedRemValue := 4*3.0 + 5*4.0 + 8*5.0
	if remaining["manga_a"].Value != expectedRemValue {
		t.Errorf("Expected remaining manga_a value=%.2f, got %.2f", expectedRemValue, remaining["manga_a"].Value)
	}

	// manga_b: bought 10, sold 5 -> remaining 5 @ 5.0 = 25
	if remaining["manga_b"].Quantity != 5 {
		t.Errorf("Expected remaining manga_b quantity=5, got %d", remaining["manga_b"].Quantity)
	}
	if remaining["manga_b"].Value != 25.0 {
		t.Errorf("Expected remaining manga_b value=25.0, got %.2f", remaining["manga_b"].Value)
	}

	// Test average sold prices
	avgSoldPrices := stock.AverageSoldPrices(soldStock)
	if avgSoldPrices["manga_a"] != 4.0 {
		t.Errorf("Expected avg sold price manga_a=4.0, got %.2f", avgSoldPrices["manga_a"])
	}
	if avgSoldPrices["manga_b"] != 6.0 {
		t.Errorf("Expected avg sold price manga_b=6.0, got %.2f", avgSoldPrices["manga_b"])
	}
}
