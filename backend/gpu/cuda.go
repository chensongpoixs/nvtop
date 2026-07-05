package gpu

/*
#cgo CFLAGS: -I/usr/local/cuda-12.4/targets/x86_64-linux/include
#cgo LDFLAGS: -lcuda

#include <cuda.h>
#include <stdlib.h>
*/
import "C"
import "fmt"

var cudaInitialized bool

// CUDAInit initializes the CUDA Driver API.
// Must be called once before any CUDA device attribute queries.
func CUDAInit() error {
	result := C.cuInit(0)
	if result != C.CUDA_SUCCESS {
		return fmt.Errorf("cuInit failed: %d", int(result))
	}
	cudaInitialized = true
	return nil
}

// getCUDAProps populates CUDA device attribute fields on the GPUInfo struct.
// Must be called after CUDAInit(). Errors are silent — fields stay at zero values
// and the frontend will display "-" for unavailable data.
func getCUDAProps(info *GPUInfo) {
	if !cudaInitialized {
		return
	}

	var device C.CUdevice
	result := C.cuDeviceGet(&device, C.int(info.Index))
	if result != C.CUDA_SUCCESS {
		return
	}

	// Helper: read an integer device attribute
	getAttr := func(attr C.CUdevice_attribute) int {
		var val C.int
		if C.cuDeviceGetAttribute(&val, attr, device) == C.CUDA_SUCCESS {
			return int(val)
		}
		return 0
	}

	info.SMs = getAttr(C.CU_DEVICE_ATTRIBUTE_MULTIPROCESSOR_COUNT)
	info.L2CacheSizeKB = getAttr(C.CU_DEVICE_ATTRIBUTE_L2_CACHE_SIZE) / 1024
	info.SharedMemPerBlockKB = getAttr(C.CU_DEVICE_ATTRIBUTE_MAX_SHARED_MEMORY_PER_BLOCK) / 1024
	info.SharedMemPerSMKB = getAttr(C.CU_DEVICE_ATTRIBUTE_MAX_SHARED_MEMORY_PER_MULTIPROCESSOR) / 1024
	info.RegistersPerBlock = getAttr(C.CU_DEVICE_ATTRIBUTE_MAX_REGISTERS_PER_BLOCK)
	info.MaxThreadsPerBlock = getAttr(C.CU_DEVICE_ATTRIBUTE_MAX_THREADS_PER_BLOCK)
	info.MaxThreadsPerSM = getAttr(C.CU_DEVICE_ATTRIBUTE_MAX_THREADS_PER_MULTIPROCESSOR)
	info.WarpSize = getAttr(C.CU_DEVICE_ATTRIBUTE_WARP_SIZE)
	info.ConcurrentKernels = getAttr(C.CU_DEVICE_ATTRIBUTE_CONCURRENT_KERNELS) != 0
	info.CopyEngines = getAttr(C.CU_DEVICE_ATTRIBUTE_ASYNC_ENGINE_COUNT)
	info.ComputePreemption = getAttr(C.CU_DEVICE_ATTRIBUTE_COMPUTE_PREEMPTION_SUPPORTED) != 0
	info.CooperativeLaunch = getAttr(C.CU_DEVICE_ATTRIBUTE_COOPERATIVE_LAUNCH) != 0
	info.CooperativeMultiDev = getAttr(C.CU_DEVICE_ATTRIBUTE_COOPERATIVE_MULTI_DEVICE_LAUNCH) != 0
	info.ManagedMemory = getAttr(C.CU_DEVICE_ATTRIBUTE_MANAGED_MEMORY) != 0
	info.UnifiedAddressing = getAttr(C.CU_DEVICE_ATTRIBUTE_UNIFIED_ADDRESSING) != 0
	info.Integrated = getAttr(C.CU_DEVICE_ATTRIBUTE_INTEGRATED) != 0
	info.KernelTimeout = getAttr(C.CU_DEVICE_ATTRIBUTE_KERNEL_EXEC_TIMEOUT) != 0
}
