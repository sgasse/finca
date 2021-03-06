package analyze

import (
	"errors"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sgasse/finca/av"
	"github.com/sgasse/finca/sim"
)

var (
	DefaultFixedFees = 56.0
	DefaultVarFees   = 0.015
	FixedFees        = DefaultFixedFees
	VarFees          = DefaultVarFees
)

type SimResults struct {
	Dates      []string
	TimeSeries map[string][]float64
	IRR        map[string]float64
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

func addSimResult(simRes *SimResults, strat sim.Strategy, name string) error {
	fixedFees := 0.0
	varFees := 0.0

	if FixedFees != DefaultFixedFees || VarFees != DefaultVarFees {
		fixedFees = FixedFees
		varFees = VarFees
	} else if _, ok := strat.(*sim.MidMonth); ok {
		varFees = DefaultVarFees
	} else {
		fixedFees = DefaultFixedFees
	}

	pValues, dates, irr := sim.SimulateStratOnRef(startDate, symbol, strat, fixedFees, varFees)

	if len(simRes.Dates) == 0 {
		simRes.Dates = dates
	} else if len(simRes.Dates) != len(dates) {
		return errors.New("Simulation dates do not agree")
	}

	simRes.TimeSeries[name] = pValues
	if irr >= 400.0 {
		irr = 0
	}
	simRes.IRR[name] = irr

	return nil
}

func addSimResults(simRes *SimResults, strats map[string]sim.Strategy) error {
	for name, strat := range strats {
		if err := addSimResult(simRes, strat, name); err != nil {
			return err
		}
	}
	return nil
}

func newSimRes() SimResults {
	return SimResults{
		TimeSeries: make(map[string][]float64),
		IRR:        make(map[string]float64),
	}
}

func maybeSetParams(r *http.Request) error {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return err
	}
	log.Println("Got params: ", params)

	prevSymbol := symbol
	inSym, ok := params["symbol"]
	if ok {
		symbol = inSym[0]
	}

	sDate, err := getStartDate(symbol)
	if err != nil {
		symbol = prevSymbol
		return err
	}

	startDate = sDate

	param, cFees := params["fixedFees"]
	if cFees {
		fFees, err := strconv.ParseFloat(param[0], 64)
		if err != nil {
			return err
		}
		FixedFees = fFees
	}

	param, cFees = params["varFees"]
	if cFees {
		vFees, err := strconv.ParseFloat(param[0], 64)
		if err != nil {
			return err
		}
		VarFees = vFees
	}
	return nil
}

func roundTo(digits float64, number float64) float64 {
	factor := math.Pow(10, digits)
	return math.Round(number*factor) / factor
}
