package sim

import "time"

type Strategy interface {
	tick(time.Time, portfolio)
}

type MidMonth struct {
	lastInvested time.Time
	minDay       int
}

type FixedMonths struct {
	investMonths map[time.Month]bool
	lastInvested time.Time
	minDay       int
}

func NewMonthlyStrategy(startDate time.Time) Strategy {
	return &MidMonth{
		lastInvested: startDate.Add(-31 * 24 * time.Hour),
		minDay:       14,
	}
}

func (mm *MidMonth) tick(date time.Time, p portfolio) {
	if !investedThisMonth(date, mm.lastInvested) {
		if date.Day() > mm.minDay {
			// Attempt invest
			err := p.rebalance(p.getCashBalance(), date)
			if err != nil {
				return
			}

			mm.lastInvested = date
		}
	}
}

func NewFixedMonthsStrategy(startDate time.Time, months []time.Month) Strategy {
	invMonths := map[time.Month]bool{}
	for _, m := range months {
		invMonths[m] = true
	}
	return &FixedMonths{
		investMonths: invMonths,
		lastInvested: startDate.Add(-365 * 24 * time.Hour),
		minDay:       14,
	}
}

func (fm *FixedMonths) tick(date time.Time, p portfolio) {
	if _, ok := fm.investMonths[date.Month()]; ok {
		if !investedThisMonth(date, fm.lastInvested) {
			if date.Day() >= fm.minDay {
				// Attempt invest
				err := p.rebalance(p.getCashBalance(), date)
				if err != nil {
					return
				}

				fm.lastInvested = date
			}
		}
	}
}

func investedThisMonth(date time.Time, lastInvested time.Time) bool {
	if lastInvested.Month() == date.Month() {
		return true
	}
	return false
}
