package sim

import (
	"errors"
	"testing"
	"time"
)

type fakePortfolio struct {
	cash                 float64
	calledRebalance      bool
	calledGetCashBalance bool
	calledTransact       bool
	reinvestArg          float64
	dateArg              time.Time
	amountArg            float64
	errOnRebalance       bool
}

func (fp *fakePortfolio) rebalance(reinvest float64, date time.Time) error {
	fp.calledRebalance = true
	fp.reinvestArg = reinvest
	fp.dateArg = date
	if fp.errOnRebalance {
		return errors.New("Error")
	}
	return nil
}

func (fp *fakePortfolio) getCashBalance() float64 {
	fp.calledGetCashBalance = true
	return fp.cash
}

func (fp *fakePortfolio) transact(amount float64) {
	fp.calledTransact = true
	fp.amountArg = amount
	fp.cash += amount
}

func TestNewStrategy(t *testing.T) {
	startDate := time.Date(2020, 6, 1, 12, 0, 0, 0, time.UTC)
	strat := NewMonthlyStrategy(startDate)

	if midM, ok := strat.(*MidMonth); ok {
		lastInvDate := midM.lastInvested.Format("2006-01-02")
		expectedDate := "2020-05-01"
		if lastInvDate != expectedDate {
			t.Error("Expected lastInvested to be initialized to ", expectedDate, ", got ", lastInvDate)
		}
	}
}

func TestStrategyTick(t *testing.T) {
	startDate := time.Date(2020, 6, 1, 12, 0, 0, 0, time.UTC)
	strat := NewMonthlyStrategy(startDate)

	if midM, ok := strat.(*MidMonth); ok {
		fp1 := &fakePortfolio{cash: 333.0}
		midM.tick(time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC), fp1)
		assertEqual(t, fp1.calledRebalance, false, "Should not have called rebalance")

		investDay := time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC)
		midM.tick(investDay, fp1)
		assertEqual(t, fp1.calledRebalance, true, "Should have called rebalance")
		assertEqual(t, fp1.calledGetCashBalance, true, "Should have called getCashBalance")
		assertEqual(t, fp1.reinvestArg, 333.0, "Wrong reinvest sum")
		assertEqual(t, fp1.dateArg, investDay, "Wrong investment day")

		fp2 := &fakePortfolio{cash: 123.0, errOnRebalance: true}
		midM.tick(time.Date(2020, 8, 16, 0, 0, 0, 0, time.UTC), fp2)
		assertEqual(t, fp2.calledRebalance, true, "Should have called rebalance")
		assertEqual(t, fp2.calledGetCashBalance, true, "Should have called getCashBalanace")
		assertEqual(t, midM.lastInvested, investDay, "lastInvested should not have changed")
	}
}
