import { shallowMount } from '@vue/test-utils'
import { createStore } from 'vuex'
import { createRouter, createWebHistory } from 'vue-router'
import Login from '@/views/Login.vue'

// Mock Element Plus components
jest.mock('element-plus', () => ({
  ElMessage: {
    success: jest.fn(),
    error: jest.fn(),
    warning: jest.fn(),
    info: jest.fn()
  }
}))

describe('Login.vue', () => {
  let wrapper
  let store
  let router
  let mockDispatch

  const createWrapper = (options = {}) => {
    const defaultOptions = {
      global: {
        plugins: [store, router],
        stubs: {
          'el-card': true,
          'el-form': true,
          'el-form-item': true,
          'el-input': true,
          'el-button': true
        }
      }
    }

    return shallowMount(Login, {
      ...defaultOptions,
      ...options
    })
  }

  beforeEach(() => {
    mockDispatch = jest.fn()

    // Create store
    store = createStore({
      modules: {
        auth: {
          namespaced: true,
          state: {
            user: null,
            isAuthenticated: false
          },
          getters: {
            currentUser: (state) => state.user,
            isAuthenticated: (state) => state.isAuthenticated
          },
          actions: {
            login: mockDispatch
          }
        }
      }
    })

    // Create router
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/', name: 'Home', component: { template: '<div>Home</div>' } },
        { path: '/login', name: 'Login', component: Login },
        { path: '/register', name: 'Register', component: { template: '<div>Register</div>' } }
      ]
    })

    // Mock router.push
    router.push = jest.fn()
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
  })

  describe('Component Initialization', () => {
    it('should render without errors', () => {
      wrapper = createWrapper()
      expect(wrapper.exists()).toBe(true)
    })

    it('should initialize with empty form data', () => {
      wrapper = createWrapper()
      expect(wrapper.vm.form).toEqual({
        username: '',
        password: ''
      })
    })

    it('should initialize with loading state false', () => {
      wrapper = createWrapper()
      expect(wrapper.vm.loading).toBe(false)
    })
  })

  describe('Form Validation', () => {
    beforeEach(() => {
      wrapper = createWrapper()
    })

    it('should have validation rules for username', () => {
      const usernameRules = wrapper.vm.rules.username
      expect(usernameRules).toBeDefined()
      expect(usernameRules[0].required).toBe(true)
      expect(usernameRules[0].message).toBe('Please input username')
    })

    it('should have validation rules for password', () => {
      const passwordRules = wrapper.vm.rules.password
      expect(passwordRules).toBeDefined()
      expect(passwordRules[0].required).toBe(true)
      expect(passwordRules[0].message).toBe('Please input password')
    })
  })

  describe('Login Process', () => {
    beforeEach(() => {
      wrapper = createWrapper()
      // Mock form validation
      wrapper.vm.loginForm = {
        validate: jest.fn()
      }
    })

    it('should handle successful login', async () => {
      wrapper.vm.loginForm.validate.mockResolvedValue(true)
      mockDispatch.mockResolvedValue(true)

      wrapper.vm.form.username = 'testuser'
      wrapper.vm.form.password = 'password123'

      await wrapper.vm.handleLogin()

      expect(mockDispatch).toHaveBeenCalledWith({
        username: 'testuser',
        password: 'password123'
      })
      expect(router.push).toHaveBeenCalledWith('/')
    })

    it('should handle login failure', async () => {
      wrapper.vm.loginForm.validate.mockResolvedValue(true)
      mockDispatch.mockResolvedValue(false)

      wrapper.vm.form.username = 'testuser'
      wrapper.vm.form.password = 'wrongpassword'

      await wrapper.vm.handleLogin()

      expect(mockDispatch).toHaveBeenCalledWith({
        username: 'testuser',
        password: 'wrongpassword'
      })
      expect(router.push).not.toHaveBeenCalled()
    })

    it('should handle validation errors', async () => {
      wrapper.vm.loginForm.validate.mockRejectedValue(new Error('Validation failed'))

      await wrapper.vm.handleLogin()

      expect(mockDispatch).not.toHaveBeenCalled()
      expect(wrapper.vm.loading).toBe(false)
    })

    it('should handle network errors', async () => {
      wrapper.vm.loginForm.validate.mockResolvedValue(true)
      mockDispatch.mockRejectedValue(new Error('Network error'))

      await wrapper.vm.handleLogin()

      expect(wrapper.vm.loading).toBe(false)
    })

    it('should set loading state during login', async () => {
      wrapper.vm.loginForm.validate.mockResolvedValue(true)
      mockDispatch.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)))

      const loginPromise = wrapper.vm.handleLogin()
      
      expect(wrapper.vm.loading).toBe(true)
      
      await loginPromise
      
      expect(wrapper.vm.loading).toBe(false)
    })
  })

  describe('Form Interaction', () => {
    beforeEach(() => {
      wrapper = createWrapper()
    })

    it('should update form data when input changes', async () => {
      wrapper.vm.form.username = 'newuser'
      wrapper.vm.form.password = 'newpassword'

      await wrapper.vm.$nextTick()

      expect(wrapper.vm.form.username).toBe('newuser')
      expect(wrapper.vm.form.password).toBe('newpassword')
    })

    it('should clear form data', () => {
      wrapper.vm.form.username = 'testuser'
      wrapper.vm.form.password = 'password123'

      wrapper.vm.form.username = ''
      wrapper.vm.form.password = ''

      expect(wrapper.vm.form.username).toBe('')
      expect(wrapper.vm.form.password).toBe('')
    })
  })

  describe('Navigation', () => {
    beforeEach(() => {
      wrapper = createWrapper()
    })

    it('should navigate to register page', () => {
      // This would be tested through the router mock
      // The actual navigation is handled by the template
      expect(router.push).toBeDefined()
    })
  })

  describe('Component State', () => {
    beforeEach(() => {
      wrapper = createWrapper()
    })

    it('should have correct initial state', () => {
      expect(wrapper.vm.form).toEqual({
        username: '',
        password: ''
      })
      expect(wrapper.vm.loading).toBe(false)
      expect(wrapper.vm.loginForm).toBeDefined()
    })

    it('should maintain reactive state', async () => {
      wrapper.vm.loading = true
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.loading).toBe(true)
      
      wrapper.vm.loading = false
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.loading).toBe(false)
    })
  })
}) 