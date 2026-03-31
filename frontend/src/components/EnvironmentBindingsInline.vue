<script setup>
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import api, { unwrap } from '../services/api'
import { diffEnvironmentIds } from './environmentBindingSync'
import './EnvironmentBindingsInline.css'

const props = defineProps({
  targetType: { type: String, required: true },
  targetId: { type: [Number, String], default: null },
  targetName: { type: String, default: '' },
  hostId: { type: [Number, String], default: null },
  registerSaveHook: { type: Function, default: null },
  saveLoading: { type: Boolean, default: false }
})

const emit = defineEmits(['updated'])

const loading = ref(false)
const allEnvironments = ref([])
const bindings = ref([])
const boundEnvironmentIds = ref([])
const initialEnvironmentIds = ref([])
const inheritedEnvironments = ref([])
const loadSeq = ref(0)

const endpoint = computed(() => (props.targetType === 'host' ? '/host-environments' : '/service-environments'))
const idKey = computed(() => (props.targetType === 'host' ? 'host_id' : 'service_id'))

const environmentName = (id) => {
  const hit = allEnvironments.value.find((x) => x.id === id)
  return hit ? hit.name : `#${id}`
}

const sourceText = computed(() => {
  if (props.targetType !== 'service') return ''
  if (boundEnvironmentIds.value.length > 0) return '使用服务自身环境标签'
  if (props.hostId && inheritedEnvironments.value.length > 0) return '继承自主机'
  return '无环境标签'
})

const displayedServiceEnvironments = computed(() => {
  if (props.targetType !== 'service') return []
  if (boundEnvironmentIds.value.length > 0) return boundEnvironmentIds.value
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
  const ids = bindings.value.map((x) => x.environment_id)
  boundEnvironmentIds.value = [...ids]
  initialEnvironmentIds.value = [...ids]
}

const loadInheritedForService = async () => {
  inheritedEnvironments.value = []
  if (props.targetType !== 'service' || !props.hostId) return
  const data = unwrap(await api.get('/host-environments', { params: { host_id: props.hostId, page: 1, page_size: 200 } }))
  inheritedEnvironments.value = (data.list || []).map((x) => x.environment_id)
}

const load = async () => {
  if (!props.targetId) {
    bindings.value = []
    boundEnvironmentIds.value = []
    initialEnvironmentIds.value = []
    inheritedEnvironments.value = []
    return
  }
  const currentSeq = loadSeq.value + 1
  loadSeq.value = currentSeq
  loading.value = true
  try {
    await Promise.all([loadAllEnvironments(), loadBindings(), loadInheritedForService()])
  } catch (e) {
    ElMessage.error(e.response?.data?.message || '加载失败')
  } finally {
    if (currentSeq === loadSeq.value) {
      loading.value = false
    }
  }
}

const collectDelta = () => {
  const { added, removed } = diffEnvironmentIds(initialEnvironmentIds.value, boundEnvironmentIds.value)
  return {
    changed: added.length > 0 || removed.length > 0,
    delta: { added, removed }
  }
}

const applyDelta = async () => {
  if (!props.targetId) return
  const { added, removed } = collectDelta().delta
  const ops = [
    ...added.map((environmentId) => api.post(endpoint.value, { [idKey.value]: props.targetId, environment_id: environmentId })),
    ...removed.map((environmentId) => api.delete(endpoint.value, { data: { [idKey.value]: props.targetId, environment_id: environmentId } }))
  ]
  const results = await Promise.allSettled(ops)
  const rejected = results.find((x) => x.status === 'rejected')
  if (rejected) throw rejected.reason
  emit('updated')
}

const commit = () => {
  initialEnvironmentIds.value = [...boundEnvironmentIds.value]
  bindings.value = initialEnvironmentIds.value.map((environment_id) => ({ environment_id }))
}

const rollback = () => {
  boundEnvironmentIds.value = [...initialEnvironmentIds.value]
}

watch(
  () => [props.targetId, props.hostId],
  async () => {
    await load()
  },
  { immediate: true }
)

watch(
  () => props.registerSaveHook,
  () => {
    if (!props.registerSaveHook) return
    props.registerSaveHook('environmentBindings', {
      collectDelta,
      applyDelta,
      commit,
      rollback,
      refresh: load
    })
  },
  { immediate: true }
)
</script>

<template>
  <el-divider />
  <div style="font-weight: 600; margin-bottom: 8px">环境标签设置</div>
  <div v-loading="loading">
    <el-form inline class="env-bindings-form">
      <el-form-item label="环境">
        <el-select
          v-model="boundEnvironmentIds"
          multiple
          clearable
          filterable
          :disabled="loading || saveLoading"
          class="env-binding-select"
          placeholder="请选择环境"
          style="width: 320px"
        >
          <el-option v-for="env in allEnvironments" :key="env.id" :label="env.name" :value="env.id" />
        </el-select>
      </el-form-item>
    </el-form>

    <el-alert v-if="targetType === 'service'" :title="`当前生效来源：${sourceText}`" type="info" show-icon :closable="false" />

    <template v-if="targetType === 'service' && displayedServiceEnvironments.length > 0">
      <div style="margin-top: 12px">{{ boundEnvironmentIds.length > 0 ? '当前选择环境：' : '当前继承环境：' }}</div>
      <el-tag v-for="eid in displayedServiceEnvironments" :key="eid" style="margin-right: 8px; margin-top: 8px">{{ environmentName(eid) }}</el-tag>
    </template>
  </div>
</template>
