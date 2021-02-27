package sim

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPortfolio struct {
	mock.Mock
	cash float64
}

func (m *mockPortfolio) rebalance(reinvest float64, date time.Time) error {
	args := m.Called(reinvest, date)
	return args.Error(0)
}

func (m *mockPortfolio) getCashBalance() float64 {
	args := m.Called()
	return args.Get(0).(float64)
}

func (m *mockPortfolio) transact(amount float64) {
	_ = m.Called(amount)
}

func (m *mockPortfolio) Evaluate(date time.Time) float64 {
	args := m.Called(date)
	return args.Get(0).(float64)
}

func TestNewStrategy(t *testing.T) {
	startDate := time.Date(2020, 6, 1, 12, 0, 0, 0, time.UTC)
	strat := NewMonthlyStrategy(startDate)

	if midM, ok := strat.(*MidMonth); ok {
		expectedDate := "2020-05-01"

		assert.Equal(t, "2020-05-01", midM.lastInvested.Format("2006-01-02"),
			fmt.Sprint("Expected last invest date to be initialized to ", expectedDate))
	}
}

func TestStrategyTick(t *testing.T) {
	startDate := time.Date(2020, 6, 1, 12, 0, 0, 0, time.UTC)
	strat := NewMonthlyStrategy(startDate)

	if midM, ok := strat.(*MidMonth); ok {
		mp1 := &mockPortfolio{cash: 333.0}
		evalDay := time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC)
		midM.tick(evalDay, mp1)
		mp1.AssertNotCalled(t, "rebalance", evalDay)

		investDay := time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC)
		mp1.On("rebalance", mp1.cash, investDay).Return(nil)
		mp1.On("getCashBalance").Return(mp1.cash)
		midM.tick(investDay, mp1)
		mp1.AssertExpectations(t)

		mp2 := &mockPortfolio{cash: 123.3}
		evalDay = time.Date(2020, 8, 16, 0, 0, 0, 0, time.UTC)
		mp2.On("rebalance", mp2.cash, evalDay).Return(errors.New("Test error"))
		mp2.On("getCashBalance").Return(mp2.cash)
		midM.tick(evalDay, mp2)
		mp2.AssertExpectations(t)
		assert.Equal(t, investDay, midM.lastInvested, "The last investment day should not have changed")
	}
}
