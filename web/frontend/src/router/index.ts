import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '../store/auth'
import { getSetupStatus } from '../api/setup'
import { exchangeSAMLCode } from '../api/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/setup',
    name: 'Setup',
    component: () => import('../views/Setup.vue'),
    meta: { setup: true },
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { guest: true },
  },
  {
    path: '/',
    component: () => import('../views/Layout.vue'),
    meta: { requiresAuth: true },
    children: [
      { path: '', name: 'Dashboard', component: () => import('../views/Dashboard.vue') },
      { path: 'topology', name: 'Topology', component: () => import('../views/Topology.vue') },
      { path: 'environments', name: 'Environments', component: () => import('../views/Environments.vue') },
      { path: 'nodes', name: 'Nodes', component: () => import('../views/Nodes.vue') },
      { path: 'nodes/:id', name: 'NodeDetail', component: () => import('../views/NodeDetail.vue') },
      { path: 'aggregation-groups', name: 'AggregationGroups', component: () => import('../views/AggregationGroups.vue') },
      { path: 'pipelines', name: 'Pipelines', component: () => import('../views/Pipelines.vue') },
      { path: 'configs', name: 'Configs', component: () => import('../views/Configs.vue') },
      { path: 'configs/:id', name: 'ConfigDetail', component: () => import('../views/ConfigDetail.vue') },
      { path: 'deploys', name: 'Deploys', component: () => import('../views/Deploys.vue') },
      { path: 'runtime', name: 'Runtime', component: () => import('../views/Runtime.vue') },
      { path: 'agent-policies', name: 'AgentPolicies', component: () => import('../views/AgentPolicies.vue') },
      { path: 'users', name: 'Users', component: () => import('../views/Users.vue') },
      { path: 'roles', name: 'Roles', component: () => import('../views/Roles.vue') },
      { path: 'audit', name: 'AuditLogs', component: () => import('../views/AuditLogs.vue') },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Cache setup status to avoid repeated API calls within the same session
let setupStatusCache: boolean | null = null

// Reset cache after setup completes (called by Setup.vue after initialization)
export function resetSetupCache() {
  setupStatusCache = true
}

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  // Handle SAML one-time code exchange
  const samlCode = to.query.saml_code as string
  if (samlCode) {
    try {
      const res = await exchangeSAMLCode(samlCode)
      await auth.loginWithToken(res.data.token)
    } catch {
      // Code invalid/expired — fall through to login
    }
    // Strip saml_code from URL regardless of success
    return { path: to.path, query: {}, replace: true }
  }

  // Check setup status if not yet cached
  if (setupStatusCache === null) {
    try {
      const res = await getSetupStatus()
      setupStatusCache = res.data.initialized
    } catch {
      // If API is unreachable, assume initialized to avoid blocking
      setupStatusCache = true
    }
  }

  // System not initialized: force setup page
  if (!setupStatusCache && !to.meta.setup) {
    return '/setup'
  }

  // System initialized: block access to setup page
  if (setupStatusCache && to.meta.setup) {
    return '/login'
  }

  // Normal auth guards
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return '/login'
  }
  if (to.meta.guest && auth.isAuthenticated) {
    return '/'
  }
})

export default router
