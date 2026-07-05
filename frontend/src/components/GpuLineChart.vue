<template>
  <div>
    <h4 class="text-xs font-medium text-gray-500 uppercase mb-3">History (1h)</h4>
    <div class="h-48">
      <canvas ref="canvasRef"></canvas>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { Chart, LineController, LineElement, PointElement, LinearScale, TimeScale, CategoryScale, Filler, Legend, Tooltip } from 'chart.js'

Chart.register(LineController, LineElement, PointElement, LinearScale, CategoryScale, Filler, Legend, Tooltip)

const props = defineProps({
  history: { type: Array, default: () => [] },
})

const canvasRef = ref(null)
let chart = null

function createChart() {
  if (!canvasRef.value) return
  const ctx = canvasRef.value.getContext('2d')

  chart = new Chart(ctx, {
    type: 'line',
    data: {
      labels: [],
      datasets: [
        {
          label: 'GPU %',
          data: [],
          borderColor: '#16a34a',
          backgroundColor: 'rgba(22, 163, 74, 0.08)',
          fill: true,
          tension: 0.3,
          pointRadius: 0,
          borderWidth: 1.5,
          order: 1,
        },
        {
          label: 'Mem %',
          data: [],
          borderColor: '#7c3aed',
          backgroundColor: 'rgba(124, 58, 237, 0.05)',
          fill: true,
          tension: 0.3,
          pointRadius: 0,
          borderWidth: 1.5,
          order: 2,
        },
        {
          label: 'Temp °C',
          data: [],
          borderColor: '#f59e0b',
          backgroundColor: 'transparent',
          fill: false,
          tension: 0.3,
          pointRadius: 0,
          borderWidth: 1.5,
          borderDash: [4, 2],
          order: 3,
        },
        {
          label: 'Mem Temp °C',
          data: [],
          borderColor: '#ef4444',
          backgroundColor: 'transparent',
          fill: false,
          tension: 0.3,
          pointRadius: 0,
          borderWidth: 1.5,
          borderDash: [3, 3],
          order: 4,
          hidden: true,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: { duration: 200 },
      interaction: {
        intersect: false,
        mode: 'index',
      },
      scales: {
        x: {
          display: true,
          ticks: { maxTicksLimit: 6, font: { size: 10 }, color: '#9ca3af' },
          grid: { color: '#f3f4f6' },
        },
        y: {
          display: true,
          min: 0,
          max: 100,
          ticks: {
            stepSize: 25,
            font: { size: 10 },
            color: '#9ca3af',
            callback: (v) => v + '%',
          },
          grid: { color: '#f3f4f6' },
        },
      },
      plugins: {
        legend: {
          position: 'top',
          align: 'end',
          labels: {
            boxWidth: 10,
            boxHeight: 2,
            padding: 16,
            font: { size: 11 },
            color: '#6b7280',
          },
        },
        tooltip: {
          backgroundColor: '#1f2937',
          titleFont: { size: 11 },
          bodyFont: { size: 11 },
          padding: 8,
        },
      },
    },
  })
}

watch(() => props.history, (hist) => {
  if (!chart) return
  chart.data.labels = hist.map((h) => {
    const d = new Date(h.time * 1000)
    return d.getHours().toString().padStart(2, '0') + ':' +
           d.getMinutes().toString().padStart(2, '0') + ':' +
           d.getSeconds().toString().padStart(2, '0')
  })
  chart.data.datasets[0].data = hist.map((h) => h.gpu)
  chart.data.datasets[1].data = hist.map((h) => h.mem)
  chart.data.datasets[2].data = hist.map((h) => h.temp)
  chart.data.datasets[3].data = hist.map((h) => h.memTemp)
  chart.update('none')
})

onMounted(() => {
  createChart()
})

onUnmounted(() => {
  if (chart) {
    chart.destroy()
    chart = null
  }
})
</script>
