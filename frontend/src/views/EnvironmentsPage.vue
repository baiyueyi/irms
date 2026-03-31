<script setup>
import CrudSimple from '../components/CrudSimple.vue'

const statusOptions = [
  { label: 'active', value: 'active' },
  { label: 'inactive', value: 'inactive' }
]

const columns = [
  { prop: 'id', label: 'ID' },
  { prop: 'code', label: '编码', creatable: true },
  { prop: 'name', label: '名称', creatable: true },
  { prop: 'status', label: '状态', creatable: true },
  { prop: 'description', label: '描述', creatable: true }
]

const formModel = { code: '', name: '', status: 'active', description: '' }
const formFields = [
  { key: 'code', label: '编码' },
  { key: 'name', label: '名称' },
  { key: 'status', label: '状态', inputType: 'select', options: statusOptions },
  { key: 'description', label: '描述' }
]
const filters = [{ key: 'status', label: '状态', options: statusOptions }]
</script>

<template>
  <CrudSimple
    title="环境管理"
    endpoint="/environments"
    :id-field="'id'"
    :columns="columns"
    :form-model="formModel"
    :editable-fields="['code', 'name', 'status', 'description']"
    :form-fields="formFields"
    :filters="filters"
  />
</template>
