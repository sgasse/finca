package stockdata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	apiTimeout  = 13 * time.Second
	rateLimitOk = make(chan bool, 1)
	avAPIKey    string
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

func GetTsDailyAdj(symbol string) tsDailyAdjResp {
	<-rateLimitOk
	log.Print("Fetching daily adjusted time series data for symbol ", symbol)

	// TODO: outputsize=full
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY_ADJUSTED&symbol=%s&outputsize=compact&apikey=%s",
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
