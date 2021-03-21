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

type WithDrawdown struct {
	relVal    float64
	refSymbol string
	priceP    priceProvider
	lastTop   float64
}

type MinDrawdown struct {
	WithDrawdown
}

type AdaptivePeriodic struct {
	waitTime     time.Duration
	lastInvested time.Time
	WithDrawdown
}

func NewMonthlyStrategy(startDate time.Time) Strategy {
	return &MidMonth{
		lastInvested: startDate.Add(-31 * 24 * time.Hour),
		minDay:       14,
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

func NewMinDrawdown(relVal float64, refSymbol string, priceP priceProvider) Strategy {
	return &MinDrawdown{WithDrawdown{relVal, refSymbol, priceP, 0.0}}
}

func NewAdaptivePeriodic(startDate time.Time, waitTime time.Duration,
	relVal float64, refSymbol string, priceP priceProvider) Strategy {

	return &AdaptivePeriodic{
		waitTime:     waitTime,
		lastInvested: startDate.Add(-waitTime),
		WithDrawdown: WithDrawdown{
			relVal:    relVal,
			refSymbol: refSymbol,
			priceP:    priceP,
		},
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
	drawdownReached, curVal := s.drawdownTick(date)

	if drawdownReached {
		// Attempt invest
		err := p.rebalance(p.getCashBalance(), date)
		if err != nil {
			return
		}

		s.lastTop = curVal
	}
}

func (s *AdaptivePeriodic) tick(date time.Time, p Portfolio) {
	waitOver := date.Sub(s.lastInvested) >= s.waitTime
	drawdownReached, curVal := s.drawdownTick(date)

	if waitOver || drawdownReached {
		// Attempt invest
		err := p.rebalance(p.getCashBalance(), date)
		if err != nil {
			return
		}

		s.lastInvested = date
		s.lastTop = curVal
	}
}

func (wd *WithDrawdown) drawdownTick(date time.Time) (reached bool, curVal float64) {
	curVal, err := wd.priceP.GetPrice(wd.refSymbol, date)
	if err != nil {
		return
	}

	if curVal > wd.lastTop {
		wd.lastTop = curVal
		return
	}

	if curVal/wd.lastTop <= wd.relVal {
		reached = true
		return
	}

	return
}

func investedThisMonth(date time.Time, lastInvested time.Time) bool {
	if lastInvested.Year() == date.Year() &&
		lastInvested.Month() == date.Month() {
		return true
	}
	return false
}
