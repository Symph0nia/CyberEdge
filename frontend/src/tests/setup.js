import { config } from '@vue/test-utils'
import { vi } from 'vitest'

// Mock Ant Design Vue components globally
const mockAntdComponents = [
  'a-input', 'a-input-search', 'a-input-password', 'a-button', 'a-form', 'a-form-item',
  'a-card', 'a-table', 'a-modal', 'a-avatar', 'a-badge', 'a-tag', 'a-alert', 'a-spin',
  'a-steps', 'a-step', 'a-switch', 'a-divider', 'a-space', 'a-tooltip', 'a-popconfirm'
]

const stubs = {}
mockAntdComponents.forEach(component => {
  stubs[component] = {
    template: `<div class="${component}"><slot /></div>`,
    props: ['loading', 'disabled', 'type', 'size', 'placeholder', 'value', 'modelValue'],
    emits: ['update:modelValue', 'click', 'submit', 'change']
  }
})

// Mock Ant Design Vue icons
stubs['SafetyOutlined'] = { template: '<div class="safety-outlined" />' }
stubs['UserOutlined'] = { template: '<div class="user-outlined" />' }
stubs['LockOutlined'] = { template: '<div class="lock-outlined" />' }

config.global.stubs = stubs

// Mock axios
const mockAxios = {
  get: vi.fn(() => Promise.resolve({ data: {} })),
  post: vi.fn(() => Promise.resolve({ data: {} })),
  put: vi.fn(() => Promise.resolve({ data: {} })),
  delete: vi.fn(() => Promise.resolve({ data: {} })),
  defaults: {
    headers: {
      common: {}
    }
  }
}

vi.mock('axios', () => ({
  default: mockAxios,
  ...mockAxios
}))

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
}
global.localStorage = localStorageMock