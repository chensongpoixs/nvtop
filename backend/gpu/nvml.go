package gpu

/*
#cgo CFLAGS: -I/usr/local/cuda-12.4/targets/x86_64-linux/include
#cgo LDFLAGS: -L/lib/x86_64-linux-gnu -l:libnvidia-ml.so.1 -ldl

#include <nvml.h>
#include <stdlib.h>

// Helper to get unsigned int value from a field value
static unsigned int getFieldValueUint(nvmlFieldValue_t *fv) {
	if (fv->nvmlReturn == NVML_SUCCESS) {
		return fv->value.uiVal;
	}
	return 0;
}
*/
import "C"

import (
	"fmt"
)

var initialized bool

// Init initializes the NVML library
func Init() error {
	result := C.nvmlInit()
	if result != C.NVML_SUCCESS {
		return fmt.Errorf("nvmlInit failed: %s", errorString(result))
	}
	initialized = true
	return nil
}

// Shutdown shuts down the NVML library
func Shutdown() error {
	if !initialized {
		return nil
	}
	result := C.nvmlShutdown()
	if result != C.NVML_SUCCESS {
		return fmt.Errorf("nvmlShutdown failed: %s", errorString(result))
	}
	initialized = false
	return nil
}

func errorString(result C.nvmlReturn_t) string {
	return C.GoString(C.nvmlErrorString(result))
}

// GetDeviceCount returns the number of NVIDIA GPUs
func GetDeviceCount() (int, error) {
	var count C.uint
	result := C.nvmlDeviceGetCount(&count)
	if result != C.NVML_SUCCESS {
		return 0, fmt.Errorf("nvmlDeviceGetCount failed: %s", errorString(result))
	}
	return int(count), nil
}

// GetDriverVersion returns the NVIDIA driver version
func GetDriverVersion() (string, error) {
	var buf [C.NVML_SYSTEM_DRIVER_VERSION_BUFFER_SIZE]C.char
	result := C.nvmlSystemGetDriverVersion(&buf[0], C.NVML_SYSTEM_DRIVER_VERSION_BUFFER_SIZE)
	if result != C.NVML_SUCCESS {
		return "", fmt.Errorf("nvmlSystemGetDriverVersion failed: %s", errorString(result))
	}
	return C.GoString(&buf[0]), nil
}

// GetCUDAVersion returns the CUDA driver version
func GetCUDAVersion() (string, error) {
	var version C.int
	result := C.nvmlSystemGetCudaDriverVersion_v2(&version)
	if result != C.NVML_SUCCESS {
		// Fallback to v1
		result = C.nvmlSystemGetCudaDriverVersion(&version)
		if result != C.NVML_SUCCESS {
			return "", fmt.Errorf("nvmlSystemGetCudaDriverVersion failed: %s", errorString(result))
		}
	}
	major := int(version) / 1000
	minor := (int(version) % 1000) / 10
	return fmt.Sprintf("%d.%d", major, minor), nil
}

// GetAllGPUInfo queries all GPUs and returns their information
func GetAllGPUInfo() ([]GPUInfo, error) {
	if !initialized {
		if err := Init(); err != nil {
			return nil, err
		}
	}

	count, err := GetDeviceCount()
	if err != nil {
		return nil, err
	}

	gpus := make([]GPUInfo, 0, count)
	for i := 0; i < count; i++ {
		info, err := getGPUInfo(i)
		if err != nil {
			continue
		}
		gpus = append(gpus, info)
	}
	return gpus, nil
}

func getGPUInfo(index int) (GPUInfo, error) {
	info := GPUInfo{Index: index}

	var device C.nvmlDevice_t
	result := C.nvmlDeviceGetHandleByIndex(C.uint(index), &device)
	if result != C.NVML_SUCCESS {
		return info, fmt.Errorf("nvmlDeviceGetHandleByIndex(%d) failed: %s", index, errorString(result))
	}

	// Name
	var name [C.NVML_DEVICE_NAME_BUFFER_SIZE]C.char
	if result := C.nvmlDeviceGetName(device, &name[0], C.NVML_DEVICE_NAME_BUFFER_SIZE); result == C.NVML_SUCCESS {
		info.Name = C.GoString(&name[0])
	}

	// UUID
	var uuid [C.NVML_DEVICE_UUID_BUFFER_SIZE]C.char
	if result := C.nvmlDeviceGetUUID(device, &uuid[0], C.NVML_DEVICE_UUID_BUFFER_SIZE); result == C.NVML_SUCCESS {
		info.UUID = C.GoString(&uuid[0])
	}

	// Utilization
	var util C.nvmlUtilization_t
	if result := C.nvmlDeviceGetUtilizationRates(device, &util); result == C.NVML_SUCCESS {
		info.UtilizationGPU = int(util.gpu)
		info.UtilizationMemory = int(util.memory)
	}

	// Memory
	var mem C.nvmlMemory_t
	if result := C.nvmlDeviceGetMemoryInfo(device, &mem); result == C.NVML_SUCCESS {
		info.MemoryTotalMB = uint64(mem.total) / (1024 * 1024)
		info.MemoryUsedMB = uint64(mem.used) / (1024 * 1024)
	}

	// Temperature
	var temp C.uint
	if result := C.nvmlDeviceGetTemperature(device, C.NVML_TEMPERATURE_GPU, &temp); result == C.NVML_SUCCESS {
		info.TemperatureC = uint(temp)
	}

	// Power
	var power C.uint
	if result := C.nvmlDeviceGetPowerUsage(device, &power); result == C.NVML_SUCCESS {
		info.PowerW = uint(power) / 1000
	}
	var powerLimit C.uint
	if result := C.nvmlDeviceGetPowerManagementLimit(device, &powerLimit); result == C.NVML_SUCCESS {
		info.PowerLimitW = uint(powerLimit) / 1000
	}

	// Fan speed
	var fan C.uint
	if result := C.nvmlDeviceGetFanSpeed(device, &fan); result == C.NVML_SUCCESS {
		info.FanSpeed = int(fan)
	}

	// Clock speeds
	var clock C.uint
	if result := C.nvmlDeviceGetClockInfo(device, C.NVML_CLOCK_GRAPHICS, &clock); result == C.NVML_SUCCESS {
		info.ClockCoreMHz = uint(clock)
	}
	if result := C.nvmlDeviceGetClockInfo(device, C.NVML_CLOCK_MEM, &clock); result == C.NVML_SUCCESS {
		info.ClockMemoryMHz = uint(clock)
	}

	// PCIe throughput
	var pcieRx, pcieTx C.uint
	if result := C.nvmlDeviceGetPcieThroughput(device, C.NVML_PCIE_UTIL_RX_BYTES, &pcieRx); result == C.NVML_SUCCESS {
		info.PCIeRxMbps = uint(pcieRx) * 8 / 1000
	}
	if result := C.nvmlDeviceGetPcieThroughput(device, C.NVML_PCIE_UTIL_TX_BYTES, &pcieTx); result == C.NVML_SUCCESS {
		info.PCIeTxMbps = uint(pcieTx) * 8 / 1000
	}

	// Encoder/Decoder utilization
	var encUtil, decUtil C.uint
	var encSamplingPeriod, decSamplingPeriod C.uint
	if result := C.nvmlDeviceGetEncoderUtilization(device, &encUtil, &encSamplingPeriod); result == C.NVML_SUCCESS {
		info.EncoderUtil = int(encUtil)
	}
	if result := C.nvmlDeviceGetDecoderUtilization(device, &decUtil, &decSamplingPeriod); result == C.NVML_SUCCESS {
		info.DecoderUtil = int(decUtil)
	}

	// --- Advanced metrics ---

	// PCIe link negotiation rate (current vs max)
	var pcieCurrGen, pcieMaxGen, pcieCurrWidth, pcieMaxWidth C.uint
	if result := C.nvmlDeviceGetCurrPcieLinkGeneration(device, &pcieCurrGen); result == C.NVML_SUCCESS {
		info.PCIeCurrentGen = int(pcieCurrGen)
	}
	if result := C.nvmlDeviceGetMaxPcieLinkGeneration(device, &pcieMaxGen); result == C.NVML_SUCCESS {
		info.PCIeMaxGen = int(pcieMaxGen)
	}
	if result := C.nvmlDeviceGetCurrPcieLinkWidth(device, &pcieCurrWidth); result == C.NVML_SUCCESS {
		info.PCIeCurrentWidth = int(pcieCurrWidth)
	}
	if result := C.nvmlDeviceGetMaxPcieLinkWidth(device, &pcieMaxWidth); result == C.NVML_SUCCESS {
		info.PCIeMaxWidth = int(pcieMaxWidth)
	}

	// Clocks throttle reasons
	var throttleReasons C.ulonglong
	if result := C.nvmlDeviceGetCurrentClocksThrottleReasons(device, &throttleReasons); result == C.NVML_SUCCESS {
		info.ClocksThrottleReasons = uint64(throttleReasons)
		info.ClocksThrottleReasonsText = parseThrottleReasons(uint64(throttleReasons))
	}

	// Memory temperature via Field Values API (fieldId 195 = NVML_FI_DEV_TEMPERATURE_MEM_MAX_TLIMIT)
	var memTempField [1]C.nvmlFieldValue_t
	memTempField[0].fieldId = 195
	memTempField[0].scopeId = 0
	if result := C.nvmlDeviceGetFieldValues(device, 1, &memTempField[0]); result == C.NVML_SUCCESS {
		info.MemoryTemperatureC = int(C.getFieldValueUint(&memTempField[0]))
	}

	// Performance state (P0=highest perf, P15=lowest)
	var pState C.nvmlPstates_t
	if result := C.nvmlDeviceGetPerformanceState(device, &pState); result == C.NVML_SUCCESS {
		info.PerformanceState = int(pState)
	}

	// Memory bus width and max memory clock
	var memBusWidth C.uint
	if result := C.nvmlDeviceGetMemoryBusWidth(device, &memBusWidth); result == C.NVML_SUCCESS {
		info.MemoryBusWidth = int(memBusWidth)
	}
	var maxMemClock C.uint
	if result := C.nvmlDeviceGetMaxClockInfo(device, C.NVML_CLOCK_MEM, &maxMemClock); result == C.NVML_SUCCESS {
		info.MaxMemoryClockMHz = int(maxMemClock)
	}
	// Current memory bandwidth: width(bits) × current_clock(MHz) × 2(DDR) / 8(bytes) / 1000(→GB/s)
	if info.MemoryBusWidth > 0 && info.ClockMemoryMHz > 0 {
		info.MemoryBandwidthCurrentGBps = float64(info.MemoryBusWidth) * float64(info.ClockMemoryMHz) * 2 / 8 / 1000
	}
	// Max theoretical memory bandwidth: width(bits) × max_clock(MHz) × 2 / 8 / 1000
	if info.MemoryBusWidth > 0 && info.MaxMemoryClockMHz > 0 {
		info.MemoryBandwidthGBps = float64(info.MemoryBusWidth) * float64(info.MaxMemoryClockMHz) * 2 / 8 / 1000
	}

	// BAR1 memory (PCIe BAR for GPU memory mapping, relevant for CUDA UVM)
	var bar1 C.nvmlBAR1Memory_t
	if result := C.nvmlDeviceGetBAR1MemoryInfo(device, &bar1); result == C.NVML_SUCCESS {
		info.BAR1TotalMB = uint64(bar1.bar1Total) / (1024 * 1024)
		info.BAR1UsedMB = uint64(bar1.bar1Used) / (1024 * 1024)
	}

	// ECC mode and error counts
	var eccCurrent, eccPending C.nvmlEnableState_t
	if result := C.nvmlDeviceGetEccMode(device, &eccCurrent, &eccPending); result == C.NVML_SUCCESS {
		if eccCurrent == C.NVML_FEATURE_ENABLED {
			info.ECCMode = "Enabled"
		} else if eccCurrent == C.NVML_FEATURE_DISABLED {
			info.ECCMode = "Disabled"
		}
	}
	var eccErrors C.ulonglong
	if result := C.nvmlDeviceGetTotalEccErrors(device, C.NVML_MEMORY_ERROR_TYPE_UNCORRECTED, C.NVML_AGGREGATE_ECC, &eccErrors); result == C.NVML_SUCCESS {
		info.ECCErrorsCount = uint64(eccErrors)
	}

	// Compute mode
	var compMode C.nvmlComputeMode_t
	if result := C.nvmlDeviceGetComputeMode(device, &compMode); result == C.NVML_SUCCESS {
		info.ComputeMode = computeModeString(compMode)
	}

	// NVLink state
	var nvlinkActive, nvlinkMax int
	for link := C.uint(0); link < C.NVML_NVLINK_MAX_LINKS; link++ {
		var isActive C.nvmlEnableState_t
		if result := C.nvmlDeviceGetNvLinkState(device, link, &isActive); result != C.NVML_SUCCESS {
			break
		}
		nvlinkMax++
		if isActive == C.NVML_FEATURE_ENABLED {
			nvlinkActive++
		}
	}
	info.NVLinkActiveLinks = nvlinkActive
	info.NVLinkMaxLinks = nvlinkMax

	// --- DeviceQuery-style static device properties ---

	// CUDA Compute Capability
	var ccMajor, ccMinor C.int
	if result := C.nvmlDeviceGetCudaComputeCapability(device, &ccMajor, &ccMinor); result == C.NVML_SUCCESS {
		info.ComputeCapability = fmt.Sprintf("%d.%d", int(ccMajor), int(ccMinor))
	}

	// Number of CUDA Cores
	var numCores C.uint
	if result := C.nvmlDeviceGetNumGpuCores(device, &numCores); result == C.NVML_SUCCESS {
		info.NumCores = int(numCores)
	}

	// Max SM Clock
	var maxSMClock C.uint
	if result := C.nvmlDeviceGetMaxClockInfo(device, C.NVML_CLOCK_SM, &maxSMClock); result == C.NVML_SUCCESS {
		info.MaxSMClockMHz = int(maxSMClock)
	}

	// VBIOS Version
	var vbios [C.NVML_DEVICE_VBIOS_VERSION_BUFFER_SIZE]C.char
	if result := C.nvmlDeviceGetVbiosVersion(device, &vbios[0], C.NVML_DEVICE_VBIOS_VERSION_BUFFER_SIZE); result == C.NVML_SUCCESS {
		info.VBIOSVersion = C.GoString(&vbios[0])
	}

	// GPU Brand
	var brand C.nvmlBrandType_t
	if result := C.nvmlDeviceGetBrand(device, &brand); result == C.NVML_SUCCESS {
		info.Brand = brandToString(brand)
	}

	// GPU Architecture (Kepler / Maxwell / Pascal / Volta / Turing / Ampere / Ada / Hopper / Blackwell)
	var arch C.nvmlDeviceArchitecture_t
	if result := C.nvmlDeviceGetArchitecture(device, &arch); result == C.NVML_SUCCESS {
		info.Architecture = archToString(arch)
	}

	// CUDA Driver API device attributes (SM count, L2 cache, warp size, thread limits, etc.)
	getCUDAProps(&info)

	// Running processes
	info.Processes = getProcesses(device)

	return info, nil
}

func getProcesses(device C.nvmlDevice_t) []GPUProcess {
	const maxProcs = 256
	procMap := make(map[uint]*GPUProcess)

	// Compute processes using v3 API (via macro nvmlDeviceGetComputeRunningProcesses -> _v3)
	var computeProcs [maxProcs]C.nvmlProcessInfo_t
	var computeCount C.uint = maxProcs

	result := C.nvmlDeviceGetComputeRunningProcesses(device, &computeCount, &computeProcs[0])
	if result == C.NVML_SUCCESS {
		for i := C.uint(0); i < computeCount && i < maxProcs; i++ {
			pid := uint(computeProcs[i].pid)
			procMap[pid] = &GPUProcess{
				PID:          pid,
				Name:         getProcessName(pid),
				MemoryUsedMB: uint64(computeProcs[i].usedGpuMemory) / (1024 * 1024),
				Type:         "C",
			}
		}
	}

	// Graphics processes using v3 API
	var gfxProcs [maxProcs]C.nvmlProcessInfo_t
	var gfxCount C.uint = maxProcs

	result = C.nvmlDeviceGetGraphicsRunningProcesses(device, &gfxCount, &gfxProcs[0])
	if result == C.NVML_SUCCESS {
		for i := C.uint(0); i < gfxCount && i < maxProcs; i++ {
			pid := uint(gfxProcs[i].pid)
			if existing, ok := procMap[pid]; ok {
				existing.Type = "C+G"
				existing.MemoryUsedMB += uint64(gfxProcs[i].usedGpuMemory) / (1024 * 1024)
			} else {
				procMap[pid] = &GPUProcess{
					PID:          pid,
					Name:         getProcessName(pid),
					MemoryUsedMB: uint64(gfxProcs[i].usedGpuMemory) / (1024 * 1024),
					Type:         "G",
				}
			}
		}
	}

	procs := make([]GPUProcess, 0, len(procMap))
	for _, p := range procMap {
		procs = append(procs, *p)
	}
	return procs
}

func getProcessName(pid uint) string {
	var name [256]C.char
	result := C.nvmlSystemGetProcessName(C.uint(pid), &name[0], 256)
	if result == C.NVML_SUCCESS {
		return C.GoString(&name[0])
	}
	return fmt.Sprintf("PID %d", pid)
}

// parseThrottleReasons converts a clocks throttle reasons bitmask to human-readable strings.
// Bit definitions from nvml.h:
//
//	bit 0: GpuIdle           - GPU is idle and clocks are lowered
//	bit 1: AppClocks         - Applications clocks setting is limiting
//	bit 2: SwPowerCap        - Software power cap is limiting
//	bit 3: HwSlowdown        - Hardware slowdown (thermal or power brake)
//	bit 4: SyncBoost         - Sync boost is active (raising clocks)
//	bit 5: SwThermal         - Software thermal slowdown
//	bit 6: HwThermal         - Hardware thermal slowdown
//	bit 7: HwPowerBrake      - Hardware power brake slowdown
//	bit 8: DisplayClock      - Display clock setting is limiting
//	bit 9: NwSlowdown        - Network slowdown (NVLink/Switch related)
func parseThrottleReasons(reasons uint64) []string {
	if reasons == 0 {
		return nil
	}

	bitNames := []struct {
		bit  uint64
		name string
	}{
		{1 << 0, "GPU Idle"},
		{1 << 1, "App Clock Setting"},
		{1 << 2, "SW Power Cap"},
		{1 << 3, "HW Slowdown"},
		{1 << 4, "Sync Boost"},
		{1 << 5, "SW Thermal"},
		{1 << 6, "HW Thermal"},
		{1 << 7, "HW Power Brake"},
		{1 << 8, "Display Clock"},
		{1 << 9, "NVLink Slowdown"},
	}

	var result []string
	for _, b := range bitNames {
		if reasons&b.bit != 0 {
			result = append(result, b.name)
		}
	}
	return result
}

// brandToString converts an NVML brand type enum to a human-readable string.
func brandToString(brand C.nvmlBrandType_t) string {
	switch brand {
	case C.NVML_BRAND_GEFORCE:
		return "GeForce"
	case C.NVML_BRAND_QUADRO:
		return "Quadro"
	case C.NVML_BRAND_TESLA:
		return "Tesla"
	case C.NVML_BRAND_NVS:
		return "NVS"
	case C.NVML_BRAND_GRID:
		return "GRID"
	case C.NVML_BRAND_TITAN:
		return "TITAN"
	case C.NVML_BRAND_QUADRO_RTX:
		return "Quadro RTX"
	case C.NVML_BRAND_NVIDIA_RTX:
		return "NVIDIA RTX"
	case C.NVML_BRAND_NVIDIA:
		return "NVIDIA"
	case C.NVML_BRAND_GEFORCE_RTX:
		return "GeForce RTX"
	case C.NVML_BRAND_TITAN_RTX:
		return "TITAN RTX"
	case C.NVML_BRAND_NVIDIA_VAPPS:
		return "NVIDIA vApps"
	case C.NVML_BRAND_NVIDIA_VPC:
		return "NVIDIA vPC"
	case C.NVML_BRAND_NVIDIA_VCS:
		return "NVIDIA vCS"
	case C.NVML_BRAND_NVIDIA_VWS:
		return "NVIDIA vWS"
	case C.NVML_BRAND_NVIDIA_CLOUD_GAMING:
		return "NVIDIA Cloud Gaming"
	default:
		return fmt.Sprintf("Unknown (%d)", int(brand))
	}
}

// archToString converts an NVML device architecture enum to a human-readable string.
func archToString(arch C.nvmlDeviceArchitecture_t) string {
	switch arch {
	case C.NVML_DEVICE_ARCH_KEPLER:
		return "Kepler"
	case C.NVML_DEVICE_ARCH_MAXWELL:
		return "Maxwell"
	case C.NVML_DEVICE_ARCH_PASCAL:
		return "Pascal"
	case C.NVML_DEVICE_ARCH_VOLTA:
		return "Volta"
	case C.NVML_DEVICE_ARCH_TURING:
		return "Turing"
	case C.NVML_DEVICE_ARCH_AMPERE:
		return "Ampere"
	case C.NVML_DEVICE_ARCH_ADA:
		return "Ada"
	case C.NVML_DEVICE_ARCH_HOPPER:
		return "Hopper"
	// CUDA 13.0+ defines NVML_DEVICE_ARCH_BLACKWELL = 10
	default:
		if int(arch) == 10 {
			return "Blackwell"
		}
		return fmt.Sprintf("Unknown (%d)", int(arch))
	}
}

// computeModeString converts an NVML compute mode enum to a human-readable string.
func computeModeString(mode C.nvmlComputeMode_t) string {
	switch mode {
	case C.NVML_COMPUTEMODE_DEFAULT:
		return "Default"
	case C.NVML_COMPUTEMODE_EXCLUSIVE_THREAD:
		return "Exclusive Thread"
	case C.NVML_COMPUTEMODE_PROHIBITED:
		return "Prohibited"
	case C.NVML_COMPUTEMODE_EXCLUSIVE_PROCESS:
		return "Exclusive Process"
	default:
		return fmt.Sprintf("Unknown (%d)", int(mode))
	}
}
