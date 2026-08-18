<script setup lang="ts">
import { useGaugeChart, type GaugeChartProps } from './GaugeChart'

const props = withDefaults(defineProps<GaugeChartProps>(), {
    max: 100,
    title: '',
    color: '#4a9eff',
    width: '100%',
    height: '300px',
})

const { chartRef } = useGaugeChart(props)
</script>

<template>
    <div ref="chartRef" :style="{ width, height }"></div>
</template>
