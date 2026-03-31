import { reactive, ref } from 'vue'

const deepClone = (value) => {
  if (typeof structuredClone === 'function') {
    try {
      return structuredClone(value)
    } catch (e) {
      if (import.meta?.env?.DEV) {
        console.warn('[clone-fallback]', e)
      }
    }
  }
  return JSON.parse(JSON.stringify(value))
}

export function useEditModal({ idField, editableFields, fieldMap, normalizeForRequest }) {
  const editDialogVisible = ref(false)
  const editingId = ref(null)
  const editForm = reactive({})
  const editCache = ref({})
  const editInitial = ref({})
  const editSaveHooks = reactive({})

  const registerSaveHook = (name, hook) => {
    if (!name || !hook) return
    editSaveHooks[name] = hook
  }

  const openEdit = async (row) => {
    editingId.value = row[idField]
    editCache.value = deepClone(row || {})
    Object.keys(editForm).forEach((k) => delete editForm[k])
    editableFields.forEach((k) => {
      const field = fieldMap()[k]
      const value = editCache.value[k]
      if (value === null || value === undefined) {
        editForm[k] = field?.inputType === 'select' ? null : ''
        return
      }
      editForm[k] = deepClone(value)
    })
    const initial = {}
    editableFields.forEach((k) => {
      initial[k] = deepClone(editForm[k])
    })
    editInitial.value = normalizeForRequest(initial, fieldMap())
    for (const item of Object.values(editSaveHooks)) {
      if (item.refresh) {
        await item.refresh()
      }
    }
    editDialogVisible.value = true
  }

  const rollbackEditForm = () => {
    editableFields.forEach((k) => {
      const field = fieldMap()[k]
      const value = editInitial.value[k]
      if (value === null || value === undefined) {
        editForm[k] = field?.inputType === 'select' ? null : ''
      } else {
        editForm[k] = deepClone(value)
      }
    })
  }

  const clearEditState = () => {
    Object.keys(editForm).forEach((k) => delete editForm[k])
    editInitial.value = {}
    editingId.value = null
    editDialogVisible.value = false
    editCache.value = {}
  }

  const collectFormDelta = () => {
    const currentRaw = {}
    const initialRaw = {}
    editableFields.forEach((k) => {
      currentRaw[k] = editForm[k]
      initialRaw[k] = editInitial.value[k]
    })
    const current = normalizeForRequest(currentRaw, fieldMap())
    const initial = normalizeForRequest(initialRaw, fieldMap())
    const delta = {}
    Object.keys(current).forEach((key) => {
      if (JSON.stringify(current[key]) !== JSON.stringify(initial[key])) {
        delta[key] = current[key]
      }
    })
    return delta
  }

  const snapshotCurrentAsInitial = () => {
    const nextInitial = {}
    editableFields.forEach((k) => {
      nextInitial[k] = deepClone(editForm[k])
    })
    editInitial.value = normalizeForRequest(nextInitial, fieldMap())
  }

  return {
    editDialogVisible,
    editingId,
    editForm,
    editCache,
    editInitial,
    editSaveHooks,
    registerSaveHook,
    openEdit,
    rollbackEditForm,
    clearEditState,
    collectFormDelta,
    snapshotCurrentAsInitial
  }
}
