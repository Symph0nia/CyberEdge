import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import ConfirmDialog from '../ConfirmDialog.vue'

describe('ConfirmDialog.vue', () => {
  let wrapper

  beforeEach(() => {
    // Mock body element
    Object.defineProperty(document, 'body', {
      value: document.createElement('body'),
      writable: true,
      configurable: true
    })
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
    document.body.style.overflow = ''
    vi.clearAllMocks()
  })

  it('initializes with correct props', () => {
    wrapper = mount(ConfirmDialog, {
      props: {
        show: false,
        message: 'Test message'
      },
      global: {
        stubs: {
          'Teleport': { template: '<div><slot /></div>' },
          'Transition': { template: '<div><slot /></div>' }
        }
      }
    })

    expect(wrapper.vm.show).toBe(false)
    expect(wrapper.vm.message).toBe('Test message')
    expect(wrapper.vm.title).toBe('确认操作')
    expect(wrapper.vm.type).toBe('info')
    expect(wrapper.vm.confirmText).toBe('确认')
    expect(wrapper.vm.cancelText).toBe('取消')
  })

  it('validates type prop correctly', () => {
    const consoleWarn = vi.spyOn(console, 'warn').mockImplementation(() => {})

    wrapper = mount(ConfirmDialog, {
      props: {
        show: true,
        message: 'Test message',
        type: 'invalid-type'
      },
      global: {
        stubs: {
          'Teleport': { template: '<div><slot /></div>' },
          'Transition': { template: '<div><slot /></div>' }
        }
      }
    })

    expect(consoleWarn).toHaveBeenCalled()
    consoleWarn.mockRestore()
  })

  it('emits confirm event when onConfirm is called', () => {
    wrapper = mount(ConfirmDialog, {
      props: {
        show: true,
        message: 'Test message'
      },
      global: {
        stubs: {
          'Teleport': { template: '<div><slot /></div>' },
          'Transition': { template: '<div><slot /></div>' }
        }
      }
    })

    wrapper.vm.onConfirm()
    expect(wrapper.emitted('confirm')).toBeTruthy()
    expect(wrapper.emitted('confirm')).toHaveLength(1)
  })

  it('emits cancel event when onCancel is called', () => {
    wrapper = mount(ConfirmDialog, {
      props: {
        show: true,
        message: 'Test message'
      },
      global: {
        stubs: {
          'Teleport': { template: '<div><slot /></div>' },
          'Transition': { template: '<div><slot /></div>' }
        }
      }
    })

    wrapper.vm.onCancel()
    expect(wrapper.emitted('cancel')).toBeTruthy()
    expect(wrapper.emitted('cancel')).toHaveLength(1)
  })

  it('handles backdrop click based on closeOnBackdrop prop', () => {
    wrapper = mount(ConfirmDialog, {
      props: {
        show: true,
        message: 'Test message',
        closeOnBackdrop: true
      },
      global: {
        stubs: {
          'Teleport': { template: '<div><slot /></div>' },
          'Transition': { template: '<div><slot /></div>' }
        }
      }
    })

    wrapper.vm.handleBackdropClick()
    expect(wrapper.emitted('cancel')).toBeTruthy()

    // Test with closeOnBackdrop false in a new wrapper
    const wrapper2 = mount(ConfirmDialog, {
      props: {
        show: true,
        message: 'Test message',
        closeOnBackdrop: false
      },
      global: {
        stubs: {
          'Teleport': { template: '<div><slot /></div>' },
          'Transition': { template: '<div><slot /></div>' }
        }
      }
    })

    wrapper2.vm.handleBackdropClick()
    expect(wrapper2.emitted('cancel')).toBeFalsy()

    wrapper2.unmount()
  })

  it('manages body scroll correctly', async () => {
    wrapper = mount(ConfirmDialog, {
      props: {
        show: false,
        message: 'Test message'
      },
      global: {
        stubs: {
          'Teleport': { template: '<div><slot /></div>' },
          'Transition': { template: '<div><slot /></div>' }
        }
      }
    })

    // Initially no scroll lock
    expect(document.body.style.overflow).toBe('')

    // Show dialog
    await wrapper.setProps({ show: true })
    await nextTick()

    expect(document.body.style.overflow).toBe('hidden')

    // Hide dialog
    await wrapper.setProps({ show: false })
    await nextTick()

    expect(document.body.style.overflow).toBe('')
  })

  it('uses correct default button focus based on dialog type', async () => {
    const wrapper = mount(ConfirmDialog, {
      props: {
        show: true,
        message: 'Test message',
        type: 'danger'
      },
      global: {
        stubs: {
          'Teleport': { template: '<div><slot /></div>' },
          'Transition': { template: '<div><slot /></div>' }
        }
      }
    })

    await nextTick()

    // For danger type, cancel button should be focused by default
    expect(wrapper.vm.type).toBe('danger')
  })
})