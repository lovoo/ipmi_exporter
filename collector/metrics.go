package collector

import "github.com/prometheus/client_golang/prometheus"

const namespace = "ipmi"

var (
	temperatures = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "temperatures"),
		"Contains the collected temperatures from IPMI",
		[]string{"sensor"},
		nil,
	)

	fanspeed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "fan_speed"),
		"Fan Speed in RPM",
		[]string{"fan"},
		nil,
	)

	voltages = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "voltages"),
		"Contains the voltages from IPMI",
		[]string{"sensor"},
		nil,
	)

	current = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "current"),
		"Contains the current from IPMI",
		[]string{"sensor"},
		nil,
	)

	intrusion = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "intrusion_status"),
		"Indicates if a chassis is open",
		nil,
		nil,
	)

	powersupply = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "power_supply_status"),
		"Indicates if a power supply is operational",
		[]string{"PSU"},
		nil,
	)

	exectime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "exec_time"),
		"How much time in milliseconds was spent executing ipmitools",
		nil,
		nil,
	)
)
