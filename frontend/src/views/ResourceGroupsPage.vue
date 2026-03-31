<script setup>
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api, { unwrap } from '../services/api'

const typeOptions = [
  { label: 'host_group', value: 'host' },
  { label: 'service_group', value: 'service' }
]

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const typeFilter = ref('')

const createVisible = ref(false)
const editVisible = ref(false)
const editingId = ref(null)

const memberOptions = ref([])
const memberDialogVisible = ref(false)
const memberDialogTitle = ref('')
const memberNames = ref([])

const createForm = reactive({ name: '', type: 'host', description: '', member_keys: [] })
const editForm = reactive({ name: '', type: 'host', description: '', member_keys: [] })

const load = async () => {
  loading.value = true
  try {
    const data = unwrap(await api.get('/resource-groups', { params: { page: page.value, page_size: pageSize.value, keyword: keyword.value, type: typeFilter.value } }))
    list.value = data.list || []
    total.value = data.pagination?.total || 0
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const loadResourceOptions = async (type, query = '') => {
  const data = unwrap(await api.get('/resources', { params: { page: 1, page_size: 300, keyword: query, type } }))
  memberOptions.value = (data.list || []).map((x) => ({ key: x.key, label: x.name }))
}

const getMembers = async (groupId) => {
  const data = unwrap(await api.get('/resource-group-members', { params: { resource_group_id: groupId, page: 1, page_size: 500 } }))
  return data.list || []
}

const syncMembers = async (groupId, targetKeys) => {
  const current = await getMembers(groupId)
  const currentKeys = new Set(current.map((x) => x.resource_key))
  const target = new Set(targetKeys)
  for (const rk of target) {
    if (!currentKeys.has(rk)) {
      await api.post('/resource-group-members', { resource_key: rk, resource_group_id: groupId })
    }
  }
  for (const row of current) {
    if (!target.has(row.resource_key)) {
      await api.delete('/resource-group-members', { data: { resource_key: row.resource_key, resource_group_id: groupId } })
    }
  }
}

const openCreate = async () => {
  createForm.name = ''
  createForm.type = 'host'
  createForm.description = ''
  createForm.member_keys = []
  createVisible.value = true
  try {
    await loadResourceOptions('host', '')
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '加载资源失败')
  }
}

const createOne = async () => {
  try {
    const resp = unwrap(await api.post('/resource-groups', { name: createForm.name, type: createForm.type, description: createForm.description }))
    await syncMembers(resp.id, createForm.member_keys)
    ElMessage.success('创建成功')
    createVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '创建失败')
  }
}

const openEdit = async (row) => {
  editingId.value = row.id
  editForm.name = row.name || ''
  editForm.type = row.type || 'host'
  editForm.description = row.description || ''
  editVisible.value = true
  try {
    await loadResourceOptions(editForm.type, '')
    const members = await getMembers(row.id)
    editForm.member_keys = members.map((x) => x.resource_key)
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '加载成员失败')
  }
}

const saveEdit = async () => {
  try {
    await api.put(`/resource-groups/${editingId.value}`, { name: editForm.name, type: editForm.type, description: editForm.description })
    await syncMembers(editingId.value, editForm.member_keys)
    ElMessage.success('保存成功')
    editVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '保存失败')
  }
}

const removeOne = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该资源组？', '提示', { type: 'warning' })
    const members = await getMembers(row.id)
    for (const m of members) {
      await api.delete('/resource-group-members', { data: { resource_key: m.resource_key, resource_group_id: row.id } })
    }
    await api.delete(`/resource-groups/${row.id}`)
    ElMessage.success('删除成功')
    await load()
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') {
      ElMessage.error(e.response?.data?.message || '删除失败')
    }
  }
}

const showMembers = async (row) => {
  try {
    const members = await getMembers(row.id)
    memberNames.value = members.map((x) => x.resource_name || `#${x.resource_key}`)
    memberDialogTitle.value = `${row.name} - 成员`
    memberDialogVisible.value = true
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '加载成员失败')
  }
}

const onCreateTypeChange = async (v) => {
  createForm.member_keys = []
  await loadResourceOptions(v, '')
}

const onEditTypeChange = async (v) => {
  editForm.member_keys = []
  await loadResourceOptions(v, '')
}

onMounted(load)
</script>

<template>
  <el-card>
    <template #header>
      <div class="row-between">
        <span>资源组管理</span>
      </div>
    </template>
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="关键字" style="width: 240px" />
      <el-select v-model="typeFilter" clearable placeholder="类型" style="width: 180px">
        <el-option v-for="op in typeOptions" :key="op.value" :label="op.label" :value="op.value" />
      </el-select>
      <el-button @click="load">查询</el-button>
      <el-button type="primary" @click="openCreate">新增</el-button>
    </div>
    <el-table :data="list" v-loading="loading">
      <el-table-column prop="id" label="ID" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="type" label="类型" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="member_count" label="成员数" width="90" />
      <el-table-column label="操作" width="240">
        <template #default="{ row }">
          <el-button link type="primary" @click="showMembers(row)">成员</el-button>
          <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="removeOne(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pager-wrap">
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" layout="prev, pager, next, total" :total="total" @current-change="load" />
    </div>
  </el-card>

  <el-dialog v-model="createVisible" title="新增资源组" width="760px">
    <el-form label-position="top">
      <el-form-item label="名称"><el-input v-model="createForm.name" /></el-form-item>
      <el-form-item label="类型">
        <el-select v-model="createForm.type" style="width: 100%" @change="onCreateTypeChange">
          <el-option v-for="op in typeOptions" :key="op.value" :label="op.label" :value="op.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="描述"><el-input v-model="createForm.description" /></el-form-item>
      <el-form-item label="成员选择">
        <el-transfer v-model="createForm.member_keys" filterable :data="memberOptions" :titles="['可选资源', '已选成员']" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="createVisible = false">取消</el-button>
      <el-button type="primary" @click="createOne">确认</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="editVisible" title="编辑资源组" width="760px">
    <el-form label-position="top">
      <el-form-item label="名称"><el-input v-model="editForm.name" /></el-form-item>
      <el-form-item label="类型">
        <el-select v-model="editForm.type" style="width: 100%" @change="onEditTypeChange">
          <el-option v-for="op in typeOptions" :key="op.value" :label="op.label" :value="op.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="描述"><el-input v-model="editForm.description" /></el-form-item>
      <el-form-item label="成员选择">
        <el-transfer v-model="editForm.member_keys" filterable :data="memberOptions" :titles="['可选资源', '已选成员']" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="editVisible = false">取消</el-button>
      <el-button type="primary" @click="saveEdit">保存</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="memberDialogVisible" :title="memberDialogTitle" width="520px">
    <el-empty v-if="memberNames.length === 0" description="暂无成员" />
    <div v-else>
      <el-tag v-for="name in memberNames" :key="name" style="margin-right: 8px; margin-bottom: 8px">{{ name }}</el-tag>
    </div>
    <template #footer>
      <el-button @click="memberDialogVisible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>
