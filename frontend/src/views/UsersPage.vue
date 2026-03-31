<script setup>
import CrudSimple from '../components/CrudSimple.vue'

const columns = [
  { prop: 'id', label: 'ID' },
  { prop: 'username', label: '用户名', creatable: true },
  { prop: 'role', label: '角色', creatable: true },
  { prop: 'status', label: '状态', creatable: true },
  { prop: 'must_change_password', label: '首次改密' }
]

const formModel = { username: '', password: '', role: 'user', status: 'enabled' }
const roleOptions = [{ label: '超管', value: 'super_admin' }, { label: '普通用户', value: 'user' }]
const statusOptions = [{ label: '启用', value: 'enabled' }, { label: '禁用', value: 'disabled' }]
const formFields = [
  { key: 'username', label: '用户名' },
  { key: 'password', label: '密码' },
  { key: 'role', label: '角色', inputType: 'select', options: roleOptions },
  { key: 'status', label: '状态', inputType: 'select', options: statusOptions }
]
const filters = [
  { key: 'role', label: '角色', options: roleOptions },
  { key: 'status', label: '状态', options: statusOptions }
]
</script>

<template>
  <CrudSimple title="用户管理" endpoint="/users" :columns="columns" :form-model="formModel" :editable-fields="['role', 'status']" :form-fields="formFields" :filters="filters" />
</template>
