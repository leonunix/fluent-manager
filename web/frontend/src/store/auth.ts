import { defineStore } from 'pinia'
import { login as apiLogin, getProfile } from '../api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || null as string | null,
    user: JSON.parse(localStorage.getItem('user') || 'null') as any,
  }),

  getters: {
    isAuthenticated: (state) => !!state.token,
    permissions: (state) => {
      const perms = new Set<string>()
      // Direct role permissions
      for (const role of state.user?.roles || []) {
        for (const p of role.permissions || []) {
          perms.add(p.name)
        }
      }
      // Group-inherited role permissions
      for (const group of state.user?.groups || []) {
        for (const role of group.roles || []) {
          for (const p of role.permissions || []) {
            perms.add(p.name)
          }
        }
      }
      return [...perms]
    },
  },

  actions: {
    async login(username: string, password: string, authSource = 'local') {
      const { data } = await apiLogin({ username, password, auth_source: authSource })
      this.token = data.token
      this.user = data.user
      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))

      // Fetch full profile with roles/permissions
      await this.fetchProfile()
    },

    async loginWithToken(token: string) {
      this.token = token
      localStorage.setItem('token', token)
      await this.fetchProfile()
    },

    async fetchProfile() {
      const { data } = await getProfile()
      this.user = data
      localStorage.setItem('user', JSON.stringify(data))
    },

    logout() {
      this.token = null
      this.user = null
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    },

    hasPermission(resource: string, action: string) {
      return this.permissions.includes(`${resource}:${action}`)
    },
  },
})
