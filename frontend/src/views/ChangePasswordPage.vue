<script setup>
import { reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const form = reactive({ oldPassword: '', newPassword: '' })

const onSubmit = async () => {
  try {
    await auth.changePassword(form.oldPassword, form.newPassword)
    ElMessage.success('密码修改成功')
    router.push(auth.user?.role === 'super_admin' ? '/admin/users' : '/my-resources')
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '修改失败')
  }
}
</script>

<template>
  <div class="center-page">
    <el-card class="auth-card">
      <template #header>首次登录请修改密码</template>
      <el-form @submit.prevent="onSubmit" label-position="top">
        <el-form-item label="旧密码">
          <el-input v-model="form.oldPassword" type="password" show-password />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="form.newPassword" type="password" show-password />
        </el-form-item>
        <el-button type="primary" style="width: 100%" @click="onSubmit">确认修改</el-button>
      </el-form>
    </el-card>
  </div>
</template>

