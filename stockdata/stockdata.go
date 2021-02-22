package stockdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
	"time"
)

var (
	avAPIKey    string
	SigChan     = make(chan os.Signal, 1)
	rateLimitOk = make(chan bool, 1)
	apiTimeout  = 13 * time.Second
	cache       = struct {
		sync.RWMutex
		m map[string]tsDailyAdjResp
	}{m: make(map[string]tsDailyAdjResp)}
	cacheFile = ".avCache.json"
	cachePath string
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

func init() {
	avAPIKey := os.Getenv("AV_API_KEY")
	if avAPIKey == "" {
		log.Fatal("You must specify your API key from AlphaVantage as AV_API_KEY.")
	}
	LaunchAV(avAPIKey)
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

	// Check for price on the exact date or up to one week previously
	for i := 0; i <= 7; i++ {
		dateS := date.Add(-time.Duration(i) * 24 * time.Hour).Format("2006-01-02")
		dailyData, ok := tsData.TimeSeries[dateS]
		if ok {
			return dailyData.AdjustedClose, nil
		}
	}
	return 0.0, errors.New(fmt.Sprint("Could not find a price for ", symbol))
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

func LaunchAV(inAvAPIKey string) {
	signal.Notify(SigChan, os.Interrupt)

	avAPIKey = inAvAPIKey
	go limitQueryRate()

	cwd, pathErr := os.Getwd()

	if pathErr != nil {
		log.Println(pathErr, ", will not load/save cache.")
		return
	}

	cachePath = path.Join(cwd, cacheFile)
	loadCache(cachePath)
	go shutdown(cachePath)

}

func limitQueryRate() {
	for {
		rateLimitOk <- true
		time.Sleep(apiTimeout)
	}
}

func shutdown(path string) {
	<-SigChan
	log.Println("Saving cache on disk")
	saveCache(path)
	os.Exit(0)
}

func saveCache(path string) {
	cache.RLock()
	cacheJSON, jsonErr := json.MarshalIndent(cache.m, "", "    ")
	cache.RUnlock()

	if jsonErr != nil {
		log.Println(jsonErr)
		return
	}

	ioutil.WriteFile(path, cacheJSON, 0644)
	log.Println("Wrote cache to ", path)
}

func loadCache(path string) {
	cacheStr, readErr := ioutil.ReadFile(path)

	if readErr != nil {
		log.Println(readErr, ", will continue without loaded cache.")
		return
	}

	cache.Lock()
	jsonErr := json.Unmarshal(cacheStr, &cache.m)

	if jsonErr != nil {
		log.Println(jsonErr)
	} else {
		log.Println("Cache loaded, symbols:")
		for k := range cache.m {
			log.Print(k)
		}
	}

	cache.Unlock()
}
