<template>
  <div class="calendar-container" v-loading="loading" element-loading-text="Loading calendar data..." element-loading-background="rgba(0, 0, 0, 0.8)">
    <!-- Header -->
    <div class="calendar-header">
      <h1 class="calendar-title">
        <el-icon class="calendar-icon"><CalendarIcon /></el-icon>
        Server Availability Calendar
      </h1>
      <p class="calendar-subtitle">View server reservations and availability across time</p>
    </div>

    <!-- Server Filter -->
    <div class="calendar-filters">
      <el-card class="filter-card">
        <div class="filter-content">
          <div class="filter-item">
            <label>Filter by Server:</label>
            <el-select v-model="selectedServerIds" multiple placeholder="All Servers" @change="onServerFilterChange">
              <el-option
                v-for="server in servers"
                :key="server.id"
                :label="server.name"
                :value="server.id"
              />
            </el-select>
          </div>
          <div class="filter-item">
            <label>View:</label>
            <el-radio-group v-model="calendarView" @change="onViewChange">
              <el-radio-button value="month">Month</el-radio-button>
              <el-radio-button value="week">Week</el-radio-button>
              <el-radio-button value="day">Day</el-radio-button>
            </el-radio-group>
          </div>
        </div>
      </el-card>
    </div>

    <!-- Calendar Component -->
    <div class="calendar-wrapper">
      <el-card class="calendar-card">
        <vue-cal
          ref="calendar"
          :time="true"
          :events="calendarEvents"
          :disable-views="['years']"
          :selected-date="selectedDate"
          :active-view="calendarView"
          :cell-focus="true"
          :drag-to-create-event="false"
          :resize-x="false"
          :resize-y="false"
          :editable-events="false"
          :events-count-on-year-view="true"
          :show-all-day-events="true"
          :split-days="[]"
          :min-event-width="0"
          :overlaps-per-time-step="3"
          @event-click="onEventClick"
          @cell-click="onCellClick"
          @view-change="onViewChange"
          class="custom-calendar"
        />
      </el-card>
    </div>

    <!-- Legend -->
    <div class="calendar-legend">
      <el-card class="legend-card">
        <div class="legend-title">Legend</div>
        <div class="legend-items">
          <div class="legend-item">
            <div class="legend-color reserved"></div>
            <span>Reserved</span>
          </div>
          <div class="legend-item">
            <div class="legend-color available"></div>
            <span>Available</span>
          </div>
          <div class="legend-item">
            <div class="legend-color expired"></div>
            <span>Expired</span>
          </div>
          <div class="legend-item">
            <div class="legend-color cancelled"></div>
            <span>Cancelled</span>
          </div>
        </div>
      </el-card>
    </div>

    <!-- Event Details Dialog -->
    <el-dialog
      v-model="eventDialogVisible"
      :title="eventDetails.title"
      width="600px"
      :before-close="closeEventDialog"
    >
      <div v-if="eventDetails.reservation" class="event-details">
        <div class="detail-row">
          <strong>Server:</strong> {{ eventDetails.reservation.server_name }}
        </div>
        <div class="detail-row">
          <strong>User:</strong> {{ eventDetails.reservation.username }}
        </div>
        <div class="detail-row">
          <strong>Start Time:</strong> {{ formatDateTime(eventDetails.reservation.start_time) }}
        </div>
        <div class="detail-row">
          <strong>End Time:</strong> {{ formatDateTime(eventDetails.reservation.end_time) }}
        </div>
        <div class="detail-row">
          <strong>Status:</strong> 
          <el-tag :type="getStatusType(eventDetails.reservation.status)">
            {{ eventDetails.reservation.status }}
          </el-tag>
        </div>
        <div class="detail-row">
          <strong>Duration:</strong> {{ calculateDuration(eventDetails.reservation.start_time, eventDetails.reservation.end_time) }}
        </div>
      </div>
      <template #footer>
        <el-button @click="closeEventDialog">Close</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import VueCal from 'vue-cal'
import 'vue-cal/dist/vuecal.css'
import api from '@/config/api'
import { Calendar as CalendarIcon } from '@element-plus/icons-vue'

export default {
  name: 'Calendar',
  components: {
    VueCal,
    CalendarIcon
  },
  data() {
    return {
      loading: false,
      servers: [],
      reservations: [],
      selectedServerIds: [],
      calendarView: 'month',
      selectedDate: new Date(),
      eventDialogVisible: false,
      eventDetails: {
        title: '',
        reservation: null
      },
      filterDebounceTimer: null
    }
  },
  computed: {
    calendarEvents() {
      console.log('🔄 CalendarEvents computed property called')
      // Ensure reservations is always an array
      const reservations = Array.isArray(this.reservations) ? this.reservations : []
      
      console.log('🔄 Calendar filtering - Total reservations:', reservations.length)
      console.log('🔄 this.reservations value:', this.reservations)
      console.log('🔄 this.reservations type:', typeof this.reservations)
      console.log('🔄 Selected server IDs:', this.selectedServerIds)
      
      if (reservations.length === 0) {
        console.log('❌ No reservations to display')
        return []
      }
      
      // Early return if no servers loaded yet
      if (this.servers.length === 0) {
        console.log('⚠️ No servers loaded yet, showing all reservations')
        return this.mapReservationsToEvents(reservations)
      }
      
      console.log('🔄 Available servers:', this.servers.map(s => ({ id: s.id, name: s.name })))
      
      const filteredReservations = reservations.filter(reservation => {
        // If no servers selected, show all reservations
        if (this.selectedServerIds.length === 0) {
          return true
        }
        
        // Filter by selected servers with type safety
        const reservationServerId = parseInt(reservation.server_id)
        const selectedIds = this.selectedServerIds.map(id => parseInt(id))
        const isIncluded = selectedIds.includes(reservationServerId)
        
        console.log(`🔄 Reservation ${reservation.id}: server_id=${reservation.server_id} (${reservationServerId}), selected=${selectedIds}, included=${isIncluded}`)
        
        return isIncluded
      })
      
      console.log('🔄 Filtered reservations:', filteredReservations.length)
      
      return this.mapReservationsToEvents(filteredReservations)
    }
  },
  async mounted() {
    // Add global error handler for ResizeObserver errors
    const originalError = window.onerror
    window.onerror = (message, source, lineno, colno, error) => {
      // Ignore ResizeObserver errors as they're harmless but noisy
      if (message && message.includes('ResizeObserver loop')) {
        console.debug('Suppressed harmless ResizeObserver error')
        return true // Prevent the error from being logged
      }
      // Call original error handler for other errors
      if (originalError) {
        return originalError(message, source, lineno, colno, error)
      }
      return false
    }
    
    await this.fetchServers()
    await this.fetchReservations()
  },
  beforeUnmount() {
    // Clean up debounce timer
    if (this.filterDebounceTimer) {
      clearTimeout(this.filterDebounceTimer)
    }
  },
  methods: {
    async fetchServers() {
      try {
        console.log('Making API call to /api/servers...')
        const response = await api.get('/api/servers')
        console.log('Servers API response:', response.data)
        console.log('Response status:', response.status)
        console.log('Response data type:', typeof response.data)
        console.log('Response data keys:', Object.keys(response.data || {}))
        
        // Handle different possible response formats
        let servers = []
        if (Array.isArray(response.data)) {
          // Direct array response
          servers = response.data
          console.log('Servers format: direct array')
        } else if (response.data && Array.isArray(response.data.servers)) {
          // Wrapped in servers property
          servers = response.data.servers
          console.log('Servers format: wrapped in servers property')
        } else if (response.data && response.data.data && Array.isArray(response.data.data)) {
          // Double wrapped
          servers = response.data.data
          console.log('Servers format: double wrapped')
        } else {
          console.warn('Unexpected servers response format:', response.data)
          servers = []
        }
        
        this.servers = servers
        console.log('Final loaded servers:', this.servers)
        console.log('Number of servers loaded:', this.servers.length)
        
        if (this.servers.length === 0) {
          console.warn('No servers were loaded - check API response format')
          // Check if user is authenticated
          const token = localStorage.getItem('token')
          if (!token) {
            console.error('No authentication token found!')
            this.$message.error('Please log in to view servers')
            return
          }
        }
      } catch (error) {
        console.error('Error fetching servers:', error)
        console.error('Error response:', error.response?.data)
        console.error('Error status:', error.response?.status)
        
        if (error.response?.status === 401) {
          console.error('Authentication failed - user not logged in or token expired')
          this.$message.error('Authentication required. Please log in.')
        } else if (error.response?.status === 403) {
          console.error('Access forbidden - user does not have permission')
          this.$message.error('Access denied. Insufficient permissions.')
        } else if (error.response?.status === 404) {
          console.error('API endpoint not found')
          this.$message.error('API endpoint not found')
        } else {
          this.$message.error('Failed to load servers')
        }
      }
    },
    async fetchReservations() {
      this.loading = true
      try {
        console.log('Making API call to /api/reservations...')
        console.log('Current auth token exists:', !!this.$store.getters['auth/token'])
        const response = await api.get('/api/reservations')
        console.log('Reservations API response received!')
        console.log('Full response object:', response)
        console.log('Response data:', response.data)
        console.log('Response status:', response.status)
        console.log('Response headers:', response.headers)
        console.log('Response data type:', typeof response.data)
        console.log('Response data is array:', Array.isArray(response.data))
        
        // Handle both possible response formats
        if (Array.isArray(response.data)) {
          this.reservations = response.data
          console.log('✅ Reservations format: direct array')
          console.log('✅ Assigned reservations:', this.reservations)
        } else if (response.data && Array.isArray(response.data.reservations)) {
          this.reservations = response.data.reservations
          console.log('✅ Reservations format: wrapped in reservations property')
          console.log('✅ Assigned reservations:', this.reservations)
        } else {
          this.reservations = []
          console.log('❌ Reservations format: unknown, defaulting to empty array')
          console.log('❌ Response data structure:', JSON.stringify(response.data, null, 2))
        }
        
        console.log('Final reservations count:', this.reservations.length)
        if (this.reservations.length > 0) {
          console.log('Sample reservation:', this.reservations[0])
        } else {
          console.log('❌ No reservations found in response')
        }
      } catch (error) {
        console.error('❌ Error fetching reservations:', error)
        console.error('❌ Error response:', error.response?.data)
        console.error('❌ Error status:', error.response?.status)
        console.error('❌ Error headers:', error.response?.headers)
        
        if (error.response?.status === 401) {
          console.error('❌ Authentication failed for reservations')
          this.$message.error('Authentication required. Please log in.')
        } else if (error.response?.status === 403) {
          console.error('❌ Access forbidden for reservations')
          this.$message.error('Access denied. Insufficient permissions.')
        } else {
          this.$message.error('Failed to load reservations')
        }
        this.reservations = [] // Ensure it's always an array even on error
      } finally {
        this.loading = false
      }
    },
    onEventClick(event) {
      console.log('📅 Event clicked:', event)
      
      // Ensure server name is available for the dialog
      let serverName = event.reservation.server_name || 'Unknown Server'
      if (!serverName || serverName === 'Unknown Server') {
        const server = this.servers.find(s => s.id === parseInt(event.reservation.server_id))
        serverName = server ? server.name : `Server ${event.reservation.server_id}`
      }
      
      this.eventDetails = {
        title: `Reservation Details - ${serverName}`,
        reservation: {
          ...event.reservation,
          server_name: serverName
        }
      }
      this.eventDialogVisible = true
    },
    onCellClick(date) {
      // Could implement "create reservation" functionality here
      console.log('Cell clicked:', date)
    },
    onViewChange(view) {
      this.calendarView = view.view
    },
    closeEventDialog() {
      this.eventDialogVisible = false
      this.eventDetails = {
        title: '',
        reservation: null
      }
    },
    formatDateTime(dateTime) {
      return new Date(dateTime).toLocaleString()
    },
    calculateDuration(start, end) {
      const startTime = new Date(start)
      const endTime = new Date(end)
      const diffMs = endTime - startTime
      const hours = Math.floor(diffMs / (1000 * 60 * 60))
      const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60))
      return `${hours}h ${minutes}m`
    },
    getStatusType(status) {
      switch (status) {
        case 'active': return 'success'
        case 'cancelled': return 'warning'
        case 'expired': return 'danger'
        default: return 'info'
      }
    },
    onServerFilterChange(newServerIds) {
      // Clear existing timer
      if (this.filterDebounceTimer) {
        clearTimeout(this.filterDebounceTimer)
      }
      
      // Debounce the filter change to prevent rapid updates
      this.filterDebounceTimer = setTimeout(() => {
        try {
          console.log('Server filter changed:', newServerIds)
          this.selectedServerIds = newServerIds
          
          // Force Vue to update in next tick to avoid resize observer conflicts
          this.$nextTick(() => {
            // Only update if component is still mounted and calendar ref exists
            try {
              if (this.$refs.calendar && 
                  this.$refs.calendar.$el && 
                  this.$refs.calendar.$el.parentNode && 
                  !this._isBeingDestroyed && 
                  !this._isDestroyed) {
                this.$refs.calendar.$forceUpdate()
              }
            } catch (error) {
              // Silently ignore common harmless errors during component lifecycle
              console.debug('Calendar update skipped due to component state:', error.message)
            }
          })
        } catch (error) {
          console.error('Error updating server filter:', error)
        }
      }, 150) // 150ms debounce
    },
    mapReservationsToEvents(reservations) {
      console.log('📅 Mapping reservations to events:', reservations.length)
      
      const events = reservations.map(reservation => {
        const startTime = new Date(reservation.start_time)
        const endTime = new Date(reservation.end_time)
        
        // Get server name - prioritize server_name from reservation, fallback to lookup by server_id
        let serverName = reservation.server_name || 'Unknown Server'
        if (!reservation.server_name && reservation.server_id) {
          const server = this.servers.find(s => s.id === parseInt(reservation.server_id))
          serverName = server ? server.name : `Server ${reservation.server_id}`
        }
        
        // Get username - prioritize username from reservation, fallback to user_id
        const username = reservation.username || `User ${reservation.user_id}`
        
        console.log(`📅 Mapping reservation ${reservation.id}:`, {
          server_name: reservation.server_name,
          server_id: reservation.server_id,
          resolved_server_name: serverName,
          username: reservation.username,
          user_id: reservation.user_id,
          resolved_username: username,
          start_time: reservation.start_time,
          end_time: reservation.end_time,
          status: reservation.status,
          startTime: startTime,
          endTime: endTime
        })
        
        return {
          id: reservation.id,
          title: `${serverName} - ${username}`,
          start: startTime,
          end: endTime,
          class: `event-${reservation.status}`,
          reservation: {
            ...reservation,
            server_name: serverName,
            username: username
          }
        }
      })
      
      console.log('📅 Generated calendar events:', events.length)
      if (events.length > 0) {
        console.log('📅 Sample event:', events[0])
      }
      
      return events
    }
  }
}
</script>

<style scoped>
.calendar-container {
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  min-height: 100vh;
}

.calendar-header {
  text-align: center;
  margin-bottom: 30px;
  color: white;
}

.calendar-title {
  font-size: 2.5rem;
  font-weight: 600;
  margin-bottom: 10px;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 15px;
}

.calendar-icon {
  color: #74b9ff;
  font-size: 2.2rem;
}

.calendar-subtitle {
  font-size: 1.2rem;
  opacity: 0.9;
  margin: 0;
}

.calendar-filters {
  margin-bottom: 20px;
}

.filter-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: none;
  border-radius: 15px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.filter-content {
  display: flex;
  gap: 30px;
  align-items: center;
  flex-wrap: wrap;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.filter-item label {
  font-weight: 600;
  color: #ffffff;
  white-space: nowrap;
}

.calendar-wrapper {
  margin-bottom: 20px;
}

.calendar-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: none;
  border-radius: 15px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.custom-calendar {
  height: 600px;
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
}

.calendar-legend {
  display: flex;
  justify-content: center;
}

.legend-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border: none;
  border-radius: 15px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  padding: 20px;
}

.legend-title {
  font-size: 1.1rem;
  font-weight: 600;
  margin-bottom: 15px;
  text-align: center;
  color: #ffffff;
}

.legend-items {
  display: flex;
  gap: 20px;
  justify-content: center;
  flex-wrap: wrap;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #ffffff;
}

.legend-color {
  width: 20px;
  height: 20px;
  border-radius: 4px;
  border: 1px solid #ddd;
}

.legend-color.reserved {
  background: #409EFF;
}

.legend-color.available {
  background: #67C23A;
}

.legend-color.expired {
  background: #F56C6C;
}

.legend-color.cancelled {
  background: #E6A23C;
}

.event-details {
  padding: 20px 0;
  line-height: 1.8;
  color: #ffffff;
}

.detail-row {
  margin-bottom: 15px;
  font-size: 16px;
  color: #ffffff;
}

.detail-row strong {
  color: #ffffff;
  min-width: 120px;
  display: inline-block;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

/* Enhanced Calendar Text Contrast */
:deep(.vuecal__title-bar) {
  background: rgba(44, 62, 80, 0.95) !important;
  color: #ffffff !important;
  font-weight: 600;
  border-bottom: 2px solid #34495e;
}

:deep(.vuecal__title) {
  color: #ffffff !important;
  font-weight: 700;
  font-size: 1.3em;
}

:deep(.vuecal__arrow) {
  color: #74b9ff !important;
}

:deep(.vuecal__arrow:hover) {
  background: rgba(116, 185, 255, 0.2) !important;
}

/* Calendar Body */
:deep(.vuecal__body) {
  background: #2c3e50 !important;
}

/* Weekdays Header */
:deep(.vuecal__weekdays-headings) {
  background: #34495e !important;
  border-bottom: 2px solid #4a6583;
}

:deep(.vuecal__heading) {
  color: #ffffff !important;
  font-weight: 700;
  font-size: 14px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Calendar Cells */
:deep(.vuecal__cell) {
  background: #2c3e50 !important;
  border: 1px solid #4a6583 !important;
  color: #ffffff !important;
}

:deep(.vuecal__cell:hover) {
  background: rgba(116, 185, 255, 0.15) !important;
}

:deep(.vuecal__cell.today) {
  background: rgba(116, 185, 255, 0.25) !important;
  border-color: #74b9ff !important;
}

:deep(.vuecal__cell.selected) {
  background: rgba(116, 185, 255, 0.3) !important;
}

/* Cell Content */
:deep(.vuecal__cell-content) {
  color: #ffffff !important;
  font-weight: 600;
}

:deep(.vuecal__cell-date) {
  color: #ffffff !important;
  font-weight: 700;
  font-size: 16px;
}

/* Today's date */
:deep(.vuecal__cell.today .vuecal__cell-date) {
  color: #74b9ff !important;
  font-weight: 800;
  background: rgba(116, 185, 255, 0.2);
  border-radius: 50%;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 2px auto;
}

/* Other month dates */
:deep(.vuecal__cell.out-of-scope) {
  background: #1e2f3d !important;
  color: #7f8c8d !important;
}

:deep(.vuecal__cell.out-of-scope .vuecal__cell-date) {
  color: #7f8c8d !important;
}

/* Week view specific */
:deep(.vuecal__time-column) {
  background: #34495e !important;
  border-right: 2px solid #4a6583 !important;
}

:deep(.vuecal__time-cell) {
  color: #ecf0f1 !important;
  font-weight: 600;
  border-bottom: 1px solid #4a6583 !important;
}

/* Time labels */
:deep(.vuecal__time-cell-label) {
  color: #ecf0f1 !important;
  font-weight: 600;
}

/* All day events row */
:deep(.vuecal__all-day) {
  background: rgba(52, 73, 94, 0.9) !important;
  border-bottom: 2px solid #4a6583 !important;
}

/* Scrollbar styling for calendar */
:deep(.vuecal__bg) {
  background: #2c3e50 !important;
}

/* Ensure no text is invisible */
:deep(.vuecal *) {
  color: inherit;
}

/* Make sure all calendar text has proper contrast */
:deep(.vuecal__cell-events-count) {
  background: #74b9ff !important;
  color: white !important;
  font-weight: 700;
  border-radius: 12px;
  font-size: 11px;
  padding: 2px 6px;
  min-width: 18px;
  text-align: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

/* No event cell styling */
:deep(.vuecal__no-event) {
  color: #bdc3c7 !important;
}

/* Month view specific enhancements */
:deep(.vuecal--month-view .vuecal__cell) {
  min-height: 100px;
}

:deep(.vuecal--month-view .vuecal__cell-date) {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 1;
}

/* Week/Day view enhancements */
:deep(.vuecal--week-view .vuecal__bg),
:deep(.vuecal--day-view .vuecal__bg) {
  background: linear-gradient(90deg, #34495e 0%, #2c3e50 100%) !important;
}

/* View buttons in title bar */
:deep(.vuecal__view-btn) {
  background: rgba(116, 185, 255, 0.2) !important;
  color: #74b9ff !important;
  border: 1px solid #74b9ff !important;
  font-weight: 600;
}

:deep(.vuecal__view-btn.vuecal__view-btn--active) {
  background: #74b9ff !important;
  color: white !important;
}

:deep(.vuecal__view-btn:hover) {
  background: rgba(116, 185, 255, 0.3) !important;
}

/* Additional dark theme elements */
:deep(.vuecal__flex) {
  background: #2c3e50 !important;
}

/* Grid lines */
:deep(.vuecal__time-column .vuecal__time-cell) {
  border-color: #4a6583 !important;
}

/* Week numbers */
:deep(.vuecal__week-number) {
  color: #bdc3c7 !important;
  background: #34495e !important;
}

/* Current time indicator */
:deep(.vuecal__now-line) {
  border-color: #e74c3c !important;
}

/* Event drag placeholder */
:deep(.vuecal__event-drag-placeholder) {
  background: rgba(116, 185, 255, 0.3) !important;
  border: 2px dashed #74b9ff !important;
}

/* Responsive Design */
@media (max-width: 768px) {
  .calendar-container {
    padding: 10px;
  }
  
  .calendar-title {
    font-size: 2rem;
  }
  
  .filter-content {
    flex-direction: column;
    gap: 15px;
  }
  
  .custom-calendar {
    height: 400px;
  }
  
  .legend-items {
    flex-direction: column;
    gap: 10px;
  }
}

/* Calendar Event Styles */
:deep(.vuecal__event.event-active) {
  background: #409EFF;
  border-color: #409EFF;
  color: white;
  font-weight: 600;
  text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.5);
}

:deep(.vuecal__event.event-cancelled) {
  background: #E6A23C;
  border-color: #E6A23C;
  color: white;
  font-weight: 600;
  text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.5);
}

:deep(.vuecal__event.event-expired) {
  background: #F56C6C;
  border-color: #F56C6C;
  color: white;
  font-weight: 600;
  text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.5);
}

:deep(.vuecal__event) {
  cursor: pointer;
  border-radius: 6px;
  font-size: 13px;
  padding: 4px 8px;
  transition: all 0.3s ease;
  color: white !important;
  font-weight: 600;
  text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.5);
}

:deep(.vuecal__event:hover) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

/* Vue-Cal Event Display */
:deep(.vuecal__event-title) {
  color: white !important;
  font-weight: 600;
  text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.5);
}

/* Vue-Cal Today Cell */
:deep(.vuecal__cell.vuecal__cell--today) {
  background: rgba(116, 185, 255, 0.1) !important;
  border: 2px solid #74b9ff !important;
}

/* Vue-Cal Selected Cell */
:deep(.vuecal__cell.vuecal__cell--selected) {
  background: rgba(116, 185, 255, 0.2) !important;
}

/* Vue-Cal Cell Hover */
:deep(.vuecal__cell:hover) {
  background: rgba(255, 255, 255, 0.05) !important;
}
</style> 