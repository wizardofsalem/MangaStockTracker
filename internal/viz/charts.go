package viz

import (
	"fmt"
	"os"
	"sort"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"salemcodex.com/mangastocktracker/internal/stock"
)

func GenerateCharts(remaining map[string]stock.QuantityValue, soldStock map[string]stock.QuantityValue, stockMap map[string]stock.QuantityValue, totalSoldValue, totalBoughtValue, profit float64, outputPath string) error {
	page := components.NewPage()
	page.AddCharts(
		createRevenueChart(totalSoldValue, totalBoughtValue, profit),
		createStockWorthChart(remaining),
		createProfitPerTitleChart(soldStock, stockMap),
		createStockQuantityChart(remaining),
		createStockContributionPie(remaining),
	)

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	
	if err := page.Render(f); err != nil {
		return err
	}
	
	table := createTargetMangaTable(remaining)
	_, err = f.WriteString(table)
	return err
}

func createRevenueChart(revenue, cost, profit float64) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Total Sold: Revenue vs Cost"}))
	bar.SetXAxis([]string{"Sold Items"}).
		AddSeries("Cost of Sold", []opts.BarData{{Value: cost}}).
		AddSeries("Revenue", []opts.BarData{{Value: revenue}}).
		AddSeries("Profit", []opts.BarData{{Value: profit}})
	return bar
}

func createStockWorthChart(remaining map[string]stock.QuantityValue) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Current Stock Worth by Title"}),
		charts.WithInitializationOpts(opts.Initialization{Height: "1200px"}),
		charts.WithGridOpts(opts.Grid{Left: "25%"}),
	)
	bar.XYReversal()
	
	type item struct {
		name  string
		value float64
	}
	var items []item
	for name, qv := range remaining {
		items = append(items, item{name, qv.Value})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].value > items[j].value })
	
	var names []string
	var values []opts.BarData
	for _, it := range items {
		names = append(names, it.name)
		values = append(values, opts.BarData{Value: it.value})
	}
	
	bar.SetXAxis(names).AddSeries("Value", values)
	return bar
}

func createProfitPerTitleChart(soldStock, stockMap map[string]stock.QuantityValue) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Profit on Sold Items by Title"}),
		charts.WithInitializationOpts(opts.Initialization{Height: "1200px"}),
		charts.WithGridOpts(opts.Grid{Left: "25%"}),
	)
	bar.XYReversal()
	
	type item struct {
		name   string
		profit float64
	}
	var items []item
	for name, sold := range soldStock {
		costBasis := stockMap[name].AverageValue * float64(sold.Quantity)
		profit := sold.Value - costBasis
		items = append(items, item{name, profit})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].profit > items[j].profit })
	
	var names []string
	var profits []opts.BarData
	for _, it := range items {
		names = append(names, it.name)
		profits = append(profits, opts.BarData{Value: it.profit})
	}
	
	bar.SetXAxis(names).AddSeries("Profit", profits)
	return bar
}

func createStockQuantityChart(remaining map[string]stock.QuantityValue) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Current Stock Quantity by Title"}),
		charts.WithInitializationOpts(opts.Initialization{Height: "1200px"}),
		charts.WithGridOpts(opts.Grid{Left: "25%"}),
	)
	bar.XYReversal()
	
	type item struct {
		name string
		qty  uint64
	}
	var items []item
	for name, qv := range remaining {
		items = append(items, item{name, qv.Quantity})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].qty > items[j].qty })
	
	var names []string
	var quantities []opts.BarData
	for _, it := range items {
		names = append(names, it.name)
		quantities = append(quantities, opts.BarData{Value: it.qty})
	}
	
	bar.SetXAxis(names).AddSeries("Quantity", quantities)
	return bar
}

func createStockContributionPie(remaining map[string]stock.QuantityValue) *charts.Pie {
	pie := charts.NewPie()
	pie.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Stock Quantity Distribution"}))
	
	var items []opts.PieData
	for name, qv := range remaining {
		items = append(items, opts.PieData{Name: name, Value: qv.Quantity})
	}
	
	pie.AddSeries("Quantity", items)
	return pie
}

func createTargetMangaTable(remaining map[string]stock.QuantityValue) string {
	targets := []string{
		"one piece", "naruto", "bleach", "mha", "attack on titan", "death note",
		"demon slayer", "jjk", "chainsaw man", "tg", "fullmetal alchemist", "hxh",
		"dragon ball", "one punch", "spy x family", "berserk", "vinland saga", "haikyu",
		"mob psycho 100", "blue exorcist", "fairy tail", "black clover", "promised neverland",
		"tokyo revengers", "vagabond de", "slam dunk", "assassination", "fire force", "dr stone", "saop",
	}
	
	html := `<div style="margin: 20px;"><h2>Popular Manga Target List</h2><table border="1" style="border-collapse: collapse; width: 50%;"><tr><th style="padding: 8px;">Title</th><th style="padding: 8px;">Current Stock</th></tr>`
	
	for _, title := range targets {
		qty := uint64(0)
		for name, qv := range remaining {
			if name == title {
				qty = qv.Quantity
				break
			}
		}
		html += fmt.Sprintf(`<tr><td style="padding: 8px;">%s</td><td style="padding: 8px; text-align: center;">%d</td></tr>`, title, qty)
	}
	
	html += `</table></div>`
	return html
}
