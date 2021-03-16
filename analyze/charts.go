package analyze

import (
	"bytes"
	"html/template"
)

type chartData struct {
	Charts template.HTML
}

type chartRes struct {
	chart template.HTML
	err   error
}

func xyTemplate(symbol string, dates []string, series []float64, tplFile string) (template.HTML, error) {
	data := struct {
		Symbol string
		Dates  []string
		Series []float64
	}{
		Symbol: symbol,
		Dates:  dates,
		Series: series,
	}
	return templateChart(data, tplFile)
}

func multiSeriesChart(symbol string, name string, dates []string, series interface{}, tplFile string) (template.HTML, error) {
	data := struct {
		Symbol string
		Name   string
		Dates  []string
		Series interface{}
	}{
		Symbol: symbol,
		Name:   name,
		Dates:  dates,
		Series: series,
	}
	return templateChart(data, tplFile)
}

func templateChart(data interface{}, tplFile string) (template.HTML, error) {
	t, err := template.ParseFiles(tplFile)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		return "", err
	}

	return template.HTML(tpl.String()), nil
}

func combineCharts(chRes []chartRes) (chData chartData, err error) {
	var validCharts []template.HTML

	for _, res := range chRes {
		if res.err != nil {
			err = res.err
			return
		}
		validCharts = append(validCharts, res.chart)
	}

	return chartData{concatCharts(validCharts)}, nil
}

func wrapCR(chart template.HTML, err error) chartRes {
	return chartRes{chart, err}
}

func concatCharts(charts []template.HTML) template.HTML {
	res := template.HTML("")
	for _, s := range charts {
		res = res + s + "\n\n"
	}
	return res
}
