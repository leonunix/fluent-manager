<template>
  <div class="topology-graph-container">
    <div class="d-flex justify-content-end mb-2 gap-2">
      <button class="btn btn-sm btn-outline-secondary" @click="resetZoom">
        <i class="bi bi-arrows-fullscreen"></i> 重置视图
      </button>
    </div>
    <div ref="chartRef" class="chart-area"></div>

    <!-- Detail Panel -->
    <div v-if="selectedNode" class="detail-panel card border-0 shadow">
      <div class="card-header bg-white d-flex justify-content-between align-items-center py-2">
        <span>
          <i :class="selectedNode.icon" class="me-1"></i>
          <strong>{{ selectedNode.label }}</strong>
        </span>
        <button class="btn btn-sm btn-close" @click="selectedNode = null"></button>
      </div>
      <div class="card-body py-2">
        <table class="table table-sm table-borderless mb-0">
          <tr v-for="(val, key) in selectedNode.details" :key="key">
            <td class="text-muted" style="width:80px">{{ key }}</td>
            <td>{{ val }}</td>
          </tr>
        </table>
        <router-link v-if="selectedNode.type === 'cluster'"
                     :to="`/nodes?cluster_id=${selectedNode.id}`"
                     class="btn btn-sm btn-success mt-2 w-100">
          <i class="bi bi-hdd-network me-1"></i>查看节点
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import * as echarts from 'echarts/core'
import { TreeChart } from 'echarts/charts'
import {
  TooltipComponent,
  LegendComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([TreeChart, TooltipComponent, LegendComponent, CanvasRenderer])

const props = defineProps({
  tree: { type: Array, default: () => [] },
})

const router = useRouter()
const chartRef = ref(null)
const chartInstance = shallowRef(null)
const selectedNode = ref(null)

// Color constants
const COLORS = {
  dc: '#0d6efd',
  region: '#0dcaf0',
  cluster: '#198754',
  clusterWarn: '#ffc107',
  clusterError: '#dc3545',
}

function clusterColor(cl) {
  if (cl.node_count === 0) return '#6c757d'
  if (cl.online_count === cl.node_count) return COLORS.cluster
  if (cl.online_count === 0) return COLORS.clusterError
  return COLORS.clusterWarn
}

function toEChartsData(tree) {
  if (!tree || !tree.length) {
    return {
      name: '基础设施',
      children: [],
      itemStyle: { color: '#6c757d' },
      label: { color: '#333' },
    }
  }
  return {
    name: '基础设施',
    value: { type: 'root' },
    itemStyle: { color: '#6c757d', borderColor: '#6c757d' },
    label: { color: '#333', fontWeight: 'bold', fontSize: 14 },
    children: tree.map(dc => ({
      name: dc.alias || dc.name,
      value: { type: 'dc', id: dc.id, data: dc },
      itemStyle: { color: COLORS.dc, borderColor: COLORS.dc },
      label: { color: COLORS.dc, fontWeight: 'bold' },
      children: (dc.regions || []).map(r => ({
        name: r.alias || r.name,
        value: { type: 'region', id: r.id, data: r, dc },
        itemStyle: { color: COLORS.region, borderColor: COLORS.region },
        label: { color: '#0a8a9a' },
        children: (r.clusters || []).map(cl => {
          const color = clusterColor(cl)
          const envLabel = cl.environment ? ` [${cl.environment}]` : ''
          return {
            name: `${cl.alias || cl.name}${envLabel}\n${cl.online_count}/${cl.node_count}`,
            value: { type: 'cluster', id: cl.id, data: cl, dc, region: r },
            itemStyle: { color, borderColor: color },
            label: { color: '#333' },
          }
        }),
      })),
    })),
  }
}

function getOption() {
  return {
    tooltip: {
      trigger: 'item',
      formatter: (params) => {
        const v = params.data?.value
        if (!v || v.type === 'root') return '基础设施拓扑'
        if (v.type === 'dc') {
          const d = v.data
          return `<strong>${d.alias || d.name}</strong><br/>供应商: ${d.provider || '-'}<br/>区域数: ${(d.regions || []).length}`
        }
        if (v.type === 'region') {
          const r = v.data
          return `<strong>${r.alias || r.name}</strong><br/>集群数: ${(r.clusters || []).length}`
        }
        if (v.type === 'cluster') {
          const cl = v.data
          return `<strong>${cl.alias || cl.name}</strong><br/>环境: ${cl.environment || '-'}<br/>节点: ${cl.online_count} 在线 / ${cl.node_count} 总计`
        }
        return params.name
      },
    },
    series: [
      {
        type: 'tree',
        data: [toEChartsData(props.tree)],
        orient: 'TB',
        layout: 'orthogonal',
        roam: true,
        zoom: 0.9,
        symbolSize: [80, 36],
        symbol: 'roundRect',
        edgeShape: 'polyline',
        edgeForkPosition: '50%',
        initialTreeDepth: -1,
        animationDuration: 500,
        animationDurationUpdate: 400,
        label: {
          position: 'inside',
          verticalAlign: 'middle',
          fontSize: 12,
          lineHeight: 16,
        },
        lineStyle: {
          color: '#ccc',
          width: 2,
          curveness: 0,
        },
        itemStyle: {
          borderWidth: 2,
        },
        emphasis: {
          focus: 'ancestor',
          itemStyle: {
            shadowBlur: 10,
            shadowColor: 'rgba(0,0,0,0.3)',
          },
        },
        leaves: {
          label: {
            position: 'inside',
            verticalAlign: 'middle',
          },
        },
      },
    ],
  }
}

function handleClick(params) {
  const v = params.data?.value
  if (!v || v.type === 'root') return

  if (v.type === 'dc') {
    selectedNode.value = {
      type: 'dc', id: v.id, label: v.data.alias || v.data.name,
      icon: 'bi bi-building text-primary',
      details: {
        '名称': v.data.name,
        '别名': v.data.alias || '-',
        '供应商': v.data.provider || '-',
      },
    }
  } else if (v.type === 'region') {
    selectedNode.value = {
      type: 'region', id: v.id, label: v.data.alias || v.data.name,
      icon: 'bi bi-globe2 text-info',
      details: {
        '名称': v.data.name,
        '别名': v.data.alias || '-',
        '数据中心': v.dc?.alias || v.dc?.name || '-',
        '集群数': (v.data.clusters || []).length,
      },
    }
  } else if (v.type === 'cluster') {
    selectedNode.value = {
      type: 'cluster', id: v.id, label: v.data.alias || v.data.name,
      icon: 'bi bi-diagram-3 text-success',
      details: {
        '名称': v.data.name,
        '环境': v.data.environment || '-',
        '区域': v.region?.alias || v.region?.name || '-',
        '在线节点': `${v.data.online_count} / ${v.data.node_count}`,
      },
    }
  }
}

function resetZoom() {
  if (chartInstance.value) {
    chartInstance.value.setOption(getOption())
  }
}

let resizeObserver = null

onMounted(() => {
  if (!chartRef.value) return
  const instance = echarts.init(chartRef.value)
  chartInstance.value = instance
  instance.setOption(getOption())
  instance.on('click', handleClick)
  instance.on('dblclick', (params) => {
    const v = params.data?.value
    if (v?.type === 'cluster') {
      router.push(`/nodes?cluster_id=${v.id}`)
    }
  })

  resizeObserver = new ResizeObserver(() => {
    instance.resize()
  })
  resizeObserver.observe(chartRef.value)
})

onBeforeUnmount(() => {
  if (resizeObserver && chartRef.value) {
    resizeObserver.unobserve(chartRef.value)
  }
  if (chartInstance.value) {
    chartInstance.value.dispose()
  }
})

watch(() => props.tree, () => {
  if (chartInstance.value) {
    chartInstance.value.setOption(getOption())
  }
}, { deep: true })
</script>

<style scoped>
.topology-graph-container {
  position: relative;
}
.chart-area {
  width: 100%;
  height: 600px;
  background: #fafbfc;
  border: 1px solid #e9ecef;
  border-radius: 8px;
}
.detail-panel {
  position: absolute;
  top: 50px;
  right: 16px;
  width: 280px;
  z-index: 10;
}
</style>
