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
	p, err := NewMultiPortfolio(startCash, stocks, goalRatios)
	assert.Nil(t, err)

	if mp, ok := p.(*multiPortfolio); ok {
		assert.Equal(t, startCash, mp.cash, "Cash in portfolio differs")
		assert.Equal(t, stocks, mp.stocks, "Stock map differs")
		assert.Equal(t, goalRatios, mp.goalRatios, "Goal ratio map differs")
	}

}
