package main

import (
	"log"
	"time"

	"encoding/csv"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type metric struct {
	metricsname string
	value       float64
	unit        string
	addr        string
}

// Exporter implements the prometheus.Collector interface. It exposes the metrics
// of a ipmi node.
type Exporter struct {
	IpmiBinary   string
	Waitgroup    *sync.WaitGroup
	metrics      map[string]*prometheus.GaugeVec
	duration     prometheus.Gauge
	totalScrapes prometheus.Counter
	namespace    string
}

// NewExporter instantiates a new ipmi Exporter.
func NewExporter(ipmiBinary string, wg *sync.WaitGroup) *Exporter {
	e := Exporter{
		IpmiBinary: ipmiBinary,
		Waitgroup:  wg,
		namespace:  "ipmi",
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: *namespace,
			Name:      "exporter_last_scrape_duration_seconds",
			Help:      "The last scrape duration.",
		}),
	}

	e.metrics = map[string]*prometheus.GaugeVec{}

	e.collect()
	return &e
}

type error interface {
	Error() string
}

func executeCommand(cmd string, wg *sync.WaitGroup) (string, error) {
	parts := strings.Fields(cmd)
	out, err := exec.Command(parts[0], parts[1]).Output()
	if err != nil {
		fmt.Println("error occured")
		fmt.Printf("%s", err)
	}
	wg.Done()
	return string(out), err
}

func convertValue(strfloat string, strunit string) (value float64, err error) {
	if strfloat != "na" {
		if strunit == "discrete" {
			strfloat = strings.Replace(strfloat, "0x", "", -1)
			parsedValue, err := strconv.ParseUint(strfloat, 16, 32)
			if err != nil {
				log.Printf("could not translate hex: %v, %v", parsedValue, err)
			}
			value = float64(parsedValue)
		} else {
			value, err = strconv.ParseFloat(strfloat, 64)
		}
	}
	return value, err
}

func convertOutput(result [][]string) (metrics []metric, err error) {
	for i := range result {
		var value float64
		var currentMetric metric

		for n := range result[i] {
			result[i][n] = strings.TrimSpace(result[i][n])
		}
		result[i][0] = strings.ToLower(result[i][0])
		result[i][0] = strings.Replace(result[i][0], " ", "_", -1)
		result[i][0] = strings.Replace(result[i][0], "-", "_", -1)
		result[i][0] = strings.Replace(result[i][0], ".", "_", -1)

		value, err = convertValue(result[i][1], result[i][2])
		if err != nil {
			log.Printf("could not parse ipmi output: %s", err)
		}

		currentMetric.value = value
		currentMetric.unit = result[i][2]
		currentMetric.metricsname = result[i][0]

		metrics = append(metrics, currentMetric)
	}
	return metrics, err
}

func splitAoutput(output string) ([][]string, error) {

	r := csv.NewReader(strings.NewReader(output))
	r.Comma = '|'
	r.Comment = '#'
	result, err := r.ReadAll()
	if err != nil {
		log.Printf("could not parse ipmi output: %v", err)
	}
	return result, err
}

func createMetrics(e *Exporter, metric []metric) {
	for n := range metric {
		e.metrics[metric[n].metricsname] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   "ipmi",
			Name:        metric[n].metricsname,
			Help:        metric[n].metricsname,
			ConstLabels: map[string]string{"unit": metric[n].unit},
		}, []string{"addr"})
		var labels prometheus.Labels = map[string]string{"addr": "localhost"}
		e.metrics[metric[n].metricsname].With(labels).Set(metric[n].value)
	}
}

// Describe Describes all the registered stats metrics from the ipmi node.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.metrics {
		m.Describe(ch)
	}

	ch <- e.duration.Desc()
}

// Collect collects all the registered stats metrics from the ipmi node.
func (e *Exporter) Collect(metrics chan<- prometheus.Metric) {
	e.collect()
	for _, m := range e.metrics {
		m.Collect(metrics)
	}

	metrics <- e.duration
}

func (e *Exporter) collect() {
	e.Waitgroup.Add(1)

	now := time.Now().UnixNano()

	output, err := executeCommand(e.IpmiBinary, e.Waitgroup)
	splitted, err := splitAoutput(string(output))
	convertedOutput, err := convertOutput(splitted)
	createMetrics(e, convertedOutput)

	e.duration.Set(float64(time.Now().UnixNano()-now) / 1000000000)

	e.Waitgroup.Wait()
	if err != nil {
		log.Printf("could not retrieve ipmi metrics: %v", err)
		return
	}
}
