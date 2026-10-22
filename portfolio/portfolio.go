// Package portfolio implements the portfolio dashboard
package portfolio

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/zentooling/golang-web-server/infra"
	"github.com/zentooling/golang-web-server/models"
	"github.com/zentooling/golang-web-server/routes"
)

type Service struct {
	env infra.ILair
}

type PortfolioData struct {
	routes.PageData
	Holdings *[]models.Portfolio
	Script   template.HTML
	// Script  string
	Element template.HTML
}

func NewService(env infra.ILair) *Service {
	return &Service{env: env}
}

// ShowPortfolio Display current portfolio and performance
func (svc Service) ShowPortfolio(c *gin.Context) {
	pd := routes.DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter)
	snippet := portfolioPerfChart().RenderSnippet()

	element := snippet.Element
	script := snippet.Script
	option := snippet.Option

	fmt.Printf("element %s\n", element)
	fmt.Printf("script %s\n", script)
	fmt.Printf("option %s\n", option)

	conf := infra.LairInstance().GetConfig()

	db, err := infra.ConnectToDatabase(conf)
	if err != nil {
		slog.Error("Run", "error", err)
		os.Exit(2)
	}

	repo := NewHoldingsRepository(db)

	ctxt := context.Background()

	holdings, err := repo.GetAll(ctxt)

	pd.Title = pd.Trans("Portfolio")
	portfolioData := PortfolioData{
		PageData: pd,
		Script:   template.HTML(script),
		Element:  template.HTML(element),
		Holdings: holdings,
	}
	c.HTML(http.StatusOK, "portfolio.gohtml", portfolioData)
}

// percentChangeZeroStartSeries computes the percentage change of each series value
// relative to the first (baseline=0) element in the series.
func percentChangeZeroStartSeries(series []float64) ([]float64, error) {

	sz := len(series)
	if sz == 0 {
		return nil, errors.New("the stock series cannot be empty")
	}

	// Define the starting baseline price
	baseValue := series[0]
	if baseValue == 0 {
		return nil, errors.New("cannot calculate percentage change: initial stock value is zero")
	}

	// Initialize result slice with the exact capacity needed
	pctChanges := make([]float64, sz)

	// The starting percentage change is always 0.0
	pctChanges[0] = 0.0

	// Calculate the change relative to the initial value for the rest of the series
	for i := 1; i < sz; i++ {
		pctChanges[i] = ((series[i] - baseValue) / baseValue) * 100
	}

	return pctChanges, nil
}

// percentChangeSeries calculates the percentage change between each consecutive pair
// in a slice of stock values.
// Formula: ((current - previous) / previous) * 100
func percentChangeSeries(prices []float64) ([]float64, error) {

	sz := len(prices)
	if sz < 2 {
		return nil, fmt.Errorf("slice must contain at least 2 prices")
	}

	// since change is calculated as delta from previous value
	// we need slice 1 less in size for results
	changes := make([]float64, sz-1)
	// starting here at 1 !
	for i := 1; i < sz; i++ {
		prev := prices[i-1]
		curr := prices[i]
		if prev == 0 {
			return nil, fmt.Errorf("previous price cannot be zero at index %d", i-1)
		}
		change := ((curr - prev) / prev) * 100.0
		changes[i-1] = change
	}

	return changes, nil
}

func floatToLineData(prices []float64) []opts.LineData {

	sz := len(prices)
	// preallocate for perf
	lineData := make([]opts.LineData, sz)
	for i := range sz {
		ld := opts.LineData{
			Value: prices[i],
		}
		lineData[i] = ld
	}

	return lineData
}

func portfolioPerfChart() *charts.Line {

	dates := []string{"1/10/26", "2/10/26", "3/10/26", "4/10/26", "5/10/26", "6/10/26", "7/10/26", "1/11/26", "2/11/26", "3/11/26", "4/11/26", "5/11/26", "6/10/26", "7/10/26", "1/10/26", "2/10/26", "3/10/26", "4/10/26", "5/10/26", "6/10/26", "7/10/26", "1/10/26", "2/10/26", "3/10/26", "4/10/26", "5/10/26", "6/10/26", "7/10/26"}

	rawStockA := []float64{150.2, 152.45, 151.1, 153.9, 155.2, 154.0, 156.8, 150.2, 152.45, 151.1, 153.9, 155.2, 154.0, 156.8, 150.2, 152.45, 151.1, 153.9, 155.2, 154.0, 156.8, 150.2, 152.45, 151.1, 153.9, 155.2, 154.0, 156.8}
	rawStockB := []float64{250.2, 345.45, 275.1, 153.9, 255.2, 354.0, 156.8, 250.2, 345.45, 275.1, 153.9, 255.2, 354.0, 156.8, 250.2, 345.45, 275.1, 153.9, 255.2, 354.0, 156.8, 250.2, 345.45, 275.1, 153.9, 255.2, 354.0, 156.8}

	percentageA, err := percentChangeZeroStartSeries(rawStockA)
	if err != nil {
		panic(err)
	}

	fmt.Println(percentageA)

	percentageB, err2 := percentChangeZeroStartSeries(rawStockB)
	if err2 != nil {
		panic(err2)
	}
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			// Title:    "Performance vs S&P 500",
			Subtitle: "relative to baseline 0%",
			// Left:     "center",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:            opts.Bool(true),
			Trigger:         "axis",
			BackgroundColor: "rgba(255, 255, 155, 0.95)",
			AxisPointer:     &opts.AxisPointer{Type: "cross"},
			Formatter: opts.FuncOpts(`function (params) {
				let res = '<strong>' + params[0].name + '</strong><br/>';
				params.forEach(function (item) {
					res += item.marker + item.seriesName + ': ' +  (item.value > 0 ? '+' : '') + parseFloat(item.value).toFixed(2) + '%<br/>';
				});
				return res;
			}`),
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: opts.Bool(true),
			// Top: "12%",
			// auto populated? Data:           []string{"Category A", "Category B"},
		}),
		charts.WithGridOpts(opts.Grid{
			Left:  "%20",
			Right: "%20",
			// Bottom:       "15%",
			ContainLabel: opts.Bool(true),
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "category",
			Name: "Date",
			// BoundaryGap: opts.Bool(true),
			Scale: opts.Bool(true),
		}),
		charts.WithYAxisOpts(opts.YAxis{
			// Type: "value",
			//SplitLine: &opts.SplitLine{
			//	LineStyle: &opts.LineStyle{
			//		Type: "dashed",
			//	},
			//},
			AxisLabel: &opts.AxisLabel{
				Formatter: "{value}%",
			},
		}),
		charts.WithDataZoomOpts(
			opts.DataZoom{
				Type:  "slider",
				Start: 0,
				End:   100,
			},
			//opts.DataZoom{
			//	Type:  "inside",
			//	Start: 0,
			//	End:   100,
			//}
		),

		charts.WithInitializationOpts(opts.Initialization{
			Width:   "100%", // fill the parent width
			Height:  "400px",
			ChartID: "perf_chart", // use this id to listen for resize events - see the html template
			Theme:   types.ThemePurplePassion,
		}),
	)

	lineDataA := floatToLineData(percentageA)
	lineDataB := floatToLineData(percentageB)

	fmt.Println(lineDataA)

	line.SetXAxis(dates)

	line.AddSeries("Mobetta", lineDataA)

	line.AddSeries("S&P 500", lineDataB)

	// call after series loaded!
	line.SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: opts.Bool(true),
		}),
	)
	return line
}
