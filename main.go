package main

import (
	"os"

	"github.com/sgasse/finca/example"
	"github.com/sgasse/finca/stockdata"
)

func main() {
	example.SimulateMonthly()
	example.SimulateBiYearly()

	stockdata.SigChan <- os.Interrupt
	// sending os.Interrupt to SigChan will exit
	// the line below is never reached but waits for stockdata.shutdown()
	<-stockdata.SigChan
}

// TODO
// Custom unmarshall?
// Catch non-existent symbols/query response error
// Find earliest possible date for all prices in portfolio
// Plotting of performance?
// Dividends?
