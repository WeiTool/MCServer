<script setup lang="ts">
import { useIoChart, type IoChartProps } from './IoChart'

const props = withDefaults(defineProps<IoChartProps>(), {
    read: 0,
    write: 0,
    width: '100%',
    height: '200px',
})

const { chartRef } = useIoChart(props)
</script>

<template>
    <div ref="chartRef" :style="{ width, height }"></div>
</template>
