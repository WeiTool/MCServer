<script setup lang="ts">
import { useTpsChart, type TpsChartProps } from './TpsChart'

const props = withDefaults(defineProps<TpsChartProps>(), {
    data: () => [],
    title: 'TPS',
    color: '#4a9eff',
    width: '100%',
    height: '220px',
})

const { chartRef } = useTpsChart(props)
</script>

<template>
    <div ref="chartRef" :style="{ width, height }"></div>
</template>
