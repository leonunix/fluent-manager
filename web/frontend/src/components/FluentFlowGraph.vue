<template>
  <div class="fm-flow-graph">
    <div class="d-flex flex-wrap align-items-center gap-3 mb-3 small text-muted">
      <span><i class="bi bi-circle-fill me-1" style="color:#2563eb"></i>{{ t('flow_graph.cluster') }}</span>
      <span><i class="bi bi-circle-fill me-1" style="color:#0d9488"></i>{{ t('flow_graph.aggregation_group') }}</span>
      <span><i class="bi bi-circle-fill me-1" style="color:#f97316"></i>{{ t('flow_graph.output') }}</span>
      <span><i class="bi bi-circle-fill me-1" style="color:#7c3aed"></i>{{ t('flow_graph.selector') }}</span>
    </div>
    <v-chart :option="option" :autoresize="true" style="height: 520px;" />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from '../i18n'
import { use } from 'echarts/core'
import { GraphChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'

use([CanvasRenderer, GraphChart, TooltipComponent, LegendComponent])

const props = defineProps({
  graph: {
    type: Object,
    default: () => ({ nodes: [], edges: [] }),
  },
})
const { t } = useI18n()

const categories = [
  { name: 'cluster' },
  { name: 'aggregation_group' },
  { name: 'output' },
  { name: 'selector' },
]

function nodeColor(type, health) {
  if (type === 'cluster') return health === 'healthy' ? '#2563eb' : '#60a5fa'
  if (type === 'aggregation_group') return health === 'healthy' ? '#0d9488' : '#5eead4'
  if (type === 'output') return health === 'healthy' ? '#f97316' : '#fdba74'
  return '#7c3aed'
}

const option = computed(() => ({
  tooltip: {
    formatter(params) {
      if (params.dataType === 'edge') {
        return `<b>${params.data.label}</b><br/>${t('flow_graph.edge_protocol')}: ${params.data.protocol || '-'}`
      }
      return `<b>${params.data.name}</b><br/>${t('flow_graph.node_type')}: ${params.data.node_type}<br/>${t('flow_graph.node_status')}: ${params.data.health || '-'}<br/>${params.data.description || ''}`
    },
  },
  legend: [{ data: categories.map((item) => item.name) }],
  series: [{
    type: 'graph',
    layout: 'force',
    roam: true,
    draggable: true,
    force: {
      repulsion: 320,
      edgeLength: [120, 220],
    },
    label: {
      show: true,
      position: 'right',
      formatter: '{b}',
      fontSize: 12,
    },
    lineStyle: {
      color: '#94a3b8',
      width: 2,
      curveness: 0.15,
    },
    edgeLabel: {
      show: true,
      formatter: (params) => params.data.protocol || '',
      fontSize: 10,
      color: '#64748b',
    },
    categories,
    data: (props.graph?.nodes || []).map((node) => ({
      id: node.id,
      name: node.label,
      node_type: node.node_type,
      health: node.health,
      description: node.description,
      category: categories.findIndex((item) => item.name === node.node_type),
      symbolSize: node.node_type === 'aggregation_group' ? 44 : 36,
      itemStyle: {
        color: nodeColor(node.node_type, node.health),
      },
    })),
    links: (props.graph?.edges || []).map((edge) => ({
      source: edge.source,
      target: edge.target,
      label: edge.label,
      protocol: edge.protocol,
    })),
  }],
}))
</script>

<style scoped>
.fm-flow-graph {
  background: #fff;
  border-radius: 12px;
}
</style>
