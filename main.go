package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// Exporter is a prometheus exporter
type Exporter struct {
	gauge       prometheus.Gauge
	gaugeVec    prometheus.GaugeVec
	testCounter prometheus.CounterVec
	cpuUsage    prometheus.GaugeVec
	memoryUsage prometheus.GaugeVec
}

func NewExporter(metricsPrefix string) *Exporter {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: metricsPrefix,
		Name:      "Gauge",
		Help:      "metric without label",
	})

	gaugeVec := *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsPrefix,
		Name:      "GaugeVec",
		Help:      "metric with label",
	}, []string{"myLabel"})

	testCounter := *prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricsPrefix,
		Name:      "auto_increment_counter",
		Help:      "+1",
	}, []string{"Label"})

	cpuUsage := *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsPrefix,
		Name:      "cpu_usage",
		Help:      "This is cpu usage stats",
	}, []string{"Label"})

	memoryUsage := *prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsPrefix,
		Name:      "memory_usage",
		Help:      "This is memory usage stats",
	}, []string{"Label"})

	return &Exporter{
		gauge:       gauge,
		gaugeVec:    gaugeVec,
		testCounter: testCounter,
		cpuUsage:    cpuUsage,
		memoryUsage: memoryUsage,
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.gauge.Set(float64(0))
	e.gaugeVec.WithLabelValues("label1").Set(float64(0))
	e.gauge.Collect(ch)
	e.gaugeVec.Collect(ch)
	e.testCounter.WithLabelValues("labe1").Inc()
	e.testCounter.Collect(ch)
	cpuPercent, _ := cpu.Percent(0, false)
	e.cpuUsage.WithLabelValues("label1").Set(float64(cpuPercent[0]))
	e.cpuUsage.Collect(ch)
	// fmt.Println(fmt.Sprintf("CPU: %s %", cpuPercent))
	memoryStats, _ := mem.VirtualMemory()
	e.memoryUsage.WithLabelValues("label1").Set(memoryStats.UsedPercent)
	e.memoryUsage.Collect(ch)
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.gauge.Describe(ch)
	e.gaugeVec.Describe(ch)
	e.testCounter.Describe(ch)
	e.cpuUsage.Describe(ch)
	e.memoryUsage.Describe(ch)
}

func main() {

	listenAddress := ":8081"
	metricsEndpoint := "/metrics"

	fmt.Println("Exporter starting...")

	exporter := NewExporter("howard_exporter")
	newRegistry := prometheus.NewRegistry()
	newRegistry.MustRegister(exporter)
	handler := promhttp.HandlerFor(newRegistry, promhttp.HandlerOpts{})

	// 以下包含預設的 prometheus collector, 使用: prometheus.Handler()
	// prometheus.MustRegister(exporter)
	// prometheus.Unregister(prometheus.NewGoCollector()) // remove default collector

	http.Handle(metricsEndpoint, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			 <head><title>Endpoint Exporter</title></head>
			 <body>
			 <h1>Endpoint Exporter</h1>
			 <p><a href='` + metricsEndpoint + `'>Metrics</a></p>
			 </body>
			 </html>`))
	})

	// fmt.Println(http.ListenAndServe(listenAddress, nil))
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
