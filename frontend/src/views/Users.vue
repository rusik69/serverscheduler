<template>
  <div class="users-container">
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <h2>
                <el-icon class="header-icon"><User /></el-icon>
                User Management
              </h2>
              <el-button type="primary" @click="showAddDialog" size="large" class="add-btn">
                <el-icon><Plus /></el-icon>
                Add User
              </el-button>
            </div>
          </template>
          
          <el-table :data="users" v-loading="loading" class="modern-table">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="username" label="Username" min-width="150">
              <template #default="{ row }">
                <div class="username-info">
                  <el-icon class="username-icon"><User /></el-icon>
                  <span>{{ row.username }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="role" label="Role" width="120">
              <template #default="{ row }">
                <el-tag 
                  :type="getRoleType(row.role)"
                  class="role-tag"
                  effect="dark"
                >
                                   <el-icon>
                   <Star v-if="row.role === 'root'" />
                   <User v-else />
                 </el-icon>
                  {{ row.role }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Actions" width="200" fixed="right">
              <template #default="{ row }">
                <div class="action-buttons">
                  <el-button 
                    size="small" 
                    type="primary" 
                    @click="editUser(row)"
                    class="action-btn"
                  >
                    <el-icon><Edit /></el-icon>
                    Edit
                  </el-button>
                  <el-button 
                    size="small" 
                    type="danger" 
                    @click="deleteUser(row)"
                    :disabled="row.username === currentUser.username"
                    class="action-btn"
                  >
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

    <!-- Add/Edit User Dialog -->
    <el-dialog
      :title="dialogTitle"
      v-model="dialogVisible"
      width="500px"
    >
      <el-form :model="userForm" :rules="rules" ref="userFormRef" label-width="120px">
        <el-form-item label="Username" prop="username">
          <el-input v-model="userForm.username" placeholder="Enter username" />
        </el-form-item>
        <el-form-item label="Password" prop="password">
          <el-input 
            v-model="userForm.password" 
            type="password" 
            :placeholder="isEditing ? 'Leave empty to keep current password' : 'Enter password'"
            show-password 
          />
        </el-form-item>
        <el-form-item label="Role" prop="role">
          <el-select 
            v-model="userForm.role" 
            placeholder="Select role"
            :disabled="isEditingSelf"
          >
            <el-option 
              label="User" 
              value="user"
              :disabled="false"
            >
                             <div class="role-option">
                 <el-icon><User /></el-icon>
                 <span>User - Basic access</span>
               </div>
             </el-option>

             <el-option 
               label="Root" 
               value="root"
               :disabled="false"
             >
               <div class="role-option">
                 <el-icon><Star /></el-icon>
                 <span>Root - Full access</span>
               </div>
            </el-option>
          </el-select>
          <div v-if="isEditingSelf" class="form-help-text">
            <el-text type="info" size="small">You cannot change your own role</el-text>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">Cancel</el-button>
          <el-button type="primary" @click="handleUserSubmit" :loading="submitting">
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
  User, 
  Plus, 
  Edit, 
  Delete, 
  Star 
} from '@element-plus/icons-vue'
import { useStore } from 'vuex'
import apiClient from '@/config/api'

export default {
  name: 'Users',
  components: {
    User,
    Plus,
    Edit,
    Delete,
    Star
  },
  setup() {
    const store = useStore()
    const users = ref([])
    const loading = ref(false)
    const dialogVisible = ref(false)
    const isEditing = ref(false)
    const submitting = ref(false)
    const userFormRef = ref(null)

    const currentUser = computed(() => store.getters['auth/user'] || {})

    const userForm = reactive({
      id: null,
      username: '',
      password: '',
      role: 'user'
    })

    const rules = {
      username: [
        { required: true, message: 'Please input username', trigger: 'blur' },
        { min: 3, max: 50, message: 'Username must be between 3 and 50 characters', trigger: 'blur' }
      ],
      password: [
        { 
          validator: (rule, value, callback) => {
            if (!isEditing.value && !value) {
              callback(new Error('Please input password'))
            } else if (value && value.length < 6) {
              callback(new Error('Password must be at least 6 characters'))
            } else {
              callback()
            }
          }, 
          trigger: 'blur' 
        }
      ],
      role: [
        { required: true, message: 'Please select role', trigger: 'change' }
      ]
    }

    const fetchUsers = async () => {
      loading.value = true
      try {
        const response = await apiClient.get('/api/users')
        users.value = response.data.users || []
      } catch (error) {
        console.error('Error fetching users:', error)
        if (error.response?.status === 403) {
          ElMessage.error('Access denied. Root privileges required.')
        } else {
          ElMessage.error('Failed to fetch users')
        }
        users.value = []
      } finally {
        loading.value = false
      }
    }

    const showAddDialog = () => {
      isEditing.value = false
      userForm.id = null
      userForm.username = ''
      userForm.password = ''
      userForm.role = 'user'
      dialogVisible.value = true
    }

    const editUser = (user) => {
      isEditing.value = true
      userForm.id = user.id
      userForm.username = user.username
      userForm.password = ''
      userForm.role = user.role
      dialogVisible.value = true
    }

    const deleteUser = async (user) => {
      try {
        await ElMessageBox.confirm(
          `Are you sure you want to delete user "${user.username}"?`,
          'Warning',
          {
            confirmButtonText: 'Delete',
            cancelButtonText: 'Cancel',
            type: 'warning'
          }
        )
        
        await apiClient.delete(`/api/users/${user.id}`)
        ElMessage.success('User deleted successfully')
        fetchUsers()
      } catch (error) {
        if (error !== 'cancel') {
          console.error('Error deleting user:', error)
          if (error.response?.data?.error) {
            ElMessage.error(error.response.data.error)
          } else {
            ElMessage.error('Failed to delete user')
          }
        }
      }
    }

    const handleUserSubmit = async () => {
      if (!userFormRef.value) return

      try {
        await userFormRef.value.validate()
        submitting.value = true

        const submitData = {
          username: userForm.username,
          role: userForm.role
        }

        // Only include password if it's provided
        if (userForm.password) {
          submitData.password = userForm.password
        }

        if (isEditing.value) {
          await apiClient.put(`/api/users/${userForm.id}`, submitData)
          ElMessage.success('User updated successfully')
        } else {
          // Password is required for creating new users
          if (!userForm.password) {
            ElMessage.error('Password is required for new users')
            return
          }
          await apiClient.post('/api/users', submitData)
          ElMessage.success('User created successfully')
        }

        dialogVisible.value = false
        fetchUsers()
      } catch (error) {
        console.error('Error saving user:', error)
        if (error.response?.data?.error) {
          ElMessage.error(error.response.data.error)
        } else {
          ElMessage.error(`Failed to ${isEditing.value ? 'update' : 'create'} user`)
        }
      } finally {
        submitting.value = false
      }
    }

    const getRoleType = (role) => {
      switch (role) {
        case 'root':
          return 'danger'
        case 'user':
          return 'info'
        default:
          return 'info'
      }
    }

    const isEditingSelf = computed(() => {
      return isEditing.value && userForm.username === currentUser.value.username
    })

    onMounted(fetchUsers)

    return {
      users,
      loading,
      dialogVisible,
      isEditing,
      submitting,
      userForm,
      userFormRef,
      rules,
      currentUser,
      dialogTitle: computed(() => isEditing.value ? 'Edit User' : 'Add User'),
      isEditingSelf,
      showAddDialog,
      editUser,
      deleteUser,
      handleUserSubmit,
      getRoleType
    }
  }
}
</script>

<style scoped>
.users-container {
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

.username-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #ffffff;
  font-weight: 600;
}

.username-icon {
  color: #74b9ff;
  font-size: 1.1rem;
}

.role-tag {
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

:deep(.role-tag .el-tag__content) {
  color: #ffffff !important;
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

.action-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

/* Dialog Styling */
.role-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-help-text {
  margin-top: 4px;
}

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
  
  .username-info span {
    font-size: 0.875rem;
  }
}
</style> 