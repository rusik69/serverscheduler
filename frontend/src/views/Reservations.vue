<template>
  <div class="reservations-container">
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <h2>Reservations</h2>
              <el-button type="primary" @click="showAddDialog">New Reservation</el-button>
            </div>
          </template>
          
          <el-table :data="reservations" v-loading="loading">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="server_name" label="Server" />
            <el-table-column prop="start_time" label="Start Time">
              <template #default="{ row }">
                {{ formatDate(row.start_time) }}
              </template>
            </el-table-column>
            <el-table-column prop="end_time" label="End Time">
              <template #default="{ row }">
                {{ formatDate(row.end_time) }}
              </template>
            </el-table-column>
            <el-table-column prop="status" label="Status">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Actions" width="200">
              <template #default="{ row }">
                <el-button-group>
                  <el-button 
                    size="small" 
                    type="danger" 
                    @click="cancelReservation(row)"
                    :disabled="row.status === 'cancelled'"
                  >
                    Cancel
                  </el-button>
                </el-button-group>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- Add Reservation Dialog -->
    <el-dialog
      title="New Reservation"
      v-model="dialogVisible"
      width="500px"
    >
      <el-form :model="reservationForm" :rules="rules" ref="reservationFormRef" label-width="100px">
        <el-form-item label="Server" prop="server_id">
          <el-select v-model="reservationForm.server_id" placeholder="Select server">
            <el-option
              v-for="server in availableServers"
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
            Create
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from 'axios'

export default {
  name: 'Reservations',
  setup() {
    const reservations = ref([])
    const availableServers = ref([])
    const loading = ref(false)
    const dialogVisible = ref(false)
    const submitting = ref(false)
    const reservationFormRef = ref(null)

    const reservationForm = reactive({
      server_id: '',
      start_time: '',
      end_time: ''
    })

    const rules = {
      server_id: [
        { required: true, message: 'Please select a server', trigger: 'change' }
      ],
      start_time: [
        { required: true, message: 'Please select start time', trigger: 'change' }
      ],
      end_time: [
        { required: true, message: 'Please select end time', trigger: 'change' }
      ]
    }

    const fetchReservations = async () => {
      loading.value = true
      try {
        const response = await axios.get('http://localhost:8080/api/reservations')
        reservations.value = response.data
      } catch (error) {
        console.error('Error fetching reservations:', error)
        ElMessage.error('Failed to fetch reservations')
      } finally {
        loading.value = false
      }
    }

    const fetchAvailableServers = async () => {
      try {
        const response = await axios.get('http://localhost:8080/api/servers')
        availableServers.value = response.data.filter(server => server.status === 'available')
      } catch (error) {
        console.error('Error fetching servers:', error)
        ElMessage.error('Failed to fetch available servers')
      }
    }

    const showAddDialog = () => {
      reservationForm.server_id = ''
      reservationForm.start_time = ''
      reservationForm.end_time = ''
      dialogVisible.value = true
    }

    const cancelReservation = async (reservation) => {
      try {
        await ElMessageBox.confirm(
          'Are you sure you want to cancel this reservation?',
          'Warning',
          {
            confirmButtonText: 'Cancel Reservation',
            cancelButtonText: 'No',
            type: 'warning'
          }
        )
        
        await axios.delete(`http://localhost:8080/api/reservations/${reservation.id}`)
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

        await axios.post('http://localhost:8080/api/reservations', reservationForm)
        ElMessage.success('Reservation created successfully')
        dialogVisible.value = false
        fetchReservations()
      } catch (error) {
        console.error('Error creating reservation:', error)
        ElMessage.error('Failed to create reservation')
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

    onMounted(() => {
      fetchReservations()
      fetchAvailableServers()
    })

    return {
      reservations,
      availableServers,
      loading,
      dialogVisible,
      submitting,
      reservationForm,
      reservationFormRef,
      rules,
      showAddDialog,
      cancelReservation,
      handleReservationSubmit,
      formatDate,
      getStatusType,
      disabledDate,
      disabledHours
    }
  }
}
</script>

<style scoped>
.reservations-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h2 {
  margin: 0;
}
</style> 