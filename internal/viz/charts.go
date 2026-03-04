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
		createProfitMarginChart(soldStock, stockMap),
		createInventoryValueChart(remaining),
		createSalesVelocityChart(soldStock, stockMap),
		createStockWorthChart(remaining),
		createProfitPerTitleChart(soldStock, stockMap),
		createProfitPercentChart(soldStock, stockMap),
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
	
	table := createTargetMangaTables(remaining, soldStock, stockMap)
	analysis := createAnalysisTables(remaining, soldStock, stockMap)
	pricing := createPricingOptimization(remaining, soldStock, stockMap)
	_, err = f.WriteString(analysis + pricing + table)
	return err
}

func createRevenueChart(revenue, cost, profit float64) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Total Sold: Revenue vs Cost"}))
	bar.SetXAxis([]string{"Sold Items"}).
		AddSeries("Total Cost", []opts.BarData{{Value: cost}}).
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
	sort.Slice(items, func(i, j int) bool { return items[i].value < items[j].value })
	
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
		name    string
		profit  float64
		percent float64
	}
	var items []item
	for name, sold := range soldStock {
		costBasis := stockMap[name].AverageValue * float64(sold.Quantity)
		profit := sold.Value - costBasis
		percent := 0.0
		if costBasis > 0 {
			percent = (profit / costBasis) * 100
		}
		items = append(items, item{name, profit, percent})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].profit < items[j].profit })
	
	var names []string
	var profits []opts.BarData
	for _, it := range items {
		names = append(names, it.name)
		profits = append(profits, opts.BarData{
			Value: it.profit,
			Name:  fmt.Sprintf("£%.2f (%.1f%%)", it.profit, it.percent),
		})
	}
	
	bar.SetXAxis(names).AddSeries("Profit", profits)
	return bar
}

func createProfitPercentChart(soldStock, stockMap map[string]stock.QuantityValue) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Profit % on Sold Items by Title"}),
		charts.WithInitializationOpts(opts.Initialization{Height: "1200px"}),
		charts.WithGridOpts(opts.Grid{Left: "25%"}),
	)
	bar.XYReversal()
	
	type item struct {
		name    string
		percent float64
	}
	var items []item
	for name, sold := range soldStock {
		costBasis := stockMap[name].AverageValue * float64(sold.Quantity)
		profit := sold.Value - costBasis
		percent := 0.0
		if costBasis > 0 {
			percent = (profit / costBasis) * 100
		}
		items = append(items, item{name, percent})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].percent < items[j].percent })
	
	var names []string
	var percents []opts.BarData
	for _, it := range items {
		names = append(names, it.name)
		percents = append(percents, opts.BarData{Value: it.percent})
	}
	
	bar.SetXAxis(names).AddSeries("Profit %", percents)
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
	sort.Slice(items, func(i, j int) bool { return items[i].qty < items[j].qty })
	
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

func createProfitMarginChart(soldStock, stockMap map[string]stock.QuantityValue) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Top Performers by Profit Margin %"}),
		charts.WithInitializationOpts(opts.Initialization{Height: "600px"}),
	)
	bar.XYReversal()
	
	type item struct {
		name   string
		margin float64
		sold   uint64
	}
	var items []item
	for name, sold := range soldStock {
		if sold.Quantity < 3 {
			continue
		}
		costBasis := stockMap[name].AverageValue * float64(sold.Quantity)
		if costBasis > 0 {
			margin := ((sold.Value - costBasis) / costBasis) * 100
			items = append(items, item{name, margin, sold.Quantity})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].margin < items[j].margin })
	
	if len(items) > 15 {
		items = items[len(items)-15:]
	}
	
	var names []string
	var margins []opts.BarData
	for _, it := range items {
		names = append(names, it.name)
		margins = append(margins, opts.BarData{Value: it.margin})
	}
	
	bar.SetXAxis(names).AddSeries("Margin %", margins)
	return bar
}

func createInventoryValueChart(remaining map[string]stock.QuantityValue) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Capital Tied Up - Top 10"}),
		charts.WithInitializationOpts(opts.Initialization{Height: "500px"}),
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
	sort.Slice(items, func(i, j int) bool { return items[i].value < items[j].value })
	
	if len(items) > 10 {
		items = items[len(items)-10:]
	}
	
	var names []string
	var values []opts.BarData
	for _, it := range items {
		names = append(names, it.name)
		values = append(values, opts.BarData{Value: it.value})
	}
	
	bar.SetXAxis(names).AddSeries("Value £", values)
	return bar
}

func createSalesVelocityChart(soldStock, stockMap map[string]stock.QuantityValue) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Sales Velocity - Top Sellers by Volume"}),
		charts.WithInitializationOpts(opts.Initialization{Height: "600px"}),
	)
	bar.XYReversal()
	
	type item struct {
		name string
		qty  uint64
	}
	var items []item
	for name, sold := range soldStock {
		items = append(items, item{name, sold.Quantity})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].qty < items[j].qty })
	
	if len(items) > 15 {
		items = items[len(items)-15:]
	}
	
	var names []string
	var qtys []opts.BarData
	for _, it := range items {
		names = append(names, it.name)
		qtys = append(qtys, opts.BarData{Value: it.qty})
	}
	
	bar.SetXAxis(names).AddSeries("Units Sold", qtys)
	return bar
}

func createAnalysisTables(remaining map[string]stock.QuantityValue, soldStock map[string]stock.QuantityValue, stockMap map[string]stock.QuantityValue) string {
	type analysis struct {
		name         string
		bought       uint64
		sold         uint64
		remaining    uint64
		avgCost      float64
		avgSale      float64
		margin       float64
		capitalTied  float64
		sellThrough  float64
	}
	
	var items []analysis
	for name, stock := range stockMap {
		a := analysis{
			name:        name,
			bought:      stock.Quantity,
			avgCost:     stock.AverageValue,
			capitalTied: 0,
		}
		
		if rem, ok := remaining[name]; ok {
			a.remaining = rem.Quantity
			a.capitalTied = rem.Value
		}
		
		if sold, ok := soldStock[name]; ok {
			a.sold = sold.Quantity
			if sold.Quantity > 0 {
				a.avgSale = sold.Value / float64(sold.Quantity)
			}
		}
		
		if a.bought > 0 {
			a.sellThrough = (float64(a.sold) / float64(a.bought)) * 100
		}
		
		if a.avgCost > 0 && a.avgSale > 0 {
			a.margin = ((a.avgSale - a.avgCost) / a.avgCost) * 100
		}
		
		items = append(items, a)
	}
	
	// Top performers
	topPerformers := make([]analysis, len(items))
	copy(topPerformers, items)
	sort.Slice(topPerformers, func(i, j int) bool {
		if topPerformers[i].sold < 3 {
			return false
		}
		if topPerformers[j].sold < 3 {
			return false
		}
		return topPerformers[i].margin > topPerformers[j].margin
	})
	
	// Problem inventory
	problemInv := make([]analysis, len(items))
	copy(problemInv, items)
	sort.Slice(problemInv, func(i, j int) bool {
		return problemInv[i].capitalTied > problemInv[j].capitalTied
	})
	
	html := `<div style="margin: 20px;">
<h2>📊 Performance Analysis</h2>
<style>
.analysis-table {
  border-collapse: collapse;
  width: 95%;
  font-family: Arial, sans-serif;
  font-size: 13px;
  margin-bottom: 30px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}
.analysis-table th {
  background: linear-gradient(to bottom, #2563eb, #1e40af);
  color: white;
  padding: 10px 8px;
  text-align: left;
  font-weight: 600;
  font-size: 12px;
}
.analysis-table td {
  padding: 8px;
  border-bottom: 1px solid #e2e8f0;
  font-size: 12px;
}
.analysis-table tr:hover {
  background-color: #f0f9ff;
}
.good { color: #059669; font-weight: 600; }
.bad { color: #dc2626; font-weight: 600; }
.warning { color: #d97706; font-weight: 600; }
</style>

<h3>🏆 Top Performers (Min 3 sold)</h3>
<table class="analysis-table">
<tr>
<th>Title</th>
<th>Bought</th>
<th>Sold</th>
<th>Sell-Through %</th>
<th>Avg Cost</th>
<th>Avg Sale</th>
<th>Margin %</th>
</tr>`
	
	count := 0
	for _, a := range topPerformers {
		if a.sold < 3 || count >= 10 {
			break
		}
		marginClass := "good"
		if a.margin < 10 {
			marginClass = "warning"
		}
		html += fmt.Sprintf(`<tr><td>%s</td><td>%d</td><td>%d</td><td>%.0f%%</td><td>£%.2f</td><td>£%.2f</td><td class="%s">%+.1f%%</td></tr>`,
			a.name, a.bought, a.sold, a.sellThrough, a.avgCost, a.avgSale, marginClass, a.margin)
		count++
	}
	
	html += `</table>

<h3>⚠️ Capital Tied Up - Action Needed</h3>
<table class="analysis-table">
<tr>
<th>Title</th>
<th>Remaining</th>
<th>Capital Tied</th>
<th>Sold</th>
<th>Sell-Through %</th>
<th>Recommendation</th>
</tr>`
	
	for i := 0; i < 10 && i < len(problemInv); i++ {
		a := problemInv[i]
		if a.capitalTied < 10 {
			break
		}
		
		rec := "Monitor"
		recClass := ""
		if a.sellThrough < 30 && a.remaining > 10 {
			rec = "Price drop or bundle"
			recClass = "bad"
		} else if a.sellThrough < 50 {
			rec = "Increase marketing"
			recClass = "warning"
		} else {
			rec = "Restock when low"
			recClass = "good"
		}
		
		html += fmt.Sprintf(`<tr><td>%s</td><td>%d</td><td class="warning">£%.2f</td><td>%d</td><td>%.0f%%</td><td class="%s">%s</td></tr>`,
			a.name, a.remaining, a.capitalTied, a.sold, a.sellThrough, recClass, rec)
	}
	
	html += `</table></div>`
	return html
}

func createTargetMangaTables(remaining map[string]stock.QuantityValue, soldStock map[string]stock.QuantityValue, stockMap map[string]stock.QuantityValue) string {
	popularTargets := []string{
		"one piece", "naruto", "bleach", "mha", "attack on titan", "death note",
		"demon slayer", "jjk", "chainsaw man", "tg", "fullmetal alchemist", "hxh",
		"dragon ball", "one punch", "spy x family", "berserk", "vinland saga", "haikyu",
		"mob psycho 100", "blue exorcist", "fairy tail", "black clover", "promised neverland",
		"tokyo revengers", "vagabond de", "slam dunk", "assassination", "fire force", "dr stone", "saop",
	}
	
	// Collect all active manga (from bought or sold)
	activeSet := make(map[string]bool)
	for name := range stockMap {
		activeSet[name] = true
	}
	for name := range soldStock {
		activeSet[name] = true
	}
	
	var active []string
	for name := range activeSet {
		active = append(active, name)
	}
	sort.Strings(active)
	
	// Find popular targets not yet bought/sold
	var consider []string
	for _, title := range popularTargets {
		if !activeSet[title] {
			consider = append(consider, title)
		}
	}
	
	html := `<div style="margin: 20px;">
<h2>📦 Complete Inventory Overview</h2>
<style>
.manga-table {
  border-collapse: collapse;
  width: 70%;
  font-family: Arial, sans-serif;
  font-size: 14px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}
.manga-table th {
  background: linear-gradient(to bottom, #4a5568, #2d3748);
  color: white;
  padding: 12px 8px;
  text-align: left;
  font-weight: 600;
  cursor: pointer;
  user-select: none;
}
.manga-table th:hover {
  background: linear-gradient(to bottom, #5a6578, #3d4758);
}
.manga-table td {
  padding: 10px 8px;
  border-bottom: 1px solid #e2e8f0;
}
.manga-table tr:hover {
  background-color: #f7fafc;
}
.manga-table tr:nth-child(even) {
  background-color: #f9fafb;
}
.manga-table tr:nth-child(even):hover {
  background-color: #f0f4f8;
}
.manga-table th::after {
  content: ' ⇅';
  opacity: 0.3;
}
</style>
<table class="manga-table" id="mangaTable">
<thead>
<tr>
<th onclick="sortTable(0)">Title</th>
<th onclick="sortTable(1)">Stock</th>
<th onclick="sortTable(2)">Avg Cost</th>
<th onclick="sortTable(3)">Sold</th>
<th onclick="sortTable(4)">Avg Sale</th>
<th onclick="sortTable(5)">Margin %</th>
</tr>
</thead>
<tbody>`
	
	for _, title := range active {
		currentQty := uint64(0)
		if qv, ok := remaining[title]; ok {
			currentQty = qv.Quantity
		}
		
		avgBought := 0.0
		if qv, ok := stockMap[title]; ok {
			avgBought = qv.AverageValue
		}
		
		qtySold := uint64(0)
		avgSale := 0.0
		if qv, ok := soldStock[title]; ok {
			qtySold = qv.Quantity
			if qtySold > 0 {
				avgSale = qv.Value / float64(qtySold)
			}
		}
		
		profitPercent := ""
		profitValue := 0.0
		if avgBought > 0 && avgSale > 0 {
			profitValue = ((avgSale - avgBought) / avgBought) * 100
			profitPercent = fmt.Sprintf("%+.1f%%", profitValue)
		}
		
		html += fmt.Sprintf(`<tr><td>%s</td><td style="text-align: center;">%d</td><td style="text-align: center;">£%.2f</td><td style="text-align: center;">%d</td><td style="text-align: center;">£%.2f</td><td style="text-align: center;">%s</td></tr>`, 
			title, currentQty, avgBought, qtySold, avgSale, profitPercent)
	}
	
	html += `</tbody></table>
<script>
let sortDirections = {};
function sortTable(col) {
  const table = document.getElementById("mangaTable");
  const tbody = table.querySelector("tbody");
  const rows = Array.from(tbody.querySelectorAll("tr"));
  
  const dir = sortDirections[col] || 1;
  sortDirections[col] = -dir;
  
  rows.sort((a, b) => {
    let aVal = a.cells[col].textContent.trim();
    let bVal = b.cells[col].textContent.trim();
    
    if (col === 0) {
      return dir * aVal.localeCompare(bVal);
    }
    
    aVal = parseFloat(aVal.replace(/[£%+,]/g, '')) || 0;
    bVal = parseFloat(bVal.replace(/[£%+,]/g, '')) || 0;
    return dir * (aVal - bVal);
  });
  
  rows.forEach(row => tbody.appendChild(row));
}
</script>
</div>`
	
	html += `<div style="margin: 20px;"><h2>Manga to Consider</h2><table border="1" style="border-collapse: collapse; width: 50%;"><tr><th style="padding: 8px;">Title</th></tr>`
	
	for _, title := range consider {
		html += fmt.Sprintf(`<tr><td style="padding: 8px;">%s</td></tr>`, title)
	}
	
	html += `</table></div>`
	return html
}

func createPricingOptimization(remaining map[string]stock.QuantityValue, soldStock map[string]stock.QuantityValue, stockMap map[string]stock.QuantityValue) string {
	type priceOpt struct {
		name            string
		currentPrice    float64
		avgCost         float64
		currentMargin   float64
		qtySold         uint64
		qtyRemaining    uint64
		sellThrough     float64
		suggestedPrice  float64
		suggestedMargin float64
		reasoning       string
	}
	
	var items []priceOpt
	
	for name, stock := range stockMap {
		opt := priceOpt{
			name:    name,
			avgCost: stock.AverageValue,
		}
		
		if rem, ok := remaining[name]; ok {
			opt.qtyRemaining = rem.Quantity
		}
		
		if sold, ok := soldStock[name]; ok {
			opt.qtySold = sold.Quantity
			if sold.Quantity > 0 {
				opt.currentPrice = sold.Value / float64(sold.Quantity)
			}
		}
		
		if stock.Quantity > 0 {
			opt.sellThrough = (float64(opt.qtySold) / float64(stock.Quantity)) * 100
		}
		
		if opt.avgCost > 0 && opt.currentPrice > 0 {
			opt.currentMargin = ((opt.currentPrice - opt.avgCost) / opt.avgCost) * 100
		}
		
		// Optimization logic
		if opt.qtySold < 3 {
			continue // Not enough data
		}
		
		// High inventory + low sell-through = reduce price
		if opt.qtyRemaining > 15 && opt.sellThrough < 40 {
			opt.suggestedPrice = opt.avgCost * 1.15 // 15% margin
			opt.reasoning = "High stock, slow sales - reduce to move inventory"
		} else if opt.qtyRemaining > 10 && opt.sellThrough < 50 {
			opt.suggestedPrice = opt.avgCost * 1.20 // 20% margin
			opt.reasoning = "Moderate stock - slight reduction to increase velocity"
		} else if opt.sellThrough > 70 && opt.currentMargin < 30 {
			// High demand, low margin = increase price
			opt.suggestedPrice = opt.avgCost * 1.35 // 35% margin
			opt.reasoning = "High demand - test higher price"
		} else if opt.sellThrough > 60 && opt.currentMargin < 25 {
			opt.suggestedPrice = opt.avgCost * 1.30 // 30% margin
			opt.reasoning = "Good velocity - room to increase margin"
		} else if opt.qtyRemaining < 5 && opt.sellThrough > 50 {
			// Low stock, proven seller = premium pricing
			opt.suggestedPrice = opt.avgCost * 1.40 // 40% margin
			opt.reasoning = "Low stock, proven seller - premium pricing"
		} else {
			opt.suggestedPrice = opt.currentPrice
			opt.reasoning = "Current pricing optimal"
		}
		
		if opt.suggestedPrice > 0 && opt.avgCost > 0 {
			opt.suggestedMargin = ((opt.suggestedPrice - opt.avgCost) / opt.avgCost) * 100
		}
		
		// Only show items where suggestion differs from current
		if opt.suggestedPrice != opt.currentPrice {
			items = append(items, opt)
		}
	}
	
	// Sort by potential impact (high inventory items first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].qtyRemaining > items[j].qtyRemaining
	})
	
	html := `<div style="margin: 20px;">
<h2>💰 Pricing Optimization Recommendations</h2>
<style>
.price-table {
  border-collapse: collapse;
  width: 95%;
  font-family: Arial, sans-serif;
  font-size: 13px;
  margin-bottom: 30px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}
.price-table th {
  background: linear-gradient(to bottom, #059669, #047857);
  color: white;
  padding: 10px 8px;
  text-align: left;
  font-weight: 600;
  font-size: 12px;
}
.price-table td {
  padding: 8px;
  border-bottom: 1px solid #e2e8f0;
  font-size: 12px;
}
.price-table tr:hover {
  background-color: #f0fdf4;
}
.price-up { color: #dc2626; }
.price-down { color: #059669; }
.price-same { color: #6b7280; }
</style>

<table class="price-table">
<tr>
<th>Title</th>
<th>Stock</th>
<th>Sold</th>
<th>Sell-Through</th>
<th>Current Price</th>
<th>Current Margin</th>
<th>Suggested Price</th>
<th>New Margin</th>
<th>Change</th>
<th>Reasoning</th>
</tr>`
	
	for _, opt := range items {
		priceChange := opt.suggestedPrice - opt.currentPrice
		changeClass := "price-same"
		changeSymbol := "→"
		if priceChange > 0 {
			changeClass = "price-up"
			changeSymbol = "↑"
		} else if priceChange < 0 {
			changeClass = "price-down"
			changeSymbol = "↓"
		}
		
		html += fmt.Sprintf(`<tr>
<td>%s</td>
<td>%d</td>
<td>%d</td>
<td>%.0f%%</td>
<td>£%.2f</td>
<td>%.0f%%</td>
<td class="%s"><strong>£%.2f</strong></td>
<td class="%s"><strong>%.0f%%</strong></td>
<td class="%s">%s £%.2f</td>
<td style="font-size: 11px;">%s</td>
</tr>`,
			opt.name, opt.qtyRemaining, opt.qtySold, opt.sellThrough,
			opt.currentPrice, opt.currentMargin,
			changeClass, opt.suggestedPrice,
			changeClass, opt.suggestedMargin,
			changeClass, changeSymbol, priceChange,
			opt.reasoning)
	}
	
	html += `</table>
<p style="font-size: 12px; color: #6b7280; margin-top: 10px;">
<strong>Note:</strong> Suggestions based on inventory levels and sales velocity. Test price changes gradually and monitor results.
</p>
</div>`
	
	return html
}
