<script setup>
import { reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const form = reactive({ username: '', password: '' })

const onSubmit = async () => {
  try {
    const user = await auth.login(form.username, form.password)
    if (user.must_change_password) {
      router.push('/change-password')
      return
    }
    router.push(user.role === 'super_admin' ? '/admin/users' : '/my-resources')
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '登录失败')
  }
}
</script>

<template>
  <div class="center-page">
    <el-card class="auth-card">
      <template #header>IRMS 登录</template>
      <el-form @submit.prevent="onSubmit" label-position="top">
        <el-form-item label="用户名">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-button type="primary" style="width: 100%" @click="onSubmit">登录</el-button>
      </el-form>
    </el-card>
  </div>
</template>

