package sim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiPortfolio(t *testing.T) {
	sIBM := &Stock{Symbol: "IBM"}
	sOther := &Stock{Symbol: "H411.DE"}

	stocks := map[*Stock]int64{
		sIBM:   10,
		sOther: 23,
	}

	goalRatios := map[*Stock]float64{
		sIBM:   0.6,
		sOther: 0.4,
	}

	startCash := 1201.67
	p, err := NewMultiPortfolio(startCash, stocks, goalRatios, 56.0, 0.015)
	assert.Nil(t, err)

	if mp, ok := p.(*multiPortfolio); ok {
		assert.Equal(t, startCash, mp.cash, "Cash in portfolio differs")
		assert.Equal(t, stocks, mp.stocks, "Stock map differs")
		assert.Equal(t, goalRatios, mp.goalRatios, "Goal ratio map differs")
	}

}

func TestCalcGoalSharesAdjPrice(t *testing.T) {
	price := 80.0
	refGoalShares := int64(10)
	fixedFees := 6.0
	varFees := 0.015

	totalMoney := price*float64(refGoalShares)*(1+varFees) + fixedFees
	refAdjPrice := totalMoney / float64(refGoalShares)

	goalShares, adjPrice := calcGoalSharesAdjPrice(totalMoney, price, fixedFees, varFees)

	assert.Equal(t, refGoalShares, goalShares, "Number of goalShares wrong")
	assert.Equal(t, refAdjPrice, adjPrice, "Adjusted price wrong")
}
