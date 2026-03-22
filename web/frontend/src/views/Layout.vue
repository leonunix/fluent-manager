<template>
  <div class="d-flex">
    <!-- Sidebar -->
    <nav class="bg-dark text-white vh-100 p-3 d-flex flex-column" style="width: 220px; position: fixed;">
      <h5 class="mb-4">
        <i class="bi bi-diagram-3-fill"></i> Fluent Manager
      </h5>
      <ul class="nav flex-column flex-grow-1">
        <li class="nav-item">
          <router-link to="/" class="nav-link text-white" active-class="bg-primary rounded">
            <i class="bi bi-speedometer2 me-2"></i>仪表盘
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/topology" class="nav-link text-white" active-class="bg-primary rounded">
            <i class="bi bi-diagram-3 me-2"></i>基础设施拓扑
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/nodes" class="nav-link text-white" active-class="bg-primary rounded">
            <i class="bi bi-hdd-network me-2"></i>节点管理
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/configs" class="nav-link text-white" active-class="bg-primary rounded">
            <i class="bi bi-file-earmark-code me-2"></i>配置管理
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/deploys" class="nav-link text-white" active-class="bg-primary rounded">
            <i class="bi bi-rocket me-2"></i>部署任务
          </router-link>
        </li>
        <hr class="text-secondary">
        <li class="nav-item">
          <router-link to="/users" class="nav-link text-white" active-class="bg-primary rounded">
            <i class="bi bi-people me-2"></i>用户管理
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/roles" class="nav-link text-white" active-class="bg-primary rounded">
            <i class="bi bi-shield-lock me-2"></i>角色权限
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/audit" class="nav-link text-white" active-class="bg-primary rounded">
            <i class="bi bi-journal-text me-2"></i>审计日志
          </router-link>
        </li>
      </ul>
      <div class="mt-auto border-top pt-3 border-secondary">
        <div class="d-flex align-items-center">
          <i class="bi bi-person-circle me-2"></i>
          <span class="small">{{ auth.user?.username }}</span>
          <button class="btn btn-sm btn-outline-light ms-auto" @click="handleLogout">
            <i class="bi bi-box-arrow-right"></i>
          </button>
        </div>
      </div>
    </nav>

    <!-- Main content -->
    <main class="flex-grow-1 bg-light min-vh-100" style="margin-left: 220px;">
      <div class="p-4">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'

const router = useRouter()
const auth = useAuthStore()

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>
