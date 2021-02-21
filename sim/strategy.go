package sim

import "time"

type Strategy interface {
	tick(time.Time, portfolio)
}

type MidMonth struct {
	lastInvested time.Time
}

func (mm *MidMonth) tick(date time.Time, p portfolio) {
	if mm.lastInvested.Month() != date.Month() {
		if date.Day() > 14 {
			// Attempt invest
			err := p.rebalance(p.getCashBalance(), date)
			if err != nil {
				return
			}

			mm.lastInvested = date
		}
	}

}

func NewStrategy(startDate time.Time) Strategy {
	return &MidMonth{
		lastInvested: startDate.Add(-31 * 24 * time.Hour),
	}
}
