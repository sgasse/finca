package example

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sgasse/finca/sim"
	"github.com/sgasse/finca/stockdata"
)

func setup() {
	avAPIKey := os.Getenv("AV_API_KEY")
	if avAPIKey == "" {
		log.Fatal("You must specify your API key from AlphaVantage as AV_API_KEY.")
	}
	go stockdata.LaunchRequester(avAPIKey)
}

func printPrice(symbol string, date string) {
	dateT, _ := time.Parse("2006-01-02", date)
	price, err := stockdata.GetPrice(symbol, dateT)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Price for ", symbol, " on ", date, ": ", price)
}

func PrintPrices() {
	setup()
	printPrice("IBM", "2021-02-01")
	printPrice("IBM", "2021-02-02")
	printPrice("H411.DE", "2021-02-02")
}

func SimulateMonthly() {
	setup()

	startDate := time.Date(2010, 1, 1, 10, 0, 0, 0, time.UTC)
	p := sim.NewSinglePortfolio(sim.Stock{"EUNL.DE", "", "", 1}, 0.0)
	inc := sim.NewIncome(startDate, 1000.0)
	strat := sim.NewMonthlyStrategy(startDate)
	sim.Simulate(p, inc, startDate, strat)

	p.Evaluate(time.Now())
}

func SimulateBiYearly() {
	setup()

	startDate := time.Date(2010, 1, 1, 10, 0, 0, 0, time.UTC)
	p := sim.NewSinglePortfolio(sim.Stock{"EUNL.DE", "", "", 1}, 0.0)
	inc := sim.NewIncome(startDate, 1000.0)
	strat := sim.NewFixedMonthsStrategy(startDate, []time.Month{2, 8})

	sim.Simulate(p, inc, startDate, strat)

	p.Evaluate(time.Now())
}
