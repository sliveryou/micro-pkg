package promcollector

import (
	stdnet "net"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v4/net"
)

// NetCollectorOpts 网络指标收集器配置
type NetCollectorOpts struct {
	Namespace       string // 命名空间
	ReportErrors    bool   // 是否报告错误
	CheckInterfaces bool   // 是否检查网卡，开启后将跳过环回和未运行的网卡
}

// NetCollector 网络指标收集器
type NetCollector struct {
	BytesSent   *prometheus.Desc
	BytesRecv   *prometheus.Desc
	PacketsSent *prometheus.Desc
	PacketsRecv *prometheus.Desc
	ErrIn       *prometheus.Desc
	ErrOut      *prometheus.Desc
	DropIn      *prometheus.Desc
	DropOut     *prometheus.Desc

	checkInterfaces bool
	BaseCollector
}

// NewNetCollector 新建网络指标收集器
func NewNetCollector(opts ...NetCollectorOpts) *NetCollector {
	var opt NetCollectorOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	bc := BaseCollector{reportErrors: opt.ReportErrors}
	ns := bc.getNamespace(opt.Namespace)

	return &NetCollector{
		BaseCollector:   bc,
		checkInterfaces: opt.CheckInterfaces,
		BytesSent: prometheus.NewDesc(
			ns+"net_bytes_sent", "Number of bytes sent by the network interface.",
			[]string{"interface"}, nil),
		BytesRecv: prometheus.NewDesc(
			ns+"net_bytes_recv", "Number of bytes received by the network interface.",
			[]string{"interface"}, nil),
		PacketsSent: prometheus.NewDesc(
			ns+"net_packets_sent", "Number of packets sent by the network interface.",
			[]string{"interface"}, nil),
		PacketsRecv: prometheus.NewDesc(
			ns+"net_packets_recv", "Number of packets received by the network interface.",
			[]string{"interface"}, nil),
		ErrIn: prometheus.NewDesc(
			ns+"net_err_in", "Number of receive errors detected by the network interface.",
			[]string{"interface"}, nil),
		ErrOut: prometheus.NewDesc(
			ns+"net_err_out", "Number of transmit errors detected by the network interface.",
			[]string{"interface"}, nil),
		DropIn: prometheus.NewDesc(
			ns+"net_drop_in", "Number of received packets dropped by the network interface.",
			[]string{"interface"}, nil),
		DropOut: prometheus.NewDesc(
			ns+"net_drop_out", "Number of transmitted packets dropped by the network interface.",
			[]string{"interface"}, nil),
	}
}

// Describe 实现 Describe 方法
func (c *NetCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.BytesSent
	ch <- c.BytesRecv
	ch <- c.PacketsSent
	ch <- c.PacketsRecv
	ch <- c.ErrIn
	ch <- c.ErrOut
	ch <- c.DropIn
	ch <- c.DropOut
}

// Collect 实现 Collect 方法
func (c *NetCollector) Collect(ch chan<- prometheus.Metric) {
	interfaces, err := stdnet.Interfaces()
	if err != nil {
		c.reportError(ch, err)
		return
	}

	interfacesByName := map[string]stdnet.Interface{}
	for _, iface := range interfaces {
		interfacesByName[iface.Name] = iface
	}

	counts, err := net.IOCounters(true)
	if err != nil {
		c.reportError(ch, err)
		return
	}

	for _, count := range counts {
		if c.checkInterfaces {
			iface, ok := interfacesByName[count.Name]
			if !ok {
				continue
			}
			if iface.Flags&stdnet.FlagLoopback == stdnet.FlagLoopback {
				continue
			}
			if iface.Flags&stdnet.FlagUp == 0 {
				continue
			}
		}

		labels := []string{count.Name}

		ch <- prometheus.MustNewConstMetric(c.BytesSent, prometheus.GaugeValue, float64(count.BytesSent), labels...)
		ch <- prometheus.MustNewConstMetric(c.BytesRecv, prometheus.GaugeValue, float64(count.BytesRecv), labels...)
		ch <- prometheus.MustNewConstMetric(c.PacketsSent, prometheus.GaugeValue, float64(count.PacketsSent), labels...)
		ch <- prometheus.MustNewConstMetric(c.PacketsRecv, prometheus.GaugeValue, float64(count.PacketsRecv), labels...)
		ch <- prometheus.MustNewConstMetric(c.ErrIn, prometheus.GaugeValue, float64(count.Errin), labels...)
		ch <- prometheus.MustNewConstMetric(c.ErrOut, prometheus.GaugeValue, float64(count.Errout), labels...)
		ch <- prometheus.MustNewConstMetric(c.DropIn, prometheus.GaugeValue, float64(count.Dropin), labels...)
		ch <- prometheus.MustNewConstMetric(c.DropOut, prometheus.GaugeValue, float64(count.Dropout), labels...)
	}
}
