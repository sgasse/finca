package sim

import (
	"time"
)

type Income interface {
	tick(time.Time) float64
}

type MonthlyIncome struct {
	lastPaid      time.Time
	monthlyAmount float64
}

func (mi *MonthlyIncome) tick(date time.Time) float64 {
	if mi.lastPaid.Month() != date.Month() {
		// Pay out
		mi.lastPaid = date
		//log.Println("On simDay ", date)
		//log.Println("Monthly payout paid ", mi.monthlyAmount)
		return mi.monthlyAmount
	}
	return 0.0

}

func NewIncome(startDate time.Time, amount float64) Income {
	return &MonthlyIncome{
		lastPaid:      startDate.Add(-31 * 24 * time.Hour),
		monthlyAmount: amount,
	}
}
