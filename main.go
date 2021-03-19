package main

import (
	"log"
	"os"

	"github.com/sgasse/finca/analyze"
	"github.com/sgasse/finca/av"
)

func main() {
	avAPIKey := os.Getenv("AV_API_KEY")
	if avAPIKey == "" {
		log.Fatal("You must specify your API key from AlphaVantage as AV_API_KEY.")
	}
	av.LaunchAV(avAPIKey)
	analyze.LaunchVisualizer()
}
