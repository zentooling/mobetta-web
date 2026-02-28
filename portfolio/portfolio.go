// Package portfolio implements the portfolio dashboard
package portfolio

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zentooling/golang-web-server/infra"
	"github.com/zentooling/golang-web-server/routes"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

type Service struct {
	env infra.ILair
}

type PortfolioData struct {
	routes.PageData
	Script template.HTML
	// Script  string
	Element template.HTML
}

func NewService(env infra.ILair) *Service {
	return &Service{env: env}
}

// ShowPortfolio Display current pportfolio and performance
func (svc Service) ShowPortfolio(c *gin.Context) {
	pd := routes.DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter)

	// build chart
	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		// TextAlign: "center",
		Title: "My Bar Chart",
	}))
	bar.SetXAxis([]string{"Jan", "Feb", "Mar"}).
		AddSeries("Sales", []opts.BarData{
			{Value: 25}, {Value: 50}, {Value: 75},
		})

	snippet := bar.RenderSnippet()

	element := snippet.Element
	script := snippet.Script
	option := snippet.Option
	css := `
	<style>
	#chart-container {
		/* Set a fixed size for the container */
		width: 600px;
		height: 400px;

		/* Center the container itself within its parent (e.g., the body or another wrapper) */
		// margin: 0 auto;

		/* Or if using a flex parent on the page: */
		display: flex; 
		justify-content: center; 
		align-items: center;
	}
	</style>
		`

	fmt.Printf("element %s\n", element)
	fmt.Printf("script %s\n", script)
	fmt.Printf("option %s\n", option)

	pd.Title = pd.Trans("Portfolio")
	portfolioData := PortfolioData{
		PageData: pd,
		Script:   template.HTML(script),
		Element:  template.HTML(element + css),
	}
	c.HTML(http.StatusOK, "portfolio.gohtml", portfolioData)
}
