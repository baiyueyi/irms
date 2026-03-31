import { defineStore } from 'pinia'
import api, { unwrap } from '../services/api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('irms_token') || '',
    user: null,
    initialized: false,
    allowedPageRoutePaths: [],
    pagePermissionsLoaded: false
  }),
  actions: {
    async login(username, password) {
      const data = unwrap(await api.post('/auth/login', { username, password }))
      this.token = data.token
      this.user = data.user
      this.initialized = true
      this.allowedPageRoutePaths = []
      this.pagePermissionsLoaded = false
      localStorage.setItem('irms_token', data.token)
      return data.user
    },
    async fetchMe() {
      const data = unwrap(await api.get('/me'))
      this.user = data
      this.initialized = true
      return data
    },
    async fetchPagePermissions(force = false) {
      if (this.pagePermissionsLoaded && !force) {
        return this.allowedPageRoutePaths
      }
      const data = unwrap(await api.get('/permissions/resources'))
      const routes = (data.list || [])
        .filter((x) => x.resource_type === 'page')
        .map((x) => x.route_path)
        .filter(Boolean)
      this.allowedPageRoutePaths = Array.from(new Set(routes))
      this.pagePermissionsLoaded = true
      return this.allowedPageRoutePaths
    },
    canAccessPage(path) {
      return this.allowedPageRoutePaths.includes(path)
    },
    async changePassword(oldPassword, newPassword) {
      await api.post('/auth/change-password', { old_password: oldPassword, new_password: newPassword })
      await this.fetchMe()
    },
    logout() {
      this.token = ''
      this.user = null
      this.initialized = true
      this.allowedPageRoutePaths = []
      this.pagePermissionsLoaded = false
      localStorage.removeItem('irms_token')
    }
  }
})
