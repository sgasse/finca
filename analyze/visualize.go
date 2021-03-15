package analyze

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sgasse/finca/sim"
)

var (
	startDate = time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC)
	symbol    = "SPY"
)

// LaunchVisualizer creates a server mux to visualize simulation callbacks
// with charts in the browser. A custom port can be set with the
// environment variable `ANALYZER_PORT`.
func LaunchVisualizer() {
	port := os.Getenv("ANALYZER_PORT")
	if port == "" {
		port = "3310"
	}

	mux := http.NewServeMux()

	mux.Handle("/compare", chartHandler(compareStrats))
	mux.Handle("/showStock", chartHandler(showStock))
	mux.Handle("/biyearly", chartHandler(biyearly))
	mux.Handle("/drawdown", chartHandler(drawdown))
	http.ListenAndServe(":"+port, mux)
}

// A chartHandler wraps a HTTP handler that might return an error. If the
// wrapped handler does return an error, this error is written to the HTTP
// response with the error code 500.
type chartHandler func(http.ResponseWriter, *http.Request) error

// ServeHTTP tries to serve a HTTP request with the wrapped handler. If
// this handler errors, the error is returned as response with error code
// 500. This makes the chartHandler interface implement `http.Handler`.
func (fn chartHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func compareStrats(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		err := maybeSetSymbol(r)
		if err != nil {
			return err
		}

		simRes := newSimRes()

		strats := map[string]sim.Strategy{
			"Monthly":     sim.NewMonthlyStrategy(startDate),
			"NoInvest":    &sim.NoInvest{},
			"30%Drawdown": &sim.MinDrawdown{LastTop: 0.0, RelVal: 0.7, RefSymbol: symbol},
			"55%Drawdown": &sim.MinDrawdown{LastTop: 0.0, RelVal: 0.45, RefSymbol: symbol},
		}

		if err = addSimResults(&simRes, strats); err != nil {
			return err
		}

		tsComp, err := multiSeriesChart(symbol, "hybrid_strats", simRes.Dates, simRes.TimeSeries, "templates/timeSeriesComp.html")
		if err != nil {
			return err
		}
		irrChart, err := multiSeriesChart(symbol, "hybrid_strats", simRes.Dates, simRes.IRR, "templates/barComp.html")
		if err != nil {
			return err
		}

		dates, stockTs, stockRelChange, stockDrawdown := evalSingleStockData(startDate, symbol)

		pChart, err := xyTemplate(symbol, dates, stockTs, "templates/stockprice.html")
		if err != nil {
			return err
		}
		ddChart, err := xyTemplate(symbol, dates, stockDrawdown, "templates/drawdown.html")
		if err != nil {
			return err
		}
		relChangeChart, err := xyTemplate(symbol, dates, stockRelChange, "templates/relChange.html")
		if err != nil {
			return err
		}

		data := chartData{concatCharts([]template.HTML{tsComp, irrChart, pChart, relChangeChart, ddChart})}

		t, err := template.ParseFiles("templates/compare.html")
		if err != nil {
			return err
		}

		t.Execute(w, &data)
	}
	return nil
}

func showStock(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		err := maybeSetSymbol(r)
		if err != nil {
			return err
		}

		dates, stockTs, stockRelChange, stockDrawdown := evalSingleStockData(startDate, symbol)

		pChart, err := xyTemplate(symbol, dates, stockTs, "templates/stockprice.html")
		if err != nil {
			return err
		}
		ddChart, err := xyTemplate(symbol, dates, stockDrawdown, "templates/drawdown.html")
		if err != nil {
			return err
		}
		relChangeChart, err := xyTemplate(symbol, dates, stockRelChange, "templates/relChange.html")
		if err != nil {
			return err
		}

		data := chartData{concatCharts([]template.HTML{pChart, relChangeChart, ddChart})}

		t, err := template.ParseFiles("templates/compare.html")
		if err != nil {
			return err
		}

		t.Execute(w, &data)
	}
	return nil
}

func biyearly(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		err := maybeSetSymbol(r)
		if err != nil {
			return err
		}

		simRes := newSimRes()

		if err = addSimResult(&simRes, &sim.NoInvest{}, "NoInvest"); err != nil {
			return err
		}

		for i := 1; i <= 6; i++ {
			strat := sim.NewFixedMonthsStrategy(startDate, []time.Month{time.Month(i),
				time.Month(i + 6)})
			name := fmt.Sprint(time.Month(i), "/", time.Month(i+6))

			if err = addSimResult(&simRes, strat, name); err != nil {
				return err
			}
		}

		biyearlyComp, err := multiSeriesChart(symbol, "biyearly_strats", simRes.Dates, simRes.TimeSeries, "templates/timeSeriesComp.html")
		if err != nil {
			return err
		}
		irrChart, err := multiSeriesChart(symbol, "hybrid_strats", simRes.Dates, simRes.IRR, "templates/barComp.html")
		if err != nil {
			return err
		}

		t, err := template.ParseFiles("templates/compare.html")
		if err != nil {
			return err
		}

		t.Execute(w, &chartData{concatCharts([]template.HTML{biyearlyComp, irrChart})})
	}
	return nil
}

func drawdown(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		err := maybeSetSymbol(r)
		if err != nil {
			return err
		}

		simRes := newSimRes()

		if err = addSimResult(&simRes, &sim.NoInvest{}, "NoInvest"); err != nil {
			return err
		}

		for relVal := 0.95; relVal >= 0.3; relVal -= 0.05 {
			perc := (1.0 - relVal) * 100
			err = addSimResult(
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
			return err
		}
		irrChart, err := multiSeriesChart(symbol, "hybrid_strats", simRes.Dates, simRes.IRR, "templates/barComp.html")
		if err != nil {
			return err
		}

		t, err := template.ParseFiles("templates/compare.html")
		if err != nil {
			return err
		}

		t.Execute(w, &chartData{concatCharts([]template.HTML{ddTsComp, irrChart})})
	}
	return nil
}
