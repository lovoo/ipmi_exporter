package main

import (
	"encoding/csv"
	"os/exec"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
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
	IpmiBinary string

	metrics   map[string]*prometheus.GaugeVec
	namespace string
}

// NewExporter instantiates a new ipmi Exporter.
func NewExporter(ipmiBinary string) *Exporter {
	e := Exporter{
		IpmiBinary: ipmiBinary,
		namespace:  "ipmi",
	}

	e.metrics = map[string]*prometheus.GaugeVec{}

	e.collect()
	return &e
}

type error interface {
	Error() string
}

func executeCommand(cmd string) (string, error) {
	parts := strings.Fields(cmd)
	out, err := exec.Command(parts[0], parts[1]).Output()
	if err != nil {
		log.Errorf("error while calling ipmitool: %v", err)
	}
	return string(out), err
}

func convertValue(strfloat string, strunit string) (value float64, err error) {
	if strfloat != "na" {
		if strunit == "discrete" {
			strfloat = strings.Replace(strfloat, "0x", "", -1)
			parsedValue, err := strconv.ParseUint(strfloat, 16, 32)
			if err != nil {
				log.Errorf("could not translate hex: %v, %v", parsedValue, err)
			}
			value = float64(parsedValue)
		} else {
			value, err = strconv.ParseFloat(strfloat, 64)
		}
	}
	return value, err
}

func convertOutput(result [][]string) (metrics []metric, err error) {
	for _, res := range result {
		var value float64
		var currentMetric metric

		for n := range res {
			res[n] = strings.TrimSpace(res[n])
		}
		res[0] = strings.ToLower(res[0])
		res[0] = strings.Replace(res[0], " ", "_", -1)
		res[0] = strings.Replace(res[0], "-", "_", -1)
		res[0] = strings.Replace(res[0], ".", "_", -1)

		value, err = convertValue(res[1], res[2])
		if err != nil {
			log.Errorf("could not parse ipmi output: %s", err)
		}

		currentMetric.value = value
		currentMetric.unit = res[2]
		currentMetric.metricsname = res[0]

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
		log.Errorf("could not parse ipmi output: %v", err)
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
}

// Collect collects all the registered stats metrics from the ipmi node.
func (e *Exporter) Collect(metrics chan<- prometheus.Metric) {
	e.collect()
	for _, m := range e.metrics {
		m.Collect(metrics)
	}
}

func (e *Exporter) collect() {
	output, err := executeCommand(e.IpmiBinary)
	if err != nil {
		log.Errorln(err)
	}
	splitted, err := splitAoutput(string(output))
	if err != nil {
		log.Errorln(err)
	}
	convertedOutput, err := convertOutput(splitted)
	if err != nil {
		log.Errorln(err)
	}
	createMetrics(e, convertedOutput)
}
