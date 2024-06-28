package sysctl

import (
	"fmt"
	"testing"

	humanize "github.com/dustin/go-humanize"
	"github.com/stretchr/testify/require"
)

func TestGetSystem(t *testing.T) {
	s, err := GetSystem()
	require.NoError(t, err)

	fmt.Printf("%+v\n", s)
}

func TestGetSystem2(t *testing.T) {
	s, err := GetSystem(Params{
		IP:   "127.0.0.1",
		Path: ".",
	})
	require.NoError(t, err)

	fmt.Printf("%+v\n", s)
}

func TestGetHost(t *testing.T) {
	h, err := GetHost()
	require.NoError(t, err)

	fmt.Println("机器名称：" + h.Name)
	fmt.Println("机器系统：" + h.OS)
	fmt.Println("机器架构：" + h.Arch)
}

func TestGetCPU(t *testing.T) {
	c, err := GetCPU()
	require.NoError(t, err)

	fmt.Println("CPU 型号：" + c.ModelName)
	fmt.Printf("CPU 核心数量：%d\n", c.Num)
	fmt.Printf("CPU 使用率：%f%%\n", c.Usage)
	fmt.Printf("CPU 1分钟内平均负载：%f%%\n", c.LoadAvg1m)
	fmt.Printf("CPU 5分钟内平均负载：%f%%\n", c.LoadAvg5m)
	fmt.Printf("CPU 15分钟内平均负载：%f%%\n", c.LoadAvg15m)
}

func TestGetMemory(t *testing.T) {
	m, err := GetMemory()
	require.NoError(t, err)

	fmt.Printf("内存大小：%d B = %s\n", m.Size, humanize.IBytes(m.Size))
	fmt.Printf("已用内存：%d B = %s\n", m.Used, humanize.IBytes(m.Used))
	fmt.Printf("空闲内存：%d B = %s\n", m.Free, humanize.IBytes(m.Free))
	fmt.Printf("内存使用率：%f%%\n", m.Usage)
	fmt.Printf("交换内存大小：%d B = %s\n", m.SwapSize, humanize.IBytes(m.SwapSize))
	fmt.Printf("已用交换内存：%d B = %s\n", m.SwapUsed, humanize.IBytes(m.SwapUsed))
	fmt.Printf("空闲交换内存：%d B = %s\n", m.SwapFree, humanize.IBytes(m.SwapFree))
	fmt.Printf("交换内存使用率：%f%%\n", m.SwapUsage)
}

func TestGetLocalIP(t *testing.T) {
	t.Log(GetLocalIP())
}

func TestGetInterfaceNameByIP(t *testing.T) {
	t.Log(GetInterfaceNameByIP("127.0.0.1"))
	t.Log(GetInterfaceNameByIP(GetLocalIP()))
	t.Log(GetInterfaceNameByIP("192.168.2.109"))
	t.Log(GetInterfaceNameByIP("192.168.64.1"))
}

func TestGetNetwork(t *testing.T) {
	n, err := GetNetwork()
	require.NoError(t, err)

	fmt.Println("网卡名称：" + n.Name)
	fmt.Printf("发送字节数：%d B = %s\n", n.BytesSent, humanize.Bytes(n.BytesSent))
	fmt.Printf("接收字节数：%d B = %s\n", n.BytesRecv, humanize.Bytes(n.BytesRecv))
	fmt.Printf("发送速率：%d bit/s\n", n.SpeedSent)
	fmt.Printf("接收速率：%d bit/s\n", n.SpeedRecv)
	fmt.Printf("发送数据包数：%d\n", n.PacketsSent)
	fmt.Printf("接收数据包数：%d\n", n.PacketsRecv)
	fmt.Printf("传入数据包错误总数：%d\n", n.ErrIn)
	fmt.Printf("传出数据包错误总数：%d\n", n.ErrOut)
	fmt.Printf("传入数据包丢弃总数：%d\n", n.DropIn)
	fmt.Printf("传出数据包丢弃总数：%d\n", n.DropOut)
}

func TestGetDisk(t *testing.T) {
	d, err := GetDisk()
	require.NoError(t, err)

	fmt.Println("路径：" + d.Path)
	fmt.Println("文件系统类型：" + d.FsType)
	fmt.Printf("硬盘大小：%d B = %s\n", d.Size, humanize.Bytes(d.Size))
	fmt.Printf("已用硬盘：%d B = %s\n", d.Used, humanize.Bytes(d.Used))
	fmt.Printf("空闲硬盘：%d B = %s\n", d.Free, humanize.Bytes(d.Free))
	fmt.Printf("硬盘使用率：%f%%\n", d.Usage)
	fmt.Printf("读取速率：%d Byte/s\n", d.ReadSpeed)
	fmt.Printf("写入速率：%d Byte/s\n", d.WriteSpeed)
}
