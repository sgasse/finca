package sim

import (
	"errors"
	"math"
	"time"
)

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
