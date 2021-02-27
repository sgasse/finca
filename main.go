package main

import (
	"os"

	"github.com/sgasse/finca/av"
	"github.com/sgasse/finca/example"
)

func main() {
	example.SimulateMonthly()
	example.SimulateBiYearly()

	av.SigChan <- os.Interrupt
	// sending os.Interrupt to SigChan will exit
	// the line below is never reached but waits for av.shutdown()
	<-av.SigChan
}

// TODO
// Custom unmarshall?
// Catch non-existent symbols/query response error
// Find earliest possible date for all prices in portfolio
// Plotting of performance?
// Dividends?
// Declare cache outdated
// Migrate to testify?
