<script setup>
import { computed, reactive, ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api, { unwrap } from '../services/api'

const route = useRoute()
const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const dialogVisible = ref(false)
const dialogMode = ref('create')
const activeTab = ref('subject')
const editingId = ref(null)

const filters = reactive({
  subject_type: '',
  object_type: '',
  permission: '',
  object_id: '',
  subject_id: ''
})

const subjectTypeOptions = [
  { label: 'user', value: 'user' },
  { label: 'group', value: 'group' }
]
const objectTypeOptions = [
  { label: 'page', value: 'page' },
  { label: 'host', value: 'host' },
  { label: 'service', value: 'service' },
  { label: 'host_group', value: 'host_group' },
  { label: 'service_group', value: 'service_group' }
]
const permissionOptions = [
  { label: 'ReadOnly', value: 'ReadOnly' },
  { label: 'ReadWrite', value: 'ReadWrite' }
]

const form = reactive({
  subject_type: 'user',
  subject_option: null,
  subject_label: '',
  object_type: 'page',
  object_option: null,
  object_label: '',
  permission: 'ReadOnly'
})
const subjectOptions = ref([])
const objectOptions = ref([])
const subjectLoading = ref(false)
const objectLoading = ref(false)

const load = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value,
      subject_type: filters.subject_type,
      object_type: filters.object_type,
      permission: filters.permission,
      object_id: filters.object_id,
      subject_id: filters.subject_id
    }
    const data = unwrap(await api.get('/grants', { params }))
    list.value = data.list || []
    total.value = data.pagination?.total || 0
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '查询失败')
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.subject_type = 'user'
  form.subject_option = null
  form.subject_label = ''
  form.object_type = 'page'
  form.object_option = null
  form.object_label = ''
  form.permission = 'ReadOnly'
  subjectOptions.value = []
  objectOptions.value = []
}

const openCreate = () => {
  resetForm()
  dialogMode.value = 'create'
  editingId.value = null
  activeTab.value = 'subject'
  dialogVisible.value = true
}

const openEdit = async (row) => {
  resetForm()
  dialogMode.value = 'edit'
  editingId.value = row.id
  form.subject_type = row.subject_type_display === 'group' ? 'group' : 'user'
  form.subject_option = row.subject_id
  form.subject_label = row.subject_name
  form.object_type = row.object_type_display
  form.object_option = row.object_id
  form.object_label = row.object_name
  form.permission = row.permission
  activeTab.value = 'permission'
  dialogVisible.value = true
}

const fetchSubjectOptions = async (query = '') => {
  subjectLoading.value = true
  try {
    if (form.subject_type === 'user') {
      const data = unwrap(await api.get('/users', { params: { page: 1, page_size: 20, keyword: query } }))
      subjectOptions.value = (data.list || []).map((x) => ({ label: x.username, value: x.id }))
      return
    }
    const data = unwrap(await api.get('/user-groups', { params: { page: 1, page_size: 20, keyword: query } }))
    subjectOptions.value = (data.list || []).map((x) => ({ label: x.name, value: x.id }))
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '主体查询失败')
  } finally {
    subjectLoading.value = false
  }
}

const fetchObjectOptions = async (query = '') => {
  objectLoading.value = true
  try {
    if (form.object_type === 'page') {
      const data = unwrap(await api.get('/pages', { params: { page: 1, page_size: 20, keyword: query } }))
      objectOptions.value = (data.list || []).map((x) => ({ label: x.name, value: x.id }))
      return
    }
    if (form.object_type === 'host') {
      const data = unwrap(await api.get('/hosts', { params: { page: 1, page_size: 20, keyword: query } }))
      objectOptions.value = (data.list || []).map((x) => ({ label: x.name, value: x.id }))
      return
    }
    if (form.object_type === 'service') {
      const data = unwrap(await api.get('/services', { params: { page: 1, page_size: 20, keyword: query } }))
      objectOptions.value = (data.list || []).map((x) => ({ label: x.name, value: x.id }))
      return
    }
    const groupType = form.object_type === 'host_group' ? 'host' : 'service'
    const data = unwrap(await api.get('/resource-groups', { params: { page: 1, page_size: 20, keyword: query, type: groupType } }))
    objectOptions.value = (data.list || []).map((x) => ({ label: x.name, value: x.id }))
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '客体查询失败')
  } finally {
    objectLoading.value = false
  }
}

const createGrant = async () => {
  if (!form.subject_option || !form.object_option) {
    ElMessage.error('请先选择主体和客体')
    return
  }
  try {
    await api.post('/grants', {
      subject_type: form.subject_type,
      subject_id: form.subject_option,
      object_type: form.object_type,
      object_id: form.object_option,
      permission: form.permission
    })
    ElMessage.success('创建成功')
    dialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '创建失败')
  }
}

const updateGrant = async () => {
  try {
    await api.put(`/grants/${editingId.value}`, { permission: form.permission })
    ElMessage.success('更新成功')
    dialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '更新失败')
  }
}

const removeOne = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该记录？', '提示', { type: 'warning' })
    await api.delete(`/grants/${row.id}`)
    ElMessage.success('删除成功')
    await load()
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') {
      ElMessage.error(e.response?.data?.message || '删除失败')
    }
  }
}

watch(
  () => form.subject_type,
  async () => {
    form.subject_option = null
    subjectOptions.value = []
    await fetchSubjectOptions('')
  }
)

watch(
  () => form.object_type,
  async () => {
    form.object_option = null
    objectOptions.value = []
    await fetchObjectOptions('')
  }
)

const onSubjectSelect = (v) => {
  const hit = subjectOptions.value.find((x) => x.value === v)
  form.subject_label = hit ? hit.label : ''
}

const onObjectSelect = (v) => {
  const hit = objectOptions.value.find((x) => x.value === v)
  form.object_label = hit ? hit.label : ''
}

const previewText = computed(() => {
  const subjectTypeLabel = form.subject_type === 'group' ? '用户组' : '用户'
  const objectTypeMap = {
    page: '页面资源',
    host: '主机',
    service: '服务',
    host_group: '主机组',
    service_group: '服务组'
  }
  const objectTypeLabel = objectTypeMap[form.object_type] || form.object_type
  const subjectName = form.subject_label || '-'
  const objectName = form.object_label || '-'
  return `将[${subjectTypeLabel} ${subjectName}]授予[${objectTypeLabel} ${objectName}]的[${form.permission}]权限`
})

onMounted(() => {
  if (typeof route.query.keyword === 'string') keyword.value = route.query.keyword
  if (typeof route.query.object_type === 'string') filters.object_type = route.query.object_type
  if (typeof route.query.permission === 'string') filters.permission = route.query.permission
  if (typeof route.query.object_id === 'string') filters.object_id = route.query.object_id
  if (typeof route.query.subject_id === 'string') filters.subject_id = route.query.subject_id
  load()
})
</script>

<template>
  <el-card>
    <template #header>
      <div class="row-between">
        <span>权限管理</span>
      </div>
    </template>
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="关键字" style="width: 240px" />
      <el-select v-model="filters.subject_type" clearable placeholder="主体类型" style="width: 160px">
        <el-option v-for="op in subjectTypeOptions" :key="op.value" :label="op.label" :value="op.value" />
      </el-select>
      <el-select v-model="filters.object_type" clearable placeholder="客体类型" style="width: 180px">
        <el-option v-for="op in objectTypeOptions" :key="op.value" :label="op.label" :value="op.value" />
      </el-select>
      <el-select v-model="filters.permission" clearable placeholder="权限" style="width: 160px">
        <el-option v-for="op in permissionOptions" :key="op.value" :label="op.label" :value="op.value" />
      </el-select>
      <el-button @click="load">查询</el-button>
      <el-tag v-if="filters.object_id" type="info">object_id={{ filters.object_id }}</el-tag>
      <el-button type="primary" @click="openCreate">新增</el-button>
    </div>
    <el-table :data="list" v-loading="loading">
      <el-table-column prop="subject_name" label="主体名称" />
      <el-table-column prop="subject_type_display" label="主体类型" />
      <el-table-column prop="object_name" label="客体名称" />
      <el-table-column prop="object_type_display" label="客体类型" />
      <el-table-column prop="permission" label="权限" />
      <el-table-column prop="updated_at" label="更新时间" width="180" />
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="removeOne(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pager-wrap">
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" layout="prev, pager, next, total" :total="total" @current-change="load" />
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogMode === 'create' ? '新增授权' : '编辑授权'" width="760px">
      <el-tabs v-model="activeTab">
        <el-tab-pane name="subject" label="主体">
          <el-form label-position="top">
            <el-form-item label="主体类型">
              <el-select v-model="form.subject_type" style="width: 100%" :disabled="dialogMode === 'edit'">
                <el-option v-for="op in subjectTypeOptions" :key="op.value" :label="op.label" :value="op.value" />
              </el-select>
            </el-form-item>
            <el-form-item label="主体名称">
              <el-select
                v-model="form.subject_option"
                filterable
                remote
                reserve-keyword
                :remote-method="fetchSubjectOptions"
                :loading="subjectLoading"
                style="width: 100%"
                :disabled="dialogMode === 'edit'"
                @change="onSubjectSelect"
              >
                <el-option v-for="op in subjectOptions" :key="op.value" :label="op.label" :value="op.value" />
              </el-select>
            </el-form-item>
            <el-tag v-if="form.subject_label">{{ form.subject_label }}</el-tag>
          </el-form>
        </el-tab-pane>
        <el-tab-pane name="object" label="客体">
          <el-form label-position="top">
            <el-form-item label="客体类型">
              <el-select v-model="form.object_type" style="width: 100%" :disabled="dialogMode === 'edit'">
                <el-option v-for="op in objectTypeOptions" :key="op.value" :label="op.label" :value="op.value" />
              </el-select>
            </el-form-item>
            <el-form-item label="客体名称">
              <el-select
                v-model="form.object_option"
                filterable
                remote
                reserve-keyword
                :remote-method="fetchObjectOptions"
                :loading="objectLoading"
                style="width: 100%"
                :disabled="dialogMode === 'edit'"
                @change="onObjectSelect"
              >
                <el-option v-for="op in objectOptions" :key="op.value" :label="op.label" :value="op.value" />
              </el-select>
            </el-form-item>
            <el-tag v-if="form.object_label">{{ form.object_label }}</el-tag>
          </el-form>
        </el-tab-pane>
        <el-tab-pane name="permission" label="权限">
          <el-form label-position="top">
            <el-form-item label="权限">
              <el-select v-model="form.permission" style="width: 100%">
                <el-option v-for="op in permissionOptions" :key="op.value" :label="op.label" :value="op.value" />
              </el-select>
            </el-form-item>
            <el-alert v-if="dialogMode === 'edit'" title="编辑模式仅允许修改权限，主体与客体不可变更" type="info" :closable="false" />
          </el-form>
        </el-tab-pane>
        <el-tab-pane name="confirm" label="确认">
          <el-alert :title="previewText" type="success" :closable="false" />
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button v-if="dialogMode === 'create'" type="primary" @click="createGrant">确认新增</el-button>
        <el-button v-else type="primary" @click="updateGrant">确认保存</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>
