<script setup>
import { ref, onMounted } from 'vue'
import CrudSimple from '../components/CrudSimple.vue'
import CredentialsDialog from '../components/CredentialsDialog.vue'
import EnvironmentBindingsInline from '../components/EnvironmentBindingsInline.vue'
import api, { unwrap } from '../services/api'

const statusOptions = [
  { label: 'active', value: 'active' },
  { label: 'inactive', value: 'inactive' }
]
const providerOptions = [
  { label: 'physical', value: 'physical' },
  { label: 'vm', value: 'vm' },
  { label: 'cloud_instance', value: 'cloud_instance' },
  { label: 'other', value: 'other' }
]
const cloudVendorOptions = ref([
  { label: 'AWS', value: 'aws' },
  { label: '阿里云', value: 'aliyun' },
  { label: '腾讯云', value: 'tencent' },
  { label: '华为云', value: 'huawei' },
  { label: 'Azure', value: 'azure' },
  { label: 'GCP', value: 'gcp' }
])

const columns = [
  { prop: 'id', label: 'ID' },
  { prop: 'name', label: '名称', creatable: true },
  { prop: 'hostname', label: '主机名', creatable: true },
  { prop: 'primary_address', label: '主地址', creatable: true },
  { prop: 'provider_kind', label: '资源类型', creatable: true },
  { prop: 'location', label: '位置' },
  { prop: 'environments', label: '环境标签', renderType: 'envTags', maxTags: 2 },
  { prop: 'status', label: '状态', creatable: true },
  { prop: 'description', label: '描述', creatable: true }
]

const formModel = {
  name: '',
  hostname: '',
  primary_address: '',
  provider_kind: 'physical',
  cloud_vendor: '',
  cloud_instance_id: '',
  os_type: '',
  location_id: '',
  status: 'active',
  description: ''
}
const formFields = [
  { key: 'name', label: '名称' },
  { key: 'hostname', label: '主机名' },
  { key: 'primary_address', label: '主地址' },
  { key: 'provider_kind', label: '资源类型', inputType: 'select', options: providerOptions },
  { key: 'cloud_vendor', label: '云厂商', inputType: 'select', filterable: true, options: cloudVendorOptions.value },
  { key: 'cloud_instance_id', label: '云实例ID' },
  { key: 'os_type', label: '系统类型' },
  { key: 'location_id', label: '位置', inputType: 'select', nullable: true, filterable: true, options: [] },
  { key: 'status', label: '状态', inputType: 'select', options: statusOptions },
  { key: 'description', label: '描述' }
]
const filters = [
  { key: 'provider_kind', label: '资源类型', options: providerOptions },
  { key: 'status', label: '状态', options: statusOptions }
]

const credentialVisible = ref(false)
const currentHostId = ref('')
const currentHostName = ref('')
const crudRefreshToken = ref(0)

const openCredentialDialog = (row) => {
  currentHostId.value = row.id
  currentHostName.value = row.name
  credentialVisible.value = true
}

const onUpdated = () => {
  crudRefreshToken.value += 1
}

const syncCloudVendorField = () => {
  const field = formFields.find((x) => x.key === 'cloud_vendor')
  if (field) field.options = cloudVendorOptions.value
}

const loadLocations = async () => {
  try {
    const data = unwrap(await api.get('/locations', { params: { page: 1, page_size: 200 } }))
    const options = (data.list || []).map((x) => ({ label: x.name, value: x.id }))
    const field = formFields.find((x) => x.key === 'location_id')
    if (field) field.options = options
  } catch (e) {
  }
}

const loadCloudVendors = async () => {
  try {
    const [hostData, serviceData] = await Promise.all([
      api.get('/hosts', { params: { page: 1, page_size: 200 } }).then(unwrap),
      api.get('/services', { params: { page: 1, page_size: 200 } }).then(unwrap)
    ])
    const preset = cloudVendorOptions.value.map((x) => x.value)
    const merged = new Set(preset)
    ;(hostData.list || []).forEach((x) => {
      if (x.cloud_vendor) merged.add(String(x.cloud_vendor))
    })
    ;(serviceData.list || []).forEach((x) => {
      if (x.cloud_vendor) merged.add(String(x.cloud_vendor))
    })
    cloudVendorOptions.value = Array.from(merged).map((x) => ({ label: x, value: x }))
    syncCloudVendorField()
  } catch (e) {
    syncCloudVendorField()
  }
}

onMounted(async () => {
  await Promise.all([loadLocations(), loadCloudVendors()])
})
</script>

<template>
  <CrudSimple
    title="主机管理"
    endpoint="/hosts"
    :delta-update="true"
    :id-field="'id'"
    :columns="columns"
    :form-model="formModel"
    :editable-fields="['name', 'hostname', 'primary_address', 'provider_kind', 'cloud_vendor', 'cloud_instance_id', 'os_type', 'location_id', 'status', 'description']"
    :form-fields="formFields"
    :filters="filters"
    :refresh-token="crudRefreshToken"
  >
    <template #row-actions="{ row }">
      <el-button link type="primary" @click="openCredentialDialog(row)">凭据</el-button>
    </template>
    <template #edit-extra="{ editForm, editingId, registerSaveHook, saveLoading }">
      <EnvironmentBindingsInline
        target-type="host"
        :target-id="editingId"
        :target-name="editForm.name"
        :register-save-hook="registerSaveHook"
        :save-loading="saveLoading"
        @updated="onUpdated"
      />
    </template>
  </CrudSimple>

  <CredentialsDialog
    v-model="credentialVisible"
    resource-type="host"
    :resource-id="currentHostId"
    :resource-name="currentHostName"
    @updated="onUpdated"
  />
</template>
