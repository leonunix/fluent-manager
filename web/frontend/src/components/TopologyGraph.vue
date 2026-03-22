<template>
  <div class="topology-graph-container">
    <div class="d-flex align-items-center mb-2">
      <div class="btn-group btn-group-sm me-3">
        <button class="btn btn-outline-secondary" @click="zoomLevel = Math.min(zoomLevel + 0.2, 2)"><i class="bi bi-zoom-in"></i></button>
        <button class="btn btn-outline-secondary" @click="zoomLevel = Math.max(zoomLevel - 0.2, 0.4)"><i class="bi bi-zoom-out"></i></button>
        <button class="btn btn-outline-secondary" @click="zoomLevel = 1"><i class="bi bi-arrows-fullscreen"></i></button>
      </div>
      <div class="btn-group btn-group-sm me-3">
        <button class="btn" :class="layout === 'TB' ? 'btn-primary' : 'btn-outline-primary'" @click="layout = 'TB'">从上到下</button>
        <button class="btn" :class="layout === 'LR' ? 'btn-primary' : 'btn-outline-primary'" @click="layout = 'LR'">从左到右</button>
      </div>
      <div class="d-flex align-items-center gap-3 ms-auto small text-muted">
        <span><i class="bi bi-square-fill text-primary me-1"></i>数据中心</span>
        <span><i class="bi bi-circle-fill me-1" style="color:#0dcaf0"></i>区域</span>
        <span><i class="bi bi-diamond-fill text-success me-1"></i>集群</span>
        <span><i class="bi bi-diamond-fill me-1" style="color:#6f42c1"></i>默认集群</span>
        <span><i class="bi bi-diamond-fill me-1" style="color:#fd7e14"></i>异常集群</span>
      </div>
    </div>
    <v-chart
      ref="chartRef"
      :option="chartOption"
      :autoresize="true"
      :style="{ width: '100%', height: '560px', transform: `scale(${zoomLevel})`, transformOrigin: 'top left' }"
      @click="handleClick"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { use } from 'echarts/core'
import { TreeChart } from 'echarts/charts'
import { TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'

use([CanvasRenderer, TreeChart, TooltipComponent])

const props = defineProps({
  tree: { type: Array, default: () => [] },
})

const emit = defineEmits(['select'])

const chartRef = ref(null)
const layout = ref('TB')
const zoomLevel = ref(1)

function buildTreeData(tree) {
  if (!tree || tree.length === 0) {
    return [{
      name: '暂无数据中心，请在管理视图中创建',
      itemStyle: { color: '#ccc', borderColor: '#ccc' },
      label: { color: '#999' },
    }]
  }

  const dcNodes = tree.map(dc => ({
    name: dc.alias || dc.name,
    value: { type: 'dc', id: dc.id, data: dc },
    itemStyle: { color: '#0d6efd', borderColor: '#0d6efd' },
    label: { fontWeight: 'bold', fontSize: 13 },
    symbol: 'roundRect',
    symbolSize: [18, 18],
    children: (dc.regions || []).map(r => ({
      name: r.alias || r.name,
      value: { type: 'region', id: r.id, data: r, dc },
      itemStyle: { color: '#0dcaf0', borderColor: '#0dcaf0' },
      symbol: 'circle',
      symbolSize: 14,
      children: (r.clusters || []).map(cl => {
        const envTag = cl.environment ? ` [${cl.environment}]` : ''
        const statsTag = ` ${cl.online_count}/${cl.node_count}`
        const isHealthy = cl.node_count === 0 || cl.online_count === cl.node_count
        return {
          name: (cl.alias || cl.name) + envTag + statsTag,
          value: { type: 'cluster', id: cl.id, data: cl, region: r, dc },
          itemStyle: {
            color: cl.is_default ? '#6f42c1' : isHealthy ? '#198754' : '#fd7e14',
            borderColor: cl.is_default ? '#6f42c1' : isHealthy ? '#198754' : '#fd7e14',
          },
          symbol: 'diamond',
          symbolSize: cl.is_default ? 16 : 12,
        }
      }),
    })),
  }))

  if (dcNodes.length === 1) return dcNodes
  return [{
    name: '基础设施',
    itemStyle: { color: '#6c757d', borderColor: '#6c757d' },
    label: { fontWeight: 'bold', fontSize: 14, color: '#333' },
    symbol: 'emptyCircle',
    symbolSize: 20,
    children: dcNodes,
  }]
}

const chartOption = computed(() => ({
  tooltip: {
    trigger: 'item',
    backgroundColor: 'rgba(255,255,255,0.95)',
    borderColor: '#ddd',
    textStyle: { color: '#333', fontSize: 12 },
    formatter(params) {
      const v = params.data?.value
      if (!v) return params.name
      if (v.type === 'dc') {
        const d = v.data
        return `<b style="color:#0d6efd">■ 数据中心</b><br/><b>${d.alias || d.name}</b><br/>` +
          `标识: ${d.name}<br/>` +
          `供应商: ${d.provider || '-'}`
      }
      if (v.type === 'region') {
        const d = v.data
        return `<b style="color:#0dcaf0">● 区域</b><br/><b>${d.alias || d.name}</b><br/>` +
          `标识: ${d.name}`
      }
      if (v.type === 'cluster') {
        const cl = v.data
        return `<b style="color:#198754">◆ 集群</b><br/><b>${cl.alias || cl.name}</b><br/>` +
          (cl.environment ? `环境: <span style="color:${cl.env_color}">● ${cl.environment}</span><br/>` : '') +
          `节点: <b>${cl.online_count}</b> 在线 / ${cl.node_count} 总计` +
          (cl.is_default ? '<br/><b style="color:#6f42c1">★ 默认集群</b>' : '') +
          '<br/><small style="color:#999">点击查看详情</small>'
      }
      return params.name
    },
  },
  series: [{
    type: 'tree',
    data: buildTreeData(props.tree),
    orient: layout.value,
    top: layout.value === 'TB' ? '5%' : '10%',
    left: layout.value === 'LR' ? '5%' : '10%',
    bottom: layout.value === 'TB' ? '15%' : '10%',
    right: layout.value === 'LR' ? '20%' : '10%',
    edgeShape: 'polyline',
    edgeForkPosition: '63%',
    initialTreeDepth: 4,
    roam: true,
    lineStyle: { color: '#c4c8cc', width: 1.5, curveness: 0.3 },
    label: {
      position: layout.value === 'TB' ? 'bottom' : 'right',
      fontSize: 11,
      color: '#444',
      distance: 8,
    },
    leaves: {
      label: {
        position: layout.value === 'TB' ? 'bottom' : 'right',
      },
    },
    expandAndCollapse: true,
    animationDuration: 500,
    animationDurationUpdate: 500,
    emphasis: {
      focus: 'ancestor',
      itemStyle: { borderWidth: 3 },
    },
  }],
}))

function handleClick(params) {
  const v = params.data?.value
  if (v && v.type) {
    emit('select', v)
  }
}
</script>

<style scoped>
.topology-graph-container {
  background: #fff;
  border-radius: 8px;
  padding: 12px;
  border: 1px solid #e9ecef;
  overflow: auto;
}
</style>
