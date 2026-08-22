<script setup lang="ts">
import { useGcChart, type GcChartProps } from './GcChart'

const props = withDefaults(defineProps<GcChartProps>(), {
    data: () => [],
    width: '100%',
    height: '220px',
})

const { chartRef } = useGcChart(props)
</script>

<template>
    <div ref="chartRef" :style="{ width, height }"></div>
</template>
