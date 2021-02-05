package infrastructure

import (
	"context"
	"godtop/domain"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type hostService struct {
}

func CreateHostService() *hostService {
	return &hostService{}
}

//GetInfo returns info about the host system
func (h hostService) GetInfo(ctx context.Context) *domain.HostInfo {
	cpuUsage := getCpuUsage()
	swapUsed, swapTotal := getSwapMemInfo()
	memUsed, memTotal := getMemInfo()
	storageUsed, storageTotal := getStorageInfo()

	return &domain.HostInfo{
		CpuUsage:        cpuUsage,
		UsedSwapMemory:  swapUsed,
		TotalSwapMemory: swapTotal,
		UsedMemory:      memUsed,
		TotalMemory:     memTotal,
		UsedStorage:     storageUsed,
		TotalStorage:    storageTotal,
	}
}

func getCpuUsage() float64 {
	result, err := cpu.Percent(0, false)
	if err != nil || len(result) < 0 {
		return 0
	}
	return result[0]
}

func getSwapMemInfo() (uint64, uint64) {
	info, err := mem.SwapMemory()
	if err != nil {
		return 0, 0
	}
	return info.Used, info.Total
}

func getMemInfo() (uint64, uint64) {
	info, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0
	}
	return info.Used, info.Total
}

func getStorageInfo() (uint64, uint64) {
	var used, total uint64

	devices, err := disk.Partitions(false)
	if err != nil {
		return 0, 0
	}

	for _, device := range devices {
		info, err := disk.Usage(device.Mountpoint)
		if err != nil {
			continue
		}

		used += info.Used
		total += info.Total
	}

	return used, total
}
