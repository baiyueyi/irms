import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { routes } from './routes'

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.initialized && auth.token) {
    try {
      await auth.fetchMe()
    } catch (e) {
      auth.logout()
      if (!to.meta.public) {
        return '/login'
      }
      return true
    }
  }
  if (to.meta.public) {
    return true
  }
  if (!auth.token) {
    return '/login'
  }
  if (auth.user?.must_change_password && to.path !== '/change-password') {
    return '/change-password'
  }
  if (to.meta.superAdminOnly && auth.user?.role !== 'super_admin') {
    return '/my-resources'
  }
  if (auth.user?.role !== 'super_admin' && !to.meta.public && to.path !== '/change-password') {
    try {
      await auth.fetchPagePermissions()
    } catch (e) {
      auth.logout()
      return '/login'
    }
    if (!auth.canAccessPage(to.path)) {
      if (auth.canAccessPage('/my-resources') && to.path !== '/my-resources') {
        return '/my-resources'
      }
      return '/login'
    }
  }
  if (to.path === '/login') {
    return auth.user?.role === 'super_admin' ? '/admin/pages' : '/my-resources'
  }
  return true
})

export default router
