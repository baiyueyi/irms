<script setup>
import CrudSimple from '../components/CrudSimple.vue'

const statusOptions = [
  { label: 'active', value: 'active' },
  { label: 'inactive', value: 'inactive' }
]

const locationTypeOptions = [
  { label: 'room', value: 'room' },
  { label: 'rack', value: 'rack' },
  { label: 'idc', value: 'idc' },
  { label: 'region', value: 'region' },
  { label: 'office', value: 'office' },
  { label: 'other', value: 'other' }
]

const columns = [
  { prop: 'id', label: 'ID' },
  { prop: 'code', label: '编码', creatable: true },
  { prop: 'name', label: '名称', creatable: true },
  { prop: 'location_type', label: '位置类型', creatable: true },
  { prop: 'address', label: '地址', creatable: true },
  { prop: 'status', label: '状态', creatable: true },
  { prop: 'description', label: '描述', creatable: true }
]

const formModel = { code: '', name: '', location_type: 'room', address: '', status: 'active', description: '' }
const formFields = [
  { key: 'code', label: '编码' },
  { key: 'name', label: '名称' },
  { key: 'location_type', label: '位置类型', inputType: 'select', options: locationTypeOptions },
  { key: 'address', label: '地址' },
  { key: 'status', label: '状态', inputType: 'select', options: statusOptions },
  { key: 'description', label: '描述' }
]
const filters = [
  { key: 'location_type', label: '位置类型', options: locationTypeOptions },
  { key: 'status', label: '状态', options: statusOptions }
]
</script>

<template>
  <CrudSimple
    title="位置管理"
    endpoint="/locations"
    :id-field="'id'"
    :columns="columns"
    :form-model="formModel"
    :editable-fields="['code', 'name', 'location_type', 'address', 'status', 'description']"
    :form-fields="formFields"
    :filters="filters"
  />
</template>
