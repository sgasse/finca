package sim

import (
	"time"

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
