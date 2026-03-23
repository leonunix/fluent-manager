import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '../store/auth'

const routes: RouteRecordRaw[] = [
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

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return '/login'
  }
  if (to.meta.guest && auth.isAuthenticated) {
    return '/'
  }
})

export default router
