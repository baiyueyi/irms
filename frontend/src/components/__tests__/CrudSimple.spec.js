import { shallowMount } from '@vue/test-utils'
import { describe, expect, it, beforeEach, vi } from 'vitest'
import CrudSimple from '../CrudSimple.vue'

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

const mountCrud = () =>
  shallowMount(CrudSimple, {
    props: {
      title: '测试',
      endpoint: '/items',
      idField: 'id',
      columns: [{ prop: 'id', label: 'ID' }, { prop: 'name', label: '名称', creatable: true }],
      formModel: { name: '' },
      editableFields: ['name'],
      formFields: [{ key: 'name', label: '名称' }],
      deltaUpdate: true
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
        'el-tag': { template: '<span><slot /></span>' },
        'el-tooltip': { template: '<span><slot /></span>' },
        'el-button': { template: '<button><slot /></button>' },
        'el-table': { template: '<table><slot /></table>' },
        'el-table-column': { template: '<div><slot /></div>' },
        'el-pagination': true,
        'el-form': { template: '<form><slot /></form>' },
        'el-form-item': { template: '<div><slot /></div>' },
        'el-dialog': { template: '<div><slot /><slot name="footer" /></div>' }
      }
    }
  })

describe('CrudSimple edit isolation', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('keeps edit cache isolated and cancel does not pollute list', async () => {
    api.get.mockResolvedValueOnce({
      list: [{ id: 1, name: 'n1', environments: [1, 2] }],
      pagination: { total: 1 }
    })
    const wrapper = mountCrud()
    await flush()
    const row = wrapper.vm.list[0]
    await wrapper.vm.openEdit(row)
    wrapper.vm.editCache.environments.push(99)
    expect(row.environments).toEqual([1, 2])
    await wrapper.vm.cancelEdit()
    expect(wrapper.vm.list[0].environments).toEqual([1, 2])
  })

  it('refreshes list from latest api data after save success', async () => {
    api.get
      .mockResolvedValueOnce({
        list: [{ id: 1, name: 'old' }],
        pagination: { total: 1 }
      })
      .mockResolvedValueOnce({
        list: [{ id: 1, name: 'latest-from-api' }],
        pagination: { total: 1 }
      })
    api.put.mockResolvedValueOnce({ code: 'OK' })
    const wrapper = mountCrud()
    await flush()
    await wrapper.vm.openEdit(wrapper.vm.list[0])
    wrapper.vm.editForm.name = 'changed'
    await wrapper.vm.saveEdit()
    expect(api.put).toHaveBeenCalledWith('/items/1', { name: 'changed' })
    expect(wrapper.vm.list[0].name).toBe('latest-from-api')
  })
})
