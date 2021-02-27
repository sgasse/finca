package example

import (
	"fmt"
	"log"
	"time"

	"github.com/sgasse/finca/av"
	"github.com/sgasse/finca/sim"
)

func printPrice(symbol string, date string) {
	dateT, _ := time.Parse("2006-01-02", date)
	price, err := av.GetPrice(symbol, dateT)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Price for ", symbol, " on ", date, ": ", price)
}

func PrintPrices() {
	printPrice("IBM", "2021-02-01")
	printPrice("IBM", "2021-02-02")
	printPrice("H411.DE", "2021-02-02")
	dateT, _ := time.Parse("2006-01-02", "2021-02-02")
	_, _ = av.GetPrice("H411.DE", dateT)
	_, _ = av.GetPrice("CEMS.DE", dateT)
	_, _ = av.GetPrice("IQQ6.DE", dateT)
	_, _ = av.GetPrice("IUS9.DE", dateT)
	_, _ = av.GetPrice("EUNM.DE", dateT)
	_, _ = av.GetPrice("SXRG.DE", dateT)
	_, _ = av.GetPrice("UBU5.DE", dateT)
	_, _ = av.GetPrice("DX2J.DE", dateT)
	_, _ = av.GetPrice("SPYI.DE", dateT)
}

func getRefPortfolio() sim.Portfolio {
	sACWI := &sim.Stock{Symbol: "SPYI.DE"}

	stocks := map[*sim.Stock]int64{
		sACWI: 0,
	}

	goalRatios := map[*sim.Stock]float64{
		sACWI: 1.0,
	}

	startCash := 0.0
	p, err := sim.NewMultiPortfolio(startCash, stocks, goalRatios)
	if err != nil {
		log.Fatal(err)
	}

	return p
}

func SimulateMonthly() {
	startDate := time.Date(2011, 6, 1, 10, 0, 0, 0, time.UTC)
	p := getRefPortfolio()

	inc := sim.NewIncome(startDate, 1000.0)

	strat := sim.NewMonthlyStrategy(startDate)

	err := sim.Simulate(startDate, p, inc, strat)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Simulation results for monthly strategy")
	p.Evaluate(time.Now(), true)
}

func LoopBiYearly() {
	for i := 1; i <= 6; i++ {
		SimulateBiYearly(time.Month(i))
	}
}

func SimulateBiYearly(firstMonth time.Month) {
	startDate := time.Date(2011, 6, 1, 10, 0, 0, 0, time.UTC)
	p := getRefPortfolio()

	inc := sim.NewIncome(startDate, 1000.0)

	strat := sim.NewFixedMonthsStrategy(startDate, []time.Month{firstMonth, firstMonth + 6})

	err := sim.Simulate(startDate, p, inc, strat)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Simulation results for bi-monthly strategy of ", firstMonth, " and ", firstMonth+6)
	p.Evaluate(time.Now(), true)
}
