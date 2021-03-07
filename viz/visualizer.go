package viz

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sgasse/finca/sim"
)

type chartData struct {
	Dates         []string
	ValueOverTime map[string][]float64
	EndValues     map[string]float64
	IRR           map[string]float64
}

func compareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/compare.html")
		if err != nil {
			log.Fatal(err)
		}

		cData := chartData{
			ValueOverTime: make(map[string][]float64),
			EndValues:     make(map[string]float64),
			IRR:           make(map[string]float64),
		}

		startDate := time.Date(2011, 6, 1, 10, 0, 0, 0, time.UTC)

		monthly := sim.NewMonthlyStrategy(startDate)
		pValues, dates, irr := sim.SimulateStrategyOnRef(startDate, monthly)
		cData.Dates = dates
		cData.ValueOverTime["MonthlyInvest"] = pValues
		cData.IRR["MonthlyInvest"] = irr

		for i := 1; i <= 6; i++ {
			strat := sim.NewFixedMonthsStrategy(startDate, []time.Month{time.Month(i), time.Month(i + 6)})
			name := fmt.Sprint("Invest", time.Month(i), "And", time.Month(i+6))

			pValues, dates, irr := sim.SimulateStrategyOnRef(startDate, strat)
			cData.Dates = dates
			cData.ValueOverTime[name] = pValues
			cData.IRR[name] = irr
		}

		noInvest := &sim.NoInvest{}
		pValues, dates, irr = sim.SimulateStrategyOnRef(startDate, noInvest)
		cData.Dates = dates
		cData.ValueOverTime["NoInvest"] = pValues
		cData.IRR["NoInvest"] = irr

		t.Execute(w, &cData)
	}
}

func LaunchVisualizer() {
	port := os.Getenv("BALANCER_PORT")
	if port == "" {
		port = "3310"
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("assets"))

	mux.HandleFunc("/compare", compareHandler)
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.ListenAndServe(":"+port, mux)
}
