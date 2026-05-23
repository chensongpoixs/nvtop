package api

import (
	"encoding/json"
	"net/http"
	"time"

	"nvtop-server/gpu"
)

// HandleGPUSnapshot returns current GPU and system info as JSON
func HandleGPUSnapshot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	gpus, err := gpu.GetAllGPUInfo()
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	driverVer, _ := gpu.GetDriverVersion()
	cudaVer, _ := gpu.GetCUDAVersion()

	snapshot := gpu.Snapshot{
		Timestamp:     time.Now().Unix(),
		DriverVersion: driverVer,
		CUDAVersion:   cudaVer,
		GPUs:          gpus,
		System:        gpu.GetSystemInfo(),
	}

	json.NewEncoder(w).Encode(snapshot)
}
