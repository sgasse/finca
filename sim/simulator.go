package sim

import (
	"errors"
	"time"
)

func Simulate(start time.Time, p Portfolio, inc Income, strat Strategy) error {
	if time.Now().Sub(start) < 0 {
		return errors.New("Start lies in the future")
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

		// Increase day
		simDay = simDay.Add(time.Duration(24 * time.Hour))
	}

	return nil
}
