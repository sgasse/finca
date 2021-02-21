package sim

import (
	"errors"
	"log"
	"math"
	"time"

	"github.com/sgasse/finca/stockdata"
)

type Stock struct {
	Symbol string
	WKN    string
	ISIN   string
	Volume int64
}

type portfolio struct {
	stocks            []Stock
	cash              float64
	receivedDividends float64
}

func NewPortfolio(stocks []Stock, cash float64) portfolio {
	return portfolio{stocks: stocks, cash: cash, receivedDividends: 0.0}
}

func (p *portfolio) rebalance(reinvest float64, date time.Time) error {
	// TODO abstract

	// Get price for only stock
	price, err := stockdata.GetPrice(p.stocks[0].Symbol, date)
	if err != nil {
		return err
	}

	newStocks := math.Floor(p.cash / price)
	expense := newStocks * price

	// Commit
	p.stocks[0].Volume += int64(newStocks)
	p.cash -= expense
	log.Println("On simDay", date)
	log.Println("Rebalancing bought ", int64(newStocks), " new stocks for ", expense)
	log.Println("Volume: ", p.stocks[0].Volume, " | Cash: ", p.cash)

	return nil
}

func Simulate(p portfolio, inc Income, start time.Time, strat Strategy) error {
	if time.Now().Sub(start) < 0 {
		return errors.New("Start lies in the future")
	}

	simDay := start
	// Simulate until reaching the current date
	for time.Now().Sub(simDay) > 0 {
		// Tick income
		p.cash += inc.tick(simDay)

		// Tick invest strategy
		strat.tick(simDay, &p)

		// Increase day
		simDay = simDay.Add(time.Duration(24 * time.Hour))
	}

	return nil
}
