<script setup>
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { adminMenuItems } from '../router/routes'

const router = useRouter()
const auth = useAuthStore()

const logout = () => {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <el-container style="min-height: 100vh">
    <el-aside width="220px" class="menu-panel">
      <h3>IRMS 管理端</h3>
      <el-menu router :default-active="$route.path">
        <el-menu-item v-for="item in adminMenuItems" :key="item.index" :index="item.index">
          {{ item.title }}
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header-bar">
        <span>{{ auth.user?.username }}</span>
        <el-button size="small" @click="logout">退出</el-button>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>
