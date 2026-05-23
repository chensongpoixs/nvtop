package gpu

/*
#cgo CFLAGS: -I/usr/local/cuda-12.4/targets/x86_64-linux/include
#cgo LDFLAGS: -L/lib/x86_64-linux-gnu -l:libnvidia-ml.so.1 -ldl

#include <nvml.h>
#include <stdlib.h>
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
