import { mount } from '@vue/test-utils'
import { h } from 'vue'
import { describe, expect, it, vi, beforeEach } from 'vitest'
import CrudSimple from '../CrudSimple.vue'
import EnvironmentBindingsInline from '../EnvironmentBindingsInline.vue'

vi.mock('../../services/api', () => {
  const api = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }
  return {
    default: api,
    unwrap: (x) => x
  }
})

vi.mock('element-plus', () => ({
  ElMessage: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
    warning: vi.fn()
  },
  ElMessageBox: {
    confirm: vi.fn()
  }
}))

const { default: api } = await import('../../services/api')

const flush = async () => {
  await Promise.resolve()
  await Promise.resolve()
}

describe('CrudSimple + EnvironmentBindingsInline integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('does not change list before save; refreshes list only after save success', async () => {
    let servicesStage = 0
    api.get.mockImplementation(async (url) => {
      if (url === '/services') {
        servicesStage += 1
        if (servicesStage === 1) {
          return {
            list: [{ id: 1, name: 'svc-1', host_id: 10, environments: ['Env1'], environment_source_label: '继承自主机' }],
            pagination: { total: 1 }
          }
        }
        return {
          list: [{ id: 1, name: 'svc-1', host_id: 10, environments: ['Env2', 'Env3'], environment_source_label: '使用服务自身环境标签' }],
          pagination: { total: 1 }
        }
      }
      if (url === '/environments') {
        return { list: [{ id: 1, name: 'Env1' }, { id: 2, name: 'Env2' }, { id: 3, name: 'Env3' }] }
      }
      if (url === '/service-environments') {
        return { list: [{ environment_id: 1 }] }
      }
      if (url === '/host-environments') {
        return { list: [{ environment_id: 2 }, { environment_id: 3 }] }
      }
      return { list: [], pagination: { total: 0 } }
    })
    api.put.mockResolvedValue({ code: 'OK' })
    api.post.mockResolvedValue({ code: 'OK' })
    api.delete.mockResolvedValue({ code: 'OK' })

    const wrapper = mount(CrudSimple, {
      props: {
        title: '服务管理',
        endpoint: '/services',
        idField: 'id',
        columns: [
          { prop: 'id', label: 'ID' },
          { prop: 'name', label: '名称', creatable: true },
          { prop: 'environments', label: '环境标签' },
          { prop: 'environment_source_label', label: '环境来源' }
        ],
        formModel: { name: '' },
        editableFields: ['name'],
        formFields: [{ key: 'name', label: '名称' }],
        deltaUpdate: true
      },
      slots: {
        'edit-extra': ({ editForm, editingId, registerSaveHook, saveLoading }) => {
          return h(EnvironmentBindingsInline, {
            targetType: 'service',
            targetId: editingId,
            targetName: editForm.name,
            hostId: editForm.host_id,
            registerSaveHook,
            saveLoading
          })
        }
      },
      global: {
        directives: {
          loading: () => {}
        },
        stubs: {
          'el-card': { template: '<div><slot name="header" /><slot /></div>' },
          'el-input': { template: '<input />' },
          'el-select': { template: '<select><slot /></select>' },
          'el-option': true,
          'el-button': { template: '<button><slot /></button>' },
          'el-table': { template: '<table><slot /></table>' },
          'el-table-column': { template: '<div><slot /></div>' },
          'el-pagination': true,
          'el-form': { template: '<form><slot /></form>' },
          'el-form-item': { template: '<div><slot /></div>' },
          'el-dialog': { props: ['modelValue'], template: '<div v-if="modelValue"><slot /><slot name="footer" /></div>' },
          'el-divider': true,
          'el-alert': true,
          'el-tag': { template: '<span><slot /></span>' },
          'el-tooltip': { template: '<span><slot /></span>' }
        }
      }
    })

    await flush()
    expect(wrapper.vm.list).toHaveLength(1)
    const original = JSON.parse(JSON.stringify(wrapper.vm.list[0].environments))

    await wrapper.vm.openEdit(wrapper.vm.list[0])
    await flush()
    const env1 = wrapper.findComponent(EnvironmentBindingsInline)
    expect(env1.exists()).toBe(true)
    env1.vm.boundEnvironmentIds = [2, 3]
    await flush()

    expect(wrapper.vm.list[0].environments).toEqual(original)

    await wrapper.vm.cancelEdit()
    expect(wrapper.vm.list[0].environments).toEqual(original)

    await wrapper.vm.openEdit(wrapper.vm.list[0])
    await flush()
    const env2 = wrapper.findComponent(EnvironmentBindingsInline)
    expect(env2.exists()).toBe(true)
    env2.vm.boundEnvironmentIds = [2, 3]
    await flush()

    await wrapper.vm.saveEdit()
    expect(wrapper.vm.list[0].environments).toEqual(['Env2', 'Env3'])
  })
})
