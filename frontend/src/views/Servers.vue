<template>
  <div class="servers-container">
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <h2>Servers</h2>
              <el-button type="primary" @click="showAddDialog">Add Server</el-button>
            </div>
          </template>
          
          <el-table :data="servers" v-loading="loading">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="name" label="Name" />
            <el-table-column prop="status" label="Status">
              <template #default="{ row }">
                <el-tag :type="row.status === 'available' ? 'success' : 'danger'">
                  {{ row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Actions" width="200">
              <template #default="{ row }">
                <el-button-group>
                  <el-button size="small" @click="editServer(row)">Edit</el-button>
                  <el-button size="small" type="danger" @click="deleteServer(row)">Delete</el-button>
                </el-button-group>
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
      <el-form :model="serverForm" :rules="rules" ref="serverFormRef" label-width="100px">
        <el-form-item label="Name" prop="name">
          <el-input v-model="serverForm.name" />
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from 'axios'

export default {
  name: 'Servers',
  setup() {
    const servers = ref([])
    const loading = ref(false)
    const dialogVisible = ref(false)
    const isEditing = ref(false)
    const submitting = ref(false)
    const serverFormRef = ref(null)

    const serverForm = reactive({
      id: null,
      name: '',
      status: 'available'
    })

    const rules = {
      name: [
        { required: true, message: 'Please input server name', trigger: 'blur' }
      ],
      status: [
        { required: true, message: 'Please select status', trigger: 'change' }
      ]
    }

    const fetchServers = async () => {
      loading.value = true
      try {
        const response = await axios.get('http://localhost:8080/api/servers')
        servers.value = response.data
      } catch (error) {
        console.error('Error fetching servers:', error)
        ElMessage.error('Failed to fetch servers')
      } finally {
        loading.value = false
      }
    }

    const showAddDialog = () => {
      isEditing.value = false
      serverForm.id = null
      serverForm.name = ''
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
        
        await axios.delete(`http://localhost:8080/api/servers/${server.id}`)
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
          await axios.put(`http://localhost:8080/api/servers/${serverForm.id}`, serverForm)
          ElMessage.success('Server updated successfully')
        } else {
          await axios.post('http://localhost:8080/api/servers', serverForm)
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
      handleServerSubmit
    }
  }
}
</script>

<style scoped>
.servers-container {
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