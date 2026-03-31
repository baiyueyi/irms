<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api, { unwrap } from '../services/api'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()
const list = ref([])
const loading = ref(false)

const load = async () => {
  loading.value = true
  try {
    const data = unwrap(await api.get('/permissions/resources'))
    list.value = data.list || []
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const logout = () => {
  auth.logout()
  router.push('/login')
}

onMounted(load)
</script>

<template>
  <el-container style="min-height: 100vh">
    <el-header class="header-bar">
      <span>我的可访问资源 - {{ auth.user?.username }}</span>
      <el-button size="small" @click="logout">退出</el-button>
    </el-header>
    <el-main>
      <el-card>
        <el-table :data="list" v-loading="loading">
          <el-table-column prop="resource_key" label="资源Key" />
          <el-table-column prop="resource_name" label="资源名称" />
          <el-table-column prop="resource_type" label="资源类型" />
          <el-table-column prop="permission" label="权限" />
        </el-table>
      </el-card>
    </el-main>
  </el-container>
</template>

