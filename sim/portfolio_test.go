package sim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSinglePortfolio(t *testing.T) {
	stock := Stock{Symbol: "IBM", Volume: 5}
	startCash := 123.45
	p := NewSinglePortfolio(stock, startCash)

	if sp, ok := p.(*singlePortfolio); ok {
		assert.Equal(t, startCash, sp.cash, "Cash in portfolio differs")
		assert.Equal(t, "IBM", sp.stock.Symbol, "Wrong symbol in stock")
		assert.Equal(t, int64(5), sp.stock.Volume, "Wrong stock volume")

		assert.Equal(t, startCash, sp.getCashBalance(), "Wrong cash balance")

		sp.transact(10.0)
		assert.Equal(t, 133.45, sp.cash, "Transaction resulted in wrong cash balance")
		assert.Equal(t, 133.45, sp.getCashBalance(), "Transaction resulted in wrong cash balance")
	}
}
