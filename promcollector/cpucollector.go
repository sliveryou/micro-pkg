package promcollector

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v4/cpu"
)

// CPUCollectorOpts cpu 指标收集器配置
type CPUCollectorOpts struct {
	Namespace    string // 命名空间
	ReportErrors bool   // 是否报告错误
}

// CPUCollector  cpu 指标收集器
type CPUCollector struct {
	UsageActive    *prometheus.Desc
	UsageUser      *prometheus.Desc
	UsageSystem    *prometheus.Desc
	UsageIdle      *prometheus.Desc
	UsageNice      *prometheus.Desc
	UsageIowait    *prometheus.Desc
	UsageIrq       *prometheus.Desc
	UsageSoftirq   *prometheus.Desc
	UsageSteal     *prometheus.Desc
	UsageGuest     *prometheus.Desc
	UsageGuestNice *prometheus.Desc

	mu      sync.RWMutex
	lastCts *cpu.TimesStat
	BaseCollector
}

// NewCPUCollector 新建 cpu 指标收集器
func NewCPUCollector(opts ...CPUCollectorOpts) *CPUCollector {
	var opt CPUCollectorOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	bc := BaseCollector{reportErrors: opt.ReportErrors}
	ns := bc.getNamespace(opt.Namespace)

	return &CPUCollector{
		BaseCollector: bc,
		UsageActive: prometheus.NewDesc(
			ns+"cpu_usage_active", "Percentage of time that the CPU is active in any capacity.",
			[]string{"cpu"}, nil),
		UsageUser: prometheus.NewDesc(
			ns+"cpu_usage_user", "Percentage of time that the CPU is in user mode.",
			[]string{"cpu"}, nil),
		UsageSystem: prometheus.NewDesc(
			ns+"cpu_usage_system", "Percentage of time that the CPU is in system mode.",
			[]string{"cpu"}, nil),
		UsageIdle: prometheus.NewDesc(
			ns+"cpu_usage_idle", "Percentage of time that the CPU is idle.",
			[]string{"cpu"}, nil),
		UsageNice: prometheus.NewDesc(
			ns+"cpu_usage_nice", "Percentage of time that the CPU is in user mode with low-priority processes, which higher-priority processes can easily interrupt.",
			[]string{"cpu"}, nil),
		UsageIowait: prometheus.NewDesc(
			ns+"cpu_usage_iowait", "Percentage of time that the CPU is waiting for I/O operations to complete.",
			[]string{"cpu"}, nil),
		UsageIrq: prometheus.NewDesc(
			ns+"cpu_usage_irq", "Percentage of time that the CPU is servicing interrupts.",
			[]string{"cpu"}, nil),
		UsageSoftirq: prometheus.NewDesc(
			ns+"cpu_usage_softirq", "Percentage of time that the CPU is servicing software interrupts.",
			[]string{"cpu"}, nil),
		UsageSteal: prometheus.NewDesc(
			ns+"cpu_usage_steal", "Percentage of time that the CPU is in stolen time, or time spent in other operating systems in a virtualized environment.",
			[]string{"cpu"}, nil),
		UsageGuest: prometheus.NewDesc(
			ns+"cpu_usage_guest", "Percentage of time that the CPU is running a virtual CPU for a guest operating system.",
			[]string{"cpu"}, nil),
		UsageGuestNice: prometheus.NewDesc(
			ns+"cpu_usage_guest_nice", "Percentage of time that the CPU is running a virtual CPU for a guest operating system, which is low-priority and can be interrupted by other processes.",
			[]string{"cpu"}, nil),
	}
}

// Describe 实现 Describe 方法
func (c *CPUCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.UsageActive
	ch <- c.UsageUser
	ch <- c.UsageSystem
	ch <- c.UsageIdle
	ch <- c.UsageNice
	ch <- c.UsageIowait
	ch <- c.UsageIrq
	ch <- c.UsageSoftirq
	ch <- c.UsageSteal
	ch <- c.UsageGuest
	ch <- c.UsageGuestNice
}

// Collect 实现 Collect 方法
func (c *CPUCollector) Collect(ch chan<- prometheus.Metric) {
	metric, err := c.Metric()
	if err != nil {
		c.reportError(ch, err)
		return
	}

	labels := []string{"cpu-total"}

	ch <- prometheus.MustNewConstMetric(c.UsageActive, prometheus.GaugeValue, metric["usage_active"], labels...)
	ch <- prometheus.MustNewConstMetric(c.UsageUser, prometheus.GaugeValue, metric["usage_user"], labels...)
	ch <- prometheus.MustNewConstMetric(c.UsageSystem, prometheus.GaugeValue, metric["usage_system"], labels...)
	ch <- prometheus.MustNewConstMetric(c.UsageIdle, prometheus.GaugeValue, metric["usage_idle"], labels...)
	ch <- prometheus.MustNewConstMetric(c.UsageNice, prometheus.GaugeValue, metric["usage_nice"], labels...)
	ch <- prometheus.MustNewConstMetric(c.UsageIowait, prometheus.GaugeValue, metric["usage_iowait"], labels...)
	ch <- prometheus.MustNewConstMetric(c.UsageIrq, prometheus.GaugeValue, metric["usage_irq"], labels...)
	ch <- prometheus.MustNewConstMetric(c.UsageSoftirq, prometheus.GaugeValue, metric["usage_softirq"], labels...)
	ch <- prometheus.MustNewConstMetric(c.UsageSteal, prometheus.GaugeValue, metric["usage_steal"], labels...)
	ch <- prometheus.MustNewConstMetric(c.UsageGuest, prometheus.GaugeValue, metric["usage_guest"], labels...)
	ch <- prometheus.MustNewConstMetric(c.UsageGuestNice, prometheus.GaugeValue, metric["usage_guest_nice"], labels...)
}

// Metric 获取 cpu 指标
func (c *CPUCollector) Metric() (map[string]float64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	lastCts := c.lastCts
	if lastCts == nil {
		ts, err := cpu.Times(false)
		if err != nil {
			return nil, err
		}
		lastCts = &ts[0]

		time.Sleep(1 * time.Second)
	}

	ts, err := cpu.Times(false)
	if err != nil {
		return nil, err
	}
	cts := &ts[0]
	c.lastCts = cts

	total := totalTime(cts)
	lastTotal := totalTime(lastCts)
	totalDelta := total - lastTotal
	if totalDelta <= 0 {
		return map[string]float64{}, nil
	}

	return map[string]float64{
		"usage_active":     100 * ((total - cts.Idle) - (lastTotal - lastCts.Idle)) / totalDelta,
		"usage_user":       100 * (cts.User - lastCts.User - (cts.Guest - lastCts.Guest)) / totalDelta,
		"usage_system":     100 * (cts.System - lastCts.System) / totalDelta,
		"usage_idle":       100 * (cts.Idle - lastCts.Idle) / totalDelta,
		"usage_nice":       100 * (cts.Nice - lastCts.Nice - (cts.GuestNice - lastCts.GuestNice)) / totalDelta,
		"usage_iowait":     100 * (cts.Iowait - lastCts.Iowait) / totalDelta,
		"usage_irq":        100 * (cts.Irq - lastCts.Irq) / totalDelta,
		"usage_softirq":    100 * (cts.Softirq - lastCts.Softirq) / totalDelta,
		"usage_steal":      100 * (cts.Steal - lastCts.Steal) / totalDelta,
		"usage_guest":      100 * (cts.Guest - lastCts.Guest) / totalDelta,
		"usage_guest_nice": 100 * (cts.GuestNice - lastCts.GuestNice) / totalDelta,
	}, nil
}

func totalTime(t *cpu.TimesStat) float64 {
	total := t.User + t.System + t.Nice + t.Iowait + t.Irq + t.Softirq + t.Steal + t.Idle
	return total
}
