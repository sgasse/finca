package sim

import (
	"time"
)

// A Strategy is an interface which triggers reinvestments based on some
// criteria held in its internal state. It gets the current date to evaluate
// passed along with a Portfolio to trigger rebalancing if the criteria are
// met.
type Strategy interface {
	tick(time.Time, Portfolio)
}

// This struct implements the Strategy interface and triggers rebalancing
// always on `minDay` of the month or the first evaluation day after the
// `minDay`.
type MidMonth struct {
	lastInvested time.Time
	minDay       int
}

// A FixedMonths strategy triggers rebalancing always on `minDay` or the first
// evaluation day after `minDay`, but only in months given in `investMonths`.
type FixedMonths struct {
	investMonths map[time.Month]bool
	lastInvested time.Time
	minDay       int
}

// NoInvest is an empty strategy implementing the `Strategy` interface but
// never trigering any investment. It serves as baseline comparison for other
// strategies.
type NoInvest struct {
}

// WithDrawdown is a type to be embedded in other strategies. It holds the state
// and provides methods to evaluate if a certain minimum drawdown was reached as
// investment criterion.
type WithDrawdown struct {
	relVal    float64
	refSymbol string
	priceP    priceProvider
	lastTop   float64
}

// MinDrawdown triggers an investment when a certain minimum drawdown is
// reached. In the current implementation, it wraps `WithDrawdown`.
type MinDrawdown struct {
	WithDrawdown
}

// AdaptivePeriodic is a hybrid strategy which triggers investments periodically
// after `waitTime` but also invests earlier if a certain minimum drawdown is
// reached.
type AdaptivePeriodic struct {
	waitTime     time.Duration
	lastInvested time.Time
	WithDrawdown
}

// NewMonthlyStrategy creates a new strategy investing monthly on the 14th or
// the first evaluation day after the 14th.
func NewMonthlyStrategy(startDate time.Time) Strategy {
	return &MidMonth{
		lastInvested: startDate.Add(-31 * 24 * time.Hour),
		minDay:       14,
	}
}

// NewFixedMonthsStrategy creates a new strategy investing on the 14th of every
// month given in `month` or the first evaluation day after the 14th.
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

// NewMinDrawdown creates a new strategy investing when the stock behind
// `refSymbol` suffered a minimum relative drawdown to `relVal` from its last
// known top value.
func NewMinDrawdown(relVal float64, refSymbol string, priceP priceProvider) Strategy {
	return &MinDrawdown{WithDrawdown{relVal, refSymbol, priceP, 0.0}}
}

// NewAdaptivePeriodic creates a new strategy investing either when `waitTime`
// has passed since the last investment or when `refSymbol` suffered a minimum
// relative drawdown to `relVal` from its last known top value.
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
