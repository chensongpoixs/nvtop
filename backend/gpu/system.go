package gpu

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type cpuSample struct {
	total    uint64
	idle     uint64
	perCore  []uint64 // total ticks per core
	perIdle  []uint64 // idle ticks per core
}

var lastCPUSample *cpuSample

// GetSystemInfo reads /proc/stat and /proc/meminfo to get CPU and memory usage
func GetSystemInfo() SystemInfo {
	info := SystemInfo{}

	info.CPUUsagePercent, info.CPUPerCorePercent = getCPUUsage()
	memTotal, memAvailable := getMemoryInfo()
	info.MemoryTotalMB = memTotal / 1024
	info.MemoryUsedMB = (memTotal - memAvailable) / 1024
	if memTotal > 0 {
		info.MemoryUsagePercent = float64(memTotal-memAvailable) / float64(memTotal) * 100
	}

	return info
}

func getCPUUsage() (float64, []float64) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, nil
	}
	defer f.Close()

	current := &cpuSample{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			current.total, current.idle = parseCPULine(line)
		} else if strings.HasPrefix(line, "cpu") {
			// Per-core: cpu0, cpu1, ...
			total, idle := parseCPULine(line)
			current.perCore = append(current.perCore, total)
			current.perIdle = append(current.perIdle, idle)
		}
	}

	result := 0.0
	var perCore []float64

	if lastCPUSample != nil && len(current.perCore) == len(lastCPUSample.perCore) {
		tdiff := float64(current.total - lastCPUSample.total)
		idiff := float64(current.idle - lastCPUSample.idle)
		if tdiff > 0 {
			result = (tdiff - idiff) / tdiff * 100
		}

		perCore = make([]float64, len(current.perCore))
		for i := range current.perCore {
			ct := float64(current.perCore[i] - lastCPUSample.perCore[i])
			ci := float64(current.perIdle[i] - lastCPUSample.perIdle[i])
			if ct > 0 {
				perCore[i] = (ct - ci) / ct * 100
			}
		}
	}

	// Store as last sample
	_ = time.Now() // trigger the sample to be stored
	lastCPUSample = current

	return result, perCore
}

func parseCPULine(line string) (total, idle uint64) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return 0, 0
	}
	// fields[0] is "cpu" or "cpuN"
	// fields[1]=user, [2]=nice, [3]=system, [4]=idle, [5]=iowait, [6]=irq, [7]=softirq, [8]=steal...
	var sum uint64
	for i := 1; i < len(fields); i++ {
		v, _ := strconv.ParseUint(fields[i], 10, 64)
		sum += v
		if i == 4 {
			idle = v // fields[4] is idle
		}
	}
	return sum, idle
}

func getMemoryInfo() (totalKB, availableKB uint64) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			totalKB = parseMemValue(line)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			availableKB = parseMemValue(line)
		}
		if totalKB > 0 && availableKB > 0 {
			break
		}
	}
	return totalKB, availableKB
}

func parseMemValue(line string) uint64 {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0
	}
	v, _ := strconv.ParseUint(fields[1], 10, 64)
	return v
}
