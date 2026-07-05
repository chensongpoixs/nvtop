<template>
  <div class="p-4 max-w-7xl mx-auto">

    <!-- Header -->
    <header class="mb-6">
      <div class="flex items-center justify-between flex-wrap gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 tracking-tight">nvtop</h1>
          <p class="text-sm text-gray-500">NVIDIA GPU Real-time Monitor</p>
        </div>
        <div class="flex items-center gap-4 text-xs text-gray-500">
          <span v-if="snapshot?.driver_version" class="bg-gray-100 px-2 py-1 rounded">
            Driver {{ snapshot.driver_version }}
          </span>
          <span v-if="snapshot?.cuda_version" class="bg-gray-100 px-2 py-1 rounded">
            CUDA {{ snapshot.cuda_version }}
          </span>
          <span class="flex items-center gap-1.5">
            <span class="w-2 h-2 rounded-full" :class="connected ? 'bg-green-500 live-pulse' : 'bg-red-500'"></span>
            {{ connected ? 'Live' : 'Disconnected' }}
          </span>
          <span v-if="snapshot" class="bg-gray-100 px-2 py-1 rounded">
            {{ formatTime(snapshot.timestamp) }}
          </span>
        </div>
      </div>
    </header>

    <!-- System Info: CPU & Memory -->
    <section class="mb-6" v-if="snapshot?.system">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <!-- CPU Overall Gauge -->
        <div class="bg-white border border-gray-200 rounded-xl p-4 shadow-sm hover:shadow-md transition-shadow">
          <CircularGauge
            :value="snapshot.system.cpu_usage_percent"
            :max="100"
            label="CPU"
            :sub-label="snapshot.system.cpu_per_core_percent?.length + ' cores'"
            unit="%"
          />
        </div>

        <!-- Memory Gauge -->
        <div class="bg-white border border-gray-200 rounded-xl p-4 shadow-sm hover:shadow-md transition-shadow">
          <CircularGauge
            :value="snapshot.system.memory_usage_percent"
            :max="100"
            label="Memory"
            :sub-label="formatMB(snapshot.system.memory_used_mb) + ' / ' + formatMB(snapshot.system.memory_total_mb)"
            unit="%"
          />
        </div>

        <!-- CPU per-core (summary) -->
        <div class="bg-white border border-gray-200 rounded-xl p-4 shadow-sm hover:shadow-md transition-shadow md:col-span-2">
          <span class="text-xs font-medium text-gray-500 uppercase tracking-wide block mb-2">CPU Cores</span>
          <div class="flex flex-wrap gap-1.5">
            <div v-for="(pct, idx) in snapshot.system.cpu_per_core_percent" :key="idx"
              class="flex flex-col items-center">
              <div class="w-7 h-16 bg-gray-100 rounded-md relative overflow-hidden">
                <div class="absolute bottom-0 left-0 right-0 rounded-b-md transition-all duration-500"
                  :class="usageBarColor(pct)"
                  :style="{ height: pct + '%' }">
                </div>
              </div>
              <span class="text-[10px] text-gray-400 mt-0.5">C{{ idx }}</span>
              <span class="text-[10px] font-mono" :class="usageColor(pct)">{{ Math.round(pct) }}%</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- GPU Cards -->
    <div v-if="snapshot?.gpus?.length" class="space-y-4">
      <GpuCard
        v-for="gpu in snapshot.gpus"
        :key="gpu.index"
        :gpu="gpu"
        :history="historyData[gpu.index] || []"
      />
    </div>
    <div v-else class="text-center py-20 text-gray-400">
      <div class="text-5xl mb-4">GPU</div>
      <p class="text-lg">No GPU data available</p>
      <p class="text-sm mt-1" v-if="!connected">WebSocket not connected - trying to reconnect...</p>
    </div>

  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useWebSocket } from '../composables/useWebSocket.js'
import CircularGauge from './CircularGauge.vue'
import GpuCard from './GpuCard.vue'

const { data, connected } = useWebSocket()
const snapshot = ref(null)
const historyData = ref({})

const MAX_HISTORY = 3600

watch(data, (newData) => {
  if (!newData) return
  snapshot.value = newData

  if (newData.gpus) {
    newData.gpus.forEach((gpu) => {
      const existing = historyData.value[gpu.index] || []
      const updated = [...existing, {
        time: newData.timestamp,
        gpu: gpu.utilization_gpu,
        mem: gpu.utilization_memory,
        temp: gpu.temperature_c,
        power: gpu.power_w,
        memTemp: gpu.memory_temperature_c || 0,
      }]
      if (updated.length > MAX_HISTORY) {
        updated.splice(0, updated.length - MAX_HISTORY)
      }
      historyData.value[gpu.index] = updated
    })
  }
})

function formatTime(ts) {
  if (!ts) return ''
  return new Date(ts * 1000).toLocaleTimeString()
}

function formatMB(mb) {
  if (!mb) return '0 MB'
  if (mb >= 1024) return (mb / 1024).toFixed(1) + ' GB'
  return Math.round(mb) + ' MB'
}

function usageColor(pct) {
  if (pct >= 90) return 'util-high'
  if (pct >= 60) return 'util-mid'
  return 'util-low'
}

function usageBarColor(pct) {
  if (pct >= 90) return 'bg-red-500'
  if (pct >= 60) return 'bg-yellow-500'
  return 'bg-green-500'
}
</script>
