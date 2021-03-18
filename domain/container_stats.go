package domain

type ContainerStats struct {
	RxBytes     int64
	TxBytes     int64
	UsedMemory  int64
	MemoryUsage float32
	CpuUsage    float32
	DisplayName string
}
