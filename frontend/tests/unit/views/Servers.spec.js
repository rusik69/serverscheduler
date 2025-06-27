import { shallowMount } from '@vue/test-utils'
import { createStore } from 'vuex'
import { createRouter, createWebHistory } from 'vue-router'
import Servers from '@/views/Servers.vue'
import MockAdapter from 'axios-mock-adapter'
import apiClient from '@/config/api'

// Mock Element Plus components
jest.mock('element-plus', () => ({
  ElMessage: {
    success: jest.fn(),
    error: jest.fn(),
    warning: jest.fn(),
    info: jest.fn()
  },
  ElMessageBox: {
    confirm: jest.fn().mockResolvedValue('confirm')
  }
}))

describe('Servers.vue', () => {
  let wrapper
  let store
  let router
  let mockAxios

  const mockServers = [
    {
      id: 1,
      name: 'Test Server 1',
      ip_address: '192.168.1.1',
      username: 'root',
      password: 'password123',
      status: 'available'
    },
    {
      id: 2,
      name: 'Test Server 2',
      ip_address: '192.168.1.2',
      username: 'admin',
      password: 'admin123',
      status: 'reserved'
    }
  ]

  const createWrapper = (options = {}) => {
    const defaultOptions = {
      global: {
        plugins: [store, router],
        stubs: {
          'el-container': true,
          'el-header': true,
          'el-main': true,
          'el-table': true,
          'el-table-column': true,
          'el-button': true,
          'el-input': true,
          'el-form': true,
          'el-form-item': true,
          'el-select': true,
          'el-option': true,
          'el-dialog': true,
          'el-card': true,
          'el-tag': true,
          'el-tooltip': true,
          'el-space': true,
          'el-row': true,
          'el-col': true,
          'Server': true,
          'ViewIcon': true,
          'Hide': true,
          'CopyDocument': true,
          'Edit': true,
          'Delete': true,
          'Plus': true
        }
      }
    }

    return shallowMount(Servers, {
      ...defaultOptions,
      ...options
    })
  }

  beforeEach(() => {
    // Create mock axios
    mockAxios = new MockAdapter(apiClient)

    // Create store
    store = createStore({
      modules: {
        auth: {
          namespaced: true,
          state: {
            user: { id: 1, username: 'testuser', role: 'root' },
            isAuthenticated: true
          },
          getters: {
            currentUser: (state) => state.user,
            isAuthenticated: (state) => state.isAuthenticated,
            user: (state) => state.user
          }
        }
      }
    })

    // Create router
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/', name: 'Home', component: { template: '<div>Home</div>' } },
        { path: '/servers', name: 'Servers', component: Servers }
      ]
    })

    // Mock API responses
    mockAxios.onGet('/api/servers').reply(200, { servers: mockServers })
    mockAxios.onPost('/api/servers').reply(201, { id: 3, ...mockServers[0] })
    mockAxios.onPut(/\/api\/servers\/\d+/).reply(200, mockServers[0])
    mockAxios.onDelete(/\/api\/servers\/\d+/).reply(200, { message: 'Server deleted' })
  })

  afterEach(() => {
    mockAxios.restore()
    if (wrapper) {
      wrapper.unmount()
    }
  })

  describe('Component Initialization', () => {
    it('should render without errors', () => {
      wrapper = createWrapper()
      expect(wrapper.exists()).toBe(true)
    })

    it('should initialize with empty servers list', () => {
      wrapper = createWrapper()
      expect(wrapper.vm.servers).toEqual([])
    })

    it('should fetch servers on mount', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      
      // Wait for mounted hook to complete
      await new Promise(resolve => setTimeout(resolve, 0))
      expect(wrapper.vm.servers).toEqual(mockServers)
    })

    it('should handle API errors gracefully', async () => {
      mockAxios.onGet('/api/servers').reply(500, { message: 'Server error' })
      
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 0))
      
      expect(wrapper.vm.servers).toEqual([])
    })
  })

  describe('Server Management', () => {
    beforeEach(() => {
      wrapper = createWrapper()
    })

    describe('Create Server', () => {
      it('should create new server successfully', async () => {
        const newServer = {
          name: 'New Server',
          ip_address: '192.168.1.3',
          username: 'root',
          password: 'newpassword',
          status: 'available'
        }

        // Mock form validation
        wrapper.vm.serverFormRef = {
          validate: jest.fn().mockResolvedValue(true)
        }

        Object.assign(wrapper.vm.serverForm, newServer)
        await wrapper.vm.handleServerSubmit()

        expect(mockAxios.history.post).toHaveLength(1)
        expect(JSON.parse(mockAxios.history.post[0].data)).toMatchObject(newServer)
      })

      it('should handle create server error', async () => {
        mockAxios.onPost('/api/servers').reply(400, { error: 'Server already exists' })

        const newServer = {
          name: 'Existing Server',
          ip_address: '192.168.1.1',
          username: 'root',
          password: 'password',
          status: 'available'
        }

        // Mock form validation
        wrapper.vm.serverFormRef = {
          validate: jest.fn().mockResolvedValue(true)
        }

        Object.assign(wrapper.vm.serverForm, newServer)
        await wrapper.vm.handleServerSubmit()

        expect(mockAxios.history.post).toHaveLength(1)
      })
    })

    describe('Edit Server', () => {
      it('should open edit dialog with server data', async () => {
        await wrapper.vm.editServer(mockServers[0])

        expect(wrapper.vm.dialogVisible).toBe(true)
        expect(wrapper.vm.isEditing).toBe(true)
        expect(wrapper.vm.serverForm.name).toBe(mockServers[0].name)
      })

      it('should update server successfully', async () => {
        const updatedServer = { ...mockServers[0], name: 'Updated Server' }
        
        // Mock form validation
        wrapper.vm.serverFormRef = {
          validate: jest.fn().mockResolvedValue(true)
        }
        
        wrapper.vm.isEditing = true
        Object.assign(wrapper.vm.serverForm, updatedServer)
        await wrapper.vm.handleServerSubmit()

        expect(mockAxios.history.put).toHaveLength(1)
        expect(mockAxios.history.put[0].url).toBe(`/api/servers/${updatedServer.id}`)
      })
    })

    describe('Delete Server', () => {
      it('should delete server after confirmation', async () => {
        await wrapper.vm.deleteServer(mockServers[0])

        expect(mockAxios.history.delete).toHaveLength(1)
        expect(mockAxios.history.delete[0].url).toBe(`/api/servers/${mockServers[0].id}`)
      })

      it('should handle delete cancellation', async () => {
        const { ElMessageBox } = require('element-plus')
        ElMessageBox.confirm.mockRejectedValue('cancel')

        await wrapper.vm.deleteServer(mockServers[0])

        expect(mockAxios.history.delete).toHaveLength(0)
      })

      it('should handle delete error', async () => {
        mockAxios.onDelete(`/api/servers/${mockServers[0].id}`).reply(500, { error: 'Cannot delete server' })

        await wrapper.vm.deleteServer(mockServers[0])

        expect(mockAxios.history.delete).toHaveLength(1)
      })
    })
  })

  describe('Password Visibility', () => {
    beforeEach(() => {
      wrapper = createWrapper()
    })

    it('should toggle password visibility for root users', () => {
      expect(wrapper.vm.showPasswords).toBe(false)
      
      wrapper.vm.togglePasswordVisibility()
      
      expect(wrapper.vm.showPasswords).toBe(true)
    })

    it('should copy password to clipboard', async () => {
      // Mock document.execCommand
      document.execCommand = jest.fn().mockReturnValue(true)

      await wrapper.vm.copyToClipboard('password123')

      expect(document.execCommand).toHaveBeenCalledWith('copy')
    })

    it('should handle clipboard copy error', async () => {
      // Mock document.execCommand to throw error
      document.execCommand = jest.fn().mockImplementation(() => {
        throw new Error('Copy failed')
      })

      // Expect the error to be thrown during execution
      expect(() => {
        wrapper.vm.copyToClipboard('password123')
      }).toThrow('Copy failed')
    })
  })

  describe('Form Management', () => {
    beforeEach(() => {
      wrapper = createWrapper()
    })

    it('should show add dialog', () => {
      wrapper.vm.showAddDialog()

      expect(wrapper.vm.dialogVisible).toBe(true)
      expect(wrapper.vm.isEditing).toBe(false)
      expect(wrapper.vm.serverForm.id).toBe(null)
    })

    it('should reset form when showing add dialog', () => {
      wrapper.vm.showAddDialog()

      expect(wrapper.vm.serverForm).toEqual({
        id: null,
        name: '',
        ip_address: '',
        username: '',
        password: '',
        status: 'available'
      })
    })
  })

  describe('Role-based Access', () => {
    it('should show password controls for root users', () => {
      wrapper = createWrapper()
      
      expect(wrapper.vm.isRoot).toBe(true)
    })

    it('should hide password controls for regular users', () => {
      store.state.auth.user.role = 'user'
      wrapper = createWrapper()
      
      expect(wrapper.vm.isRoot).toBe(false)
    })
  })

  describe('Form Validation', () => {
    beforeEach(() => {
      wrapper = createWrapper()
    })

    it('should validate required fields', () => {
      const rules = wrapper.vm.rules
      
      expect(rules.name).toBeDefined()
      expect(rules.ip_address).toBeDefined()
      expect(rules.status).toBeDefined()
    })

    it('should validate IP address format', () => {
      const ipRule = wrapper.vm.rules.ip_address.find(rule => rule.validator)
      const callback = jest.fn()
      
      // Valid IP
      ipRule.validator(null, '192.168.1.1', callback)
      expect(callback).toHaveBeenCalledWith()
      
      callback.mockClear()
      
      // Invalid IP
      ipRule.validator(null, 'invalid-ip', callback)
      expect(callback).toHaveBeenCalledWith(expect.any(Error))
    })
  })
}) 