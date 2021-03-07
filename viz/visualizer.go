package viz

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sgasse/finca/sim"
)

type chartData struct {
	Dates  []string
	Values map[string][]float64
}

func compareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/compare.html")
		if err != nil {
			log.Fatal(err)
		}

		startDate := time.Date(2011, 6, 1, 10, 0, 0, 0, time.UTC)

		values := make(map[string][]float64)

		pValues, dates := sim.SimulateMonthly(startDate)
		values["MonthlyInvest"] = pValues

		for i := 1; i <= 6; i++ {
			pValues, dates = sim.SimulateBiYearly(startDate, time.Month(i))
			name := fmt.Sprint("Invest", time.Month(i), "And", time.Month(i+6))
			values[name] = pValues
		}

		cData := chartData{
			Dates:  dates,
			Values: values,
		}

		t.Execute(w, &cData)
	}
}

func LaunchVisualizer() {
	port := os.Getenv("BALANCER_PORT")
	if port == "" {
		port = "3310"
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("assets"))

	mux.HandleFunc("/compare", compareHandler)
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.ListenAndServe(":"+port, mux)
}
