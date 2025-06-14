<template>
  <div class="home-container">
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card>
          <template #header>
            <h2>Welcome to Server Scheduler</h2>
          </template>
          <p>Manage your server reservations efficiently with our scheduling system.</p>
          <div class="features">
            <el-row :gutter="20">
              <el-col :span="8">
                <el-card shadow="hover">
                  <h3>Server Management</h3>
                  <p>View and manage your available servers</p>
                  <el-button type="primary" @click="$router.push('/servers')">View Servers</el-button>
                </el-card>
              </el-col>
              <el-col :span="8">
                <el-card shadow="hover">
                  <h3>Reservations</h3>
                  <p>Schedule and manage server reservations</p>
                  <el-button type="primary" @click="$router.push('/reservations')">View Reservations</el-button>
                </el-card>
              </el-col>
              <el-col :span="8">
                <el-card shadow="hover">
                  <h3>User Management</h3>
                  <p>Manage user accounts and permissions</p>
                  <el-button type="primary" @click="$router.push('/users')">View Users</el-button>
                </el-card>
              </el-col>
            </el-row>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
export default {
  name: 'Home'
}
</script>

<style scoped>
.home-container {
  padding: 20px;
}

.features {
  margin-top: 20px;
}

.el-card {
  margin-bottom: 20px;
}

h3 {
  margin-top: 0;
}
</style> 