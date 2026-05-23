<template>
  <div class="max-h-48 overflow-y-auto">
    <table class="w-full text-xs" v-if="sortedProcesses.length">
      <thead class="sticky top-0 bg-gray-50">
        <tr class="text-gray-500 uppercase">
          <th class="text-left py-2 px-2 font-medium">PID</th>
          <th class="text-left py-2 px-2 font-medium">Name</th>
          <th class="text-right py-2 px-2 font-medium">GPU Memory</th>
          <th class="text-center py-2 px-2 font-medium">Type</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="proc in sortedProcesses" :key="proc.pid"
          class="border-b border-gray-50 hover:bg-gray-50 transition-colors">
          <td class="py-1.5 px-2 font-mono text-gray-500">{{ proc.pid }}</td>
          <td class="py-1.5 px-2 max-w-[240px] truncate" :title="proc.name">{{ proc.name }}</td>
          <td class="py-1.5 px-2 text-right font-mono tabular-nums" :class="proc.memory_used_mb > 1024 ? 'text-red-600 font-semibold' : 'text-gray-700'">
            {{ formatMB(proc.memory_used_mb) }}
          </td>
          <td class="py-1.5 px-2 text-center">
            <span class="px-1.5 py-0.5 rounded text-[10px] font-medium"
              :class="proc.type === 'C' ? 'bg-blue-100 text-blue-700' :
                       proc.type === 'G' ? 'bg-purple-100 text-purple-700' :
                       'bg-cyan-100 text-cyan-700'">
              {{ proc.type }}
            </span>
          </td>
        </tr>
      </tbody>
    </table>
    <div v-else class="text-center text-gray-400 py-4 text-xs">
      No processes running on this GPU
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  processes: { type: Array, default: () => [] },
})

const sortedProcesses = computed(() => {
  return [...props.processes].sort((a, b) => b.memory_used_mb - a.memory_used_mb)
})

function formatMB(mb) {
  if (!mb) return '0 MB'
  if (mb >= 1024) return (mb / 1024).toFixed(1) + ' GB'
  return Math.round(mb) + ' MB'
}
</script>
