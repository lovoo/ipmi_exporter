package collector

import (
	"bytes"
	"encoding/csv"
	"os/exec"
	"regexp"
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
	IPMIBinary string

	namespace string
}

// NewExporter instantiates a new ipmi Exporter.
func NewExporter(ipmiBinary string) *Exporter {
	return &Exporter{
		IPMIBinary: ipmiBinary,
		namespace:  "ipmi",
	}
}

func ipmiOutput(cmd string) ([]byte, error) {
	parts := strings.Fields(cmd)
	out, err := exec.Command(parts[0], parts[1]).Output()
	if err != nil {
		log.Errorf("error while calling ipmitool: %v", err)
	}
	return out, err
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


func splitOutput(impiOutput []byte) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(impiOutput))
	r.Comma = '|'
	r.Comment = '#'
	result, err := r.ReadAll()
	if err != nil {
		log.Errorf("could not parse ipmi output: %v", err)
        return result, err
	}

        keys := make(map[string]int)
        var res [][]string
        for _, v := range result {
            key := v[0]
            if _, ok := keys[key]; ok {
                keys[key] += 1
                v[0] = strings.TrimSpace(v[0]) + strconv.Itoa(keys[key])            
            } else {
                keys[key] = 1
            }
            res = append(res, v)
        }
        return res, err
}


// Describe describes all the registered stats metrics from the ipmi node.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- temperatures
	ch <- fanspeed
	ch <- voltages
	ch <- intrusion
	ch <- powersupply
        ch <- current
}

// Collect collects all the registered stats metrics from the ipmi node.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	output, err := ipmiOutput(e.IPMIBinary + " sensor")
	if err != nil {
		log.Errorln(err)
	}
	splitted, err := splitOutput(output)
	if err != nil {
		log.Errorln(err)
	}
	convertedOutput, err := convertOutput(splitted)
	if err != nil {
		log.Errorln(err)
	}

	for _, res := range convertedOutput {
		push := func(m *prometheus.Desc) {
			ch <- prometheus.MustNewConstMetric(m, prometheus.GaugeValue, res.value, res.metricsname)
		}
		switch strings.ToLower(res.unit) {
		case "degrees c":
			push(temperatures)
		case "volts":
			push(voltages)
		case "rpm":
			push(fanspeed)
                case "watts":
                        push(powersupply)
                case "amps":
                        push(current)
		}

		if matches, err := regexp.MatchString("PS.* Status", res.metricsname); matches && err == nil {
			push(powersupply)
		} else if strings.HasSuffix(res.metricsname, "Chassis Intru") {
			ch <- prometheus.MustNewConstMetric(intrusion, prometheus.GaugeValue, res.value)
		}
	}
}
