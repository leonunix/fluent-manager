<template>
  <div>
    <h4 class="mb-4">审计日志</h4>
    <div class="card border-0 shadow-sm">
      <div class="card-body p-0">
        <table class="table table-hover mb-0">
          <thead>
            <tr>
              <th>时间</th>
              <th>用户</th>
              <th>操作</th>
              <th>资源</th>
              <th>详情</th>
              <th>IP</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in logs" :key="log.id">
              <td>{{ formatTime(log.created_at) }}</td>
              <td>{{ log.username || '-' }}</td>
              <td><span class="badge bg-info">{{ log.action }}</span></td>
              <td>{{ log.resource }}</td>
              <td class="text-truncate" style="max-width: 250px;">{{ log.detail }}</td>
              <td>{{ log.ip }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <nav v-if="total > pageSize" class="mt-3">
      <ul class="pagination justify-content-center">
        <li class="page-item" :class="{ disabled: page <= 1 }">
          <a class="page-link" href="#" @click.prevent="page--; loadLogs()">上一页</a>
        </li>
        <li class="page-item disabled">
          <span class="page-link">{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
        </li>
        <li class="page-item" :class="{ disabled: page >= Math.ceil(total / pageSize) }">
          <a class="page-link" href="#" @click.prevent="page++; loadLogs()">下一页</a>
        </li>
      </ul>
    </nav>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getAuditLogs } from '../api'

const logs = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 50

function formatTime(t) { return t ? new Date(t).toLocaleString('zh-CN') : '-' }

async function loadLogs() {
  const { data } = await getAuditLogs({ page: page.value, page_size: pageSize })
  logs.value = data.data || []
  total.value = data.total
}

onMounted(loadLogs)
</script>
