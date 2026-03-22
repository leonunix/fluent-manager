<template>
  <div class="d-flex">
    <!-- Sidebar -->
    <nav class="fm-sidebar vh-100 d-flex flex-column" style="width: 240px; position: fixed; z-index: 100;">
      <!-- Logo -->
      <div class="fm-sidebar-logo">
        <div class="fm-logo-icon">
          <i class="bi bi-diagram-3-fill"></i>
        </div>
        <div>
          <div class="fm-logo-title">Fluent Manager</div>
          <div class="fm-logo-subtitle">{{ t('nav.platform') }}</div>
        </div>
      </div>

      <!-- Nav links -->
      <ul class="nav flex-column flex-grow-1 px-3">
        <li class="fm-nav-section">{{ t('nav.overview') }}</li>
        <li class="nav-item">
          <router-link to="/" class="fm-nav-link" exact-active-class="active">
            <i class="bi bi-grid-1x2 me-2"></i>{{ t('nav.dashboard') }}
          </router-link>
        </li>

        <li class="fm-nav-section">{{ t('nav.infra') }}</li>
        <li class="nav-item">
          <router-link to="/topology" class="fm-nav-link" active-class="active">
            <i class="bi bi-diagram-3 me-2"></i>{{ t('nav.topology') }}
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/environments" class="fm-nav-link" active-class="active">
            <i class="bi bi-layers me-2"></i>{{ t('nav.environments') }}
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/nodes" class="fm-nav-link" active-class="active">
            <i class="bi bi-hdd-network me-2"></i>{{ t('nav.nodes') }}
          </router-link>
        </li>

        <li class="fm-nav-section">{{ t('nav.config_deploy') }}</li>
        <li class="nav-item">
          <router-link to="/configs" class="fm-nav-link" active-class="active">
            <i class="bi bi-file-earmark-code me-2"></i>{{ t('nav.configs') }}
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/deploys" class="fm-nav-link" active-class="active">
            <i class="bi bi-rocket-takeoff me-2"></i>{{ t('nav.deploys') }}
          </router-link>
        </li>

        <li class="fm-nav-section">{{ t('nav.system') }}</li>
        <li class="nav-item">
          <router-link to="/users" class="fm-nav-link" active-class="active">
            <i class="bi bi-people me-2"></i>{{ t('nav.users') }}
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/roles" class="fm-nav-link" active-class="active">
            <i class="bi bi-shield-check me-2"></i>{{ t('nav.roles') }}
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/audit" class="fm-nav-link" active-class="active">
            <i class="bi bi-journal-text me-2"></i>{{ t('nav.audit') }}
          </router-link>
        </li>
      </ul>

      <!-- User footer -->
      <div class="fm-sidebar-footer">
        <div class="fm-user-avatar">
          {{ (auth.user?.username || 'U').charAt(0).toUpperCase() }}
        </div>
        <div class="flex-grow-1 overflow-hidden">
          <div class="fm-user-name">{{ auth.user?.display_name || auth.user?.username }}</div>
          <div class="fm-user-role">{{ auth.user?.roles?.[0]?.name || 'user' }}</div>
        </div>
        <button class="btn btn-sm fm-logout-btn" @click="handleLogout" title="退出登录">
          <i class="bi bi-box-arrow-right"></i>
        </button>
      </div>
    </nav>

    <!-- Main content -->
    <main class="flex-grow-1 min-vh-100" style="margin-left: 240px; background: var(--fm-body-bg);">
      <!-- Top bar -->
      <header class="fm-topbar">
        <div class="fm-breadcrumb">
          <i class="bi bi-geo-alt-fill text-primary me-2"></i>
          <span>{{ currentPageTitle }}</span>
        </div>
        <div class="fm-topbar-right">
          <select v-model="currentLocale" class="form-select form-select-sm fm-lang-select" @change="switchLocale">
            <option v-for="l in availableLocales" :key="l" :value="l">{{ localeNames[l] }}</option>
          </select>
          <span class="fm-time">{{ currentTime }}</span>
        </div>
      </header>
      <div class="p-4">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../store/auth'
import { useI18n } from '../i18n'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const { t, locale, setLocale, availableLocales, localeNames } = useI18n()
const currentTime = ref('')
const currentLocale = ref(locale.value)
let timer = null

const pageTitleKeys = {
  'Dashboard': 'nav.dashboard',
  'Topology': 'nav.topology',
  'Environments': 'nav.environments',
  'Nodes': 'nav.nodes',
  'NodeDetail': 'nav.nodes',
  'Configs': 'nav.configs',
  'ConfigDetail': 'nav.configs',
  'Deploys': 'nav.deploys',
  'Users': 'nav.users',
  'Roles': 'nav.roles',
  'AuditLogs': 'nav.audit',
}

const currentPageTitle = computed(() => t(pageTitleKeys[route.name] || 'nav.dashboard'))

function switchLocale() {
  setLocale(currentLocale.value)
}

function updateTime() {
  currentTime.value = new Date().toLocaleString('zh-CN', { hour12: false })
}

function handleLogout() {
  auth.logout()
  router.push('/login')
}

onMounted(() => {
  updateTime()
  timer = setInterval(updateTime, 1000)
})
onUnmounted(() => clearInterval(timer))
</script>

<style scoped>
/* Sidebar */
.fm-sidebar {
  background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
  overflow-y: auto;
  overflow-x: hidden;
}

/* Logo */
.fm-sidebar-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 20px 20px 16px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  margin-bottom: 8px;
}
.fm-logo-icon {
  width: 38px;
  height: 38px;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 1.15rem;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.35);
}
.fm-logo-title {
  font-weight: 700;
  font-size: 0.95rem;
  color: #f1f5f9;
  letter-spacing: -0.01em;
  line-height: 1.2;
}
.fm-logo-subtitle {
  font-size: 0.65rem;
  color: #64748b;
  letter-spacing: 0.05em;
}

/* Nav section labels */
.fm-nav-section {
  font-size: 0.65rem;
  font-weight: 600;
  color: #475569;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  padding: 16px 12px 6px;
  list-style: none;
}

/* Nav links */
.fm-nav-link {
  display: flex;
  align-items: center;
  padding: 9px 12px;
  margin: 1px 0;
  border-radius: 8px;
  color: #94a3b8;
  text-decoration: none;
  font-size: 0.85rem;
  font-weight: 500;
  transition: all 0.15s ease;
  border-left: 3px solid transparent;
}
.fm-nav-link:hover {
  color: #e2e8f0;
  background: rgba(59, 130, 246, 0.08);
}
.fm-nav-link.active {
  color: #fff;
  background: rgba(59, 130, 246, 0.18);
  border-left-color: #3b82f6;
}
.fm-nav-link.active i {
  color: #60a5fa;
}
.fm-nav-link i {
  font-size: 1rem;
  width: 20px;
  text-align: center;
}

/* User footer */
.fm-sidebar-footer {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  margin: 4px 12px 12px;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.08);
}
.fm-user-avatar {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 700;
  font-size: 0.8rem;
  flex-shrink: 0;
}
.fm-user-name {
  font-size: 0.8rem;
  color: #e2e8f0;
  font-weight: 600;
  line-height: 1.2;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.fm-user-role {
  font-size: 0.65rem;
  color: #64748b;
}
.fm-logout-btn {
  color: #64748b;
  padding: 4px 6px;
  line-height: 1;
}
.fm-logout-btn:hover {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
}

/* Top bar */
.fm-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 52px;
  background: #fff;
  border-bottom: 1px solid var(--fm-card-border);
  position: sticky;
  top: 0;
  z-index: 50;
}
.fm-breadcrumb {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--fm-text-primary);
}
.fm-topbar-right {
  display: flex;
  align-items: center;
  gap: 16px;
}
.fm-lang-select {
  width: auto;
  font-size: 0.8rem;
  padding: 2px 28px 2px 8px;
  border-color: #e2e8f0;
  color: var(--fm-text-secondary);
}
.fm-time {
  font-size: 0.8rem;
  color: var(--fm-text-muted);
  font-variant-numeric: tabular-nums;
}
</style>
