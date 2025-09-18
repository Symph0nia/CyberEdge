import { vi } from 'vitest'

// Mock axios全局配置
vi.mock('axios', () => ({
  default: {
    create: vi.fn(() => ({
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
      interceptors: {
        request: { use: vi.fn() },
        response: { use: vi.fn() }
      }
    })),
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }
}))

// Mock Chart.js
vi.mock('chart.js', () => ({
  Chart: {
    register: vi.fn()
  },
  registerables: [],
  CategoryScale: vi.fn(),
  LinearScale: vi.fn(),
  BarElement: vi.fn(),
  Title: vi.fn(),
  Tooltip: vi.fn(),
  Legend: vi.fn(),
  ArcElement: vi.fn(),
  LineElement: vi.fn(),
  PointElement: vi.fn()
}))

// Mock vue-chartjs
vi.mock('vue-chartjs', () => ({
  Line: vi.fn(),
  Bar: vi.fn(),
  Doughnut: vi.fn(),
  Pie: vi.fn()
}))

// 全局测试配置
global.console = {
  ...console,
  // 在测试中忽略某些console输出
  warn: vi.fn(),
  error: vi.fn()
}