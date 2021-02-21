package stockdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	apiTimeout  = 13 * time.Second
	rateLimitOk = make(chan bool, 1)
	avAPIKey    string
	cache       = struct {
		sync.RWMutex
		m map[string]tsDailyAdjResp
	}{m: make(map[string]tsDailyAdjResp)}
)

type tsDailyAdjMd struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize    string `json:"4. Output Size"`
	TimeZone      string `json:"4. Time Zone"`
}

type tsDailyAdj struct {
	Open             float64 `json:"1. open,string"`
	High             float64 `json:"2. high,string"`
	Low              float64 `json:"3. low,string"`
	Close            float64 `json:"4. close,string"`
	AdjustedClose    float64 `json:"5. adjusted close,string"`
	Volume           int64   `json:"6. volume,string"`
	DividendAmount   float64 `json:"7. dividend amount,string"`
	SplitCoefficient float64 `json:"8. split coefficient,string"`
}

type tsDailyAdjResp struct {
	MetaData   tsDailyAdjMd          `json:"Meta Data"`
	TimeSeries map[string]tsDailyAdj `json:"Time Series (Daily)"`
}

func GetPrice(symbol string, date time.Time) (float64, error) {
	// Check if data for symbol exists
	cache.RLock()
	tsData, ok := cache.m[symbol]
	if !ok {
		cache.RUnlock()
		cache.Lock()
		cache.m[symbol] = getTsDailyAdj(symbol)
		tsData = cache.m[symbol]
		cache.Unlock()
	} else {
		cache.RUnlock()
	}

	// Check if date exists
	dateS := date.Format("2006-01-02")
	dailyData, ok := tsData.TimeSeries[dateS]
	if !ok {
		return 0.0, errors.New(fmt.Sprint("No entry for date ", dateS))
	}

	return dailyData.AdjustedClose, nil
}

func getTsDailyAdj(symbol string) tsDailyAdjResp {
	<-rateLimitOk
	log.Print("Fetching daily adjusted time series data for symbol ", symbol)

	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY_ADJUSTED&symbol=%s&outputsize=full&apikey=%s",
		symbol, avAPIKey)

	client := http.Client{Timeout: time.Second * 5}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		fmt.Println(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		fmt.Println(readErr)
	}

	var resp tsDailyAdjResp

	jsonErr := json.Unmarshal(body, &resp)
	if jsonErr != nil {
		fmt.Println(jsonErr)
	}

	return resp
}

func LaunchRequester(inAvAPIKey string) {
	avAPIKey = inAvAPIKey
	go limitQueryRate()
}

func limitQueryRate() {
	for {
		rateLimitOk <- true
		time.Sleep(apiTimeout)
	}
}
