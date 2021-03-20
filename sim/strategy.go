package sim

import (
	"time"
)

type Strategy interface {
	tick(time.Time, Portfolio)
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

type NoInvest struct {
}

type MinDrawdown struct {
	LastTop   float64
	RelVal    float64
	RefSymbol string
	PriceP    priceProvider
}

func NewMonthlyStrategy(startDate time.Time) Strategy {
	return &MidMonth{
		lastInvested: startDate.Add(-31 * 24 * time.Hour),
		minDay:       14,
	}
}

func (mm *MidMonth) tick(date time.Time, p Portfolio) {
	if !investedThisMonth(date, mm.lastInvested) {
		if date.Day() >= mm.minDay {
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
		// Go 31 days back
		lastInvested: startDate.Add(-31 * 24 * time.Hour),
		minDay:       14,
	}
}

func (fm *FixedMonths) tick(date time.Time, p Portfolio) {
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

func (s *NoInvest) tick(date time.Time, p Portfolio) {
}

func (s *MinDrawdown) tick(date time.Time, p Portfolio) {
	curVal, err := s.PriceP.GetPrice(s.RefSymbol, date)
	if err != nil {
		return
	}

	if curVal > s.LastTop {
		s.LastTop = curVal
		return
	}

	if curVal/s.LastTop <= s.RelVal {
		// Drawdown reached, rebalance
		err := p.rebalance(p.getCashBalance(), date)
		if err != nil {
			return
		}

		s.LastTop = curVal
	}
}

func investedThisMonth(date time.Time, lastInvested time.Time) bool {
	if lastInvested.Year() == date.Year() &&
		lastInvested.Month() == date.Month() {
		return true
	}
	return false
}
