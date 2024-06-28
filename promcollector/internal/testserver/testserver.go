package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/sliveryou/micro-pkg/promcollector"
)

func main() {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())

	registry.MustRegister(promcollector.NewCPUCollector())
	registry.MustRegister(promcollector.NewMemCollector())
	registry.MustRegister(promcollector.NewNetCollector(promcollector.NetCollectorOpts{
		CheckInterfaces: true,
	}))
	registry.MustRegister(promcollector.NewDiskCollector())
	registry.MustRegister(promcollector.NewDiskIOCollector())

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		Registry: registry,
	}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
