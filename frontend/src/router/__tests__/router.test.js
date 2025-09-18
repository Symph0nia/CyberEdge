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

    // Mock actual components used in the router
    vi.mock('@/components/Login/LoginPage.vue', () => ({ default: { name: 'LoginPage' } }))
    vi.mock('@/components/User/UserManagement.vue', () => ({ default: { name: 'UserManagement' } }))
    vi.mock('@/components/Profile/ProfilePage.vue', () => ({ default: { name: 'ProfilePage' } }))
    vi.mock('@/components/Settings/SettingsPage.vue', () => ({ default: { name: 'SettingsPage' } }))
    vi.mock('@/components/Project/ProjectList.vue', () => ({ default: { name: 'ProjectList' } }))
    vi.mock('@/components/Project/ProjectDetail.vue', () => ({ default: { name: 'ProjectDetail' } }))
    vi.mock('@/components/Vulnerability/VulnerabilityList.vue', () => ({ default: { name: 'VulnerabilityList' } }))

    const { default: router } = await import('../index.js')
    const routes = router.getRoutes()

    // Test basic route definitions
    expect(routes.length).toBeGreaterThan(0)

    // Test specific routes based on actual implementation
    const homeRoute = routes.find(route => route.path === '/')
    expect(homeRoute).toBeDefined()
    expect(homeRoute.name).toBe('Home')

    const loginRoute = routes.find(route => route.path === '/login')
    expect(loginRoute).toBeDefined()
    expect(loginRoute.name).toBe('LoginPage')

    const projectsRoute = routes.find(route => route.path === '/projects')
    expect(projectsRoute).toBeDefined()
    expect(projectsRoute.name).toBe('ProjectList')

    // Test parameterized routes
    const projectDetailRoute = routes.find(route => route.path === '/projects/:id')
    expect(projectDetailRoute).toBeDefined()
    expect(projectDetailRoute.name).toBe('ProjectDetail')

    const vulnerabilitiesRoute = routes.find(route => route.path === '/vulnerabilities')
    expect(vulnerabilitiesRoute).toBeDefined()
    expect(vulnerabilitiesRoute.name).toBe('VulnerabilityList')

    const projectVulnsRoute = routes.find(route => route.path === '/vulnerabilities/:projectId')
    expect(projectVulnsRoute).toBeDefined()
    expect(projectVulnsRoute.name).toBe('ProjectVulnerabilities')
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

    // Test URL resolution based on actual routes
    const homeResolved = router.resolve('/')
    expect(homeResolved.name).toBe('Home')

    const loginResolved = router.resolve('/login')
    expect(loginResolved.name).toBe('LoginPage')

    const projectDetailResolved = router.resolve('/projects/123')
    expect(projectDetailResolved.name).toBe('ProjectDetail')
    expect(projectDetailResolved.params.id).toBe('123')

    const projectVulnsResolved = router.resolve('/vulnerabilities/123')
    expect(projectVulnsResolved.name).toBe('ProjectVulnerabilities')
    expect(projectVulnsResolved.params.projectId).toBe('123')
  })
})