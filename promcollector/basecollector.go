package promcollector

import "github.com/prometheus/client_golang/prometheus"

// BaseCollector 基础收集器
type BaseCollector struct {
	reportErrors bool
}

func (c *BaseCollector) getNamespace(ns string) string {
	if ns != "" {
		return ns + "_"
	}

	return ""
}

func (c *BaseCollector) reportError(ch chan<- prometheus.Metric, err error) {
	if !c.reportErrors {
		return
	}
	ch <- prometheus.NewInvalidMetric(prometheus.NewInvalidDesc(err), err)
}
