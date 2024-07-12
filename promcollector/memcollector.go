package promcollector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v4/mem"
)

// MemCollectorOpts 内存指标收集器配置
type MemCollectorOpts struct {
	Namespace    string // 命名空间
	ReportErrors bool   // 是否报告错误
}

// MemCollector 内存指标收集器
type MemCollector struct {
	TotalBytes       *prometheus.Desc
	AvailableBytes   *prometheus.Desc
	UsedBytes        *prometheus.Desc
	AvailablePercent *prometheus.Desc
	UsedPercent      *prometheus.Desc
	ActiveBytes      *prometheus.Desc
	BufferedBytes    *prometheus.Desc
	CachedBytes      *prometheus.Desc
	FreeBytes        *prometheus.Desc
	InactiveBytes    *prometheus.Desc

	BaseCollector
}

// NewMemCollector 新建内存指标收集器
func NewMemCollector(opts ...MemCollectorOpts) *MemCollector {
	var opt MemCollectorOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	bc := BaseCollector{reportErrors: opt.ReportErrors}
	ns := bc.getNamespace(opt.Namespace)

	return &MemCollector{
		BaseCollector: bc,
		TotalBytes: prometheus.NewDesc(
			ns+"mem_total_bytes", "Total number of bytes of memory.",
			nil, nil),
		AvailableBytes: prometheus.NewDesc(
			ns+"mem_available_bytes", "Number of bytes of available memory.",
			nil, nil),
		UsedBytes: prometheus.NewDesc(
			ns+"mem_used_bytes", "Number of bytes of used memory.",
			nil, nil),
		AvailablePercent: prometheus.NewDesc(
			ns+"mem_available_percent", "Percentage of available memory.",
			nil, nil),
		UsedPercent: prometheus.NewDesc(
			ns+"mem_used_percent", "Percentage of used memory.",
			nil, nil),
		ActiveBytes: prometheus.NewDesc(
			ns+"mem_active_bytes", "Number of bytes of active memory.",
			nil, nil),
		BufferedBytes: prometheus.NewDesc(
			ns+"mem_buffered_bytes", "Number of bytes of buffered memory.",
			nil, nil),
		CachedBytes: prometheus.NewDesc(
			ns+"mem_cached_bytes", "Number of bytes of cached memory.",
			nil, nil),
		FreeBytes: prometheus.NewDesc(
			ns+"mem_free_bytes", "Number of bytes of free memory.",
			nil, nil),
		InactiveBytes: prometheus.NewDesc(
			ns+"mem_inactive_bytes", "Number of bytes of inactive memory.",
			nil, nil),
	}
}

// Describe 实现 Describe 方法
func (c *MemCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.TotalBytes
	ch <- c.AvailableBytes
	ch <- c.UsedBytes
	ch <- c.AvailablePercent
	ch <- c.UsedPercent
	ch <- c.ActiveBytes
	ch <- c.BufferedBytes
	ch <- c.CachedBytes
	ch <- c.FreeBytes
	ch <- c.InactiveBytes
}

// Collect 实现 Collect 方法
func (c *MemCollector) Collect(ch chan<- prometheus.Metric) {
	metric, err := c.Metric()
	if err != nil {
		c.reportError(ch, err)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.TotalBytes, prometheus.GaugeValue, metric["total_bytes"])
	ch <- prometheus.MustNewConstMetric(c.AvailableBytes, prometheus.GaugeValue, metric["available_bytes"])
	ch <- prometheus.MustNewConstMetric(c.UsedBytes, prometheus.GaugeValue, metric["used_bytes"])
	ch <- prometheus.MustNewConstMetric(c.AvailablePercent, prometheus.GaugeValue, metric["available_percent"])
	ch <- prometheus.MustNewConstMetric(c.UsedPercent, prometheus.GaugeValue, metric["used_percent"])
	ch <- prometheus.MustNewConstMetric(c.ActiveBytes, prometheus.GaugeValue, metric["active_bytes"])
	ch <- prometheus.MustNewConstMetric(c.BufferedBytes, prometheus.GaugeValue, metric["buffered_bytes"])
	ch <- prometheus.MustNewConstMetric(c.CachedBytes, prometheus.GaugeValue, metric["cached_bytes"])
	ch <- prometheus.MustNewConstMetric(c.FreeBytes, prometheus.GaugeValue, metric["free_bytes"])
	ch <- prometheus.MustNewConstMetric(c.InactiveBytes, prometheus.GaugeValue, metric["inactive_bytes"])
}

// Metric 获取 cpu 指标
func (c *MemCollector) Metric() (map[string]float64, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return map[string]float64{
		"total_bytes":       float64(vm.Total),
		"available_bytes":   float64(vm.Available),
		"used_bytes":        float64(vm.Used),
		"available_percent": 100 * float64(vm.Available) / float64(vm.Total),
		"used_percent":      vm.UsedPercent,
		"active_bytes":      float64(vm.Active),
		"buffered_bytes":    float64(vm.Buffers),
		"cached_bytes":      float64(vm.Cached),
		"free_bytes":        float64(vm.Free),
		"inactive_bytes":    float64(vm.Inactive),
	}, nil
}
