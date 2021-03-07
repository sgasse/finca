package sim

import (
	"log"
	"time"
)

func getRefPortfolio() Portfolio {
	sACWI := &Stock{Symbol: "SPY"}

	stocks := map[*Stock]int64{
		sACWI: 0,
	}

	goalRatios := map[*Stock]float64{
		sACWI: 1.0,
	}

	startCash := 0.0
	p, err := NewMultiPortfolio(startCash, stocks, goalRatios)
	if err != nil {
		log.Fatal(err)
	}

	return p
}

func SimulateStrategyOnRef(startDate time.Time, strat Strategy) ([]float64, []string, float64) {
	p := getRefPortfolio()

	inc := NewIncome(startDate, 1000.0)

	pValues, dates, err := Simulate(startDate, p, inc, strat)
	if err != nil {
		log.Fatal(err)
	}

	irr := p.CalcIRR(time.Now())

	return pValues, dates, irr
}
