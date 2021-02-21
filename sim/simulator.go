package sim

import (
	"errors"
	"time"
)

func Simulate(p portfolio, inc Income, start time.Time, strat Strategy) error {
	if time.Now().Sub(start) < 0 {
		return errors.New("Start lies in the future")
	}

	simDay := start
	// Simulate until reaching the current date
	for time.Now().Sub(simDay) > 0 {
		// Tick income
		p.transact(inc.tick(simDay))

		// Tick invest strategy
		strat.tick(simDay, p)

		// Increase day
		simDay = simDay.Add(time.Duration(24 * time.Hour))
	}

	return nil
}
