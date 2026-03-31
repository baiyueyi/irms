<script setup>
import { reactive, ref, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api, { unwrap } from '../services/api'
import { normalizeForRequest, stringifyPayload } from './crudPayload'
import { useEditModal } from '../composables/useEditModal'

const props = defineProps({
  title: String,
  endpoint: String,
  columns: Array,
  formModel: Object,
  idField: { type: String, default: 'id' },
  editableFields: { type: Array, default: () => [] },
  deletable: { type: Boolean, default: true },
  updatable: { type: Boolean, default: true },
  deltaUpdate: { type: Boolean, default: false },
  saveDebounceMs: { type: Number, default: 500 },
  filters: { type: Array, default: () => [] },
  formFields: { type: Array, default: () => [] },
  refreshToken: { type: [String, Number], default: '' }
})

const loading = ref(false)
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const form = reactive({ ...props.formModel })
const filterValues = reactive({})
const createDialogVisible = ref(false)
const saveLoading = ref(false)
const lastSaveAt = ref(0)

props.filters.forEach((f) => {
  filterValues[f.key] = ''
})

const fieldMap = computed(() => {
  const m = {}
  props.formFields.forEach((f) => {
    m[f.key] = f
  })
  return m
})
const {
  editDialogVisible,
  editingId,
  editForm,
  editCache,
  editSaveHooks,
  registerSaveHook,
  openEdit,
  rollbackEditForm,
  clearEditState,
  collectFormDelta,
  snapshotCurrentAsInitial
} = useEditModal({
  idField: props.idField,
  editableFields: props.editableFields,
  fieldMap: () => fieldMap.value,
  normalizeForRequest
})

const createFields = computed(() => {
  if (props.formFields.length > 0) {
    return props.formFields.filter((f) => f.creatable !== false).filter((f) => (f.showWhen ? f.showWhen(form) : true))
  }
  return props.columns.filter((c) => c.creatable).map((c) => ({ key: c.prop, label: c.label, inputType: 'input' }))
})

const editFields = computed(() => {
  return props.editableFields
    .map((k) => fieldMap.value[k] || { key: k, label: k, inputType: 'input' })
    .filter((f) => (f.showWhen ? f.showWhen(editForm) : true))
})

const normalizeTagList = (value) => {
  if (Array.isArray(value)) {
    return value.map((x) => String(x)).filter((x) => x.trim() !== '')
  }
  if (value === null || value === undefined || value === '') return []
  return [String(value)]
}

const previewTags = (value, max = 2) => normalizeTagList(value).slice(0, max)
const hiddenTagCount = (value, max = 2) => Math.max(0, normalizeTagList(value).length - max)
const fullTagText = (value) => normalizeTagList(value).join(' / ')

const envSourceType = (value) => {
  const text = String(value || '')
  if (text.includes('服务自身')) return 'success'
  if (text.includes('继承')) return 'warning'
  return 'info'
}

const load = async () => {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value, keyword: keyword.value, ...filterValues }
    const data = unwrap(await api.get(props.endpoint, { params }))
    list.value = data.list || []
    total.value = data.pagination?.total || 0
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '查询失败')
  } finally {
    loading.value = false
  }
}

const openCreate = () => {
  Object.keys(form).forEach((k) => (form[k] = props.formModel[k] ?? ''))
  createDialogVisible.value = true
}

const buildRequestPayload = (source) => {
  try {
    return JSON.parse(stringifyPayload(source, fieldMap.value))
  } catch {
    ElMessage.error('请求数据序列化失败，请检查输入内容')
    return null
  }
}

const createOne = async () => {
  const payload = buildRequestPayload(form)
  if (!payload) return
  try {
    await api.post(props.endpoint, payload)
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    Object.keys(form).forEach((k) => (form[k] = props.formModel[k] ?? ''))
    await load()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '创建失败')
  }
}

const cancelEdit = async () => {
  rollbackEditForm()
  for (const item of Object.values(editSaveHooks)) {
    if (item.rollback) {
      await item.rollback()
    }
  }
  const residue = Object.values(editSaveHooks).some((item) => item.collectDelta && item.collectDelta().changed)
  if (residue) {
    console.warn('[save-audit]', {
      endpoint: props.endpoint,
      id: editingId.value,
      reason: 'cancel_with_residue',
      at: new Date().toISOString()
    })
  }
  clearEditState()
}

const saveEdit = async (retrying = false) => {
  const now = Date.now()
  if (!retrying && (saveLoading.value || now - lastSaveAt.value < props.saveDebounceMs)) return
  const formDelta = collectFormDelta()
  const hookChanges = Object.entries(editSaveHooks)
    .map(([name, hook]) => {
      const detail = hook.collectDelta ? hook.collectDelta() : { changed: false, delta: {} }
      return { name, hook, detail }
    })
    .filter((x) => x.detail?.changed)
  if (Object.keys(formDelta).length === 0 && hookChanges.length === 0) {
    ElMessage.info('无变更，无需保存')
    return
  }
  try {
    saveLoading.value = true
    console.info('[save-audit]', {
      endpoint: props.endpoint,
      id: editingId.value,
      changedFields: Object.keys(formDelta),
      extraChanges: hookChanges.map((x) => ({ name: x.name, delta: x.detail.delta || {} })),
      at: new Date().toISOString()
    })
    if (Object.keys(formDelta).length > 0) {
      if (props.deltaUpdate) {
        const deltaPayload = buildRequestPayload(formDelta)
        if (!deltaPayload) return
        await api.put(`${props.endpoint}/${editingId.value}`, deltaPayload)
      } else {
        const fullPayload = buildRequestPayload(editForm)
        if (!fullPayload) return
        await api.put(`${props.endpoint}/${editingId.value}`, fullPayload)
      }
    }
    for (const item of hookChanges) {
      if (item.hook.applyDelta) {
        await item.hook.applyDelta()
      }
    }
    snapshotCurrentAsInitial()
    for (const item of hookChanges) {
      if (item.hook.commit) {
        item.hook.commit()
      }
    }
    ElMessage.success('编辑成功')
    lastSaveAt.value = Date.now()
    clearEditState()
    await load()
  } catch (e) {
    const errorMsg = e.response?.data?.message || '编辑失败'
    console.error('[save-audit]', {
      endpoint: props.endpoint,
      id: editingId.value,
      changedFields: Object.keys(formDelta),
      error: errorMsg,
      at: new Date().toISOString()
    })
    try {
      await ElMessageBox.confirm(`${errorMsg}。点击“确定”重试，点击“取消”回滚。`, '保存失败', {
        confirmButtonText: '重试',
        cancelButtonText: '回滚',
        distinguishCancelAndClose: true,
        type: 'error'
      })
      saveLoading.value = false
      await saveEdit(true)
      return
    } catch (action) {
      if (action === 'cancel') {
        rollbackEditForm()
        for (const item of hookChanges) {
          if (item.hook.rollback) {
            await item.hook.rollback()
          }
        }
        ElMessage.info('已回滚本地修改')
      } else {
        ElMessage.warning('已保留未提交修改，可继续重试')
      }
    }
  } finally {
    saveLoading.value = false
  }
}

const removeOne = async (row) => {
  try {
    await ElMessageBox.confirm('确认删除该记录？', '提示', { type: 'warning' })
    await api.delete(`${props.endpoint}/${row[props.idField]}`)
    ElMessage.success('删除成功')
    await load()
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') {
      ElMessage.error(e.response?.data?.message || '删除失败')
    }
  }
}

watch(() => props.refreshToken, load, { immediate: true })
</script>

<template>
  <el-card>
    <template #header>
      <div class="row-between">
        <span>{{ title }}</span>
      </div>
    </template>
    <div class="toolbar">
      <el-input v-model="keyword" placeholder="关键字" style="width: 240px" />
      <el-select v-for="f in filters" :key="f.key" v-model="filterValues[f.key]" clearable :placeholder="f.label" style="width: 160px">
        <el-option v-for="op in f.options" :key="op.value" :label="op.label" :value="op.value" />
      </el-select>
      <el-button @click="load">查询</el-button>
      <el-button type="primary" @click="openCreate">新增</el-button>
    </div>
    <el-table :data="list" v-loading="loading">
      <el-table-column v-for="col in columns" :key="col.prop" :prop="col.prop" :label="col.label">
        <template #default="{ row }">
          <template v-if="col.renderType === 'envTags'">
            <div style="display: flex; align-items: center; gap: 6px; min-height: 28px; overflow: hidden">
              <el-tag
                v-for="tag in previewTags(row?.[col.prop], col.maxTags || 2)"
                :key="`${row?.[props.idField] || 'row'}-${col.prop}-${tag}`"
                size="small"
                type="info"
              >
                {{ tag }}
              </el-tag>
              <el-tooltip v-if="hiddenTagCount(row?.[col.prop], col.maxTags || 2) > 0" :content="fullTagText(row?.[col.prop])" placement="top">
                <el-tag size="small" type="info">+{{ hiddenTagCount(row?.[col.prop], col.maxTags || 2) }}</el-tag>
              </el-tooltip>
              <span v-if="previewTags(row?.[col.prop], col.maxTags || 2).length === 0" style="color: #909399">-</span>
            </div>
          </template>
          <template v-else-if="col.renderType === 'envSource'">
            <el-tag size="small" :type="envSourceType(row?.[col.prop])">{{ row?.[col.prop] || '无环境' }}</el-tag>
          </template>
          <template v-else>
            {{ row?.[col.prop] }}
          </template>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180">
        <template #default="{ row }">
          <slot name="row-actions" :row="row" :reload="load" />
          <el-button v-if="updatable" link type="primary" @click="openEdit(row)">编辑</el-button>
          <el-button v-if="deletable" link type="danger" @click="removeOne(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div class="pager-wrap">
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" layout="prev, pager, next, total" :total="total" @current-change="load" />
    </div>
    <el-dialog v-model="createDialogVisible" title="新增">
      <el-form label-position="top">
        <el-form-item v-for="f in createFields" :key="f.key" :label="f.label || f.key">
          <el-select v-if="f.inputType === 'select'" v-model="form[f.key]" :filterable="!!f.filterable" clearable style="width: 100%">
            <el-option v-for="op in f.options || []" :key="op.value" :label="op.label" :value="op.value" />
          </el-select>
          <el-input v-else v-model="form[f.key]" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createOne">确认</el-button>
      </template>
    </el-dialog>
    <el-dialog v-model="editDialogVisible" title="编辑" @close="cancelEdit">
      <el-form label-position="top">
        <el-form-item v-for="f in editFields" :key="f.key" :label="f.label || f.key">
          <el-select v-if="f.inputType === 'select'" v-model="editForm[f.key]" :filterable="!!f.filterable" clearable style="width: 100%">
            <el-option v-for="op in f.options || []" :key="op.value" :label="op.label" :value="op.value" />
          </el-select>
          <el-input v-else v-model="editForm[f.key]" />
        </el-form-item>
      </el-form>
      <slot
        name="edit-extra"
        :edit-form="editForm"
        :editing-id="editingId"
        :edit-cache="editCache"
        :reload="load"
        :register-save-hook="registerSaveHook"
        :save-loading="saveLoading"
      />
      <template #footer>
        <el-button @click="cancelEdit">取消</el-button>
        <el-button type="primary" :loading="saveLoading" :disabled="saveLoading" @click="saveEdit">保存</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>
