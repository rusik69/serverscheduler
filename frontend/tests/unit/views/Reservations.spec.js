import { shallowMount } from '@vue/test-utils'
import { createStore } from 'vuex'
import { createRouter, createWebHistory } from 'vue-router'
import Reservations from '@/views/Reservations.vue'
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

describe('Reservations.vue', () => {
  let wrapper
  let store
  let router
  let mockAxios

  // Mock data
  const mockReservations = [
    {
      id: 1,
      server_id: 1,
      server_name: 'Test Server 1',
      user_id: 1,
      username: 'testuser',
      start_time: '2024-01-01T10:00:00Z',
      end_time: '2024-01-01T12:00:00Z',
      status: 'active',
      server_ip: '192.168.1.1',
      server_username: 'root',
      server_password: 'password123'
    },
    {
      id: 2,
      server_id: 2,
      server_name: 'Test Server 2',
      user_id: 2,
      username: 'testuser2',
      start_time: '2024-01-02T14:00:00Z',
      end_time: '2024-01-02T16:00:00Z',
      status: 'pending',
      server_ip: '',
      server_username: '',
      server_password: ''
    }
  ]

  const mockServers = [
    {
      id: 1,
      name: 'Test Server 1',
      ip: '192.168.1.1',
      username: 'root',
      password: 'password123',
      status: 'available'
    },
    {
      id: 2,
      name: 'Test Server 2',
      ip: '192.168.1.2',
      username: 'admin',
      password: 'admin123',
      status: 'available'
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
          'el-date-picker': true,
          'el-time-picker': true,
          'el-dialog': true,
          'el-card': true,
          'el-tag': true,
          'el-tooltip': true,
          'el-popconfirm': true,
          'el-space': true,
          'el-row': true,
          'el-col': true,
          'Calendar': true,
          'ViewIcon': true,
          'Hide': true,
          'CopyDocument': true,
          'Edit': true,
          'Delete': true,
          'Plus': true,
          'Remove': true
        }
      }
    }

    return shallowMount(Reservations, {
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
            user: { id: 1, username: 'testuser', role: 'user' },
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
        { path: '/reservations', name: 'Reservations', component: Reservations }
      ]
    })

    // Mock API responses
    mockAxios.onGet('/api/reservations').reply(200, mockReservations)
    mockAxios.onGet('/api/servers').reply(200, { servers: mockServers })
    mockAxios.onPost('/api/reservations').reply(201, { id: 3, ...mockReservations[0] })
    mockAxios.onPut(/\/api\/reservations\/\d+/).reply(200, mockReservations[0])
    mockAxios.onDelete(/\/api\/reservations\/\d+/).reply(200, { message: 'Reservation cancelled' })
    mockAxios.onDelete(/\/api\/reservations\/\d+\/delete/).reply(200, { message: 'Reservation deleted' })
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

    it('should initialize with empty reservations list', () => {
      wrapper = createWrapper()
      expect(wrapper.vm.reservations).toEqual([])
    })

    it('should initialize with empty servers list', () => {
      wrapper = createWrapper()
      expect(wrapper.vm.servers).toEqual([])
    })

    it('should initialize with default form data', () => {
      wrapper = createWrapper()
      expect(wrapper.vm.reservationForm).toEqual({
        id: null,
        server_id: null,
        start_time: '',
        end_time: ''
      })
    })
  })

  describe('Data Loading', () => {
    it('should load reservations on mount', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      
      // Wait for mounted hook to complete
      await new Promise(resolve => setTimeout(resolve, 0))
      expect(wrapper.vm.reservations).toEqual(mockReservations)
    })

    it('should load servers on mount', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      
      // Wait for mounted hook to complete
      await new Promise(resolve => setTimeout(resolve, 0))
      expect(wrapper.vm.servers).toEqual(mockServers)
    })

    it('should handle API errors gracefully', async () => {
      mockAxios.onGet('/api/reservations').reply(500, { message: 'Server error' })
      
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 0))
      
      expect(wrapper.vm.reservations).toEqual([])
    })
  })

  describe('CRUD Operations', () => {
    beforeEach(() => {
      wrapper = createWrapper()
    })

    it('should create a new reservation', async () => {
      const newReservation = {
        server_id: 1,
        start_time: '2024-01-03T10:00:00Z',
        end_time: '2024-01-03T12:00:00Z'
      }

      Object.assign(wrapper.vm.reservationForm, newReservation)
      await wrapper.vm.handleReservationSubmit()

      expect(mockAxios.history.post).toHaveLength(1)
      expect(JSON.parse(mockAxios.history.post[0].data)).toMatchObject(newReservation)
    })

    it('should update an existing reservation', async () => {
      const updatedReservation = { ...mockReservations[0], start_time: '2024-01-01T11:00:00Z' }
      
      wrapper.vm.isEditing = true
      Object.assign(wrapper.vm.reservationForm, updatedReservation)
      await wrapper.vm.handleReservationSubmit()

      expect(mockAxios.history.put).toHaveLength(1)
      expect(mockAxios.history.put[0].url).toBe(`/api/reservations/${updatedReservation.id}`)
    })

    it('should cancel a reservation', async () => {
      await wrapper.vm.cancelReservation(mockReservations[0])

      expect(mockAxios.history.delete).toHaveLength(1)
      expect(mockAxios.history.delete[0].url).toBe(`/api/reservations/${mockReservations[0].id}`)
    })

    it('should delete a reservation permanently', async () => {
      await wrapper.vm.deleteReservation(mockReservations[0])

      expect(mockAxios.history.delete).toHaveLength(1)
      expect(mockAxios.history.delete[0].url).toBe(`/api/reservations/${mockReservations[0].id}/delete`)
    })
  })

  describe('Status Display', () => {
    beforeEach(() => {
      wrapper = createWrapper()
    })

    it('should return correct status type for active reservations', () => {
      expect(wrapper.vm.getStatusType('active')).toBe('success')
    })

    it('should return correct status type for pending reservations', () => {
      expect(wrapper.vm.getStatusType('pending')).toBe('warning')
    })

    it('should return correct status type for cancelled reservations', () => {
      expect(wrapper.vm.getStatusType('cancelled')).toBe('')
    })

    it('should return default status type for unknown status', () => {
      expect(wrapper.vm.getStatusType('unknown')).toBe('info')
    })
  })

  describe('Date Formatting', () => {
    beforeEach(() => {
      wrapper = createWrapper()
    })

    it('should format date correctly', () => {
      const date = '2024-01-01T10:00:00Z'
      const formatted = wrapper.vm.formatDate(date)
      expect(formatted).toMatch(/\d{1,2}\/\d{1,2}\/\d{4}/)
    })

    it('should handle invalid date', () => {
      const formatted = wrapper.vm.formatDate('invalid-date')
      expect(formatted).toBe('Invalid Date')
    })
  })

  describe('Role-based Access', () => {
    it('should show credentials for root users', () => {
      store.state.auth.user.role = 'root'
      wrapper = createWrapper()
      
      expect(wrapper.vm.isRoot).toBe(true)
    })

    it('should hide credentials for regular users', () => {
      wrapper = createWrapper()
      
      expect(wrapper.vm.isRoot).toBe(false)
    })
  })

  describe('Server Selection', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      wrapper.vm.servers = mockServers
    })

    it('should get server name by id', () => {
      const serverName = wrapper.vm.getServerName(1)
      expect(serverName).toBe('Test Server 1')
    })

    it('should return unknown for invalid server id', () => {
      const serverName = wrapper.vm.getServerName(999)
      expect(serverName).toBe('Unknown')
    })
  })

  describe('Utility Functions', () => {
    beforeEach(() => {
      wrapper = createWrapper()
    })

    it('should copy text to clipboard', async () => {
      // Mock document.execCommand
      document.execCommand = jest.fn().mockReturnValue(true)
      
      const text = 'test-password'
      await wrapper.vm.copyToClipboard(text)
      
      expect(document.execCommand).toHaveBeenCalledWith('copy')
    })

    it('should show add dialog', () => {
      wrapper.vm.showAddDialog()

      expect(wrapper.vm.dialogVisible).toBe(true)
      expect(wrapper.vm.isEditing).toBe(false)
      expect(wrapper.vm.reservationForm.id).toBe(null)
    })

    it('should show edit dialog', () => {
      const reservation = mockReservations[0]
      wrapper.vm.editReservation(reservation)

      expect(wrapper.vm.dialogVisible).toBe(true)
      expect(wrapper.vm.isEditing).toBe(true)
      expect(wrapper.vm.reservationForm.id).toBe(reservation.id)
    })
  })
}) 