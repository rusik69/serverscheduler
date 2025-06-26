// Jest setup file for global configuration and mocks
import { config } from '@vue/test-utils'

// Configure Vue Test Utils global stubs
config.global.stubs = {
  // Element Plus components
  'el-container': true,
  'el-header': true,
  'el-main': true,
  'el-aside': true,
  'el-menu': true,
  'el-menu-item': true,
  'el-submenu': true,
  'el-table': true,
  'el-table-column': true,
  'el-button': true,
  'el-input': true,
  'el-form': true,
  'el-form-item': true,
  'el-select': true,
  'el-option': true,
  'el-date-picker': true,
  'el-time-picker': true,
  'el-dialog': true,
  'el-card': true,
  'el-tag': true,
  'el-dropdown': true,
  'el-dropdown-menu': true,
  'el-dropdown-item': true,
  'el-divider': true,
  'el-space': true,
  'el-scrollbar': true,
  'el-row': true,
  'el-col': true,
  'el-tooltip': true,
  'el-popconfirm': true,
  'el-loading': true,
  // Icons
  'el-icon': true,
  'Calendar': true,
  'CalendarIcon': true,
  'User': true,
  'Lock': true,
  'ViewIcon': true,
  'Hide': true,
  'CopyDocument': true,
  'Edit': true,
  'Delete': true,
  'Plus': true,
  'Search': true,
  'Refresh': true,
  'Setting': true,
  'Check': true,
  'Close': true,
  'Warning': true,
  'InfoFilled': true,
  'CircleClose': true,
  'Remove': true,
  // Vue Cal
  'vue-cal': true,
  // Router
  'router-link': true,
  'router-view': true
}

// Configure Vue Test Utils
config.global.mocks = {
  $t: (msg) => msg,
  $tc: (msg) => msg,
  $te: (msg) => msg,
  $d: (msg) => msg,
  $n: (msg) => msg
}

// Mock Element Plus globally
global.ElMessage = {
  success: jest.fn(),
  error: jest.fn(),
  warning: jest.fn(),
  info: jest.fn()
}

global.ElMessageBox = {
  confirm: jest.fn().mockResolvedValue('confirm'),
  alert: jest.fn().mockResolvedValue('confirm'),
  prompt: jest.fn().mockResolvedValue({ value: 'test' })
}

global.ElLoading = {
  service: jest.fn().mockReturnValue({
    close: jest.fn()
  })
}

// Mock Vue Router
const mockRouter = {
  push: jest.fn(),
  replace: jest.fn(),
  go: jest.fn(),
  back: jest.fn(),
  forward: jest.fn(),
  currentRoute: {
    value: {
      path: '/',
      name: 'Home',
      params: {},
      query: {}
    }
  }
}

global.$router = mockRouter

// Mock Vue Store
const mockStore = {
  dispatch: jest.fn(),
  commit: jest.fn(),
  getters: {
    isAuthenticated: false,
    currentUser: null,
    userRole: 'user'
  },
  state: {
    auth: {
      isAuthenticated: false,
      user: null
    }
  }
}

global.$store = mockStore

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn()
}

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock
})

// Mock console methods to reduce noise in tests
global.console = {
  ...console,
  log: jest.fn(),
  debug: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  error: jest.fn()
}

// Mock ResizeObserver
global.ResizeObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn()
}))

// Mock IntersectionObserver
global.IntersectionObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn()
}))

// Mock clipboard API
Object.assign(navigator, {
  clipboard: {
    writeText: jest.fn().mockResolvedValue()
  }
})

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(), // deprecated
    removeListener: jest.fn(), // deprecated
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
})

// Mock requestAnimationFrame
global.requestAnimationFrame = jest.fn(cb => setTimeout(cb, 0))
global.cancelAnimationFrame = jest.fn()

// Reset all mocks before each test
beforeEach(() => {
  jest.clearAllMocks()
  localStorageMock.getItem.mockClear()
  localStorageMock.setItem.mockClear()
  localStorageMock.removeItem.mockClear()
  localStorageMock.clear.mockClear()
  
  // Reset store state
  mockStore.getters.isAuthenticated = false
  mockStore.getters.currentUser = null
  mockStore.getters.userRole = 'user'
  mockStore.state.auth.isAuthenticated = false
  mockStore.state.auth.user = null
}) 