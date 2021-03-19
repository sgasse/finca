package sim

import (
	"testing"
	"time"

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

func newMockP(callDate time.Time) *mockPortfolio {
	p := &mockPortfolio{}
	p.On("getCashBalance").Return(2000.0)
	p.On("rebalance", 2000.0, callDate).Return(nil)
	return p
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
