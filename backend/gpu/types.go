package gpu

// GPUInfo holds all information about a single GPU
type GPUInfo struct {
	Index            int          `json:"index"`
	Name             string       `json:"name"`
	UUID             string       `json:"uuid"`
	UtilizationGPU   int          `json:"utilization_gpu"`
	UtilizationMemory int         `json:"utilization_memory"`
	MemoryTotalMB    uint64       `json:"memory_total_mb"`
	MemoryUsedMB     uint64       `json:"memory_used_mb"`
	TemperatureC     uint         `json:"temperature_c"`
	PowerW           uint         `json:"power_w"`
	PowerLimitW      uint         `json:"power_limit_w"`
	FanSpeed         int          `json:"fan_speed"`
	ClockCoreMHz     uint         `json:"clock_core_mhz"`
	ClockMemoryMHz   uint         `json:"clock_memory_mhz"`
	PCIeRxMbps       uint         `json:"pcie_rx_mbps"`
	PCIeTxMbps       uint         `json:"pcie_tx_mbps"`
	EncoderUtil      int          `json:"encoder_util"`
	DecoderUtil      int          `json:"decoder_util"`
	Processes        []GPUProcess `json:"processes"`

	// Advanced metrics
	PCIeCurrentGen           int      `json:"pcie_current_gen"`
	PCIeCurrentWidth         int      `json:"pcie_current_width"`
	PCIeMaxGen               int      `json:"pcie_max_gen"`
	PCIeMaxWidth             int      `json:"pcie_max_width"`
	ClocksThrottleReasons    uint64   `json:"clocks_throttle_reasons"`
	ClocksThrottleReasonsText []string `json:"clocks_throttle_reasons_text"`
	MemoryTemperatureC       int      `json:"memory_temperature_c"`
	PerformanceState         int      `json:"performance_state"`
	MemoryBusWidth           int      `json:"memory_bus_width"`
	MaxMemoryClockMHz        int      `json:"max_memory_clock_mhz"`
	MemoryBandwidthCurrentGBps float64 `json:"memory_bandwidth_current_gbps"` // 当前运行时带宽
	MemoryBandwidthGBps      float64  `json:"memory_bandwidth_gbps"`          // 理论最大带宽
	BAR1TotalMB              uint64   `json:"bar1_total_mb"`
	BAR1UsedMB               uint64   `json:"bar1_used_mb"`
	ECCMode                  string   `json:"ecc_mode"`
	ECCErrorsCount           uint64   `json:"ecc_errors_count"`
	ComputeMode              string   `json:"compute_mode"`
	NVLinkActiveLinks        int      `json:"nvlink_active_links"`
	NVLinkMaxLinks           int      `json:"nvlink_max_links"`
}

// GPUProcess holds information about a process running on a GPU
type GPUProcess struct {
	PID          uint   `json:"pid"`
	Name         string `json:"name"`
	MemoryUsedMB uint64 `json:"memory_used_mb"`
	Type         string `json:"type"` // "C" for compute, "G" for graphics, "C+G" for both
}

// SystemInfo holds CPU and memory usage information
type SystemInfo struct {
	CPUUsagePercent    float64   `json:"cpu_usage_percent"`
	CPUPerCorePercent  []float64 `json:"cpu_per_core_percent"`
	CPUTemperatureC    float64   `json:"cpu_temperature_c"` // CPU 封装温度
	MemoryTotalMB      uint64    `json:"memory_total_mb"`
	MemoryUsedMB       uint64    `json:"memory_used_mb"`
	MemoryUsagePercent float64   `json:"memory_usage_percent"`
}

// Snapshot is the complete data payload sent to clients
type Snapshot struct {
	Timestamp     int64      `json:"timestamp"`
	DriverVersion string     `json:"driver_version"`
	CUDAVersion   string     `json:"cuda_version"`
	GPUs          []GPUInfo  `json:"gpus"`
	System        SystemInfo `json:"system"`
}
