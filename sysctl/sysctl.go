package sysctl

import (
	stdnet "net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

// Params 参数信息
type Params struct {
	IP   string // 为空时，将使用默认本地 ip 地址
	Path string // 为空时，将使用当前可执行程序所在目录
}

// System 系统信息
type System struct {
	Host    Host    // 主机信息
	CPU     CPU     // cpu 信息
	Memory  Memory  // 内存信息
	Network Network // 网络信息
	Disk    Disk    // 硬盘信息
}

// GetSystem 获取系统信息
func GetSystem(params ...Params) (*System, error) {
	var p Params
	if len(params) > 0 {
		p = params[0]
	}

	h, err := GetHost()
	if err != nil {
		return nil, errors.WithMessage(err, "get host err")
	}

	c, err := GetCPU()
	if err != nil {
		return nil, errors.WithMessage(err, "get cpu err")
	}

	m, err := GetMemory()
	if err != nil {
		return nil, errors.WithMessage(err, "get memory err")
	}

	n, err := GetNetwork(p.IP)
	if err != nil {
		return nil, errors.WithMessage(err, "get network err")
	}

	d, err := GetDisk(p.Path)
	if err != nil {
		return nil, errors.WithMessage(err, "get disk err")
	}

	return &System{
		Host:    *h,
		CPU:     *c,
		Memory:  *m,
		Network: *n,
		Disk:    *d,
	}, nil
}

// Host 主机信息
type Host struct {
	Name string // 名称
	OS   string // 系统
	Arch string // 架构
}

// GetHost 获取主机信息
func GetHost() (*Host, error) {
	hi, err := host.Info()
	if err != nil {
		return nil, errors.WithMessage(err, "host info err")
	}

	return &Host{
		Name: hi.Hostname,
		OS:   hi.OS,
		Arch: hi.KernelArch,
	}, nil
}

// CPU 信息
type CPU struct {
	Num        int     // 核心数量
	Usage      float64 // 使用率
	ModelName  string  // 型号
	LoadAvg1m  float64 // 1分钟内平均负载
	LoadAvg5m  float64 // 5分钟内平均负责
	LoadAvg15m float64 // 15分钟内平均负责
}

// GetCPU 获取 cpu 信息
func GetCPU() (*CPU, error) {
	// 获取 cpu 逻辑核心数量
	num, err := cpu.Counts(true)
	if err != nil {
		return nil, errors.WithMessage(err, "cpu counts err")
	}

	// 获取 1s 内 cpu 总使用率
	usage, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, errors.WithMessage(err, "cpu percent err")
	}

	// 获取 cpu 型号
	modelName := ""
	infos, err := cpu.Info()
	if err != nil {
		return nil, errors.WithMessage(err, "cpu info err")
	}
	for _, info := range infos {
		if info.ModelName != "" {
			modelName = info.ModelName
			break
		}
	}

	// 获取 cpu 平均负载
	l, err := load.Avg()
	if err != nil {
		return nil, errors.WithMessage(err, "load avg err")
	}

	return &CPU{
		Num:        num,
		Usage:      usage[0],
		ModelName:  modelName,
		LoadAvg1m:  l.Load1,
		LoadAvg5m:  l.Load5,
		LoadAvg15m: l.Load15,
	}, nil
}

// Memory 内存信息
type Memory struct {
	Size      uint64  // 内存大小
	Used      uint64  // 已用内存
	Free      uint64  // 空闲内存
	Usage     float64 // 内存使用率
	SwapSize  uint64  // 交换内存大小
	SwapUsed  uint64  // 已用交换内存
	SwapFree  uint64  // 空闲交换内存
	SwapUsage float64 // 交换内存使用率
}

// GetMemory 获取内存信息
func GetMemory() (*Memory, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, errors.WithMessage(err, "virtual memory err")
	}

	sm, err := mem.SwapMemory()
	if err != nil {
		return nil, errors.WithMessage(err, "swap memory err")
	}

	return &Memory{
		Size:      vm.Total,
		Used:      vm.Used,
		Free:      vm.Free,
		Usage:     vm.UsedPercent,
		SwapSize:  sm.Total,
		SwapUsed:  sm.Used,
		SwapFree:  sm.Free,
		SwapUsage: sm.UsedPercent,
	}, nil
}

// GetLocalIP 获取本地 ip 地址
func GetLocalIP() string {
	conn, err := stdnet.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	return strings.Split(conn.LocalAddr().String(), ":")[0]
}

// GetInterfaceNameByIP 获取指定 ip 对应的网卡名称
func GetInterfaceNameByIP(ip string) string {
	ifaces, err := stdnet.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if n, ok := addr.(*stdnet.IPNet); ok && n.IP.To4() != nil {
				if n.IP.String() == ip {
					return iface.Name
				}
			}
		}
	}

	return ""
}

// Network 网络信息
type Network struct {
	Name        string // 网卡名称
	BytesSent   uint64 // 发送字节数
	BytesRecv   uint64 // 接收字节数
	SpeedSent   uint64 // 发送速率，bit/s
	SpeedRecv   uint64 // 接收速率，bit/s
	PacketsSent uint64 // 发送数据包数
	PacketsRecv uint64 // 接收数据包数
	ErrIn       uint64 // 传入数据包错误总数
	ErrOut      uint64 // 传出数据表错误总数
	DropIn      uint64 // 传入数据包丢弃总数
	DropOut     uint64 // 传出数据包丢弃总数
}

// GetNetwork 获取网络信息，不传入 ip 时，将使用默认本地 ip 地址
func GetNetwork(ip ...string) (*Network, error) {
	var localIP string
	if len(ip) > 0 && ip[0] != "" {
		localIP = ip[0]
	} else {
		localIP = GetLocalIP()
	}
	iName := GetInterfaceNameByIP(localIP)

	return GetNetworkByInterfaceName(iName)
}

// GetNetworkByInterfaceName 获取指定网卡名称的网络信息
func GetNetworkByInterfaceName(interfaceName string) (*Network, error) {
	n1, err := networkByInterfaceName(interfaceName)
	if err != nil {
		return nil, err
	}

	// 等待 1s 钟
	var t uint64 = 1
	time.Sleep(time.Duration(t) * time.Second)

	n2, err := networkByInterfaceName(interfaceName)
	if err != nil {
		return nil, err
	}
	n2.SpeedSent = (n2.BytesSent - n1.BytesSent) / t * 8
	n2.SpeedRecv = (n2.BytesRecv - n1.BytesRecv) / t * 8

	return n2, nil
}

func networkByInterfaceName(interfaceName string) (*Network, error) {
	counts, err := net.IOCounters(true)
	if err != nil {
		return nil, errors.WithMessage(err, "net io counters err")
	}

	for _, count := range counts {
		if count.Name == interfaceName {
			return &Network{
				Name:        count.Name,
				BytesSent:   count.BytesSent,
				BytesRecv:   count.BytesRecv,
				PacketsSent: count.PacketsSent,
				PacketsRecv: count.PacketsRecv,
				ErrIn:       count.Errin,
				ErrOut:      count.Errout,
				DropIn:      count.Dropin,
				DropOut:     count.Dropout,
			}, nil
		}
	}

	return nil, errors.New("network not found")
}

// Disk 硬盘信息
type Disk struct {
	Path       string  // 路径
	FsType     string  // 文件系统类型
	Size       uint64  // 硬盘大小
	Used       uint64  // 已用硬盘
	Free       uint64  // 空闲硬盘
	Usage      float64 // 硬盘使用率
	ReadSpeed  uint64  // 读取速率，Byte/s
	WriteSpeed uint64  // 写入速率，Byte/s
}

// GetDisk 查询硬盘信息，不传入 path 时，将使用当前可执行程序所在目录
func GetDisk(path ...string) (*Disk, error) {
	var p string
	if len(path) > 0 && path[0] != "" {
		abs, err := filepath.Abs(path[0])
		if err != nil {
			return nil, errors.WithMessage(err, "filepath abs err")
		}

		p = abs
	} else {
		dir, err := ExecutableFolder()
		if err != nil {
			return nil, errors.WithMessage(err, "executable folder err")
		}

		p = dir
	}

	us, err := disk.Usage(p)
	if err != nil {
		return nil, errors.WithMessage(err, "disk usage err")
	}

	s1, err := getIOCountersStat()
	if err != nil {
		return nil, err
	}

	// 等待 1s 钟
	var t uint64 = 1
	time.Sleep(time.Duration(t) * time.Second)

	s2, err := getIOCountersStat()
	if err != nil {
		return nil, err
	}

	return &Disk{
		Path:       us.Path,
		FsType:     us.Fstype,
		Size:       us.Total,
		Used:       us.Used,
		Free:       us.Free,
		Usage:      us.UsedPercent,
		ReadSpeed:  (s2.ReadBytes - s1.ReadBytes) / t,
		WriteSpeed: (s2.WriteBytes - s1.WriteBytes) / t,
	}, nil
}

func getIOCountersStat() (*disk.IOCountersStat, error) {
	counters, err := disk.IOCounters()
	if err != nil {
		return nil, errors.WithMessage(err, "disk io counters err")
	}

	var stat disk.IOCountersStat
	for _, s := range counters {
		stat.ReadCount += s.ReadCount
		stat.MergedReadCount += s.MergedReadCount
		stat.WriteCount += s.WriteCount
		stat.MergedWriteCount += s.MergedWriteCount
		stat.ReadBytes += s.ReadBytes
		stat.WriteBytes += s.WriteBytes
		stat.ReadTime += s.ReadTime
		stat.WriteTime += s.WriteTime
	}

	return &stat, nil
}

// Executable 获取当前可执行程序所在路径
func Executable() (string, error) {
	e, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Clean(e), nil
}

// ExecutableFolder 获取当前可执行程序所在目录
func ExecutableFolder() (string, error) {
	e, err := Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(e), nil
}
