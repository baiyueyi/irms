<script setup>
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api, { unwrap } from '../services/api'

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')

const userOptions = ref([])
const memberDialogVisible = ref(false)
const memberDialogTitle = ref('')
const memberNames = ref([])

const createVisible = ref(false)
const editVisible = ref(false)
const editingId = ref(null)

const createForm = reactive({ name: '', description: '', member_ids: [] })
const editForm = reactive({ name: '', description: '', member_ids: [] })

const loadUsers = async (query = '') => {
  const data = unwrap(await api.get('/users', { params: { page: 1, page_size: 200, keyword: query } }))
  userOptions.value = (data.list || []).map((x) => ({ key: x.id, label: x.username }))
}

const load = async () => {
  loading.value = true
  try {
    const data = unwrap(await api.get('/user-groups', { params: { page: page.value, page_size: pageSize.value, keyword: keyword.value } }))
    list.value = data.list || []
    total.value = data.pagination?.total || 0
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const getMembers = async (groupId) => {
  const data = unwrap(await api.get('/user-group-members', { params: { user_group_id: groupId, page: 1, page_size: 500 } }))
  return data.list || []
}

const syncMembers = async (groupId, targetIds) => {
  const current = await getMembers(groupId)
  const currentIds = new Set(current.map((x) => x.user_id))
  const target = new Set(targetIds)
  for (const uid of target) {
    if (!currentIds.has(uid)) {
      await api.post('/user-group-members', { user_id: uid, user_group_id: groupId })
    }
  }
  for (const row of current) {
    if (!target.has(row.user_id)) {
      await api.delete('/user-group-members', { data: { user_id: row.user_id, user_group_id: groupId } })
    }
  }
}

const openCreate = async () => {
  createForm.name = ''
  createForm.description = ''
  createForm.member_ids = []
  createVisible.value = true
  try {
    await loadUsers('')
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '加载用户失败')
  }
}

const createOne = async () => {
  try {
    const resp = unwrap(await api.post('/user-groups', { name: createForm.name, description: createForm.description }))
    await syncMembers(resp.id, createForm.member_ids)
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
  editForm.description = row.description || ''
  editVisible.value = true
  try {
    await loadUsers('')
    const members = await getMembers(row.id)
    editForm.member_ids = members.map((x) => x.user_id)
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '加载成员失败')
  }
}

const saveEdit = async () => {
  try {
    await api.put(`/user-groups/${editingId.value}`, { name: editForm.name, description: editForm.description })
    await syncMembers(editingId.value, editForm.member_ids)
    ElMessage.success('保存成功')
    editVisible.value = false
    await load()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '保存失败')
  }
}

const removeOne = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该用户组？', '提示', { type: 'warning' })
    const members = await getMembers(row.id)
    for (const m of members) {
      await api.delete('/user-group-members', { data: { user_id: m.user_id, user_group_id: row.id } })
    }
    await api.delete(`/user-groups/${row.id}`)
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
    memberNames.value = members.map((x) => x.user_name || `#${x.user_id}`)
    memberDialogTitle.value = `${row.name} - 成员`
    memberDialogVisible.value = true
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '加载成员失败')
  }
}

onMounted(async () => {
  await Promise.all([load(), loadUsers('')])
})
</script>

<template>
  <el-card>
    <template #header>
      <div class="row-between">
        <span>用户组管理</span>
      </div>
    </template>
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="关键字" style="width: 240px" />
      <el-button @click="load">查询</el-button>
      <el-button type="primary" @click="openCreate">新增</el-button>
    </div>
    <el-table :data="list" v-loading="loading">
      <el-table-column prop="id" label="ID" />
      <el-table-column prop="name" label="名称" />
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

  <el-dialog v-model="createVisible" title="新增用户组" width="720px">
    <el-form label-position="top">
      <el-form-item label="名称"><el-input v-model="createForm.name" /></el-form-item>
      <el-form-item label="描述"><el-input v-model="createForm.description" /></el-form-item>
      <el-form-item label="成员选择">
        <el-transfer v-model="createForm.member_ids" filterable :data="userOptions" :titles="['可选用户', '已选成员']" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="createVisible = false">取消</el-button>
      <el-button type="primary" @click="createOne">确认</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="editVisible" title="编辑用户组" width="720px">
    <el-form label-position="top">
      <el-form-item label="名称"><el-input v-model="editForm.name" /></el-form-item>
      <el-form-item label="描述"><el-input v-model="editForm.description" /></el-form-item>
      <el-form-item label="成员选择">
        <el-transfer v-model="editForm.member_ids" filterable :data="userOptions" :titles="['可选用户', '已选成员']" />
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
