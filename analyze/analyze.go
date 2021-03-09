package analyze

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/sgasse/finca/av"
	"github.com/sgasse/finca/sim"
)

var (
	startDate = time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC)
	symbol    = "SPY"
)

func LaunchVisualizer() {
	port := os.Getenv("BALANCER_PORT")
	if port == "" {
		port = "3310"
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("assets"))

	mux.HandleFunc("/compare", compareHandler)
	mux.HandleFunc("/drawdown", drawdownHandler)
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.ListenAndServe(":"+port, mux)
}

type SimResults struct {
	Dates      []string
	TimeSeries map[string][]float64
	IRR        map[string]float64
}

func addSimResults(simRes *SimResults, strat sim.Strategy, name string) error {
	pValues, dates, irr := sim.SimulateStrategyOnRef(startDate, symbol, strat)
	if len(simRes.Dates) == 0 {
		simRes.Dates = dates
	} else if len(simRes.Dates) != len(dates) {
		return errors.New("Simulation dates do not agree")
	}

	simRes.TimeSeries[name] = pValues
	simRes.IRR[name] = irr

	return nil

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

func maybeUpdateSymbol(params url.Values) {
	inSym, ok := params["symbol"]
	if ok {
		symbol = inSym[0]
	}

	sDate, err := getStartDate(symbol)
	if err == nil {
		startDate = sDate
	}
}

func drawdownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		params, err := url.ParseQuery(r.URL.RawQuery)
		log.Println("Got params: ", params)

		maybeUpdateSymbol(params)

		simRes := SimResults{
			TimeSeries: make(map[string][]float64),
			IRR:        make(map[string]float64),
		}

		err = addSimResults(&simRes, &sim.NoInvest{}, "NoInvest")
		if err != nil {
			log.Fatal(err)
		}
		delete(simRes.IRR, "NoInvest")

		for relVal := 0.95; relVal >= 0.3; relVal -= 0.05 {
			perc := (1.0 - relVal) * 100
			err = addSimResults(
				&simRes,
				&sim.MinDrawdown{
					LastTop:   0.0,
					RelVal:    relVal,
					RefSymbol: symbol,
				},
				fmt.Sprintf("%.0f", perc)+"%Drawdown",
			)
			if err != nil {
				log.Fatal(err)
			}
		}

		ddTsComp, err := multiSeriesChart(symbol, "drawdown_strats", simRes.Dates, simRes.TimeSeries, "templates/timeSeriesComp.html")
		if err != nil {
			log.Fatal(err)
		}

		data := struct {
			Charts template.HTML
		}{
			Charts: ddTsComp,
		}

		t, err := template.ParseFiles("templates/compare.html")
		if err != nil {
			log.Fatal(err)
		}

		t.Execute(w, &data)
	}
}

func compareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		params, err := url.ParseQuery(r.URL.RawQuery)
		log.Println("Got params: ", params)

		maybeUpdateSymbol(params)

		simRes := SimResults{
			TimeSeries: make(map[string][]float64),
			IRR:        make(map[string]float64),
		}

		err = addSimResults(&simRes, sim.NewMonthlyStrategy(startDate), "Monthly")
		if err != nil {
			log.Fatal(err)
		}

		err = addSimResults(&simRes, &sim.NoInvest{}, "NoInvest")
		if err != nil {
			log.Fatal(err)
		}
		delete(simRes.IRR, "NoInvest")

		err = addSimResults(
			&simRes,
			&sim.MinDrawdown{
				LastTop:   0.0,
				RelVal:    0.7,
				RefSymbol: symbol,
			},
			"30%Drawdown",
		)
		if err != nil {
			log.Fatal(err)
		}

		err = addSimResults(
			&simRes,
			&sim.MinDrawdown{
				LastTop:   0.0,
				RelVal:    0.45,
				RefSymbol: symbol,
			},
			"55%Drawdown",
		)
		if err != nil {
			log.Fatal(err)
		}

		/* 		addSimResults(&cData, sim.NewMonthlyStrategy(startDate), "Monthly")
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
		*/

		tsComp, err := multiSeriesChart(symbol, "hybrid_strats", simRes.Dates, simRes.TimeSeries, "templates/timeSeriesComp.html")
		if err != nil {
			log.Fatal(err)
		}

		dates, stockTs, stockRelChange, stockDrawdown := evalSingleStockData(startDate, symbol)

		pChart, err := xyTemplate(symbol, dates, stockTs, "templates/stockprice.html")
		if err != nil {
			log.Fatal(err)
		}
		ddChart, err := xyTemplate(symbol, dates, stockDrawdown, "templates/drawdown.html")
		if err != nil {
			log.Fatal(err)
		}
		relChangeChart, err := xyTemplate(symbol, dates, stockRelChange, "templates/relChange.html")
		if err != nil {
			log.Fatal(err)
		}

		data := struct {
			Charts template.HTML
		}{
			Charts: concatCharts([]template.HTML{
				tsComp,
				pChart,
				relChangeChart,
				ddChart,
			}),
		}

		t, err := template.ParseFiles("templates/compare.html")
		if err != nil {
			log.Fatal(err)
		}

		t.Execute(w, &data)
	}
}

func concatCharts(charts []template.HTML) template.HTML {
	res := template.HTML("")
	for _, s := range charts {
		res = res + s + "\n\n"
	}
	return res
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

func templateChart(data interface{}, tplFile string) (template.HTML, error) {
	t, err := template.ParseFiles(tplFile)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		return "", err
	}

	return template.HTML(tpl.String()), nil
}

func xyTemplate(symbol string, dates []string, series []float64, tplFile string) (template.HTML, error) {
	data := struct {
		Symbol string
		Dates  []string
		Series []float64
	}{
		Symbol: symbol,
		Dates:  dates,
		Series: series,
	}
	return templateChart(data, tplFile)
}

func multiSeriesChart(symbol string, name string, dates []string, series map[string][]float64, tplFile string) (template.HTML, error) {
	data := struct {
		Symbol string
		Name   string
		Dates  []string
		Series map[string][]float64
	}{
		Symbol: symbol,
		Name:   name,
		Dates:  dates,
		Series: series,
	}
	return templateChart(data, tplFile)
}
