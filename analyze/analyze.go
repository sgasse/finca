package analyze

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/sgasse/finca/av"
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
	StockDates    []string
	StockVals     []float64
	StockRelVal   []float64
	StockMaxDD    []float64
	Symbol        string
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

func getStartDate(symbol string) (sDate time.Time, err error) {
	earliest, _, err := av.GetDateRange(symbol)
	if err != nil {
		return
	}

	sDate, err = time.Parse("2006-01-02", earliest)
	if err != nil {
		return
	}

	// Shift to beginning of next month
	sDate = time.Date(sDate.Year(), sDate.Month()+1, 1, 12, 0, 0, 0, time.UTC)
	return
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

		sDate, err := getStartDate(symbol)
		if err == nil {
			startDate = sDate
		}

		cData := chartData{
			ValueOverTime: make(map[string][]float64),
			EndValues:     make(map[string]float64),
			IRR:           make(map[string]float64),
			Symbol:        symbol,
		}

		addSimResults(&cData, sim.NewMonthlyStrategy(startDate), "Monthly")
		addSimResults(&cData, &sim.NoInvest{}, "NoInvest")

		for i := 1; i <= 6; i++ {
			strat := sim.NewFixedMonthsStrategy(startDate, []time.Month{time.Month(i),
				time.Month(i + 6)})
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

		cData.StockDates, cData.StockVals, cData.StockRelVal, cData.StockMaxDD = evalSingleStockData(startDate, symbol)

		t.Execute(w, &cData)
	}
}

func evalSingleStockData(startDate time.Time, symbol string) (dates []string, timeSeries []float64, relChange []float64, maxDD []float64) {
	curDay := startDate
	for time.Now().Sub(curDay) > 0 {
		price, err := av.GetPrice(symbol, curDay)
		if err == nil {
			price := roundTo(2, price)
			timeSeries = append(timeSeries, price)
			dates = append(dates, curDay.Format("2006-01-02"))
		}

		curDay = curDay.Add(time.Duration(24 * time.Hour))
	}

	lastMax := 0.0
	// Change to first day equal to zero
	lastPrice := timeSeries[0]

	for _, price := range timeSeries {
		if price >= lastMax {
			maxDD = append(maxDD, 0.0)
			lastMax = price
		} else {
			drawdown := roundTo(2, (price/lastMax-1.0)*100)
			maxDD = append(maxDD, drawdown)
		}

		percChange := roundTo(2, (price/lastPrice-1.0)*100)
		relChange = append(relChange, percChange)
		lastPrice = price
	}

	return
}

func roundTo(digits float64, number float64) float64 {
	factor := math.Pow(10, digits)
	return math.Round(number*factor) / factor
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
