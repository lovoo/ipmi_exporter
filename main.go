package main

import (
	"flag"
	"net/http"

	"github.com/lovoo/ipmi_exporter/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

var (
	listenAddress = flag.String("web.listen", ":9289", "Address on which to expose metrics and web interface.")
	metricsPath   = flag.String("web.path", "/metrics", "Path under which to expose metrics.")
	ipmiBinary    = flag.String("ipmi.path", "ipmitool", "Path to the ipmi binary")
)

func main() {
	flag.Parse()
	prometheus.MustRegister(collector.NewExporter(*ipmiBinary))

	handler := promhttp.Handler()
	if *metricsPath == "" || *metricsPath == "/" {
		http.Handle(*metricsPath, handler)
	} else {
		http.Handle(*metricsPath, handler)
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>
			<head><title>IPMI Exporter</title></head>
			<body>
			<h1>IPMI Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
		})
	}

	log.Infof("Starting Server: %s", *listenAddress)
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
