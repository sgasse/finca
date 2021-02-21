package sim

import (
	"testing"
	"time"
)

func TestNewStrategy(t *testing.T) {
	startDate := time.Date(2020, 6, 1, 12, 0, 0, 0, time.UTC)
	strat := NewStrategy(startDate)

	if midM, ok := strat.(*MidMonth); ok {
		lastInvDate := midM.lastInvested.Format("2006-01-02")
		expectedDate := "2020-05-01"
		if lastInvDate != expectedDate {
			t.Error("Expected lastInvested to be initialized to ", expectedDate, ", got ", lastInvDate)
		}
	}
}
