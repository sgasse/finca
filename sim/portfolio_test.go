package sim

import "testing"

func TestSinglePortfolio(t *testing.T) {
	stock := Stock{Symbol: "IBM", Volume: 5}
	startCash := 123.45
	p := NewSinglePortfolio(stock, startCash)

	if sp, ok := p.(*singlePortfolio); ok {
		assertEqual(t, sp.cash, startCash, "Cash in portfolio differs")
		assertEqual(t, sp.stock.Symbol, "IBM", "Wrong symbol in stock")
		assertEqual(t, sp.stock.Volume, int64(5), "Wrong stock volume")

		assertEqual(t, sp.getCashBalance(), startCash, "Wrong cash balance")

		sp.transact(10.0)
		assertEqual(t, sp.cash, 133.45, "Transaction resulted in wrong cash balance")
		assertEqual(t, sp.getCashBalance(), 133.45, "Transaction resulted in wrong cash balance")
	}
}
