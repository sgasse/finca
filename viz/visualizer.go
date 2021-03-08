package viz

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/sgasse/finca/sim"
)

var (
	startDate = time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC)
	symbol    = "SPY"
)

type chartData struct {
	Dates         []string
	ValueOverTime map[string][]float64
	EndValues     map[string]float64
	IRR           map[string]float64
}

func addSimResults(cData *chartData, strat sim.Strategy, name string) {
	pValues, dates, irr := sim.SimulateStrategyOnRef(startDate, symbol, strat)
	cData.Dates = dates
	cData.ValueOverTime[name] = pValues

	// TODO: Remove hack
	if irr >= 400 {
		irr = 0.0
	}
	cData.IRR[name] = irr
}

func compareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/compare.html")
		if err != nil {
			log.Fatal(err)
		}

		params, err := url.ParseQuery(r.URL.RawQuery)
		log.Println("Got params: ", params)

		inSym, ok := params["symbol"]
		if ok {
			symbol = inSym[0]
		}

		cData := chartData{
			ValueOverTime: make(map[string][]float64),
			EndValues:     make(map[string]float64),
			IRR:           make(map[string]float64),
		}

		addSimResults(&cData, sim.NewMonthlyStrategy(startDate), "Monthly")
		addSimResults(&cData, &sim.NoInvest{}, "NoInvest")

		for i := 1; i <= 6; i++ {
			strat := sim.NewFixedMonthsStrategy(startDate, []time.Month{time.Month(i), time.Month(i + 6)})
			name := fmt.Sprint(time.Month(i), "/", time.Month(i+6))

			pValues, dates, irr := sim.SimulateStrategyOnRef(startDate, symbol, strat)
			cData.Dates = dates
			cData.ValueOverTime[name] = pValues
			cData.IRR[name] = irr
		}

		relVal := 0.90
		relIn, ok := params["relVal"]
		if ok {
			relParsed, err := strconv.ParseFloat(relIn[0], 64)
			if err == nil {
				relVal = relParsed
			}
		}

		minDrawdown := &sim.MinDrawdown{
			LastTop:   0.0,
			RelVal:    relVal,
			RefSymbol: "SPY",
		}
		addSimResults(&cData, minDrawdown, fmt.Sprintf("DrawdownTo%.2f", minDrawdown.RelVal))

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
