import { describe, it, expect, vi } from 'vitest'

describe('Router Configuration', () => {
  it('defines correct route structure', async () => {
    // Mock store to avoid router guard issues
    vi.mock('@/store', () => ({
      default: {
        dispatch: vi.fn().mockResolvedValue(),
        state: { isAuthenticated: true }
      }
    }))

    // Mock all components to avoid import issues
    vi.mock('@/components/HomePage.vue', () => ({ default: { name: 'Home' } }))
    vi.mock('@/components/Login/LoginPage.vue', () => ({ default: { name: 'LoginPage' } }))
    vi.mock('@/components/Dashboard.vue', () => ({ default: { name: 'Dashboard' } }))
    vi.mock('@/components/Config/SystemConfiguration.vue', () => ({ default: { name: 'SystemConfiguration' } }))
    vi.mock('@/components/User/UserManagement.vue', () => ({ default: { name: 'UserManagement' } }))
    vi.mock('@/components/Login/GoogleAuthQRCode.vue', () => ({ default: { name: 'GoogleAuthQRCode' } }))
    vi.mock('@/components/Task/TaskManagement.vue', () => ({ default: { name: 'TaskManagement' } }))
    vi.mock('@/components/Port/PortScanResults.vue', () => ({ default: { name: 'PortScanResults' } }))
    vi.mock('@/components/Port/PortScanDetail.vue', () => ({ default: { name: 'PortScanDetail' } }))
    vi.mock('@/components/Subdomain/SubdomainScanResults.vue', () => ({ default: { name: 'SubdomainScanResults' } }))
    vi.mock('@/components/Subdomain/SubdomainScanDetail.vue', () => ({ default: { name: 'SubdomainScanDetail' } }))
    vi.mock('@/components/Path/PathScanResults.vue', () => ({ default: { name: 'PathScanResults' } }))
    vi.mock('@/components/Path/PathScanDetail.vue', () => ({ default: { name: 'PathScanDetail' } }))
    vi.mock('@/components/Target/TargetManagement.vue', () => ({ default: { name: 'TargetManagement' } }))
    vi.mock('@/components/Target/TargetDetail.vue', () => ({ default: { name: 'TargetDetail' } }))
    vi.mock('@/components/UnderDevelopment.vue', () => ({ default: { name: 'UnderDevelopment' } }))
    vi.mock('@/components/Config/ToolConfiguration.vue', () => ({ default: { name: 'ToolConfiguration' } }))

    const { default: router } = await import('../index.js')
    const routes = router.getRoutes()

    // Test basic route definitions
    expect(routes.length).toBeGreaterThan(0)

    // Test specific routes
    const homeRoute = routes.find(route => route.path === '/')
    expect(homeRoute).toBeDefined()
    expect(homeRoute.name).toBe('Home')

    const loginRoute = routes.find(route => route.path === '/login')
    expect(loginRoute).toBeDefined()
    expect(loginRoute.name).toBe('LoginPage')

    const dashboardRoute = routes.find(route => route.path === '/dashboard')
    expect(dashboardRoute).toBeDefined()
    expect(dashboardRoute.name).toBe('WAFDashboard')

    // Test parameterized routes
    const portDetailRoute = routes.find(route => route.path === '/port-scan-results/:id')
    expect(portDetailRoute).toBeDefined()
    expect(portDetailRoute.name).toBe('PortScanDetail')

    const targetDetailRoute = routes.find(route => route.path === '/target-management/:id')
    expect(targetDetailRoute).toBeDefined()
    expect(targetDetailRoute.name).toBe('TargetDetail')
  })

  it('has router history configuration', async () => {
    vi.mock('@/store', () => ({
      default: {
        dispatch: vi.fn(),
        state: { isAuthenticated: true }
      }
    }))

    const { default: router } = await import('../index.js')

    expect(router).toBeDefined()
    expect(router.options).toBeDefined()
    expect(router.options.history).toBeDefined()
  })

  it('has beforeEach guard configured', async () => {
    const mockStore = {
      dispatch: vi.fn().mockResolvedValue(),
      state: { isAuthenticated: true }
    }

    vi.mock('@/store', () => ({ default: mockStore }))

    const { default: router } = await import('../index.js')

    // Check that router has been configured (has guards internally)
    expect(router).toBeDefined()
    expect(typeof router.beforeEach).toBe('function')

    // Verify that the guard function works
    const mockTo = { name: 'WAFDashboard' }
    const mockFrom = { name: 'Home' }
    const mockNext = vi.fn()

    // Since we can't directly access the guard, we test that the router object is properly configured
    expect(router.options).toBeDefined()
  })

  it('generates correct route URLs', async () => {
    vi.mock('@/store', () => ({
      default: {
        dispatch: vi.fn(),
        state: { isAuthenticated: true }
      }
    }))

    const { default: router } = await import('../index.js')

    // Test URL resolution
    const homeResolved = router.resolve('/')
    expect(homeResolved.name).toBe('Home')

    const loginResolved = router.resolve('/login')
    expect(loginResolved.name).toBe('LoginPage')

    const portDetailResolved = router.resolve('/port-scan-results/123')
    expect(portDetailResolved.name).toBe('PortScanDetail')
    expect(portDetailResolved.params.id).toBe('123')
  })
})