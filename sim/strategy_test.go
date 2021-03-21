package sim

import (
	"errors"
	"testing"
	"time"

	"github.com/sgasse/finca/av"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testDate struct {
	date   time.Time
	invest bool
}

type mockPortfolio struct {
	mock.Mock
	cash float64
}

func (m *mockPortfolio) SetStart(date time.Time) {
	_ = m.Called(date)
}

func (m *mockPortfolio) TotalValue(date time.Time) float64 {
	args := m.Called(date)
	return args.Get(0).(float64)
}

func (m *mockPortfolio) CalcIRR(date time.Time) float64 {
	args := m.Called(date)
	return args.Get(0).(float64)
}

func (m *mockPortfolio) getCashBalance() float64 {
	args := m.Called()
	return args.Get(0).(float64)
}

func (m *mockPortfolio) transact(tr transaction) {
	_ = m.Called(tr)
}

func (m *mockPortfolio) rebalance(reinvest float64, date time.Time) error {
	args := m.Called(reinvest, date)
	return args.Error(0)
}

type mockPriceProvider struct {
	mock.Mock
}

func (m *mockPriceProvider) GetPrice(symbol string, date time.Time) (float64, error) {
	args := m.Called(symbol, date)
	return args.Get(0).(float64), args.Error(1)
}

func TestNewMonthlyStrategy(t *testing.T) {
	startDate := time.Date(2020, 06, 03, 23, 55, 1, 0, time.UTC)
	strat := NewMonthlyStrategy(startDate)
	if monStrat, ok := strat.(*MidMonth); ok {
		assert.Equal(t, 14, monStrat.minDay, "MinDay wrong")
		assert.Equal(t,
			time.Date(2020, 05, 03, 23, 55, 1, 0, time.UTC),
			monStrat.lastInvested,
			"lastInvested should be one month back.")
	}
}

func TestNewFixedMonthsStrategy(t *testing.T) {
	startDate := time.Date(2020, 06, 03, 23, 55, 1, 0, time.UTC)
	strat := NewFixedMonthsStrategy(startDate, []time.Month{3, 9})
	if fixedStrat, ok := strat.(*FixedMonths); ok {
		assert.Equal(t, 14, fixedStrat.minDay, "MinDay wrong")
		if startDate.Sub(fixedStrat.lastInvested) != (31 * 24 * time.Hour) {
			t.Error("lastInvested should be one month back.")
		}
		assert.Contains(t, fixedStrat.investMonths, time.Month(3),
			"Month March missing")
		assert.Contains(t, fixedStrat.investMonths, time.Month(9),
			"Month September missing")
	}
}

func TestNewAdaptivePeriodic(t *testing.T) {
	startDate := time.Date(2020, 05, 02, 13, 12, 59, 0, time.UTC)
	waitTime := time.Duration(182*24) * time.Hour
	strat := NewAdaptivePeriodic(
		startDate,
		waitTime,
		0.7,
		"TEST.DE",
		&av.AvProvider{})

	assert.Equal(
		t,
		strat.(*AdaptivePeriodic).lastInvested,
		startDate.Add(-waitTime),
		"lastInvested should be `waitTime` back in the past",
	)

}

func TestMidMonthTick(t *testing.T) {
	startDate := time.Date(2020, 06, 03, 23, 55, 1, 0, time.UTC)

	for simDay := startDate; simDay.Day() < 31; simDay = simDay.Add(
		time.Duration(6) * time.Hour) {

		p := newMockP(simDay)

		strat := NewMonthlyStrategy(startDate)
		strat.tick(simDay, p)

		if simDay.Day() < 14 {
			// No call before the 14th of the month
			p.AssertNotCalled(t, "getCashBalance")
			p.AssertNotCalled(t, "rebalance", 2000.0, simDay)
		} else {
			// Call on 14th and any other later day if not called before
			p.AssertExpectations(t)
		}

	}

	// Test to not call after calling once in the month
	simDay := time.Date(2020, 06, 13, 12, 55, 1, 0, time.UTC)
	strat := NewMonthlyStrategy(startDate)

	// No call on the day before the invest day
	p := newMockP(simDay)
	strat.tick(simDay, p)
	p.AssertNotCalled(t, "getCashBalance")
	p.AssertNotCalled(t, "rebalance", 2000.0, simDay)

	// Invest on the 14th
	simDay = simDay.Add(time.Duration(24) * time.Hour)
	p = newMockP(simDay)
	strat.tick(simDay, p)
	p.AssertExpectations(t)

	// Do not invest again one day later
	simDay = simDay.Add(time.Duration(24) * time.Hour)
	p = newMockP(simDay)
	strat.tick(simDay, p)
	p.AssertNotCalled(t, "getCashBalance")
	p.AssertNotCalled(t, "rebalance", 2000.0, simDay)

	// Test error on rebalance
	startDate = time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC)
	strat = NewMonthlyStrategy(startDate)
	tmpLastInv := strat.(*MidMonth).lastInvested

	date := time.Date(2020, 1, 14, 1, 1, 1, 0, time.UTC)

	p = &mockPortfolio{}
	p.On("getCashBalance").Return(2000.0)
	p.On("rebalance", 2000.0, date).Return(errors.New("Test error"))
	strat.tick(date, p)
	p.AssertExpectations(t)
	assert.Equal(t, tmpLastInv, strat.(*MidMonth).lastInvested,
		"lastInvested should remain unchanged.")
}

func TestFixedMonthsTick(t *testing.T) {
	startDate := time.Date(2020, 05, 15, 11, 24, 1, 0, time.UTC)

	expected := []testDate{
		// Invest on these days individually
		{time.Date(2020, 10, 15, 11, 24, 1, 0, time.UTC), true},
		{time.Date(2020, 10, 19, 11, 24, 1, 0, time.UTC), true},
		{time.Date(2021, 1, 14, 1, 4, 1, 0, time.UTC), true},
		{time.Date(2021, 1, 14, 11, 24, 1, 0, time.UTC), true},
		{time.Date(2021, 1, 31, 11, 24, 1, 0, time.UTC), true},
		{time.Date(2021, 4, 14, 11, 24, 1, 0, time.UTC), true},
		{time.Date(2021, 4, 30, 11, 24, 1, 0, time.UTC), true},
		// No investment on these days
		{time.Date(2020, 10, 5, 11, 24, 1, 0, time.UTC), false},
		{time.Date(2020, 10, 9, 11, 24, 1, 0, time.UTC), false},
		{time.Date(2021, 9, 14, 1, 4, 1, 0, time.UTC), false},
		{time.Date(2021, 8, 19, 11, 24, 1, 0, time.UTC), false},
		{time.Date(2021, 2, 31, 11, 24, 1, 0, time.UTC), false},
		{time.Date(2021, 4, 13, 11, 24, 1, 0, time.UTC), false},
		// this is actually May the first
		{time.Date(2021, 4, 31, 11, 24, 1, 0, time.UTC), false},
	}

	for _, td := range expected {
		// Expect investment on those days if called independently
		strat := NewFixedMonthsStrategy(startDate, []time.Month{1, 4, 10})

		p := newMockP(td.date)
		strat.tick(td.date, p)
		if td.invest {
			p.AssertExpectations(t)
		} else {
			p.AssertNotCalled(t, "getCashBalance")
			p.AssertNotCalled(t, "rebalance", 2000.0, td.date)
		}
	}

	// Go through a cycle with exemplary days
	startDate = time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC)
	strat := NewFixedMonthsStrategy(startDate, []time.Month{1, 4, 10})

	expected = []testDate{
		{time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC), false},
		{time.Date(2020, 1, 13, 1, 1, 1, 0, time.UTC), false},
		// Invest on minDay
		{time.Date(2020, 1, 14, 1, 1, 1, 0, time.UTC), true},
		// Do not invest again in the same month
		{time.Date(2020, 1, 15, 1, 1, 1, 0, time.UTC), false},
		{time.Date(2020, 1, 19, 1, 1, 1, 0, time.UTC), false},
		// Do not invest in months that are not mentioned
		{time.Date(2020, 2, 1, 1, 1, 1, 0, time.UTC), false},
		{time.Date(2020, 2, 14, 1, 1, 1, 0, time.UTC), false},
		{time.Date(2020, 2, 20, 1, 1, 1, 0, time.UTC), false},
		{time.Date(2020, 3, 20, 1, 1, 1, 0, time.UTC), false},
		// Wait for minDay in a correct month
		{time.Date(2020, 4, 1, 1, 1, 1, 0, time.UTC), false},
		{time.Date(2020, 4, 14, 1, 1, 1, 0, time.UTC), true},
		// Do not invest again in the same month
		{time.Date(2020, 4, 15, 1, 1, 1, 0, time.UTC), false},
		// Unrelated months - no investment
		{time.Date(2020, 5, 15, 1, 1, 1, 0, time.UTC), false},
		{time.Date(2020, 8, 15, 1, 1, 1, 0, time.UTC), false},
		// Beginning of correct month, wait for minDay
		{time.Date(2020, 10, 3, 1, 1, 1, 0, time.UTC), false},
		// Invest one day after minDay if minDay was not evaluated
		{time.Date(2020, 10, 15, 1, 1, 1, 0, time.UTC), true},
		// Do not invest again in the same month
		{time.Date(2020, 10, 17, 1, 1, 1, 0, time.UTC), false},
	}

	for _, td := range expected {
		p := newMockP(td.date)
		strat.tick(td.date, p)
		t.Log(td.date.Format("2006-01-02"))

		if td.invest {
			p.AssertExpectations(t)
		} else {
			p.AssertNotCalled(t, "getCashBalance")
			p.AssertNotCalled(t, "rebalance", 2000.0, td.date)
		}
	}

	// Test error on rebalance
	startDate = time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC)
	strat = NewFixedMonthsStrategy(startDate, []time.Month{1, 4, 10})
	tmpLastInv := strat.(*FixedMonths).lastInvested

	date := time.Date(2020, 1, 14, 1, 1, 1, 0, time.UTC)

	p := &mockPortfolio{}
	p.On("getCashBalance").Return(2000.0)
	p.On("rebalance", 2000.0, date).Return(errors.New("Test error"))
	strat.tick(date, p)
	p.AssertExpectations(t)
	assert.Equal(t, tmpLastInv, strat.(*FixedMonths).lastInvested,
		"lastInvested should remain unchanged.")

}

func TestNoInvestTick(t *testing.T) {
	strat := &NoInvest{}

	for date := time.Date(2020, 1, 1, 1, 1, 1, 0, time.UTC); date.Year() < 2021; date = date.Add(time.Duration(72) * time.Hour) {
		p := newMockP(date)
		strat.tick(date, p)
		p.AssertNotCalled(t, "getCashBalance")
		p.AssertNotCalled(t, "rebalance", 2000.0, date)
	}
}

func TestMinDrawDownTick(t *testing.T) {
	date := time.Date(2020, 05, 15, 11, 24, 1, 0, time.UTC)

	priceP := &mockPriceProvider{}

	strat := NewMinDrawdown(0.8, "TEST.DE", priceP)

	// Test error on fetching price
	priceP.On("GetPrice", "TEST.DE", date).Return(0.0, errors.New("Test error")).Once()
	p := newMockP(date)
	strat.tick(date, p)
	priceP.AssertExpectations(t)
	p.AssertNotCalled(t, "getCashBalance")
	p.AssertNotCalled(t, "rebalance", 2000.0, date)

	// Test new maximum value
	priceP.On("GetPrice", "TEST.DE", date).Return(100.0, nil).Once()
	p = newMockP(date)
	strat.tick(date, p)
	priceP.AssertExpectations(t)
	p.AssertNotCalled(t, "getCashBalance")
	p.AssertNotCalled(t, "rebalance", 2000.0, date)
	assert.Equal(t, 100.0, strat.(*MinDrawdown).lastTop, "New top price was not set correctly.")

	// Test no investment if drawdown not large enough
	priceP.On("GetPrice", "TEST.DE", date).Return(90.0, nil).Once()
	p = newMockP(date)
	strat.tick(date, p)
	priceP.AssertExpectations(t)
	p.AssertNotCalled(t, "getCashBalance")
	p.AssertNotCalled(t, "rebalance", 2000.0, date)
	assert.Equal(t, 100.0, strat.(*MinDrawdown).lastTop, "Top price should be unchanged.")

	// Test drawdown large enough and invest
	priceP.On("GetPrice", "TEST.DE", date).Return(79.5, nil).Once()
	p = newMockP(date)
	strat.tick(date, p)
	priceP.AssertExpectations(t)
	p.AssertExpectations(t)
	assert.Equal(t, 79.5, strat.(*MinDrawdown).lastTop, "Top price should be lowered.")

	// Test drawdown large enough and error on invest
	strat.(*MinDrawdown).lastTop = 100.0
	priceP.On("GetPrice", "TEST.DE", date).Return(79.5, nil).Once()
	p = &mockPortfolio{}
	p.On("getCashBalance").Return(2000.0)
	p.On("rebalance", 2000.0, date).Return(errors.New("Test error"))
	strat.tick(date, p)
	priceP.AssertExpectations(t)
	p.AssertExpectations(t)
	assert.Equal(t, 100.0, strat.(*MinDrawdown).lastTop, "Top price should be unchanged.")
}

func TestAdaptivePeriodic(t *testing.T) {
	date := time.Date(2020, 1, 14, 11, 24, 1, 0, time.UTC)
	waitTime := time.Duration(182*24) * time.Hour
	priceP := &mockPriceProvider{}

	strat := NewAdaptivePeriodic(date, waitTime, 0.8, "TEST.DE", priceP)

	// Invest on the start
	adaptiveInvest(date, 90.0, strat, t)

	// Do not invest in waitTime (small drawdown)
	date = time.Date(2020, 4, 15, 11, 24, 1, 0, time.UTC)
	adaptiveNoInvest(date, 85.0, strat, 90.0, t)

	// Invest after waitTime
	periodicInvDate := time.Date(2020, 7, 14, 11, 24, 1, 0, time.UTC)
	adaptiveInvest(periodicInvDate, 80.0, strat, t)

	// Take period invest's last value into accounts as lastTop
	// - no invest based on last periodic reset
	date = time.Date(2020, 9, 14, 11, 24, 1, 0, time.UTC)
	// LastTop should be 80.0 from the last investment.
	// If it was at 90.0, a price of 70.0 would trigger a reinvestment.
	adaptiveNoInvest(date, 70.0, strat, 80.0, t)

	// Invest before waitTime through large drawdown
	ddInvDate := time.Date(2020, 10, 14, 11, 24, 1, 0, time.UTC)
	adaptiveInvest(ddInvDate, 60.0, strat, t)

	// Next periodic investment should not happen waitTime after last periodic
	// but waitTime after last in general, which was a drawdown investment.
	date = periodicInvDate.Add(waitTime)
	adaptiveNoInvest(date, 90.0, strat, 90.0, t)

	// Next regular periodic investment
	date = ddInvDate.Add(waitTime)
	adaptiveInvest(date, 85.0, strat, t)
}

func TestAdaptivePeriodicFailTick(t *testing.T) {
	date := time.Date(2020, 1, 14, 11, 24, 1, 0, time.UTC)
	waitTime := time.Duration(182*24) * time.Hour
	priceP := &mockPriceProvider{}
	priceP.On("GetPrice", "TEST.DE", date).Return(100.0, nil).Once()

	p := &mockPortfolio{}
	p.On("getCashBalance").Return(2000.0)
	p.On("rebalance", 2000.0, date).Return(errors.New("Test error"))
	strat := NewAdaptivePeriodic(date, waitTime, 0.8, "TEST.DE", priceP)
	strat.tick(date, p)
	priceP.AssertExpectations(t)
	p.AssertExpectations(t)

}

func adaptiveInvest(date time.Time, price float64, strat Strategy, t *testing.T) {
	strat.(*AdaptivePeriodic).WithDrawdown.priceP.(*mockPriceProvider).On("GetPrice", "TEST.DE", date).Return(price, nil).Once()
	p := newMockP(date)
	strat.tick(date, p)
	strat.(*AdaptivePeriodic).WithDrawdown.priceP.(*mockPriceProvider).AssertExpectations(t)
	p.AssertExpectations(t)
	assert.Equal(t, strat.(*AdaptivePeriodic).lastTop, price, "lastTop wrong")
}

func adaptiveNoInvest(date time.Time, price float64, strat Strategy, lastTop float64, t *testing.T) {
	strat.(*AdaptivePeriodic).WithDrawdown.priceP.(*mockPriceProvider).On("GetPrice", "TEST.DE", date).Return(price, nil).Once()
	p := newMockP(date)
	strat.tick(date, p)
	strat.(*AdaptivePeriodic).WithDrawdown.priceP.(*mockPriceProvider).AssertExpectations(t)
	p.AssertNotCalled(t, "getCashBalance")
	p.AssertNotCalled(t, "rebalance", 2000.0, date)
	assert.Equal(t, strat.(*AdaptivePeriodic).lastTop, lastTop, "lastTop wrong")

}

func newMockP(callDate time.Time) *mockPortfolio {
	p := &mockPortfolio{}
	p.On("getCashBalance").Return(2000.0)
	p.On("rebalance", 2000.0, callDate).Return(nil)
	return p
}
