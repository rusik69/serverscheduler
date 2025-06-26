<template>
  <div class="servers-container">
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <h2>
                <el-icon class="header-icon"><Monitor /></el-icon>
                Server Management
              </h2>
              <el-button type="primary" @click="showAddDialog" size="large" class="add-btn">
                <el-icon><Plus /></el-icon>
                Add Server
              </el-button>
            </div>
          </template>
          
          <el-table :data="servers" v-loading="loading" class="modern-table">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="name" label="Server Name" min-width="150">
              <template #default="{ row }">
                <div class="server-name">
                  <el-icon class="server-icon"><Monitor /></el-icon>
                  <span>{{ row.name }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="ip_address" label="IP Address" min-width="130">
              <template #default="{ row }">
                <div class="ip-address">
                  <el-icon><Connection /></el-icon>
                  <span>{{ row.ip_address || 'Not set' }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="username" label="Username" min-width="120">
              <template #default="{ row }">
                <div class="username">
                  <el-icon><User /></el-icon>
                  <span>{{ row.username || 'Not set' }}</span>
                </div>
              </template>
            </el-table-column>
            <!-- Password column - Only visible to root users -->
            <el-table-column v-if="isRoot" label="Password" min-width="150">
              <template #default="{ row }">
                <div v-if="row.password" class="password-field">
                  <el-tag 
                    size="small" 
                    type="warning" 
                    @click="copyToClipboard(row.password)" 
                    class="clickable-tag"
                  >
                    {{ showPasswords ? row.password : '••••••••' }}
                    <el-icon><CopyDocument /></el-icon>
                  </el-tag>
                  <el-button size="small" text @click="togglePasswordVisibility">
                    <el-icon><ViewIcon v-if="!showPasswords" /><Hide v-else /></el-icon>
                  </el-button>
                </div>
                <div v-else class="no-password">
                  <el-text type="info">Not set</el-text>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="Status" width="120">
              <template #default="{ row }">
                <el-tag 
                  :type="row.status === 'available' ? 'success' : 'danger'"
                  class="status-tag"
                  effect="dark"
                >
                  <el-icon>
                    <CircleCheck v-if="row.status === 'available'" />
                    <CircleClose v-else />
                  </el-icon>
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Actions" width="180" fixed="right">
              <template #default="{ row }">
                <div class="action-buttons">
                  <el-button size="small" type="primary" @click="editServer(row)" class="action-btn">
                    <el-icon><Edit /></el-icon>
                    Edit
                  </el-button>
                  <el-button size="small" type="danger" @click="deleteServer(row)" class="action-btn">
                    <el-icon><Delete /></el-icon>
                    Delete
                  </el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- Add/Edit Server Dialog -->
    <el-dialog
      :title="dialogTitle"
      v-model="dialogVisible"
      width="500px"
    >
      <el-form :model="serverForm" :rules="rules" ref="serverFormRef" label-width="120px">
        <el-form-item label="Name" prop="name">
          <el-input v-model="serverForm.name" />
        </el-form-item>
        <el-form-item label="IP Address" prop="ip_address">
          <el-input v-model="serverForm.ip_address" placeholder="e.g., 192.168.1.100" />
        </el-form-item>
        <el-form-item label="Username" prop="username">
          <el-input v-model="serverForm.username" placeholder="Server login username" />
        </el-form-item>
        <el-form-item label="Password" prop="password">
          <el-input v-model="serverForm.password" type="password" placeholder="Server login password" show-password />
        </el-form-item>
        <el-form-item label="Status" prop="status">
          <el-select v-model="serverForm.status" placeholder="Select status">
            <el-option label="Available" value="available" />
            <el-option label="Unavailable" value="unavailable" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">Cancel</el-button>
          <el-button type="primary" @click="handleServerSubmit" :loading="submitting">
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
  Monitor, 
  Connection, 
  User, 
  CircleCheck, 
  CircleClose, 
  Edit, 
  Delete, 
  Plus,
  CopyDocument,
  View as ViewIcon,
  Hide
} from '@element-plus/icons-vue'
import apiClient from '@/config/api'

export default {
  name: 'Servers',
  components: {
    Monitor,
    Connection,
    User,
    CircleCheck,
    CircleClose,
    Edit,
    Delete,
    Plus,
    CopyDocument,
    ViewIcon,
    Hide
  },
  setup() {
    const servers = ref([])
    const loading = ref(false)
    const dialogVisible = ref(false)
    const isEditing = ref(false)
    const submitting = ref(false)
    const serverFormRef = ref(null)
    const showPasswords = ref(false)
    const store = useStore()

    const isRoot = computed(() => store.getters['auth/user']?.role === 'root')

    const serverForm = reactive({
      id: null,
      name: '',
      ip_address: '',
      username: '',
      password: '',
      status: 'available'
    })

    const validateIPAddress = (rule, value, callback) => {
      if (!value || value.trim() === '') {
        // Empty IP address is allowed
        callback()
        return
      }
      
      // Regular expression for IPv4 validation
      const ipv4Regex = /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/
      
      // More comprehensive IPv6 validation
      const ipv6Regex = /^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$/
      
      const trimmedValue = value.trim()
      
      if (ipv4Regex.test(trimmedValue) || ipv6Regex.test(trimmedValue)) {
        callback()
      } else {
        callback(new Error('Please enter a valid IP address (IPv4 or IPv6)'))
      }
    }

    const rules = {
      name: [
        { required: true, message: 'Please input server name', trigger: 'blur' }
      ],
      ip_address: [
        { validator: validateIPAddress, trigger: 'blur' }
      ],
      status: [
        { required: true, message: 'Please select status', trigger: 'change' }
      ]
    }

    const fetchServers = async () => {
      loading.value = true
      try {
        const response = await apiClient.get('/api/servers')
        servers.value = response.data.servers || []
      } catch (error) {
        console.error('Error fetching servers:', error)
        ElMessage.error('Failed to fetch servers')
        servers.value = []
      } finally {
        loading.value = false
      }
    }

    const showAddDialog = () => {
      isEditing.value = false
      serverForm.id = null
      serverForm.name = ''
      serverForm.ip_address = ''
      serverForm.username = ''
      serverForm.password = ''
      serverForm.status = 'available'
      dialogVisible.value = true
    }

    const editServer = (server) => {
      isEditing.value = true
      Object.assign(serverForm, server)
      dialogVisible.value = true
    }

    const deleteServer = async (server) => {
      try {
        await ElMessageBox.confirm(
          'Are you sure you want to delete this server?',
          'Warning',
          {
            confirmButtonText: 'Delete',
            cancelButtonText: 'Cancel',
            type: 'warning'
          }
        )
        
        await apiClient.delete(`/api/servers/${server.id}`)
        ElMessage.success('Server deleted successfully')
        fetchServers()
      } catch (error) {
        if (error !== 'cancel') {
          console.error('Error deleting server:', error)
          ElMessage.error('Failed to delete server')
        }
      }
    }

    const handleServerSubmit = async () => {
      if (!serverFormRef.value) return

      try {
        await serverFormRef.value.validate()
        submitting.value = true

        if (isEditing.value) {
          await apiClient.put(`/api/servers/${serverForm.id}`, serverForm)
          ElMessage.success('Server updated successfully')
        } else {
          await apiClient.post('/api/servers', serverForm)
          ElMessage.success('Server created successfully')
        }

        dialogVisible.value = false
        fetchServers()
      } catch (error) {
        console.error('Error saving server:', error)
        ElMessage.error('Failed to save server')
      } finally {
        submitting.value = false
      }
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

    const togglePasswordVisibility = () => {
      showPasswords.value = !showPasswords.value
    }

    onMounted(fetchServers)

    return {
      servers,
      loading,
      dialogVisible,
      isEditing,
      submitting,
      serverForm,
      serverFormRef,
      rules,
      dialogTitle: computed(() => isEditing.value ? 'Edit Server' : 'Add Server'),
      showAddDialog,
      editServer,
      deleteServer,
      handleServerSubmit,
      copyToClipboard,
      togglePasswordVisibility,
      isRoot,
      showPasswords
    }
  }
}
</script>

<style scoped>
.servers-container {
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

/* Table Styling - High Contrast Dark Theme */
.modern-table {
  border-radius: 12px !important;
  overflow: hidden;
  background: #2c3e50 !important;
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

.server-name,
.ip-address,
.username {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #ffffff;
  font-weight: 600;
}

.server-name .el-icon {
  color: #74b9ff;
  font-size: 1.1rem;
}

.ip-address .el-icon {
  color: #00d084;
  font-size: 1rem;
}

.username .el-icon {
  color: #ffd93d;
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

.action-buttons {
  display: flex;
  gap: 8px;
}

.action-btn {
  border-radius: 8px !important;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.action-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
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
  
  .action-btn {
    width: 100%;
    justify-content: center;
  }
}

@media (max-width: 640px) {
  :deep(.el-table .el-table__cell) {
    padding: 8px 4px !important;
  }
  
  .server-name span,
  .ip-address span,
  .username span {
    font-size: 0.875rem;
  }
}

/* Password Field Styling */
.password-field {
  display: flex;
  align-items: center;
  gap: 8px;
}

.clickable-tag {
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: 'Courier New', monospace;
}

.clickable-tag:hover {
  opacity: 0.8;
  transform: scale(1.05);
}

.no-password {
  text-align: center;
  color: #94a3b8;
  font-style: italic;
}
</style> 