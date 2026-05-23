<template>
  <div class="flex flex-col items-center">
    <svg :width="size" :height="size" viewBox="0 0 120 120">
      <!-- Background track -->
      <circle
        cx="60" cy="60" :r="radius"
        fill="none"
        stroke="#f3f4f6"
        :stroke-width="strokeWidth"
        class="-rotate-90 origin-center"
        style="transform-origin: 60px 60px; transform: rotate(-90deg);"
      />
      <!-- Progress arc -->
      <circle
        cx="60" cy="60" :r="radius"
        fill="none"
        :stroke="gaugeColor"
        :stroke-width="strokeWidth"
        stroke-linecap="round"
        :stroke-dasharray="circumference"
        :stroke-dashoffset="dashOffset"
        style="transform-origin: 60px 60px; transform: rotate(-90deg);"
        class="transition-all duration-700 ease-out"
      />
      <!-- Center text -->
      <text x="60" y="54" text-anchor="middle" class="font-bold" :fill="gaugeColor" font-size="22">
        {{ displayValue }}
      </text>
      <text x="60" y="72" text-anchor="middle" fill="#9ca3af" font-size="11">
        {{ unit }}
      </text>
    </svg>
    <!-- External label -->
    <span class="text-[11px] text-gray-500 font-medium text-center -mt-1">{{ label }}</span>
    <span v-if="subLabel" class="text-[10px] text-gray-400 text-center leading-tight">{{ subLabel }}</span>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  value: { type: Number, default: 0 },
  max: { type: Number, default: 100 },
  label: { type: String, default: '' },
  subLabel: { type: String, default: '' },
  unit: { type: String, default: '%' },
  size: { type: Number, default: 100 },
})

const radius = 48
const strokeWidth = 10
const circumference = 2 * Math.PI * radius

const displayValue = computed(() => {
  const v = props.value ?? 0
  if (v % 1 === 0) return Math.round(v)
  return v.toFixed(1)
})

const dashOffset = computed(() => {
  const pct = Math.min((props.value ?? 0) / props.max, 1)
  return circumference * (1 - pct)
})

const gaugeColor = computed(() => {
  const pct = (props.value ?? 0) / props.max * 100
  if (pct >= 90) return '#dc2626'
  if (pct >= 60) return '#f59e0b'
  return '#16a34a'
})
</script>
