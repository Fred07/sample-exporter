package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// Exporter is a prometheus exporter
type Exporter struct {
	gauge                prometheus.Gauge
	gaugeVec             prometheus.GaugeVec
	testCounter          prometheus.CounterVec
	cpuUsage             prometheus.GaugeVec
	memoryUsage          prometheus.GaugeVec
	histogram            prometheus.HistogramVec
	endpointsAccessTotal prometheus.CounterVec
}

func NewExporter(metricsPrefix string) *Exporter {
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

	histogram := *prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricsPrefix,
		Name:      "histogram_test",
		Buckets:   prometheus.LinearBuckets(20, 5, 5),
		Help:      "histogram test",
	}, []string{"Label"})

	endpointsAccessTotal := *prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricsPrefix,
		Name:      "endpoints_access",
		Help:      "endpoints access",
	}, []string{"path"})

	return &Exporter{
		testCounter:          testCounter,
		cpuUsage:             cpuUsage,
		memoryUsage:          memoryUsage,
		histogram:            histogram,
		endpointsAccessTotal: endpointsAccessTotal,
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.testCounter.WithLabelValues("labe1").Inc()
	e.testCounter.Collect(ch)
	cpuPercent, _ := cpu.Percent(0, false)
	e.cpuUsage.WithLabelValues("label1").Set(float64(cpuPercent[0]))
	e.cpuUsage.Collect(ch)
	// fmt.Println(fmt.Sprintf("CPU: %s %", cpuPercent))
	memoryStats, _ := mem.VirtualMemory()
	e.memoryUsage.WithLabelValues("label1").Set(memoryStats.UsedPercent)
	e.memoryUsage.Collect(ch)

	e.histogram.WithLabelValues("label_1").Observe(float64(100))
	e.histogram.WithLabelValues("label_2").Observe(float64(50))
	e.histogram.Collect(ch)

	s1 := rand.NewSource(time.Now().Unix())
	r1 := rand.New(s1)
	e.endpointsAccessTotal.WithLabelValues("/v1/me").Add(math.Ceil(r1.Float64() * 10))
	e.endpointsAccessTotal.WithLabelValues("/v1/users").Add(math.Ceil(r1.Float64() * 15))
	e.endpointsAccessTotal.WithLabelValues("/v1/channels").Add(math.Ceil(r1.Float64() * 8))
	e.endpointsAccessTotal.Collect(ch)
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.testCounter.Describe(ch)
	e.cpuUsage.Describe(ch)
	e.memoryUsage.Describe(ch)
	e.histogram.Describe(ch)
	e.endpointsAccessTotal.Describe(ch)
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
