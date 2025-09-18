import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import StatusCard from '../StatusCard.vue'

describe('StatusCard.vue', () => {
  it('renders title and value correctly', () => {
    const title = 'Test Title'
    const value = 'Test Value'

    const wrapper = mount(StatusCard, {
      props: { title, value }
    })

    expect(wrapper.find('h3').text()).toBe(title)
    expect(wrapper.find('p').text()).toBe(value)
  })

  it('displays default dash when value is empty', () => {
    const title = 'Test Title'

    const wrapper = mount(StatusCard, {
      props: { title }
    })

    expect(wrapper.find('h3').text()).toBe(title)
    expect(wrapper.find('p').text()).toBe('-')
  })

  it('displays default dash when value is null', () => {
    const title = 'Test Title'
    const value = null

    const wrapper = mount(StatusCard, {
      props: { title, value }
    })

    expect(wrapper.find('h3').text()).toBe(title)
    expect(wrapper.find('p').text()).toBe('-')
  })

  it('displays default dash when value is undefined', () => {
    const title = 'Test Title'
    const value = undefined

    const wrapper = mount(StatusCard, {
      props: { title, value }
    })

    expect(wrapper.find('h3').text()).toBe(title)
    expect(wrapper.find('p').text()).toBe('-')
  })

  it('renders with numeric value', () => {
    const title = 'Count'
    const value = '42'

    const wrapper = mount(StatusCard, {
      props: { title, value }
    })

    expect(wrapper.find('h3').text()).toBe(title)
    expect(wrapper.find('p').text()).toBe('42')
  })

  it('requires title prop', () => {
    // 测试title prop是必需的
    const consoleWarn = vi.spyOn(console, 'warn').mockImplementation(() => {})

    mount(StatusCard, {
      props: { value: 'test' }
    })

    // Vue会在控制台发出警告
    expect(consoleWarn).toHaveBeenCalled()
    consoleWarn.mockRestore()
  })

  it('has correct component name', () => {
    const wrapper = mount(StatusCard, {
      props: { title: 'Test' }
    })

    expect(wrapper.vm.$options.name).toBe('StatusCard')
  })
})