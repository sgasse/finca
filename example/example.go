package example

import (
	"fmt"
	"log"
	"os"

	"github.com/sgasse/finca/stockdata"
)

func GetPrices() {
	avAPIKey := os.Getenv("AV_API_KEY")
	if avAPIKey == "" {
		log.Fatal("You must specify your API key from AlphaVantage as AV_API_KEY.")
	}
	go stockdata.LaunchRequester(avAPIKey)
	res := stockdata.GetTsDailyAdj("IBM")
	for date, vals := range res.TimeSeries {
		fmt.Println(date, ": ", vals)
	}
}
