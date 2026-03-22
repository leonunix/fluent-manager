import { defineStore } from 'pinia'
import { login as apiLogin, getProfile } from '../api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || null,
    user: JSON.parse(localStorage.getItem('user') || 'null'),
  }),

  getters: {
    isAuthenticated: (state) => !!state.token,
    permissions: (state) => {
      if (!state.user?.roles) return []
      const perms = new Set()
      for (const role of state.user.roles) {
        if (role.permissions) {
          for (const p of role.permissions) {
            perms.add(p.name)
          }
        }
      }
      return [...perms]
    },
  },

  actions: {
    async login(username, password, authSource = 'local') {
      const { data } = await apiLogin({ username, password, auth_source: authSource })
      this.token = data.token
      this.user = data.user
      localStorage.setItem('token', data.token)
      localStorage.setItem('user', JSON.stringify(data.user))

      // Fetch full profile with roles/permissions
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

    hasPermission(resource, action) {
      return this.permissions.includes(`${resource}:${action}`)
    },
  },
})
