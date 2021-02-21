package sim

import (
	"math"
	"testing"
	"time"
)

func TestNewIncome(t *testing.T) {
	start := time.Date(2020, 2, 1, 12, 0, 0, 0, time.UTC)
	amount := 1000.0
	income := NewIncome(start, amount)

	if monthlyI, ok := income.(*MonthlyIncome); ok {
		if monthlyI.monthlyAmount != amount {
			t.Error("Monthly amount should be ", amount, " but is ", monthlyI.monthlyAmount)
		}

		lastPaidDate := monthlyI.lastPaid.Format("2006-01-02")
		expectedDate := "2020-01-01"
		if lastPaidDate != expectedDate {
			t.Error("lastPaid was initialized to ", lastPaidDate, ", expected ", expectedDate)
		}
	}
}

func TestMonthlyIncomeTick(t *testing.T) {
	startDate := time.Date(2020, 2, 1, 12, 0, 0, 0, time.UTC)
	amount := 1000.0
	mi := NewIncome(startDate, amount)

	pay := mi.tick(startDate)
	if math.Abs(pay-amount) > 1e-7 {
		t.Error("Expected payout of ", amount, ", got ", pay)
	}

	pay = mi.tick(startDate.Add(24 * time.Hour))
	if pay > 0.0 {
		t.Error("Did not expect any payout, got ", pay)
	}

	pay = mi.tick(time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC))
	if math.Abs(pay-amount) > 1e-7 {
		t.Error("Expected payout of ", amount, ", got ", pay)
	}

}
