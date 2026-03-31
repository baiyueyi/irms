<script setup>
import { onMounted, ref } from 'vue'
import CrudSimple from '../components/CrudSimple.vue'
import CredentialsDialog from '../components/CredentialsDialog.vue'
import EnvironmentBindingsInline from '../components/EnvironmentBindingsInline.vue'
import api, { unwrap } from '../services/api'

const statusOptions = [
  { label: 'active', value: 'active' },
  { label: 'inactive', value: 'inactive' }
]
const serviceKindOptions = [
  { label: 'app', value: 'app' },
  { label: 'api', value: 'api' },
  { label: 'database', value: 'database' },
  { label: 'middleware', value: 'middleware' },
  { label: 'cloud_product', value: 'cloud_product' },
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
  { prop: 'service_kind', label: '服务类型', creatable: true },
  { prop: 'host', label: '主机' },
  { prop: 'endpoint_or_identifier', label: '端点/标识', creatable: true },
  { prop: 'environments', label: '环境标签', renderType: 'envTags', maxTags: 2 },
  { prop: 'environment_source_label', label: '环境来源', renderType: 'envSource' },
  { prop: 'protocol', label: '协议', creatable: true },
  { prop: 'status', label: '状态', creatable: true },
  { prop: 'description', label: '描述', creatable: true }
]

const formModel = {
  name: '',
  service_kind: 'app',
  host_id: '',
  endpoint_or_identifier: '',
  port: '',
  protocol: '',
  cloud_vendor: '',
  cloud_product_code: '',
  status: 'active',
  description: ''
}
const formFields = [
  { key: 'name', label: '名称' },
  { key: 'service_kind', label: '服务类型', inputType: 'select', options: serviceKindOptions },
  { key: 'host_id', label: '主机ID（可空）' },
  { key: 'endpoint_or_identifier', label: '端点/标识' },
  { key: 'port', label: '端口' },
  { key: 'protocol', label: '协议' },
  { key: 'cloud_vendor', label: '云厂商', inputType: 'select', filterable: true, options: cloudVendorOptions.value },
  { key: 'cloud_product_code', label: '云产品编码' },
  { key: 'status', label: '状态', inputType: 'select', options: statusOptions },
  { key: 'description', label: '描述' }
]
const filters = [
  { key: 'service_kind', label: '服务类型', options: serviceKindOptions },
  { key: 'status', label: '状态', options: statusOptions }
]

const credentialVisible = ref(false)
const currentServiceId = ref('')
const currentServiceName = ref('')
const crudRefreshToken = ref(0)

const openCredentialDialog = (row) => {
  currentServiceId.value = row.id
  currentServiceName.value = row.name
  credentialVisible.value = true
}

const onUpdated = () => {
  crudRefreshToken.value += 1
}

const syncCloudVendorField = () => {
  const field = formFields.find((x) => x.key === 'cloud_vendor')
  if (field) field.options = cloudVendorOptions.value
}

const loadCloudVendors = async () => {
  try {
    const data = unwrap(await api.get('/services', { params: { page: 1, page_size: 200 } }))
    const merged = new Set(cloudVendorOptions.value.map((x) => x.value))
    ;(data.list || []).forEach((x) => {
      if (x.cloud_vendor) merged.add(String(x.cloud_vendor))
    })
    cloudVendorOptions.value = Array.from(merged).map((x) => ({ label: x, value: x }))
    syncCloudVendorField()
  } catch (e) {
    syncCloudVendorField()
  }
}

onMounted(loadCloudVendors)
</script>

<template>
  <CrudSimple
    title="服务管理"
    endpoint="/services"
    :id-field="'id'"
    :columns="columns"
    :form-model="formModel"
    :editable-fields="['name', 'service_kind', 'host_id', 'endpoint_or_identifier', 'port', 'protocol', 'cloud_vendor', 'cloud_product_code', 'status', 'description']"
    :form-fields="formFields"
    :filters="filters"
    :refresh-token="crudRefreshToken"
  >
    <template #row-actions="{ row }">
      <el-button link type="primary" @click="openCredentialDialog(row)">凭据</el-button>
    </template>
    <template #edit-extra="{ editForm, editingId, registerSaveHook, saveLoading }">
      <EnvironmentBindingsInline
        target-type="service"
        :target-id="editingId"
        :target-name="editForm.name"
        :host-id="editForm.host_id"
        :register-save-hook="registerSaveHook" 
        :save-loading="saveLoading"
        @updated="onUpdated"
      />
    </template>
  </CrudSimple>

  <CredentialsDialog
    v-model="credentialVisible"
    resource-type="service"
    :resource-id="currentServiceId"
    :resource-name="currentServiceName"
    @updated="onUpdated"
  />
</template>
