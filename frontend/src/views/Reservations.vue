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
          
          <el-table :data="reservations" v-loading="loading" class="modern-table">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="server_name" label="Server" min-width="150">
              <template #default="{ row }">
                <div class="server-info">
                  <el-icon class="server-icon"><Monitor /></el-icon>
                  <span>{{ row.server_name }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="username" label="User" min-width="120">
              <template #default="{ row }">
                <div class="user-info">
                  <el-icon class="user-icon"><User /></el-icon>
                  <span>{{ row.username }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="start_time" label="Start Time" min-width="180">
              <template #default="{ row }">
                <div class="time-info">
                  <el-icon class="time-icon"><Clock /></el-icon>
                  <span>{{ formatDate(row.start_time) }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="end_time" label="End Time" min-width="180">
              <template #default="{ row }">
                <div class="time-info">
                  <el-icon class="time-icon"><Clock /></el-icon>
                  <span>{{ formatDate(row.end_time) }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="Status" width="120">
              <template #default="{ row }">
                <el-tag 
                  :type="getStatusType(row.status)"
                  class="status-tag"
                  effect="dark"
                >
                  <el-icon>
                    <CircleCheck v-if="row.status === 'active'" />
                    <WarningFilled v-else-if="row.status === 'pending'" />
                    <CircleClose v-else />
                  </el-icon>
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Actions" width="200" fixed="right">
              <template #default="{ row }">
                <div class="action-buttons">
                  <el-button 
                    size="small" 
                    type="primary" 
                    @click="editReservation(row)"
                    :disabled="row.status === 'cancelled'"
                    class="edit-btn"
                  >
                    <el-icon><Edit /></el-icon>
                    Edit
                  </el-button>
                  <el-button 
                    size="small" 
                    type="danger" 
                    @click="cancelReservation(row)"
                    :disabled="row.status === 'cancelled'"
                    class="cancel-btn"
                  >
                    <el-icon><CircleClose /></el-icon>
                    Cancel
                  </el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
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
import { 
  Calendar, 
  Plus, 
  Monitor, 
  User, 
  Clock, 
  CircleCheck, 
  WarningFilled, 
  CircleClose,
  Edit 
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
    Edit
  },
  setup() {
    const reservations = ref([])
    const servers = ref([])
    const loading = ref(false)
    const dialogVisible = ref(false)
    const isEditing = ref(false)
    const submitting = ref(false)
    const reservationFormRef = ref(null)

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
          return 'danger'
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
      handleReservationSubmit,
      formatDate,
      getStatusType,
      disabledDate,
      disabledHours,
      getServerName
    }
  }
}
</script>

<style scoped>
.reservations-container {
  max-width: 1400px;
  margin: 0 auto;
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
  color: #10b981;
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

/* Table Styling */
.modern-table {
  border-radius: 12px !important;
  overflow: hidden;
}

.server-info,
.user-info,
.time-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #e2e8f0;
}

.server-icon {
  color: #60a5fa;
  font-size: 1.1rem;
}

.user-icon {
  color: #f59e0b;
  font-size: 1rem;
}

.time-icon {
  color: #8b5cf6;
  font-size: 1rem;
}

.status-tag {
  border-radius: 8px !important;
  font-weight: 500;
  padding: 4px 12px !important;
  display: flex;
  align-items: center;
  gap: 4px;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.edit-btn,
.cancel-btn {
  border-radius: 8px !important;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: all 0.3s ease;
}

.edit-btn:hover,
.cancel-btn:hover {
  transform: translateY(-1px);
}

.edit-btn:disabled,
.cancel-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
  
  .edit-btn,
  .cancel-btn {
    width: 100%;
    justify-content: center;
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
</style> 