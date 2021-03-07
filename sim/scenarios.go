package sim

import (
	"log"
	"time"
)

func getRefPortfolio() Portfolio {
	sACWI := &Stock{Symbol: "SPYI.DE"}

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

func SimulateMonthly(startDate time.Time) ([]float64, []string) {
	// startDate := time.Date(2011, 6, 1, 10, 0, 0, 0, time.UTC)
	p := getRefPortfolio()

	inc := NewIncome(startDate, 1000.0)

	strat := NewMonthlyStrategy(startDate)

	pValues, dates, err := Simulate(startDate, p, inc, strat)
	if err != nil {
		log.Fatal(err)
	}

	return pValues, dates
}

func SimulateBiYearly(startDate time.Time, firstMonth time.Month) ([]float64, []string) {
	p := getRefPortfolio()

	inc := NewIncome(startDate, 1000.0)

	strat := NewFixedMonthsStrategy(startDate, []time.Month{firstMonth, firstMonth + 6})

	pValues, dates, err := Simulate(startDate, p, inc, strat)
	if err != nil {
		log.Fatal(err)
	}

	return pValues, dates
}
