<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api, { unwrap } from '../services/api'
import { collectPageRouteCandidates } from '../router/routes'

const router = useRouter()
const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const status = ref('')
const syncing = ref(false)
const syncDialogVisible = ref(false)
const editDialogVisible = ref(false)
const manualDialogVisible = ref(false)

const editForm = reactive({
  id: 0,
  name: '',
  route_path: '',
  source: 'manual',
  status: 'active',
  description: ''
})
const manualForm = reactive({
  name: '',
  route_path: '',
  status: 'active',
  description: ''
})
const syncPreview = reactive({
  summary: null,
  new_routes: [],
  existing_routes: [],
  changed_routes: [],
  retired_routes: []
})
const syncCandidates = ref([])

const statusOptions = [
  { label: '全部状态', value: '' },
  { label: 'active', value: 'active' },
  { label: 'inactive', value: 'inactive' }
]
const sourceLabelMap = {
  menu: 'menu',
  router: 'router',
  manual: '补录'
}

const sourceTagType = (source) => {
  if (source === 'manual') return 'warning'
  if (source === 'menu') return 'success'
  return 'info'
}

const load = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value,
      status: status.value
    }
    const data = unwrap(await api.get('/pages', { params }))
    list.value = data.list || []
    total.value = data.pagination?.total || 0
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const buildSyncCandidates = () => collectPageRouteCandidates()

const openSyncPreview = async () => {
  syncing.value = true
  try {
    const routes = buildSyncCandidates()
    syncCandidates.value = routes
    const data = unwrap(await api.post('/pages/sync', { dry_run: true, routes }))
    syncPreview.summary = data.summary || null
    syncPreview.new_routes = data.new_routes || []
    syncPreview.existing_routes = data.existing_routes || []
    syncPreview.changed_routes = data.changed_routes || []
    syncPreview.retired_routes = data.retired_routes || []
    syncDialogVisible.value = true
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '扫描失败')
  } finally {
    syncing.value = false
  }
}

const applySync = async () => {
  syncing.value = true
  try {
    await api.post('/pages/sync', { dry_run: false, routes: syncCandidates.value })
    ElMessage.success('同步完成')
    syncDialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '同步失败')
  } finally {
    syncing.value = false
  }
}

const openEdit = (row) => {
  editForm.id = row.id
  editForm.name = row.name
  editForm.route_path = row.route_path
  editForm.source = row.source || 'manual'
  editForm.status = row.status
  editForm.description = row.description || ''
  editDialogVisible.value = true
}

const saveEdit = async () => {
  try {
    await api.put(`/pages/${editForm.id}`, {
      name: editForm.name,
      route_path: editForm.route_path,
      source: editForm.source,
      status: editForm.status,
      description: editForm.description
    })
    ElMessage.success('更新成功')
    editDialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '更新失败')
  }
}

const openManual = () => {
  manualForm.name = ''
  manualForm.route_path = ''
  manualForm.status = 'active'
  manualForm.description = ''
  manualDialogVisible.value = true
}

const createManual = async () => {
  try {
    await api.post('/pages', {
      name: manualForm.name,
      route_path: manualForm.route_path,
      source: 'manual',
      status: manualForm.status,
      description: manualForm.description
    })
    ElMessage.success('补录成功')
    manualDialogVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '补录失败')
  }
}

const goGrant = (row) => {
  router.push({
    path: '/admin/grants',
    query: {
      object_type: 'page',
      object_id: String(row.id),
      keyword: row.route_path
    }
  })
}

const removeOne = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该页面资源？', '提示', { type: 'warning' })
    await api.delete(`/pages/${row.id}`)
    ElMessage.success('删除成功')
    await load()
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') {
      ElMessage.error(e.response?.data?.message || '删除失败')
    }
  }
}

const previewStat = computed(() => syncPreview.summary || {})
const onPageChange = (p) => {
  page.value = p
  load()
}

onMounted(load)
</script>

<template>
  <el-card>
    <template #header>
      <div class="row-between">
        <span>页面资源管理</span>
      </div>
    </template>
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="关键字(route/name)" style="width: 240px" />
      <el-select v-model="status" style="width: 160px">
        <el-option v-for="op in statusOptions" :key="op.value" :label="op.label" :value="op.value" />
      </el-select>
      <el-button @click="load">查询</el-button>
      <el-button type="primary" :loading="syncing" @click="openSyncPreview">同步路由</el-button>
      <el-button @click="openManual">补录异常路由</el-button>
    </div>
    <el-alert type="info" :closable="false" style="margin-bottom: 12px">
      <template #title>
        page 主来源为前端 router/menu；此页面用于同步注册与状态管理，授权入口在“权限管理”。
      </template>
    </el-alert>
    <el-table :data="list" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="名称" min-width="160" />
      <el-table-column prop="route_path" label="route_path" min-width="220" />
      <el-table-column label="来源" width="120">
        <template #default="{ row }">
          <el-tag :type="sourceTagType(row.source)">{{ sourceLabelMap[row.source] || row.source }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column prop="grant_count" label="授权数" width="100" />
      <el-table-column prop="description" label="描述" min-width="180" />
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="primary" @click="goGrant(row)">去授权</el-button>
          <el-button link type="danger" @click="removeOne(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div style="margin-top: 12px; display: flex; justify-content: flex-end">
      <el-pagination
        background
        layout="prev, pager, next"
        :current-page="page"
        :page-size="pageSize"
        :total="total"
        @current-change="onPageChange"
      />
    </div>
  </el-card>

  <el-dialog v-model="syncDialogVisible" title="路由同步预览" width="920px">
    <el-space wrap>
      <el-tag>候选 {{ previewStat.input_total || 0 }}</el-tag>
      <el-tag type="success">新增 {{ previewStat.created || 0 }}</el-tag>
      <el-tag type="warning">变更 {{ previewStat.updated || 0 }}</el-tag>
      <el-tag>已存在 {{ previewStat.unchanged || 0 }}</el-tag>
      <el-tag type="danger">下线 {{ previewStat.retired || 0 }}</el-tag>
    </el-space>
    <el-divider />
    <el-collapse>
      <el-collapse-item :title="`新发现路由 (${syncPreview.new_routes.length})`" name="new">
        <el-table :data="syncPreview.new_routes" size="small">
          <el-table-column prop="name" label="名称" />
          <el-table-column prop="route_path" label="route_path" />
          <el-table-column prop="source" label="来源" width="100" />
        </el-table>
      </el-collapse-item>
      <el-collapse-item :title="`信息变化 (${syncPreview.changed_routes.length})`" name="changed">
        <el-table :data="syncPreview.changed_routes" size="small">
          <el-table-column prop="route_path" label="route_path" />
          <el-table-column prop="before.name" label="旧名称" />
          <el-table-column prop="after.name" label="新名称" />
          <el-table-column prop="before.source" label="旧来源" />
          <el-table-column prop="after.source" label="新来源" />
          <el-table-column prop="before.status" label="旧状态" />
          <el-table-column prop="after.status" label="新状态" />
        </el-table>
      </el-collapse-item>
      <el-collapse-item :title="`已不存在但仍保留 (${syncPreview.retired_routes.length})`" name="retired">
        <el-table :data="syncPreview.retired_routes" size="small">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="名称" />
          <el-table-column prop="route_path" label="route_path" />
          <el-table-column prop="status" label="状态" width="100" />
        </el-table>
      </el-collapse-item>
      <el-collapse-item :title="`已存在且无变化 (${syncPreview.existing_routes.length})`" name="existing">
        <el-table :data="syncPreview.existing_routes" size="small">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="名称" />
          <el-table-column prop="route_path" label="route_path" />
          <el-table-column prop="source" label="来源" width="100" />
        </el-table>
      </el-collapse-item>
    </el-collapse>
    <template #footer>
      <el-button @click="syncDialogVisible = false">取消</el-button>
      <el-button type="primary" :loading="syncing" @click="applySync">确认同步</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="editDialogVisible" title="编辑页面资源" width="520px">
    <el-form label-width="100px">
      <el-form-item label="名称">
        <el-input v-model="editForm.name" />
      </el-form-item>
      <el-form-item label="route_path">
        <el-input v-model="editForm.route_path" disabled />
      </el-form-item>
      <el-form-item label="来源">
        <el-input v-model="editForm.source" disabled />
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="editForm.status">
          <el-option label="active" value="active" />
          <el-option label="inactive" value="inactive" />
        </el-select>
      </el-form-item>
      <el-form-item label="描述">
        <el-input v-model="editForm.description" type="textarea" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="editDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="saveEdit">保存</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="manualDialogVisible" title="补录异常路由" width="520px">
    <el-form label-width="100px">
      <el-form-item label="名称">
        <el-input v-model="manualForm.name" />
      </el-form-item>
      <el-form-item label="route_path">
        <el-input v-model="manualForm.route_path" placeholder="/custom/path" />
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="manualForm.status">
          <el-option label="active" value="active" />
          <el-option label="inactive" value="inactive" />
        </el-select>
      </el-form-item>
      <el-form-item label="描述">
        <el-input v-model="manualForm.description" type="textarea" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="manualDialogVisible = false">取消</el-button>
      <el-button type="primary" @click="createManual">确认补录</el-button>
    </template>
  </el-dialog>
</template>
