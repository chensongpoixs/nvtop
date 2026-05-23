<template>
  <div class="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden">
    <!-- GPU Header -->
    <div class="px-5 py-3 border-b border-gray-100 flex items-center justify-between flex-wrap gap-3"
         :class="gpu.utilization_gpu >= 90 ? 'bg-red-50' : gpu.utilization_gpu >= 60 ? 'bg-yellow-50' : 'bg-gray-50'">
      <div class="flex items-center gap-3">
        <span class="text-xs font-mono text-gray-400 bg-white border border-gray-200 px-2 py-0.5 rounded">GPU {{ gpu.index }}</span>
        <span class="font-semibold text-gray-800 text-sm">{{ gpu.name }}</span>
      </div>
      <div class="flex items-center gap-5 text-sm">
        <span :class="usageColor(gpu.utilization_gpu)" class="font-bold text-xl tabular-nums">
          {{ gpu.utilization_gpu }}%
        </span>
        <button @click="expanded = !expanded"
          class="text-gray-400 hover:text-gray-600 transition-colors p-1">
          <svg :class="{ 'rotate-180': expanded }" class="w-5 h-5 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
          </svg>
        </button>
      </div>
    </div>

    <!-- Summary metrics row with circular gauges -->
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 p-4 border-b border-gray-100">
      <!-- GPU Utilization Gauge -->
      <CircularGauge
        :value="gpu.utilization_gpu"
        :max="100"
        label="GPU"
        unit="%"
        :size="90"
      />

      <!-- Memory Gauge -->
      <CircularGauge
        :value="gpu.utilization_memory"
        :max="100"
        label="Memory"
        :sub-label="formatMB(gpu.memory_used_mb) + '/' + formatMB(gpu.memory_total_mb)"
        unit="%"
        :size="90"
      />

      <!-- Temperature -->
      <div class="flex flex-col items-center justify-center">
        <span class="text-lg font-bold tabular-nums" :class="tempColor(gpu.temperature_c)">
          {{ gpu.temperature_c }}<span class="text-sm">°C</span>
        </span>
        <span class="text-[11px] text-gray-500 font-medium">Temp</span>
      </div>

      <!-- Power -->
      <div class="flex flex-col items-center justify-center">
        <span class="text-lg font-bold tabular-nums">{{ gpu.power_w }}<span class="text-sm">W</span></span>
        <span class="text-[11px] text-gray-500 font-medium">Power</span>
        <span class="text-[10px] text-gray-400">{{ gpu.power_limit_w }}W limit</span>
      </div>

      <!-- Fan Speed -->
      <div class="flex flex-col items-center justify-center">
        <span class="text-lg font-bold tabular-nums">{{ gpu.fan_speed }}<span class="text-sm">%</span></span>
        <span class="text-[11px] text-gray-500 font-medium">Fan</span>
      </div>

      <!-- Clocks -->
      <div class="flex flex-col items-center justify-center">
        <div class="text-sm font-bold tabular-nums">
          <div>{{ gpu.clock_core_mhz }}<span class="text-xs">MHz</span></div>
          <div class="text-gray-400">{{ gpu.clock_memory_mhz }}<span class="text-xs">MHz</span></div>
        </div>
        <span class="text-[11px] text-gray-500 font-medium">Core/Mem</span>
      </div>
    </div>

    <!-- Expanded detail -->
    <div v-if="expanded" class="border-t border-gray-100">
      <!-- Charts -->
      <div class="p-4">
        <GpuLineChart :history="history" />
      </div>

      <!-- Extra metrics + Processes -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-0 lg:gap-0 border-t border-gray-100">
        <div class="p-4 border-b lg:border-b-0 lg:border-r border-gray-100">
          <h4 class="text-xs font-medium text-gray-500 uppercase mb-3">Details</h4>
          <div class="grid grid-cols-2 gap-2 text-xs">
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">PCIe RX</span>
              <span class="font-mono">{{ gpu.pcie_rx_mbps }} Mbps</span>
            </div>
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">PCIe TX</span>
              <span class="font-mono">{{ gpu.pcie_tx_mbps }} Mbps</span>
            </div>
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">Encoder</span>
              <span class="font-mono">{{ gpu.encoder_util }}%</span>
            </div>
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">Decoder</span>
              <span class="font-mono">{{ gpu.decoder_util }}%</span>
            </div>
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">UUID</span>
              <span class="font-mono text-[10px] truncate max-w-[160px]" :title="gpu.uuid">{{ gpu.uuid }}</span>
            </div>
          </div>
        </div>

        <!-- Process list -->
        <div class="p-4 lg:col-span-2">
          <h4 class="text-xs font-medium text-gray-500 uppercase mb-3">
            Processes
            <span class="text-gray-300 font-normal">({{ gpu.processes?.length || 0 }})</span>
          </h4>
          <ProcessTable :processes="gpu.processes || []" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import CircularGauge from './CircularGauge.vue'
import GpuLineChart from './GpuLineChart.vue'
import ProcessTable from './ProcessTable.vue'

const props = defineProps({
  gpu: { type: Object, required: true },
  history: { type: Array, default: () => [] },
})

const expanded = ref(true)

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

function tempColor(c) {
  if (c >= 80) return 'temp-high'
  if (c >= 60) return 'temp-warm'
  return 'temp-normal'
}
</script>
