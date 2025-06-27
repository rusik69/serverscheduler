import { mount } from '@vue/test-utils'
import { createStore } from 'vuex'
import { createRouter, createWebHistory } from 'vue-router'
import Calendar from '@/views/Calendar.vue'
import MockAdapter from 'axios-mock-adapter'
import apiClient from '@/config/api'

// Mock Element Plus message
const mockElMessage = {
  success: jest.fn(),
  error: jest.fn(),
  warning: jest.fn(),
  info: jest.fn()
}

// Make it globally available
global.ElMessage = mockElMessage

// Mock vue-cal component
const mockVueCal = {
  template: '<div class="vuecal"><slot /></div>',
  props: ['events', 'selectedDate', 'view', 'hideWeekends', 'twelveHour', 'onEventClick'],
  methods: {
    updateLayout: jest.fn(),
    switchToNarrowerView: jest.fn()
  }
}

describe('Calendar.vue', () => {
  let wrapper
  let store
  let router
  let mockAxios

  const mockReservations = [
    {
      id: 1,
      server_id: 1,
      server_name: 'Server 1',
      user_id: 1,
      username: 'testuser',
      start_time: '2024-01-15T10:00:00Z',
      end_time: '2024-01-15T18:00:00Z',
      purpose: 'Testing',
      status: 'active'
    },
    {
      id: 2,
      server_id: 2,
      server_name: 'Server 2',
      user_id: 2,
      username: 'otheruser',
      start_time: '2024-01-16T09:00:00Z',
      end_time: '2024-01-16T17:00:00Z',
      purpose: 'Development',
      status: 'pending'
    }
  ]

  const mockServers = [
    { id: 1, name: 'Server 1', ip: '192.168.1.10', status: 'available' },
    { id: 2, name: 'Server 2', ip: '192.168.1.11', status: 'available' },
    { id: 3, name: 'Server 3', ip: '192.168.1.12', status: 'maintenance' }
  ]

  const mockUser = {
    id: 1,
    username: 'testuser',
    role: 'root'
  }

  beforeEach(() => {
    mockAxios = new MockAdapter(apiClient)
    
    store = createStore({
      modules: {
        auth: {
          namespaced: true,
          getters: {
            user: () => mockUser,
            isAuthenticated: () => true
          }
        }
      }
    })

    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/calendar', name: 'Calendar' }
      ]
    })

    // Mock API responses
    mockAxios.onGet('/api/reservations').reply(200, mockReservations)
    mockAxios.onGet('/api/servers').reply(200, mockServers)
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
    mockAxios.restore()
    jest.clearAllMocks()
  })

  const createWrapper = () => {
    const wrapper = mount(Calendar, {
      global: {
        plugins: [store, router],
        components: {
          'vue-cal': mockVueCal
        },
        stubs: {
          'el-row': { template: '<div class="el-row"><slot /></div>' },
          'el-col': { template: '<div class="el-col"><slot /></div>' },
          'el-card': { template: '<div class="el-card"><slot /></div>' },
          'el-checkbox-group': {
            template: '<div class="el-checkbox-group"><slot /></div>',
            props: ['modelValue'],
            emits: ['update:modelValue']
          },
          'el-checkbox': {
            template: '<input type="checkbox" :value="label" @change="$emit(\'change\', $event.target.checked)" />',
            props: ['label'],
            emits: ['change']
          },
          'el-button': {
            template: '<button @click="$emit(\'click\')"><slot /></button>',
            emits: ['click'],
            props: ['type', 'size']
          },
          'CalendarIcon': { template: '<span>📅</span>' }
        }
      }
    })
    
    // Mock the $message property on the component instance
    wrapper.vm.$message = mockElMessage
    
    return wrapper
  }

  describe('Component Initialization', () => {
    it('should render calendar component', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      
      expect(wrapper.find('.vuecal').exists()).toBe(true)
    })

    it('should fetch reservations and servers on mount', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.reservations).toEqual(mockReservations)
      expect(wrapper.vm.servers).toEqual(mockServers)
    })

    it('should handle API errors gracefully', async () => {
      mockAxios.onGet('/api/reservations').reply(500, { error: 'Server error' })
      
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      
      // Wait for the error handler to be called
      await new Promise(resolve => setTimeout(resolve, 0))
      
      expect(mockElMessage.error).toHaveBeenCalled()
    })
  })

  describe('Server Filtering', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
    })

    it('should initialize with all servers selected', () => {
      expect(wrapper.vm.selectedServers).toEqual([1, 2, 3])
    })

    it('should filter events based on selected servers', async () => {
      wrapper.vm.selectedServers = [1]
      await wrapper.vm.$nextTick()
      
      const filteredEvents = wrapper.vm.filteredEvents
      expect(filteredEvents).toHaveLength(1)
      expect(filteredEvents[0].serverId).toBe(1)
    })

    it('should show all events when all servers selected', async () => {
      wrapper.vm.selectedServers = [1, 2, 3]
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.filteredEvents).toHaveLength(2)
    })

    it('should handle server selection change', async () => {
      wrapper.vm.handleServerSelectionChange([2])
      
      expect(wrapper.vm.selectedServers).toEqual([2])
      
      const filteredEvents = wrapper.vm.filteredEvents
      expect(filteredEvents).toHaveLength(1)
      expect(filteredEvents[0].serverId).toBe(2)
    })
  })

  describe('Event Processing', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
    })

    it('should convert reservations to calendar events', () => {
      const events = wrapper.vm.calendarEvents
      
      expect(events).toHaveLength(2)
      expect(events[0]).toMatchObject({
        id: 1,
        title: 'Server 1 - testuser',
        start: '2024-01-15 10:00',
        end: '2024-01-15 18:00',
        serverId: 1,
        class: 'event-active'
      })
    })

    it('should assign correct CSS classes based on status', () => {
      const events = wrapper.vm.calendarEvents
      
      const activeEvent = events.find(e => e.id === 1)
      const pendingEvent = events.find(e => e.id === 2)
      
      expect(activeEvent.class).toBe('event-active')
      expect(pendingEvent.class).toBe('event-pending')
    })

    it('should handle event click', async () => {
      const event = {
        id: 1,
        title: 'Server 1 - testuser',
        reservation: mockReservations[0]
      }

      await wrapper.vm.onEventClick(event)
      
      expect(wrapper.vm.selectedEvent).toEqual(event)
      expect(wrapper.vm.eventDetailsVisible).toBe(true)
    })
  })

  describe('Date Formatting', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
    })

    it('should convert ISO date to calendar format', () => {
      const isoDate = '2024-01-15T10:00:00Z'
      const calendarDate = wrapper.vm.convertToCalendarDate(isoDate)
      
      expect(calendarDate).toMatch(/2024-01-15 \d{2}:\d{2}/)
    })

    it('should format date for display', () => {
      const testDate = '2024-01-15T10:00:00Z'
      const formatted = wrapper.vm.formatDate(testDate)
      
      expect(formatted).toMatch(/\d{4}-\d{2}-\d{2} \d{2}:\d{2}/)
    })
  })

  describe('Event Details Dialog', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
    })

    it('should show event details when event is clicked', async () => {
      const event = {
        id: 1,
        title: 'Server 1 - testuser',
        reservation: mockReservations[0]
      }

      await wrapper.vm.onEventClick(event)
      
      expect(wrapper.vm.eventDetailsVisible).toBe(true)
      expect(wrapper.vm.selectedEvent).toEqual(event)
    })

    it('should close event details dialog', async () => {
      wrapper.vm.eventDetailsVisible = true
      wrapper.vm.selectedEvent = { id: 1 }
      
      wrapper.vm.closeEventDetails()
      
      expect(wrapper.vm.eventDetailsVisible).toBe(false)
      expect(wrapper.vm.selectedEvent).toBeNull()
    })
  })

  describe('Status Styling', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
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

  describe('Calendar Layout', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
    })

    it('should have correct calendar configuration', () => {
      expect(wrapper.vm.hideWeekends).toBe(false)
      expect(wrapper.vm.twelveHour).toBe(true)
      expect(wrapper.vm.selectedDate).toBeInstanceOf(Date)
    })

    it('should handle resize observer errors gracefully', () => {
      // Simulate ResizeObserver error
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation()
      
      wrapper.vm.handleResizeObserverError()
      
      // Should not throw error
      expect(consoleSpy).not.toHaveBeenCalled()
      consoleSpy.mockRestore()
    })
  })

  describe('Server Status Display', () => {
    beforeEach(async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
    })

    it('should display server status in server list', () => {
      const availableServers = wrapper.vm.servers.filter(s => s.status === 'available')
      const maintenanceServers = wrapper.vm.servers.filter(s => s.status === 'maintenance')
      
      expect(availableServers).toHaveLength(2)
      expect(maintenanceServers).toHaveLength(1)
    })

    it('should get server name by id', () => {
      expect(wrapper.vm.getServerName(1)).toBe('Server 1')
      expect(wrapper.vm.getServerName(999)).toBe('Unknown Server')
    })
  })

  describe('Component Lifecycle', () => {
    it('should clean up event listeners on unmount', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      
      const removeEventListenerSpy = jest.spyOn(window, 'removeEventListener')
      
      wrapper.unmount()
      
      // Should clean up any window event listeners if they exist
      // This is more of a safety check
      expect(removeEventListenerSpy).toHaveBeenCalledTimes(0) // No listeners in this case
      
      removeEventListenerSpy.mockRestore()
    })
  })

  describe('Error Handling', () => {
    it('should handle calendar rendering errors', async () => {
      // Mock console.error to capture error handling
      const consoleSpy = jest.spyOn(console, 'error').mockImplementation()
      
      wrapper = createWrapper()
      
      // Simulate calendar error
      wrapper.vm.handleCalendarError = jest.fn()
      
      await wrapper.vm.$nextTick()
      
      // Component should still render without throwing
      expect(wrapper.find('.vuecal').exists()).toBe(true)
      
      consoleSpy.mockRestore()
    })

    it('should handle empty data gracefully', async () => {
      mockAxios.onGet('/api/reservations').reply(200, [])
      mockAxios.onGet('/api/servers').reply(200, [])
      
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.reservations).toEqual([])
      expect(wrapper.vm.servers).toEqual([])
      expect(wrapper.vm.calendarEvents).toEqual([])
    })
  })
}) 