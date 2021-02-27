package av

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
	MetaData    tsDailyAdjMd          `json:"Meta Data"`
	TimeSeries  map[string]tsDailyAdj `json:"Time Series (Daily)"`
	LastQueried time.Time             `json:"lastQueried"`
}

func init() {
	avAPIKey := os.Getenv("AV_API_KEY")
	if avAPIKey == "" {
		log.Fatal("You must specify your API key from AlphaVantage as AV_API_KEY.")
	}
	LaunchAV(avAPIKey)
}

func GetDividend(symbol string, date time.Time) (float64, error) {
	// Ensure latest data is available
	err := maybeUpdateCacheSymbol(symbol)
	if err != nil {
		return 0.0, err
	}

	cache.RLock()
	// Existance of symbol was ensured in `maybeUpdateCacheSymbol`
	tsData, _ := cache.m[symbol]
	cache.RUnlock()

	dailyData, ok := tsData.TimeSeries[date.Format("2006-01-02")]
	if !ok {
		return 0.0, nil
	}
	return dailyData.DividendAmount, nil
}

func GetPrice(symbol string, date time.Time) (float64, error) {
	// Ensure latest data is available
	err := maybeUpdateCacheSymbol(symbol)
	if err != nil {
		return 0.0, err
	}

	cache.RLock()
	// Existance of symbol was ensured in `maybeUpdateCacheSymbol`
	tsData, _ := cache.m[symbol]
	cache.RUnlock()

	// Check for price on the exact date or up to one week previously
	for i := 0; i <= 7; i++ {
		dateS := date.Add(-time.Duration(i) * 24 * time.Hour).Format("2006-01-02")
		dailyData, ok := tsData.TimeSeries[dateS]
		if ok {
			return dailyData.AdjustedClose, nil
		}
	}
	//log.Print("No price found")
	return 0.0, errors.New(fmt.Sprint("Could not find a price for ", symbol))
}

type qClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func avURL(symbol string, APIKey string) string {
	return fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY_ADJUSTED&symbol=%s&outputsize=full&apikey=%s",
		symbol, avAPIKey)

}

func maybeUpdateCacheSymbol(symbol string) error {
	// Check if data for symbol exists
	cache.RLock()
	tsData, entryFound := cache.m[symbol]
	cache.RUnlock()

	tooOld := false
	if entryFound {
		if tsData.LastQueried.Before(time.Now().Truncate(24 * time.Hour)) {
			// Cache too old, re-query
			tooOld = true
		}
	}

	if !entryFound || tooOld {
		// Try to get the price for the symbol
		client := http.Client{Timeout: time.Second * 5}
		tsData, err := getTsDailyAdj(symbol, &client)
		if err != nil {
			return err
		}
		tsData.LastQueried = time.Now()

		// Cache entry
		cache.Lock()
		cache.m[symbol] = tsData
		cache.Unlock()
	}
	return nil
}

func getTsDailyAdj(symbol string, client qClient) (resp tsDailyAdjResp, err error) {
	<-rateLimitOk
	log.Print("Fetching daily adjusted time series data for symbol ", symbol)

	url := avURL(symbol, avAPIKey)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		return
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &resp)
	return
}

func LaunchAV(inAvAPIKey string) {
	signal.Notify(SigChan, os.Interrupt)

	avAPIKey = inAvAPIKey
	go limitQueryRate(apiTimeout)

	cwd, pathErr := os.Getwd()

	if pathErr != nil {
		log.Println(pathErr, ", will not load/save cache.")
		return
	}

	cachePath = path.Join(cwd, cacheFile)
	loadCache(cachePath)
	go shutdown(cachePath)

}

func limitQueryRate(timeout time.Duration) {
	for {
		rateLimitOk <- true
		time.Sleep(timeout)
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
		for k, tsData := range cache.m {
			earliest, latest := getDateRange(tsData.TimeSeries)
			log.Println(k, ": ", len(tsData.TimeSeries), " entries, ", earliest, "-", latest)
		}
	}

	cache.Unlock()
}

func getDateRange(ts map[string]tsDailyAdj) (earliest, latest string) {
	earliest = time.Now().Format("2006-01-02")
	latest = "1900-01-01"

	for k := range ts {
		if k < earliest {
			earliest = k
		}
		if k > latest {
			latest = k
		}
	}
	return
}
