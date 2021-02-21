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
		// Maybe receive income
		p.transact(inc.tick(simDay))

		// Maybe invest
		strat.tick(simDay, p)

		// Increase day
		simDay = simDay.Add(time.Duration(24 * time.Hour))
	}

	return nil
}
