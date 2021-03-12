package analyze

import (
	"errors"
	"log"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/sgasse/finca/av"
	"github.com/sgasse/finca/sim"
)

type SimResults struct {
	Dates      []string
	TimeSeries map[string][]float64
	IRR        map[string]float64
}

func evalSingleStockData(startDate time.Time, symbol string) (dates []string, timeSeries []float64, relChange []float64, maxDD []float64) {
	curDay := startDate
	for time.Now().Sub(curDay) > 0 {
		price, err := av.GetPrice(symbol, curDay)
		if err == nil {
			price := roundTo(2, price)
			timeSeries = append(timeSeries, price)
			dates = append(dates, curDay.Format("2006-01-02"))
		}

		curDay = curDay.Add(time.Duration(24 * time.Hour))
	}

	lastMax := 0.0
	// Change to first day equal to zero
	lastPrice := timeSeries[0]

	for _, price := range timeSeries {
		if price >= lastMax {
			maxDD = append(maxDD, 0.0)
			lastMax = price
		} else {
			drawdown := roundTo(2, (price/lastMax-1.0)*100)
			maxDD = append(maxDD, drawdown)
		}

		percChange := roundTo(2, (price/lastPrice-1.0)*100)
		relChange = append(relChange, percChange)
		lastPrice = price
	}

	return
}

func getStartDate(symbol string) (sDate time.Time, err error) {
	earliest, _, err := av.GetDateRange(symbol)
	if err != nil {
		return
	}

	sDate, err = time.Parse("2006-01-02", earliest)
	if err != nil {
		return
	}

	// Shift to beginning of next month
	sDate = time.Date(sDate.Year(), sDate.Month()+1, 1, 12, 0, 0, 0, time.UTC)
	return
}

func addSimResults(simRes *SimResults, strat sim.Strategy, name string) error {
	pValues, dates, irr := sim.SimulateStrategyOnRef(startDate, symbol, strat)
	if len(simRes.Dates) == 0 {
		simRes.Dates = dates
	} else if len(simRes.Dates) != len(dates) {
		return errors.New("Simulation dates do not agree")
	}

	simRes.TimeSeries[name] = pValues
	if irr >= 400.0 {
		irr = 0
	}
	simRes.IRR[name] = irr

	return nil

}

func maybeSetSymbol(r *http.Request) error {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return err
	}
	log.Println("Got params: ", params)

	prevSymbol := symbol
	inSym, ok := params["symbol"]
	if ok {
		symbol = inSym[0]
	}

	sDate, err := getStartDate(symbol)
	if err != nil {
		symbol = prevSymbol
		return err
	}

	startDate = sDate
	return nil
}

func roundTo(digits float64, number float64) float64 {
	factor := math.Pow(10, digits)
	return math.Round(number*factor) / factor
}
