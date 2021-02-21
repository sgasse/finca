package sim

import (
	"log"
	"math"
	"time"

	"github.com/sgasse/finca/stockdata"
)

var fixedFeePerStock = 7.0

type Stock struct {
	Symbol string
	WKN    string
	ISIN   string
	Volume int64
}

type portfolio interface {
	rebalance(float64, time.Time) error
	getCashBalance() float64
	transact(float64)
	Evaluate(time.Time) float64
}

type singlePortfolio struct {
	stock             Stock
	cash              float64
	receivedDividends float64
}

type multiPortfolio struct {
	stocks            []Stock
	cash              float64
	receivedDividends float64
}

func NewSinglePortfolio(stock Stock, cash float64) portfolio {
	return &singlePortfolio{stock: stock, cash: cash, receivedDividends: 0.0}
}

func (p *singlePortfolio) rebalance(reinvest float64, date time.Time) error {
	// Get price for only stock
	price, err := stockdata.GetPrice(p.stock.Symbol, date)
	if err != nil {
		return err
	}

	buyingFees := fixedFeePerStock

	newStocks := math.Floor((reinvest - buyingFees) / price)
	expense := newStocks*price + buyingFees

	// Commit
	p.stock.Volume += int64(newStocks)
	p.cash -= expense
	//log.Println("On simDay", date)
	//log.Println("Rebalancing bought ", int64(newStocks), " new stocks for ", expense)
	//log.Println("Volume: ", p.stock.Volume, " | Cash: ", p.cash)

	return nil
}

func (p *singlePortfolio) getCashBalance() float64 {
	return p.cash
}

func (p *singlePortfolio) transact(amount float64) {
	p.cash += amount
}

func (p *singlePortfolio) Evaluate(date time.Time) float64 {
	price, err := stockdata.GetPrice(p.stock.Symbol, date)
	if err != nil {
		log.Fatal(err)
	}
	stockValue := float64(p.stock.Volume) * price
	totalValue := p.cash + stockValue

	log.Println("Portfolio:")
	log.Println(p.stock.Symbol, ": ", p.stock.Volume, " x ", price, " = ", stockValue)
	log.Println("Cash: ", p.cash)
	log.Println("Total: ", totalValue)

	return totalValue
}
