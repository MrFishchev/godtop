package domain

type HostInfo struct {
	CpuUsage        float64
	UsedSwapMemory  uint64
	TotalSwapMemory uint64
	UsedMemory      uint64
	TotalMemory     uint64
	UsedStorage     uint64
	TotalStorage    uint64
}
