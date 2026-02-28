package main

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/thedatashed/xlsxreader"
	"github.com/zentooling/golang-web-server/infra"
	"github.com/zentooling/golang-web-server/models"
	"gorm.io/gorm"
)

func main() {

	cwd, _ := os.Getwd()
	slog.Info("Run Portfolio cwd: " + cwd)
	// Create an instance of the reader by opening a target file
	xl, err := xlsxreader.OpenFile("./portfolio.xlsx")
	if err != nil {
		slog.Error("Run", "error", err)
		return
	}

	defer func(xl *xlsxreader.XlsxFileCloser) {
		err := xl.Close()
		if err != nil {
			slog.Error("Run", "error", err)
		}
	}(xl)

	records := make([]*models.Portfolio, 0)

	// Iterate on the rows of data
	for row := range xl.ReadRows(xl.Sheets[0]) {

		if row.Cells[0].Value == "ticker" {
			slog.Info("skipping header row")
			continue
		}

		if len(row.Cells) < 9 { // loose check - rows without calculated columns (pctAlloc, e.g.) indicate end of portfolio records
			slog.Info("encountered first non portfolio row. Ending iteration")
			break
		}

		records = append(records, XlsToRecord(row))
	}

	// get db connection

	conf := infra.LoadEnvVariables()

	db, err := infra.ConnectToDatabase(conf)
	if err != nil {
		slog.Error("Run", "error", err)
		os.Exit(2)
	}

	// soft delete previous records
	deletePortfolioRecords(db)

	// add new
	insertRecords(db, records)

}

func insertRecords(db *gorm.DB, records []*models.Portfolio) {
	result := db.Create(records) // pass a slice to insert multiple row

	if result.Error != nil {
		slog.Error("Run", "error", result.Error)
	}
}

func deletePortfolioRecords(db *gorm.DB) {

	db.Where("1 = 1").Delete(&models.Portfolio{})
	// SQL: DELETE FROM portfolio WHERE 1 = 1;

}

func XlsToRecord(row xlsxreader.Row) *models.Portfolio {

	ticker := row.Cells[1].Value

	rank, err := strconv.ParseFloat(row.Cells[2].Value, 32)
	if err != nil {
		slog.Error("Error during conversion: %v\n", err)
	}
	clsPrice, err := strconv.ParseFloat(row.Cells[3].Value, 32)
	if err != nil {
		slog.Error("Error during conversion: %v\n", err)
	}
	volatility, err := strconv.ParseFloat(row.Cells[4].Value, 32)
	if err != nil {
		slog.Error("Error during conversion: %v\n", err)
	}
	clsGtMa := false
	if row.Cells[5].Value == "1" {
		clsGtMa = true
	}
	gap := false
	if row.Cells[6].Value == "1" {
		gap = true
	}

	pctAlloc, err := strconv.ParseFloat(row.Cells[7].Value, 32)
	if err != nil {
		slog.Error("Error during conversion: %v\n", err)
	}
	cost, err := strconv.ParseFloat(row.Cells[8].Value, 32)
	if err != nil {
		slog.Error("Error during conversion: %v\n", err)
	}
	numShares, err := strconv.ParseFloat(row.Cells[9].Value, 32)
	if err != nil {
		slog.Error("Error during conversion: %v\n", err)
	}

	return &models.Portfolio{
		Ticker:       ticker,
		Rank:         rank,
		Close:        clsPrice,
		Volatility:   volatility,
		CloseAboveMa: clsGtMa,
		Gap:          gap,
		PctAlloc:     pctAlloc,
		Cost:         cost,
		NumShares:    numShares,
	}
}
