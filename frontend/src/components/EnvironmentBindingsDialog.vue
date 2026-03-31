<script setup>
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api, { unwrap } from '../services/api'

const props = defineProps({
  modelValue: Boolean,
  targetType: { type: String, required: true },
  targetId: { type: [Number, String], default: null },
  targetName: { type: String, default: '' },
  hostId: { type: [Number, String], default: null }
})

const emit = defineEmits(['update:modelValue', 'updated'])

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

const loading = ref(false)
const allEnvironments = ref([])
const bindings = ref([])
const selectedEnvironmentId = ref(null)
const inheritedEnvironments = ref([])

const endpoint = computed(() => (props.targetType === 'host' ? '/host-environments' : '/service-environments'))
const idKey = computed(() => (props.targetType === 'host' ? 'host_id' : 'service_id'))

const environmentName = (id) => {
  const hit = allEnvironments.value.find((x) => x.id === id)
  return hit ? hit.name : `#${id}`
}

const sourceText = computed(() => {
  if (props.targetType !== 'service') return ''
  if (bindings.value.length > 0) return '使用服务自身环境标签'
  if (props.hostId && inheritedEnvironments.value.length > 0) return '继承自主机'
  return '无环境标签'
})

const displayedServiceEnvironments = computed(() => {
  if (props.targetType !== 'service') return []
  if (bindings.value.length > 0) return bindings.value.map((x) => x.environment_id)
  return inheritedEnvironments.value
})

const loadAllEnvironments = async () => {
  const data = unwrap(await api.get('/environments', { params: { page: 1, page_size: 200 } }))
  allEnvironments.value = data.list || []
}

const loadBindings = async () => {
  if (!props.targetId) return
  const data = unwrap(await api.get(endpoint.value, { params: { [idKey.value]: props.targetId, page: 1, page_size: 200 } }))
  bindings.value = data.list || []
}

const loadInheritedForService = async () => {
  inheritedEnvironments.value = []
  if (props.targetType !== 'service' || !props.hostId) return
  const data = unwrap(await api.get('/host-environments', { params: { host_id: props.hostId, page: 1, page_size: 200 } }))
  inheritedEnvironments.value = (data.list || []).map((x) => x.environment_id)
}

const load = async () => {
  if (!props.targetId) return
  loading.value = true
  try {
    await Promise.all([loadAllEnvironments(), loadBindings(), loadInheritedForService()])
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '加载失败')
  } finally {
    loading.value = false
  }
}

const addBinding = async () => {
  if (!selectedEnvironmentId.value || !props.targetId) return
  try {
    await api.post(endpoint.value, { [idKey.value]: props.targetId, environment_id: selectedEnvironmentId.value })
    ElMessage.success('绑定成功')
    selectedEnvironmentId.value = null
    await load()
    emit('updated')
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '绑定失败')
  }
}

const removeBinding = async (row) => {
  try {
    await ElMessageBox.confirm('确认移除该环境标签？', '提示', { type: 'warning' })
    await api.delete(endpoint.value, { data: { [idKey.value]: props.targetId, environment_id: row.environment_id } })
    ElMessage.success('移除成功')
    await load()
    emit('updated')
  } catch (e) {
    if (e !== 'cancel' && e !== 'close') {
      ElMessage.error(e.response?.data?.message || '移除失败')
    }
  }
}

watch(
  () => props.modelValue,
  async (v) => {
    if (v) await load()
  }
)
</script>

<template>
  <el-dialog v-model="visible" :title="`${targetName || '对象'} - 环境标签`" width="640px">
    <div v-loading="loading">
      <el-form inline>
        <el-form-item label="环境">
          <el-select v-model="selectedEnvironmentId" filterable placeholder="请选择环境" style="width: 260px">
            <el-option v-for="env in allEnvironments" :key="env.id" :label="env.name" :value="env.id" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="addBinding">添加标签</el-button>
        </el-form-item>
      </el-form>

      <el-alert v-if="targetType === 'service'" :title="`当前生效来源：${sourceText}`" type="info" show-icon :closable="false" />

      <el-table :data="bindings" style="margin-top: 12px">
        <el-table-column label="环境名称">
          <template #default="{ row }">
            {{ environmentName(row.environment_id) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button link type="danger" @click="removeBinding(row)">移除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <template v-if="targetType === 'service' && bindings.length === 0 && displayedServiceEnvironments.length > 0">
        <div style="margin-top: 12px">当前继承环境：</div>
        <el-tag v-for="eid in displayedServiceEnvironments" :key="eid" style="margin-right: 8px; margin-top: 8px">{{ environmentName(eid) }}</el-tag>
      </template>
    </div>
    <template #footer>
      <el-button @click="visible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>
