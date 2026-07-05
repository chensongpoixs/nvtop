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

      <!-- Advanced Metrics: 4-column grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-0 border-t border-gray-100 divide-y md:divide-y-0 md:divide-x divide-gray-100">

        <!-- Column 1: Performance -->
        <div class="p-4">
          <h4 class="text-xs font-medium text-gray-500 uppercase mb-3">Performance</h4>
          <div class="space-y-2 text-xs">
            <!-- P-State -->
            <div class="flex items-center justify-between">
              <span class="text-gray-400">P-State</span>
              <span class="pstate-badge" :class="pstateClass(gpu.performance_state)">P{{ gpu.performance_state }}</span>
            </div>
            <!-- Compute Mode -->
            <div class="flex items-center justify-between">
              <span class="text-gray-400">Compute Mode</span>
              <span class="font-mono text-[10px] bg-gray-100 px-1.5 py-0.5 rounded">{{ gpu.compute_mode || '-' }}</span>
            </div>
            <!-- Clocks Throttle Reasons -->
            <div>
              <span class="text-gray-400 block mb-1">Throttle</span>
              <div class="flex flex-wrap gap-1">
                <template v-if="gpu.clocks_throttle_reasons_text?.length">
                  <span v-for="reason in gpu.clocks_throttle_reasons_text" :key="reason"
                    class="throttle-tag" :class="throttleClass(reason)">{{ reason }}</span>
                </template>
                <span v-else class="font-mono text-gray-400">-</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Column 2: Memory -->
        <div class="p-4">
          <h4 class="text-xs font-medium text-gray-500 uppercase mb-3">Memory</h4>
          <div class="space-y-2 text-xs">
            <!-- Memory Bus Width -->
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">Bus Width</span>
              <span class="font-mono">{{ dashNum(gpu.memory_bus_width, ' bit') }}</span>
            </div>
            <!-- Theoretical Bandwidth -->
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">Max BW</span>
              <span class="font-mono">{{ dashFmt(gpu.memory_bandwidth_gbps, formatBandwidth) }}</span>
            </div>
            <!-- Max Memory Clock -->
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">Max Mem Clock</span>
              <span class="font-mono">{{ dashNum(gpu.max_memory_clock_mhz, ' MHz') }}</span>
            </div>
            <!-- Memory Temperature -->
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">Mem Temp</span>
              <span class="font-mono" :class="gpu.memory_temperature_c ? tempColor(gpu.memory_temperature_c) : ''">{{ dashNum(gpu.memory_temperature_c, '°C') }}</span>
            </div>
            <!-- BAR1 Memory -->
            <div>
              <span class="text-gray-400 block mb-1">BAR1 Usage</span>
              <div class="w-full bg-gray-100 rounded-full h-2">
                <div class="h-2 rounded-full transition-all duration-500"
                  :class="gpu.bar1_total_mb ? (gpu.bar1_used_mb / gpu.bar1_total_mb > 0.8 ? 'bg-red-500' : 'bg-blue-500') : 'bg-gray-300'"
                  :style="{ width: gpu.bar1_total_mb ? bar1Percent + '%' : '0%' }">
                </div>
              </div>
              <span class="text-[10px] text-gray-400 mt-0.5">{{ gpu.bar1_total_mb ? formatMB(gpu.bar1_used_mb) + ' / ' + formatMB(gpu.bar1_total_mb) : '-' }}</span>
            </div>
          </div>
        </div>

        <!-- Column 3: I/O -->
        <div class="p-4">
          <h4 class="text-xs font-medium text-gray-500 uppercase mb-3">I/O</h4>
          <div class="space-y-2 text-xs">
            <!-- PCIe Link -->
            <div>
              <span class="text-gray-400 block mb-1">PCIe Link</span>
              <span class="pcie-link">
                <span class="pcie-current">{{ gpu.pcie_current_gen ? 'Gen' + gpu.pcie_current_gen + ' ×' + gpu.pcie_current_width : '-' }}</span>
                <span v-if="gpu.pcie_max_gen" class="pcie-max"> → Max Gen{{ gpu.pcie_max_gen }} ×{{ gpu.pcie_max_width }}</span>
              </span>
              <span v-if="pcieDegraded" class="text-[10px] text-yellow-600 block mt-0.5">⚠ Running below max</span>
            </div>
            <!-- PCIe Throughput -->
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">PCIe RX</span>
              <span class="font-mono">{{ gpu.pcie_rx_mbps }} Mbps</span>
            </div>
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">PCIe TX</span>
              <span class="font-mono">{{ gpu.pcie_tx_mbps }} Mbps</span>
            </div>
            <!-- NVLink -->
            <div>
              <span class="text-gray-400 block mb-1">NVLink</span>
              <span class="font-mono">{{ gpu.nvlink_max_links ? gpu.nvlink_active_links + ' / ' + gpu.nvlink_max_links + ' active' : '-' }}</span>
            </div>
          </div>
        </div>

        <!-- Column 4: Reliability -->
        <div class="p-4">
          <h4 class="text-xs font-medium text-gray-500 uppercase mb-3">Reliability</h4>
          <div class="space-y-2 text-xs">
            <!-- ECC -->
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">ECC</span>
              <span class="font-mono" :class="gpu.ecc_mode === 'Enabled' ? 'text-green-600' : gpu.ecc_mode === 'Disabled' ? 'text-gray-500' : ''">{{ gpu.ecc_mode || '-' }}</span>
            </div>
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">ECC Errors</span>
              <span class="font-mono" :class="gpu.ecc_errors_count > 0 ? 'ecc-warn' : ''">{{ gpu.ecc_errors_count }}</span>
            </div>
            <!-- Encoder/Decoder -->
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">Encoder</span>
              <span class="font-mono">{{ gpu.encoder_util }}%</span>
            </div>
            <div class="flex justify-between py-1 border-b border-gray-50">
              <span class="text-gray-400">Decoder</span>
              <span class="font-mono">{{ gpu.decoder_util }}%</span>
            </div>
            <!-- UUID -->
            <div class="py-1">
              <span class="text-gray-400 block">UUID</span>
              <span class="font-mono text-[10px] break-all" :title="gpu.uuid">{{ gpu.uuid?.substring(0, 24) }}…</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Process list at bottom -->
      <div class="p-4 border-t border-gray-100">
        <h4 class="text-xs font-medium text-gray-500 uppercase mb-3">
          Processes
          <span class="text-gray-300 font-normal">({{ gpu.processes?.length || 0 }})</span>
        </h4>
        <ProcessTable :processes="gpu.processes || []" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import CircularGauge from './CircularGauge.vue'
import GpuLineChart from './GpuLineChart.vue'
import ProcessTable from './ProcessTable.vue'

const props = defineProps({
  gpu: { type: Object, required: true },
  history: { type: Array, default: () => [] },
})

const expanded = ref(true)

// Computed
const bar1Percent = computed(() => {
  if (!props.gpu.bar1_total_mb) return 0
  return Math.min((props.gpu.bar1_used_mb / props.gpu.bar1_total_mb) * 100, 100)
})

const pcieDegraded = computed(() => {
  return props.gpu.pcie_current_gen && props.gpu.pcie_max_gen &&
    (props.gpu.pcie_current_gen < props.gpu.pcie_max_gen ||
     props.gpu.pcie_current_width < props.gpu.pcie_max_width)
})

// Formatting
function dashNum(val, suffix = '') {
  if (!val && val !== 0) return '-'
  if (val === 0) return '-'
  return val + suffix
}

function dashFmt(val, fmtFn) {
  if (!val && val !== 0) return '-'
  if (val === 0) return '-'
  return fmtFn(val)
}

function formatMB(mb) {
  if (!mb) return '0 MB'
  if (mb >= 1024) return (mb / 1024).toFixed(1) + ' GB'
  return Math.round(mb) + ' MB'
}

function formatBandwidth(gbps) {
  if (!gbps) return ''
  if (gbps >= 1000) return (gbps / 1000).toFixed(2) + ' TB/s'
  return gbps.toFixed(0) + ' GB/s'
}

// Colors
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

function pstateClass(ps) {
  if (ps === 0) return 'pstate-p0'
  if (ps <= 2) return 'pstate-p2'
  if (ps <= 5) return 'pstate-p5'
  if (ps <= 8) return 'pstate-p8'
  return 'pstate-other'
}

function throttleClass(reason) {
  if (reason.includes('Thermal')) return 'throttle-thermal'
  if (reason.includes('Power') || reason.includes('Brake')) return 'throttle-power'
  return 'throttle-other'
}
</script>
