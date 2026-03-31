import LoginPage from '../views/LoginPage.vue'
import ChangePasswordPage from '../views/ChangePasswordPage.vue'
import AdminLayout from '../views/AdminLayout.vue'
import UsersPage from '../views/UsersPage.vue'
import UserGroupsPage from '../views/UserGroupsPage.vue'
import ResourcesPage from '../views/ResourcesPage.vue'
import HostsPage from '../views/HostsPage.vue'
import ServicesPage from '../views/ServicesPage.vue'
import EnvironmentsPage from '../views/EnvironmentsPage.vue'
import LocationsPage from '../views/LocationsPage.vue'
import ResourceGroupsPage from '../views/ResourceGroupsPage.vue'
import GrantsPage from '../views/GrantsPage.vue'
import MyResourcesPage from '../views/MyResourcesPage.vue'

export const adminMenuItems = [
  { index: '/admin/users', title: '用户管理' },
  { index: '/admin/user-groups', title: '用户组管理' },
  { index: '/admin/pages', title: '页面资源管理' },
  { index: '/admin/hosts', title: '主机管理' },
  { index: '/admin/services', title: '服务管理' },
  { index: '/admin/environments', title: '环境管理' },
  { index: '/admin/locations', title: '位置管理' },
  { index: '/admin/resource-groups', title: '资源组管理' },
  { index: '/admin/grants', title: '权限管理' }
]

const pageNameByPath = {
  '/my-resources': '我的可访问页面与资源'
}

const adminChildren = [
  { path: '', redirect: '/admin/pages' },
  { path: 'users', component: UsersPage },
  { path: 'user-groups', component: UserGroupsPage },
  { path: 'pages', component: ResourcesPage },
  { path: 'resources', redirect: '/admin/pages' },
  { path: 'hosts', component: HostsPage },
  { path: 'services', component: ServicesPage },
  { path: 'environments', component: EnvironmentsPage },
  { path: 'locations', component: LocationsPage },
  { path: 'resource-groups', component: ResourceGroupsPage },
  { path: 'grants', component: GrantsPage }
]

export const routes = [
  { path: '/login', component: LoginPage, meta: { public: true } },
  { path: '/change-password', component: ChangePasswordPage },
  {
    path: '/admin',
    component: AdminLayout,
    meta: { superAdminOnly: true },
    children: adminChildren
  },
  { path: '/my-resources', component: MyResourcesPage },
  { path: '/', redirect: '/login' }
]

const menuPathSet = new Set(adminMenuItems.map((x) => x.index))
const menuNameMap = Object.fromEntries(adminMenuItems.map((x) => [x.index, x.title]))

function toAbsPath(parent, child) {
  if (!child) {
    return parent || '/'
  }
  if (child.startsWith('/')) {
    return child
  }
  if (!parent || parent === '/') {
    return `/${child}`
  }
  return `${parent.replace(/\/$/, '')}/${child}`
}

function shouldCollectPage(path, record) {
  if (!record.component) {
    return false
  }
  if (record.meta?.public) {
    return false
  }
  if (path === '/login' || path === '/change-password' || path === '/' || path === '/admin') {
    return false
  }
  if (path.includes('/:')) {
    return false
  }
  return path.startsWith('/admin/') || path === '/my-resources'
}

export function collectPageRouteCandidates() {
  const bucket = new Map()
  const walk = (records, parentPath) => {
    ;(records || []).forEach((record) => {
      const absPath = toAbsPath(parentPath, record.path)
      if (shouldCollectPage(absPath, record)) {
        const inMenu = menuPathSet.has(absPath)
        bucket.set(absPath, {
          name: menuNameMap[absPath] || pageNameByPath[absPath] || absPath,
          route_path: absPath,
          source: inMenu ? 'menu' : 'router'
        })
      }
      if (record.children?.length) {
        walk(record.children, absPath)
      }
    })
  }
  walk(routes, '')
  return Array.from(bucket.values()).sort((a, b) => a.route_path.localeCompare(b.route_path))
}
