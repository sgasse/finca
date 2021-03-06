package sim

import (
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/sgasse/finca/av"
)

var fixedFeePerStock = 56.0

type Portfolio interface {
	SetStart(time.Time)
	TotalValue(time.Time) float64
	CalcIRR(time.Time) float64
	getCashBalance() float64
	transact(transaction)
	rebalance(float64, time.Time) error
}

type Stock struct {
	Symbol string
	WKN    string
	ISIN   string
}

type transaction interface {
	delta() float64
}

type incomeTransaction struct {
	date   time.Time
	amount float64
}

type stockTransaction struct {
	date        time.Time
	stock       *Stock
	deltaVolume int64
	price       float64
}

type multiPortfolio struct {
	startDate    time.Time
	cash         float64
	stocks       map[*Stock]int64
	transactions []transaction
	goalRatios   map[*Stock]float64
	fixedFees    float64
	varFees      float64
}

func NewMultiPortfolio(cash float64, stocks map[*Stock]int64, goalRatios map[*Stock]float64, fixedFees float64, varFees float64) (Portfolio, error) {
	ratioSum := 0.0
	for stock, ratio := range goalRatios {
		ratioSum += ratio

		_, ok := stocks[stock]
		if !ok {
			msg := fmt.Sprint("Stock ", stock.Symbol, " found in goalRatios but not in stocks.")
			return &multiPortfolio{},
				errors.New(msg)
		}
	}

	if len(stocks) > len(goalRatios) {
		return &multiPortfolio{}, errors.New("Some stocks are missing a goal ratio.")
	}

	if math.Abs(ratioSum-1.0) > 1e-6 {
		return &multiPortfolio{}, errors.New("Goal ratios do not sum up to 1.0")
	}
	return &multiPortfolio{cash: cash, stocks: stocks, goalRatios: goalRatios, fixedFees: fixedFees, varFees: varFees}, nil
}

func (p *multiPortfolio) SetStart(date time.Time) {
	p.startDate = date
}

func (p *multiPortfolio) TotalValue(date time.Time) float64 {
	totalStockValue, err := getTotalStockValue(p.stocks, date)
	if err != nil {
		log.Fatal(err)
	}

	totalValue := p.cash + totalStockValue

	return totalValue
}

func (p *multiPortfolio) CalcIRR(date time.Time) float64 {
	totalValue := p.TotalValue(date)
	fx := buildTransactionFunc(p.transactions, totalValue, date)
	irr := bisect(fx, 5.0, 1.0, 1e-3, 100)
	// Transform to percent
	irr = (irr - 1.0) * 100
	// Round to two digits after the comma
	irr = math.Round(irr*100) / 100
	return irr
}

func (p *multiPortfolio) getCashBalance() float64 {
	return p.cash
}

func (p *multiPortfolio) transact(tr transaction) {
	p.transactions = append(p.transactions, tr)
	p.cash = p.cash + tr.delta()
	if st, ok := tr.(*stockTransaction); ok {
		p.stocks[st.stock] += st.deltaVolume
	}
}

func (p *multiPortfolio) rebalance(amount float64, date time.Time) error {
	// Naiv, safe, suboptimal rebalancing
	curTotalStockValue, err := getTotalStockValue(p.stocks, date)
	if err != nil {
		return err
	}

	totalGoalValue := curTotalStockValue + amount
	for stock, curVol := range p.stocks {
		// No error expected. All prices have to exist for the call to
		// `getTotalStockValue` to have succeeded before.
		price, _ := av.GetPrice(stock.Symbol, date)
		goalValue := p.goalRatios[stock] * totalGoalValue
		goalShares, adjustedPrice := calcGoalSharesAdjPrice(
			goalValue,
			price,
			p.fixedFees,
			p.varFees,
		)
		newShares := goalShares - curVol

		if newShares == 0 {
			return errors.New("Not enough money to buy a complete share")
		}

		tr := &stockTransaction{
			date:        date,
			stock:       stock,
			deltaVolume: newShares,
			price:       adjustedPrice,
		}
		p.transact(tr)
	}

	return nil
}

func (t *incomeTransaction) delta() float64 {
	return t.amount
}

func (t *stockTransaction) delta() float64 {
	return -float64(t.deltaVolume) * t.price
}

func getTotalStockValue(stocks map[*Stock]int64, date time.Time) (float64, error) {
	totalStockValue := 0.0
	for stock, vol := range stocks {
		price, err := av.GetPrice(stock.Symbol, date)
		if err != nil {
			return 0.0, err
		}
		totalStockValue += float64(vol) * price
	}
	return totalStockValue, nil
}

func bisect(fn func(float64) float64, high float64, low float64, prec float64, maxIter int) float64 {
	steps := 0
	x := low + (high-low)/2
	for diff := fn(x); math.Abs(diff) > prec && steps <= maxIter; {
		if diff > 0 {
			// x too large, narrow upper limit
			high = x
		} else {
			// x too small, narrow lower limit
			low = x
		}
		x = low + (high-low)/2
		diff = fn(x)

		steps++
	}

	if steps == maxIter && fn(x) > prec {
		fmt.Println("Warning: maxIter (", maxIter, ") reached without converging")
	}
	return x
}

func buildTransactionFunc(trs []transaction, cEnd float64, date time.Time) func(float64) float64 {
	yearHours := 365.0 * 24.0
	fn := func(x float64) float64 {
		res := -cEnd
		for _, tr := range trs {
			if incomeTr, ok := tr.(*incomeTransaction); ok {
				dt := date.Sub(incomeTr.date).Hours() / yearHours
				res += incomeTr.amount * math.Pow(x, dt)
			}
		}

		return res
	}
	return fn
}

func calcGoalSharesAdjPrice(goalValue, price, fixedFees, varFees float64) (int64, float64) {
	// newShares * price + fixedFees + (newShares * price)*varFees =!= goalValue
	// -> newShares = (goalValue - fixedFees) / (price * (1 + varFees))
	newShares := math.Floor((goalValue - fixedFees) / (price * (1 + varFees)))
	adjPrice := (1+varFees)*price + (fixedFees / newShares)
	return int64(newShares), adjPrice
}
