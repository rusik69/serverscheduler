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
                    :disabled="row.username === currentUser.username"
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
          <el-select v-model="userForm.role" placeholder="Select role">
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
  color: #f59e0b;
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

.username-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #e2e8f0;
}

.username-icon {
  color: #f59e0b;
  font-size: 1.1rem;
}

.role-tag {
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

.action-btn {
  border-radius: 8px !important;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: all 0.3s ease;
}

.action-btn:hover {
  transform: translateY(-1px);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Dialog Styling */
.role-option {
  display: flex;
  align-items: center;
  gap: 8px;
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