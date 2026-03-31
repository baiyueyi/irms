<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api, { unwrap } from '../services/api'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  resourceType: { type: String, required: true }, // host | service
  resourceId: { type: [String, Number], default: '' },
  resourceName: { type: String, default: '' }
})

const emit = defineEmits(['update:modelValue'])

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

const loading = ref(false)
const revealingId = ref(null)
const list = ref([])
const revealText = ref('')
const revealDialogVisible = ref(false)

const createVisible = ref(false)
const editVisible = ref(false)
const editingId = ref(null)

const baseForm = () => ({
  account_name: '',
  credential_name: '',
  credential_kind: 'password',
  username: '',
  secret: '',
  certificate_pem: '',
  private_key_pem: '',
  passphrase: '',
  status: 'active',
  description: ''
})

const createForm = reactive(baseForm())
const editForm = reactive(baseForm())
const kindOptions = [
  { label: 'password', value: 'password' },
  { label: 'certificate', value: 'certificate' }
]
const statusOptions = [
  { label: 'active', value: 'active' },
  { label: 'inactive', value: 'inactive' }
]

const endpoint = computed(() => (props.resourceType === 'host' ? '/host-credentials' : '/service-credentials'))
const queryKey = computed(() => (props.resourceType === 'host' ? 'host_id' : 'service_id'))
const title = computed(() => `${props.resourceType === 'host' ? '主机' : '服务'}凭据 - ${props.resourceName || props.resourceId}`)

const resetForm = (target) => {
  Object.assign(target, baseForm())
}

const load = async () => {
  if (!props.resourceId) {
    list.value = []
    return
  }
  loading.value = true
  try {
    const data = unwrap(await api.get(endpoint.value, { params: { [queryKey.value]: props.resourceId } }))
    list.value = data?.list || []
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '加载凭据失败')
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  resetForm(createForm)
  createVisible.value = true
}

const createOne = async () => {
  const targetId = Number(props.resourceId)
  if (!targetId) {
    ElMessage.error('资源ID无效，无法创建凭据')
    return
  }
  try {
    await api.post(endpoint.value, { ...createForm, [queryKey.value]: targetId })
    ElMessage.success('创建凭据成功')
    createVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '创建凭据失败')
  }
}

const openEdit = (row) => {
  editingId.value = row.id
  resetForm(editForm)
  Object.keys(editForm).forEach((k) => {
    if (row[k] !== undefined && row[k] !== null) {
      editForm[k] = row[k]
    }
  })
  editVisible.value = true
}

const updateOne = async () => {
  try {
    await api.put(`${endpoint.value}/${editingId.value}`, editForm)
    ElMessage.success('更新凭据成功')
    editVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '更新凭据失败')
  }
}

const removeOne = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该凭据？', '提示', { type: 'warning' })
    await api.delete(`${endpoint.value}/${row.id}`)
    ElMessage.success('删除凭据成功')
    await load()
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') {
      ElMessage.error(e.response?.data?.message || '删除凭据失败')
    }
  }
}

const revealOne = async (row) => {
  revealingId.value = row.id
  try {
    const data = unwrap(await api.post(`${endpoint.value}/${row.id}/reveal`))
    revealText.value = JSON.stringify(data || {}, null, 2)
    revealDialogVisible.value = true
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '查看明文失败')
  } finally {
    revealingId.value = null
  }
}

watch(
  () => [visible.value, props.resourceId],
  ([opened]) => {
    if (opened) {
      load()
    }
  }
)
</script>

<template>
  <el-dialog v-model="visible" :title="title" width="860px">
    <div style="margin-bottom: 12px">
      <el-button type="primary" @click="openCreate">新增凭据</el-button>
    </div>
    <el-table :data="list" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="account_name" label="账号名称" min-width="120" />
      <el-table-column prop="credential_name" label="凭据名称" min-width="120" />
      <el-table-column prop="credential_kind" label="凭据类型" width="110" />
      <el-table-column prop="username" label="用户名" min-width="120" />
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column prop="description" label="描述" min-width="140" />
      <el-table-column label="操作" width="210">
        <template #default="{ row }">
          <el-button link type="primary" :loading="revealingId === row.id" @click="revealOne(row)">查看明文</el-button>
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="removeOne(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-dialog>

  <el-dialog v-model="createVisible" title="新增凭据" width="680px">
    <el-form label-position="top">
      <el-form-item label="账号名称"><el-input v-model="createForm.account_name" /></el-form-item>
      <el-form-item label="凭据名称"><el-input v-model="createForm.credential_name" /></el-form-item>
      <el-form-item label="凭据类型">
        <el-select v-model="createForm.credential_kind" style="width: 100%">
          <el-option v-for="op in kindOptions" :key="op.value" :label="op.label" :value="op.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="用户名"><el-input v-model="createForm.username" /></el-form-item>
      <el-form-item v-if="createForm.credential_kind === 'password'" label="密码">
        <el-input v-model="createForm.secret" type="textarea" />
      </el-form-item>
      <el-form-item v-else label="证书">
        <el-input v-model="createForm.certificate_pem" type="textarea" :rows="4" />
      </el-form-item>
      <el-form-item v-if="createForm.credential_kind === 'certificate'" label="私钥">
        <el-input v-model="createForm.private_key_pem" type="textarea" :rows="4" />
      </el-form-item>
      <el-form-item v-if="createForm.credential_kind === 'certificate'" label="口令">
        <el-input v-model="createForm.passphrase" />
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="createForm.status" style="width: 100%">
          <el-option v-for="op in statusOptions" :key="op.value" :label="op.label" :value="op.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="描述"><el-input v-model="createForm.description" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="createVisible = false">取消</el-button>
      <el-button type="primary" @click="createOne">确认</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="editVisible" title="编辑凭据" width="680px">
    <el-form label-position="top">
      <el-form-item label="账号名称"><el-input v-model="editForm.account_name" /></el-form-item>
      <el-form-item label="凭据名称"><el-input v-model="editForm.credential_name" /></el-form-item>
      <el-form-item label="凭据类型">
        <el-select v-model="editForm.credential_kind" style="width: 100%">
          <el-option v-for="op in kindOptions" :key="op.value" :label="op.label" :value="op.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="用户名"><el-input v-model="editForm.username" /></el-form-item>
      <el-form-item v-if="editForm.credential_kind === 'password'" label="密码（留空表示不改）">
        <el-input v-model="editForm.secret" type="textarea" />
      </el-form-item>
      <el-form-item v-else label="证书（留空表示不改）">
        <el-input v-model="editForm.certificate_pem" type="textarea" :rows="4" />
      </el-form-item>
      <el-form-item v-if="editForm.credential_kind === 'certificate'" label="私钥（留空表示不改）">
        <el-input v-model="editForm.private_key_pem" type="textarea" :rows="4" />
      </el-form-item>
      <el-form-item v-if="editForm.credential_kind === 'certificate'" label="口令（留空表示不改）">
        <el-input v-model="editForm.passphrase" />
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="editForm.status" style="width: 100%">
          <el-option v-for="op in statusOptions" :key="op.value" :label="op.label" :value="op.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="描述"><el-input v-model="editForm.description" /></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="editVisible = false">取消</el-button>
      <el-button type="primary" @click="updateOne">保存</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="revealDialogVisible" title="凭据明文" width="680px">
    <el-input v-model="revealText" type="textarea" :rows="14" readonly />
    <template #footer>
      <el-button type="primary" @click="revealDialogVisible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>
