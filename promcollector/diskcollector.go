package promcollector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v4/disk"
)

// DiskCollectorOpts 硬盘指标收集器配置
type DiskCollectorOpts struct {
	Namespace    string // 命名空间
	ReportErrors bool   // 是否报告错误
}

// DiskCollector 硬盘指标收集器
type DiskCollector struct {
	TotalBytes        *prometheus.Desc
	FreeBytes         *prometheus.Desc
	UsedBytes         *prometheus.Desc
	UsedPercent       *prometheus.Desc
	InodesTotal       *prometheus.Desc
	InodesFree        *prometheus.Desc
	InodesUsed        *prometheus.Desc
	InodesUsedPercent *prometheus.Desc

	reportErrors bool
}

// NewDiskCollector 新建硬盘指标收集器
func NewDiskCollector(opts ...DiskCollectorOpts) *DiskCollector {
	var opt DiskCollectorOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	ns := ""
	if opt.Namespace != "" {
		ns = opt.Namespace + "_"
	}

	return &DiskCollector{
		reportErrors: opt.ReportErrors,
		TotalBytes: prometheus.NewDesc(
			ns+"disk_total_bytes", "Total number of bytes of space on the disk.",
			[]string{"path", "device", "fstype"}, nil),
		FreeBytes: prometheus.NewDesc(
			ns+"disk_free_bytes", "Number of bytes of free space on the disk.",
			[]string{"path", "device", "fstype"}, nil),
		UsedBytes: prometheus.NewDesc(
			ns+"disk_used_bytes", "Number of bytes of used space on the disk.",
			[]string{"path", "device", "fstype"}, nil),
		UsedPercent: prometheus.NewDesc(
			ns+"disk_used_percent", "Percentage of used space on the disk.",
			[]string{"path", "device", "fstype"}, nil),
		InodesTotal: prometheus.NewDesc(
			ns+"disk_inodes_total", "Total number of index nodes reserved on the disk.",
			[]string{"path", "device", "fstype"}, nil),
		InodesFree: prometheus.NewDesc(
			ns+"disk_inodes_free", "Number of index nodes available on the disk.",
			[]string{"path", "device", "fstype"}, nil),
		InodesUsed: prometheus.NewDesc(
			ns+"disk_inodes_used", "Number of index nodes used on the disk.",
			[]string{"path", "device", "fstype"}, nil),
		InodesUsedPercent: prometheus.NewDesc(
			ns+"disk_inodes_used_percent", "Percentage of index nodes used on the disk.",
			[]string{"path", "device", "fstype"}, nil),
	}
}

// Describe 实现 Describe 方法
func (c *DiskCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.TotalBytes
	ch <- c.FreeBytes
	ch <- c.UsedBytes
	ch <- c.UsedPercent
	ch <- c.InodesTotal
	ch <- c.InodesFree
	ch <- c.InodesUsed
	ch <- c.InodesUsedPercent
}

// Collect 实现 Collect 方法
func (c *DiskCollector) Collect(ch chan<- prometheus.Metric) {
	parts, err := disk.Partitions(true)
	if err != nil {
		c.reportError(ch, nil, err)
		return
	}

	unique := make(map[string]struct{})
	for _, part := range parts {
		if _, ok := unique[part.Mountpoint]; ok {
			continue
		}
		du, err := disk.Usage(part.Mountpoint)
		if err != nil {
			continue
		}
		if du.Total == 0 {
			continue
		}

		unique[part.Mountpoint] = struct{}{}
		labels := []string{du.Path, strings.ReplaceAll(part.Device, "/dev/", ""), du.Fstype}

		ch <- prometheus.MustNewConstMetric(c.TotalBytes, prometheus.GaugeValue, float64(du.Total), labels...)
		ch <- prometheus.MustNewConstMetric(c.FreeBytes, prometheus.GaugeValue, float64(du.Free), labels...)
		ch <- prometheus.MustNewConstMetric(c.UsedBytes, prometheus.GaugeValue, float64(du.Used), labels...)
		ch <- prometheus.MustNewConstMetric(c.UsedPercent, prometheus.GaugeValue, du.UsedPercent, labels...)
		ch <- prometheus.MustNewConstMetric(c.InodesTotal, prometheus.GaugeValue, float64(du.InodesTotal), labels...)
		ch <- prometheus.MustNewConstMetric(c.InodesFree, prometheus.GaugeValue, float64(du.InodesFree), labels...)
		ch <- prometheus.MustNewConstMetric(c.InodesUsed, prometheus.GaugeValue, float64(du.InodesUsed), labels...)
		ch <- prometheus.MustNewConstMetric(c.InodesUsedPercent, prometheus.GaugeValue, du.InodesUsedPercent, labels...)
	}
}

func (c *DiskCollector) reportError(ch chan<- prometheus.Metric, desc *prometheus.Desc, err error) {
	if !c.reportErrors {
		return
	}
	if desc == nil {
		desc = prometheus.NewInvalidDesc(err)
	}
	ch <- prometheus.NewInvalidMetric(desc, err)
}
