import { shallowMount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import EnvironmentBindingsInline from '../EnvironmentBindingsInline.vue'

describe('EnvironmentBindingsInline', () => {
  it('uses filterable multi-select with clearable tags style behavior', () => {
    const registerSaveHook = vi.fn()
    const wrapper = shallowMount(EnvironmentBindingsInline, {
      props: {
        targetType: 'host',
        targetId: null,
        registerSaveHook
      },
      global: {
        directives: {
          loading: () => {}
        },
        stubs: {
          'el-divider': true,
          'el-form': { template: '<form><slot /></form>' },
          'el-form-item': { template: '<div><slot /></div>' },
          'el-select': {
            props: {
              multiple: Boolean,
              clearable: Boolean,
              filterable: Boolean
            },
            template:
              '<div class="el-select-stub" :data-multiple="String(multiple)" :data-clearable="String(clearable)" :data-filterable="String(filterable)"><slot /></div>'
          },
          'el-option': true,
          'el-alert': true,
          'el-tag': true
        }
      }
    })
    const select = wrapper.find('.el-select-stub')
    expect(select.exists()).toBe(true)
    expect(select.attributes('data-multiple')).toBe('true')
    expect(select.attributes('data-clearable')).toBe('true')
    expect(select.attributes('data-filterable')).toBe('true')
    expect(registerSaveHook).toHaveBeenCalledTimes(1)
    expect(registerSaveHook.mock.calls[0][0]).toBe('environmentBindings')
  })

  it('resets unsaved selections on rollback hook', async () => {
    const registerSaveHook = vi.fn()
    const wrapper = shallowMount(EnvironmentBindingsInline, {
      props: {
        targetType: 'host',
        targetId: null,
        registerSaveHook
      },
      global: {
        directives: {
          loading: () => {}
        },
        stubs: {
          'el-divider': true,
          'el-form': { template: '<form><slot /></form>' },
          'el-form-item': { template: '<div><slot /></div>' },
          'el-select': true,
          'el-option': true,
          'el-alert': true,
          'el-tag': true
        }
      }
    })
    const hook = registerSaveHook.mock.calls[0][1]
    wrapper.vm.initialEnvironmentIds = [1, 2]
    wrapper.vm.boundEnvironmentIds = [1, 2, 3]
    expect(hook.collectDelta().changed).toBe(true)
    await hook.rollback()
    expect(wrapper.vm.boundEnvironmentIds).toEqual([1, 2])
    expect(hook.collectDelta().changed).toBe(false)
  })
})
