package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/lovoo/ipmi_exporter/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

var (
	listenAddress = flag.String("web.listen", ":9289", "Address on which to expose metrics and web interface.")
	metricsPath   = flag.String("web.path", "/metrics", "Path under which to expose metrics.")
	ipmiBinary    = flag.String("ipmi.path", "ipmitool", "Path to the ipmi binary")
	showVersion   = flag.Bool("version", false, "Show version information and exit")
	timeout       = flag.Int("ipmi.timeout", -1, "How many milliseconds to allow ipmitools to run before cancelling.")
)

func init() {
	prometheus.MustRegister(version.NewCollector("ipmi_exporter"))
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("ipmi_exporter"))
		os.Exit(0)
	}

	log.Infoln("Starting IPMI Exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	prometheus.MustRegister(collector.NewExporter(*ipmiBinary, *timeout))

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

	log.Infoln("Listening on", *listenAddress, *metricsPath)
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
