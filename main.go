package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	listenAddress = flag.String("web.listen", ":9289", "Address on which to expose metrics and web interface.")
	metricsPath   = flag.String("web.path", "/metrics", "Path under which to expose metrics.")
	ipmiBinary    = flag.String("ipmi.path", "/usr/local/bin/ipmitool sensor", "Path to the ipmi binary")
	namespace     = flag.String("namespace", "ipmi", "Namespace for the IPMI metrics.")
)

func main() {
	flag.Parse()
	prometheus.MustRegister(NewExporter(*ipmiBinary))

	log.Printf("Starting Server: %s", *listenAddress)
	handler := prometheus.Handler()
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

	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
