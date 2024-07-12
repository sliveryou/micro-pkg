package promcollector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v4/disk"
)

// DiskIOCollectorOpts 硬盘指标收集器配置
type DiskIOCollectorOpts struct {
	Namespace    string // 命名空间
	ReportErrors bool   // 是否报告错误
}

// DiskIOCollector 硬盘指标收集器
type DiskIOCollector struct {
	Reads          *prometheus.Desc
	Writes         *prometheus.Desc
	ReadBytes      *prometheus.Desc
	WriteBytes     *prometheus.Desc
	ReadTime       *prometheus.Desc
	WriteTime      *prometheus.Desc
	IoTime         *prometheus.Desc
	IopsInProgress *prometheus.Desc

	BaseCollector
}

// NewDiskIOCollector 新建硬盘指标收集器
func NewDiskIOCollector(opts ...DiskIOCollectorOpts) *DiskIOCollector {
	var opt DiskIOCollectorOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	bc := BaseCollector{reportErrors: opt.ReportErrors}
	ns := bc.getNamespace(opt.Namespace)

	return &DiskIOCollector{
		BaseCollector: bc,
		Reads: prometheus.NewDesc(
			ns+"diskio_reads", "Number of device read operations.",
			[]string{"name"}, nil),
		Writes: prometheus.NewDesc(
			ns+"diskio_writes", "Number of device write operations.",
			[]string{"name"}, nil),
		ReadBytes: prometheus.NewDesc(
			ns+"diskio_read_bytes", "Number of bytes read from the device.",
			[]string{"name"}, nil),
		WriteBytes: prometheus.NewDesc(
			ns+"diskio_write_bytes", "Number of bytes written to the device.",
			[]string{"name"}, nil),
		ReadTime: prometheus.NewDesc(
			ns+"diskio_read_time", "Number of milliseconds that read requests have waited on the device.",
			[]string{"name"}, nil),
		WriteTime: prometheus.NewDesc(
			ns+"diskio_write_time", "Number of milliseconds that write requests have waited on the device.",
			[]string{"name"}, nil),
		IoTime: prometheus.NewDesc(
			ns+"diskio_io_time", "Number of milliseconds during which the device has had I/O requests queued.",
			[]string{"name"}, nil),
		IopsInProgress: prometheus.NewDesc(
			ns+"diskio_iops_in_progress", "Number of I/O requests that have been issued to the device driver but have not yet completed.",
			[]string{"name"}, nil),
	}
}

// Describe 实现 Describe 方法
func (c *DiskIOCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Reads
	ch <- c.Writes
	ch <- c.ReadBytes
	ch <- c.WriteBytes
	ch <- c.ReadTime
	ch <- c.WriteTime
	ch <- c.IoTime
	ch <- c.IopsInProgress
}

// Collect 实现 Collect 方法
func (c *DiskIOCollector) Collect(ch chan<- prometheus.Metric) {
	counts, err := disk.IOCounters()
	if err != nil {
		c.reportError(ch, err)
		return
	}

	unique := make(map[string]struct{})
	for _, dio := range counts {
		if _, ok := unique[dio.Name]; ok {
			continue
		}

		unique[dio.Name] = struct{}{}
		labels := []string{dio.Name}

		ch <- prometheus.MustNewConstMetric(c.Reads, prometheus.GaugeValue, float64(dio.ReadCount), labels...)
		ch <- prometheus.MustNewConstMetric(c.Writes, prometheus.GaugeValue, float64(dio.WriteCount), labels...)
		ch <- prometheus.MustNewConstMetric(c.ReadBytes, prometheus.GaugeValue, float64(dio.ReadBytes), labels...)
		ch <- prometheus.MustNewConstMetric(c.WriteBytes, prometheus.GaugeValue, float64(dio.WriteBytes), labels...)
		ch <- prometheus.MustNewConstMetric(c.ReadTime, prometheus.GaugeValue, float64(dio.ReadTime), labels...)
		ch <- prometheus.MustNewConstMetric(c.WriteTime, prometheus.GaugeValue, float64(dio.WriteTime), labels...)
		ch <- prometheus.MustNewConstMetric(c.IoTime, prometheus.GaugeValue, float64(dio.IoTime), labels...)
		ch <- prometheus.MustNewConstMetric(c.IopsInProgress, prometheus.GaugeValue, float64(dio.IopsInProgress), labels...)
	}
}
