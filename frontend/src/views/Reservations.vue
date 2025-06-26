<template>
  <div class="reservations-container">
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <h2>
                <el-icon class="header-icon"><Calendar /></el-icon>
                Reservation Management
              </h2>
              <el-button type="primary" @click="showAddDialog" size="large" class="add-btn">
                <el-icon><Plus /></el-icon>
                New Reservation
              </el-button>
            </div>
          </template>
          
          <div class="table-container">
            <el-table :data="reservations" v-loading="loading" class="modern-table">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="server_name" label="Server" width="140">
              <template #default="{ row }">
                <div class="server-info">
                  <el-icon class="server-icon"><Monitor /></el-icon>
                  <span>{{ row.server_name }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="username" label="User" width="120">
              <template #default="{ row }">
                <div class="user-info">
                  <el-icon class="user-icon"><User /></el-icon>
                  <span>{{ row.username }}</span>
                </div>
              </template>
            </el-table-column>
            <!-- Server Credentials - Visible to all authenticated users -->
            <el-table-column label="Access" width="180">
              <template #default="{ row }">
                <div v-if="row.server_username || row.server_password || row.server_ip" class="credentials-compact">
                  <div class="cred-row">
                    <el-icon class="cred-icon"><Monitor /></el-icon>
                    <span class="cred-text">{{ row.server_ip || 'N/A' }}</span>
                  </div>
                  <div class="cred-row">
                    <el-icon class="cred-icon"><User /></el-icon>
                    <span class="cred-text">{{ row.server_username || 'N/A' }}</span>
                  </div>
                  <div v-if="row.server_password" class="cred-row">
                    <el-icon class="cred-icon"><Lock /></el-icon>
                    <el-button 
                      size="small" 
                      text 
                      @click="copyToClipboard(row.server_password)"
                      class="password-btn"
                    >
                      <el-icon><CopyDocument /></el-icon>
                      Copy
                    </el-button>
                  </div>
                </div>
                <div v-else class="no-credentials-compact">
                  <el-text type="info" size="small">No access</el-text>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="start_time" label="Start Time" width="160">
              <template #default="{ row }">
                <div class="time-info">
                  <el-icon class="time-icon"><Clock /></el-icon>
                  <span>{{ formatDate(row.start_time) }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="end_time" label="End Time" width="160">
              <template #default="{ row }">
                <div class="time-info">
                  <el-icon class="time-icon"><Clock /></el-icon>
                  <span>{{ formatDate(row.end_time) }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="Status" width="100">
              <template #default="{ row }">
                <el-tag 
                  :type="getStatusType(row.status)"
                  class="status-tag"
                  effect="dark"
                >
                  <el-icon>
                    <CircleCheck v-if="row.status === 'active'" />
                    <WarningFilled v-else-if="row.status === 'pending'" />
                    <Remove v-else-if="row.status === 'cancelled'" />
                    <CircleClose v-else />
                  </el-icon>
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Actions" :width="isRoot ? 240 : 200" fixed="right">
              <template #default="{ row }">
                <div class="action-buttons">
                  <el-button 
                    size="small" 
                    type="primary" 
                    @click="editReservation(row)"
                    :disabled="row.status === 'cancelled'"
                    class="action-btn compact-btn"
                  >
                    <el-icon><Edit /></el-icon>
                  </el-button>
                  <el-button 
                    size="small" 
                    type="warning" 
                    @click="cancelReservation(row)"
                    :disabled="row.status === 'cancelled'"
                    class="action-btn compact-btn"
                  >
                    <el-icon><CircleClose /></el-icon>
                  </el-button>
                  <el-button 
                    v-if="isRoot"
                    size="small" 
                    type="danger" 
                    @click="deleteReservation(row)"
                    class="action-btn compact-btn"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Add/Edit Reservation Dialog -->
    <el-dialog
      :title="dialogTitle"
      v-model="dialogVisible"
      width="500px"
    >
      <el-form :model="reservationForm" :rules="rules" ref="reservationFormRef" label-width="100px">
        <el-form-item label="Server" prop="server_id">
          <el-select v-model="reservationForm.server_id" placeholder="Select server">
            <el-option
              v-for="server in servers"
              :key="server.id"
              :label="server.name"
              :value="server.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="Start Time" prop="start_time">
          <el-date-picker
            v-model="reservationForm.start_time"
            type="datetime"
            placeholder="Select start time"
            :disabled-date="disabledDate"
            :disabled-hours="disabledHours"
          />
        </el-form-item>
        <el-form-item label="End Time" prop="end_time">
          <el-date-picker
            v-model="reservationForm.end_time"
            type="datetime"
            placeholder="Select end time"
            :disabled-date="disabledDate"
            :disabled-hours="disabledHours"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">Cancel</el-button>
          <el-button type="primary" @click="handleReservationSubmit" :loading="submitting">
            {{ isEditing ? 'Update' : 'Create' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useStore } from 'vuex'
import { 
  Calendar, 
  Plus, 
  Monitor, 
  User, 
  Clock, 
  CircleCheck, 
  WarningFilled, 
  CircleClose,
  Edit,
  Lock,
  CopyDocument,
  Delete,
  Remove
} from '@element-plus/icons-vue'
import apiClient from '@/config/api'

export default {
  name: 'Reservations',
  components: {
    Calendar,
    Plus,
    Monitor,
    User,
    Clock,
    CircleCheck,
    WarningFilled,
    CircleClose,
    Edit,
    Lock,
    CopyDocument,
    Delete,
    Remove
  },
  setup() {
    const reservations = ref([])
    const servers = ref([])
    const loading = ref(false)
    const dialogVisible = ref(false)
    const isEditing = ref(false)
    const submitting = ref(false)
    const reservationFormRef = ref(null)
  
    const store = useStore()

    const reservationForm = reactive({
      id: null,
      server_id: null,
      start_time: '',
      end_time: ''
    })

    const validateEndTime = (rule, value, callback) => {
      if (!value) {
        callback(new Error('Please select end time'))
      } else if (!reservationForm.start_time) {
        callback(new Error('Please select start time first'))
      } else if (new Date(value) <= new Date(reservationForm.start_time)) {
        callback(new Error('End time must be after start time'))
      } else {
        callback()
      }
    }

    const validateStartTime = (rule, value, callback) => {
      if (!value) {
        callback(new Error('Please select start time'))
      } else if (new Date(value) < new Date()) {
        callback(new Error('Start time cannot be in the past'))
      } else {
        // Re-validate end time when start time changes
        if (reservationForm.end_time && reservationFormRef.value) {
          reservationFormRef.value.validateField('end_time')
        }
        callback()
      }
    }

    const rules = {
      server_id: [
        { required: true, message: 'Please select a server', trigger: 'change' }
      ],
      start_time: [
        { required: true, validator: validateStartTime, trigger: 'change' }
      ],
      end_time: [
        { required: true, validator: validateEndTime, trigger: 'change' }
      ]
    }

    const fetchReservations = async () => {
      loading.value = true
      try {
        const response = await apiClient.get('/api/reservations')
        reservations.value = response.data
      } catch (error) {
        console.error('Error fetching reservations:', error)
        ElMessage.error('Failed to fetch reservations')
      } finally {
        loading.value = false
      }
    }

    const fetchServers = async () => {
      try {
        const response = await apiClient.get('/api/servers')
        servers.value = response.data.servers || []
      } catch (error) {
        console.error('Error fetching servers:', error)
        ElMessage.error('Failed to fetch servers')
        servers.value = []
      }
    }

    const showAddDialog = () => {
      isEditing.value = false
      reservationForm.id = null
      reservationForm.server_id = null
      reservationForm.start_time = ''
      reservationForm.end_time = ''
      dialogVisible.value = true
    }

    const editReservation = (reservation) => {
      isEditing.value = true
      reservationForm.id = reservation.id
      reservationForm.server_id = reservation.server_id
      reservationForm.start_time = new Date(reservation.start_time)
      reservationForm.end_time = new Date(reservation.end_time)
      dialogVisible.value = true
    }

    const cancelReservation = async (reservation) => {
      try {
        await ElMessageBox.confirm(
          'Are you sure you want to cancel this reservation?',
          'Warning',
          {
            confirmButtonText: 'Cancel Reservation',
            cancelButtonText: 'Keep Reservation',
            type: 'warning'
          }
        )
        
        await apiClient.delete(`/api/reservations/${reservation.id}`)
        ElMessage.success('Reservation cancelled successfully')
        fetchReservations()
      } catch (error) {
        if (error !== 'cancel') {
          console.error('Error cancelling reservation:', error)
          ElMessage.error('Failed to cancel reservation')
        }
      }
    }

    const deleteReservation = async (reservation) => {
      try {
        await ElMessageBox.confirm(
          `Are you sure you want to permanently delete this reservation?
          
Server: ${reservation.server_name || 'Unknown'}
User: ${reservation.username || 'Unknown'}
Time: ${formatDate(reservation.start_time)} - ${formatDate(reservation.end_time)}

This action cannot be undone.`,
          'Delete Reservation',
          {
            confirmButtonText: 'Delete Permanently',
            cancelButtonText: 'Cancel',
            type: 'error',
            dangerouslyUseHTMLString: false
          }
        )
        
        // Use the new permanent delete endpoint for root users
        await apiClient.delete(`/api/reservations/${reservation.id}/delete`)
        ElMessage.success('Reservation deleted permanently')
        fetchReservations()
      } catch (error) {
        if (error !== 'cancel') {
          console.error('Error deleting reservation:', error)
          if (error.response?.data?.error) {
            ElMessage.error(error.response.data.error)
          } else {
            ElMessage.error('Failed to delete reservation')
          }
        }
      }
    }

    const handleReservationSubmit = async () => {
      if (!reservationFormRef.value) return

      try {
        await reservationFormRef.value.validate()
        submitting.value = true

        if (isEditing.value) {
          await apiClient.put(`/api/reservations/${reservationForm.id}`, reservationForm)
          ElMessage.success('Reservation updated successfully')
        } else {
          await apiClient.post('/api/reservations', reservationForm)
          ElMessage.success('Reservation created successfully')
        }

        dialogVisible.value = false
        fetchReservations()
      } catch (error) {
        console.error('Error saving reservation:', error)
        ElMessage.error(`Failed to ${isEditing.value ? 'update' : 'create'} reservation`)
      } finally {
        submitting.value = false
      }
    }

    const formatDate = (date) => {
      return new Date(date).toLocaleString()
    }

    const getStatusType = (status) => {
      switch (status) {
        case 'active':
          return 'success'
        case 'cancelled':
          return ''
        case 'pending':
          return 'warning'
        default:
          return 'info'
      }
    }

    const disabledDate = (time) => {
      return time.getTime() < Date.now() - 8.64e7 // Disable dates before today
    }

    const disabledHours = () => {
      const hours = []
      for (let i = 0; i < 24; i++) {
        if (i < 9 || i > 17) { // Disable hours outside 9 AM to 5 PM
          hours.push(i)
        }
      }
      return hours
    }

    const getServerName = (serverId) => {
      const server = servers.value.find(s => s.id === serverId)
      return server ? server.name : 'Unknown'
    }

    const copyToClipboard = (text) => {
      const input = document.createElement('input')
      input.value = text
      document.body.appendChild(input)
      input.select()
      document.execCommand('copy')
      document.body.removeChild(input)
      ElMessage.success('Password copied to clipboard')
    }

    

    const isRoot = computed(() => store.getters['auth/user']?.role === 'root')

    onMounted(() => {
      fetchReservations()
      fetchServers()
    })

    return {
      reservations,
      servers,
      loading,
      dialogVisible,
      isEditing,
      submitting,
      reservationForm,
      reservationFormRef,
      rules,
      dialogTitle: computed(() => isEditing.value ? 'Edit Reservation' : 'New Reservation'),
      showAddDialog,
      editReservation,
      cancelReservation,
      deleteReservation,
      handleReservationSubmit,
      formatDate,
      getStatusType,
      disabledDate,
      disabledHours,
      getServerName,
      copyToClipboard,
      isRoot
    }
  }
}
</script>

<style scoped>
.reservations-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px;
}

:deep(.el-card) {
  background: rgba(44, 62, 80, 0.95) !important;
  backdrop-filter: blur(10px);
  border: none !important;
  border-radius: 15px !important;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3) !important;
}

:deep(.el-card__header) {
  background: rgba(52, 73, 94, 0.95) !important;
  border-bottom: 2px solid #4a6583 !important;
  border-radius: 15px 15px 0 0 !important;
}

:deep(.el-card__body) {
  background: transparent !important;
  padding: 0 !important;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h2 {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f1f5f9;
}

.header-icon {
  color: #74b9ff;
  font-size: 1.5rem;
}

.add-btn {
  border-radius: 12px !important;
  padding: 12px 24px !important;
  font-weight: 600;
  gap: 8px;
  display: flex;
  align-items: center;
}

/* Table Container */
.table-container {
  overflow-x: auto;
  border-radius: 12px;
}

/* Table Styling - High Contrast Dark Theme */
.modern-table {
  border-radius: 12px !important;
  overflow: hidden;
  background: #2c3e50 !important;
  min-width: 800px;
}

:deep(.el-table) {
  background: #2c3e50 !important;
  color: #ffffff !important;
}

:deep(.el-table__header-wrapper) {
  background: #34495e !important;
}

:deep(.el-table__header) {
  background: #34495e !important;
}

:deep(.el-table th.el-table__cell) {
  background: #34495e !important;
  color: #ffffff !important;
  border-bottom: 2px solid #4a6583 !important;
  font-weight: 700 !important;
  font-size: 14px !important;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

:deep(.el-table td.el-table__cell) {
  background: #2c3e50 !important;
  color: #ffffff !important;
  border-bottom: 1px solid #4a6583 !important;
  font-weight: 500;
}

:deep(.el-table__row) {
  background: #2c3e50 !important;
}

:deep(.el-table__row:hover > td.el-table__cell) {
  background: rgba(116, 185, 255, 0.15) !important;
}

:deep(.el-table__body tr.hover-row > td.el-table__cell) {
  background: rgba(116, 185, 255, 0.15) !important;
}

:deep(.el-table--enable-row-hover .el-table__body tr:hover > td) {
  background: rgba(116, 185, 255, 0.15) !important;
}

:deep(.el-table__empty-block) {
  background: #2c3e50 !important;
  color: #bdc3c7 !important;
}

:deep(.el-table__empty-text) {
  color: #bdc3c7 !important;
}

/* Ensure all table text has proper contrast */
:deep(.el-table .cell) {
  color: #ffffff !important;
  font-weight: 500;
}

/* ID column styling */
:deep(.el-table td.el-table__cell:first-child .cell) {
  color: #74b9ff !important;
  font-weight: 700;
  font-family: 'JetBrains Mono', 'Fira Code', 'Courier New', monospace;
}

.server-info,
.user-info,
.time-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #ffffff;
  font-weight: 600;
}

.server-icon {
  color: #74b9ff;
  font-size: 1.1rem;
}

.user-icon {
  color: #ffd93d;
  font-size: 1rem;
}

.time-icon {
  color: #a78bfa;
  font-size: 1rem;
}

.status-tag {
  border-radius: 8px !important;
  font-weight: 700;
  padding: 6px 12px !important;
  display: flex;
  align-items: center;
  gap: 6px;
  color: #ffffff !important;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-size: 12px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

/* Custom styling for cancelled status */
.status-tag:not(.el-tag--success):not(.el-tag--warning):not(.el-tag--info):not(.el-tag--danger) {
  background: #6c757d !important;
  border: 1px solid #6c757d !important;
  color: #ffffff !important;
}

.status-tag:not(.el-tag--success):not(.el-tag--warning):not(.el-tag--info):not(.el-tag--danger):hover {
  background: #5a6268 !important;
  border: 1px solid #5a6268 !important;
}

.action-buttons {
  display: flex;
  gap: 4px;
  justify-content: center;
}

.compact-btn {
  width: 32px !important;
  height: 32px !important;
  padding: 0 !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
}

.edit-btn,
.cancel-btn,
.delete-btn {
  border-radius: 8px !important;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.edit-btn:hover,
.cancel-btn:hover,
.delete-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
}

.edit-btn:disabled,
.cancel-btn:disabled,
.delete-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

/* Dialog Styling */
:deep(.el-dialog) {
  border-radius: 16px !important;
}

:deep(.el-dialog__header) {
  border-radius: 16px 16px 0 0 !important;
}

:deep(.el-form-item__label) {
  font-weight: 500;
}

:deep(.el-input__wrapper) {
  border-radius: 8px !important;
  transition: all 0.3s ease;
}

:deep(.el-input__wrapper:hover) {
  transform: translateY(-1px);
}

:deep(.el-select .el-input__wrapper) {
  border-radius: 8px !important;
}

:deep(.el-date-editor) {
  border-radius: 8px !important;
}

/* Responsive Design */
@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .card-header h2 {
    justify-content: center;
  }
  
  .add-btn {
    width: 100%;
    justify-content: center;
  }
  
  .action-buttons {
    flex-direction: column;
  }
  
  .compact-btn {
    width: 28px !important;
    height: 28px !important;
  }
}

@media (max-width: 640px) {
  :deep(.el-table .el-table__cell) {
    padding: 8px 4px !important;
  }
  
  .server-info span,
  .user-info span,
  .time-info span {
    font-size: 0.875rem;
  }
}

/* Server Credentials Styling - Compact */
.credentials-compact {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 4px;
}

.cred-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  line-height: 1.2;
}

.cred-icon {
  color: #74b9ff;
  font-size: 12px;
  min-width: 12px;
}

.cred-text {
  color: #ffffff;
  font-weight: 500;
  font-size: 11px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100px;
}

.password-btn {
  padding: 2px 6px !important;
  font-size: 10px !important;
  height: auto !important;
  min-height: auto !important;
  border-radius: 4px !important;
}

.password-btn .el-icon {
  font-size: 10px;
  margin-right: 2px;
}

.no-credentials-compact {
  text-align: center;
  padding: 8px 4px;
  color: #94a3b8;
  font-style: italic;
}
</style> 