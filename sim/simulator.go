package sim

import (
	"errors"
	"log"
	"math"
	"time"
)

type priceProvider interface {
	GetPrice(string, time.Time) (float64, error)
}

func Simulate(start time.Time, p Portfolio, inc Income, strat Strategy) (pValues []float64, dates []string, err error) {
	if time.Now().Sub(start) < 0 {
		err = errors.New("Start lies in the future")
		return
	}

	p.SetStart(start)

	simDay := start
	// Simulate until reaching the current date
	for time.Now().Sub(simDay) > 0 {
		// Maybe receive income
		amount := inc.tick(simDay)
		if amount != 0.0 {
			p.transact(&incomeTransaction{date: simDay, amount: amount})
		}

		// Maybe invest
		strat.tick(simDay, p)

		// Maybe evaluate
		if simDay.Day() == 1 {
			totalValue := math.Round(p.TotalValue((simDay)))
			pValues = append(pValues, totalValue)
			dates = append(dates, simDay.Format("2006/01/02"))
		}

		// Increase day
		simDay = simDay.Add(time.Duration(24 * time.Hour))
	}

	return
}

func SimulateStratOnRef(startDate time.Time, symbol string, strat Strategy, fixedFees float64, varFees float64) ([]float64, []string, float64) {
	p := getRefPortfolio(symbol, fixedFees, varFees)

	inc := NewIncome(startDate, 1000.0)

	pValues, dates, err := Simulate(startDate, p, inc, strat)
	if err != nil {
		log.Fatal(err)
	}

	irr := p.CalcIRR(time.Now())

	return pValues, dates, irr
}

func getRefPortfolio(symbol string, fixedFees float64, varFees float64) Portfolio {
	sACWI := &Stock{Symbol: symbol}

	stocks := map[*Stock]int64{
		sACWI: 0,
	}

	goalRatios := map[*Stock]float64{
		sACWI: 1.0,
	}

	startCash := 0.0
	p, err := NewMultiPortfolio(startCash, stocks, goalRatios, fixedFees, varFees)
	if err != nil {
		log.Fatal(err)
	}

	return p
}
