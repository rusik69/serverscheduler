<template>
  <div class="home-container">
    <!-- Hero Section -->
    <div class="hero-section">
      <div class="hero-content">
        <h1 class="hero-title">
          <el-icon class="hero-icon"><Cpu /></el-icon>
          Server Scheduler
        </h1>
        <p class="hero-subtitle">
          Efficiently manage and schedule your server infrastructure with our powerful scheduling system
        </p>
        <div class="hero-stats">
          <div class="stat-card">
            <el-icon class="stat-icon"><Monitor /></el-icon>
            <div class="stat-content">
              <div class="stat-number">{{ serverCount }}</div>
              <div class="stat-label">Total Servers</div>
            </div>
          </div>
          <div class="stat-card">
            <el-icon class="stat-icon"><Calendar /></el-icon>
            <div class="stat-content">
              <div class="stat-number">{{ reservationCount }}</div>
              <div class="stat-label">Active Reservations</div>
            </div>
          </div>
          <div class="stat-card">
            <el-icon class="stat-icon"><User /></el-icon>
            <div class="stat-content">
              <div class="stat-number">{{ userCount }}</div>
              <div class="stat-label">Registered Users</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Features Section -->
    <div class="features-section">
      <h2 class="section-title">Platform Features</h2>
      <el-row :gutter="24">
        <el-col :span="8" :xs="24" :sm="12" :md="8">
          <div class="feature-card" @click="$router.push('/servers')">
            <div class="feature-icon-wrapper servers">
              <el-icon class="feature-icon"><Monitor /></el-icon>
            </div>
            <h3 class="feature-title">Server Management</h3>
            <p class="feature-description">
              View, create, and manage your server infrastructure with detailed connection information
            </p>
            <div class="feature-actions">
              <el-button type="primary" size="large" @click.stop="$router.push('/servers')">
                Manage Servers
                <el-icon class="el-icon--right"><ArrowRight /></el-icon>
              </el-button>
            </div>
          </div>
        </el-col>
        
        <el-col :span="8" :xs="24" :sm="12" :md="8">
          <div class="feature-card" @click="$router.push('/reservations')">
            <div class="feature-icon-wrapper reservations">
              <el-icon class="feature-icon"><Calendar /></el-icon>
            </div>
            <h3 class="feature-title">Smart Reservations</h3>
            <p class="feature-description">
              Schedule server time slots with automatic conflict detection and user tracking
            </p>
            <div class="feature-actions">
              <el-button type="primary" size="large" @click.stop="$router.push('/reservations')">
                View Reservations
                <el-icon class="el-icon--right"><ArrowRight /></el-icon>
              </el-button>
            </div>
          </div>
        </el-col>
        
        <el-col :span="8" :xs="24" :sm="12" :md="8" v-if="isRoot">
          <div class="feature-card" @click="$router.push('/users')">
            <div class="feature-icon-wrapper users">
              <el-icon class="feature-icon"><Management /></el-icon>
            </div>
            <h3 class="feature-title">User Management</h3>
            <p class="feature-description">
              Manage system users and roles with comprehensive root controls
            </p>
            <div class="feature-actions">
              <el-button type="primary" size="large" @click.stop="$router.push('/users')">
                Manage Users
                <el-icon class="el-icon--right"><ArrowRight /></el-icon>
              </el-button>
            </div>
          </div>
        </el-col>
        
        <el-col :span="8" :xs="24" :sm="12" :md="8" v-if="!isRoot">
          <div class="feature-card">
            <div class="feature-icon-wrapper analytics">
              <el-icon class="feature-icon"><DataLine /></el-icon>
            </div>
            <h3 class="feature-title">Usage Analytics</h3>
            <p class="feature-description">
              Track server utilization and reservation patterns with detailed analytics
            </p>
            <div class="feature-actions">
              <el-button size="large" disabled>
                Coming Soon
                <el-icon class="el-icon--right"><Clock /></el-icon>
              </el-button>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <!-- Admin Panel (Root Only) -->
    <div v-if="isRoot" class="admin-panel">
      <el-card class="admin-panel-card">
        <template #header>
          <div class="admin-panel-header">
            <el-icon class="admin-icon"><Management /></el-icon>
            <span>System Administration</span>
            <el-tag type="danger" effect="dark" class="root-badge">ROOT ACCESS</el-tag>
          </div>
        </template>
        <div class="admin-content">
          <p class="admin-description">
            You have root privileges. Manage system users, monitor activity, and configure system settings.
          </p>
          <div class="admin-stats">
            <div class="admin-stat-item">
              <el-icon class="admin-stat-icon"><User /></el-icon>
              <div class="admin-stat-info">
                <div class="admin-stat-number">{{ userCount }}</div>
                <div class="admin-stat-label">Total Users</div>
              </div>
            </div>
            <div class="admin-stat-item">
              <el-icon class="admin-stat-icon"><Monitor /></el-icon>
              <div class="admin-stat-info">
                <div class="admin-stat-number">{{ serverCount }}</div>
                <div class="admin-stat-label">Managed Servers</div>
              </div>
            </div>
            <div class="admin-stat-item">
              <el-icon class="admin-stat-icon"><Calendar /></el-icon>
              <div class="admin-stat-info">
                <div class="admin-stat-number">{{ reservationCount }}</div>
                <div class="admin-stat-label">Active Bookings</div>
              </div>
            </div>
          </div>
          <div class="admin-actions">
            <el-button type="warning" size="large" @click="$router.push('/users')" class="admin-action-btn">
              <el-icon><Management /></el-icon>
              User Management
            </el-button>
            <el-button type="primary" size="large" @click="$router.push('/servers')" class="admin-action-btn">
              <el-icon><Monitor /></el-icon>
              Server Control
            </el-button>
            <el-button type="success" size="large" @click="$router.push('/reservations')" class="admin-action-btn">
              <el-icon><Calendar /></el-icon>
              Reservation Overview
            </el-button>
          </div>
        </div>
      </el-card>
    </div>

    <!-- Quick Actions -->
    <div class="quick-actions">
      <el-card class="quick-actions-card">
        <template #header>
          <div class="quick-actions-header">
            <el-icon><Lightning /></el-icon>
            <span>Quick Actions</span>
          </div>
        </template>
        <div class="quick-action-buttons">
          <el-button-group>
            <el-button type="primary" size="large" @click="$router.push('/servers')">
              <el-icon><Plus /></el-icon>
              Add Server
            </el-button>
            <el-button type="success" size="large" @click="$router.push('/reservations')">
              <el-icon><Calendar /></el-icon>
              New Reservation
            </el-button>
            <el-button v-if="isRoot" type="warning" size="large" @click="$router.push('/users')">
              <el-icon><Management /></el-icon>
              Manage Users
            </el-button>
            <el-button type="info" size="large" @click="refreshStats">
              <el-icon><Refresh /></el-icon>
              Refresh Stats
            </el-button>
          </el-button-group>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue'
import { 
  Cpu, 
  Monitor, 
  Calendar, 
  User, 
  ArrowRight, 
  DataLine, 
  Clock, 
  Lightning, 
  Plus, 
  Refresh,
  Management 
} from '@element-plus/icons-vue'
import { useStore } from 'vuex'
import apiClient from '@/config/api'

export default {
  name: 'Home',
  components: {
    Cpu,
    Monitor,
    Calendar,
    User,
    ArrowRight,
    DataLine,
    Clock,
    Lightning,
    Plus,
    Refresh,
    Management
  },
  setup() {
    const store = useStore()
    const serverCount = ref(0)
    const reservationCount = ref(0)
    const userCount = ref(1) // Default to 1 for current user

    const isRoot = computed(() => store.getters['auth/user']?.role === 'root')

    const fetchStats = async () => {
      try {
        // Fetch servers count
        const serversResponse = await apiClient.get('/api/servers')
        serverCount.value = serversResponse.data.servers?.length || 0

        // Fetch reservations count
        const reservationsResponse = await apiClient.get('/api/reservations')
        reservationCount.value = reservationsResponse.data?.length || 0

        // Fetch users count (only for root users)
        if (isRoot.value) {
          try {
            const usersResponse = await apiClient.get('/api/users')
            userCount.value = usersResponse.data.users?.length || 0
          } catch (error) {
            console.error('Error fetching users:', error)
            userCount.value = 1 // Fallback to current user
          }
        }
      } catch (error) {
        console.error('Error fetching stats:', error)
      }
    }

    const refreshStats = () => {
      fetchStats()
    }

    onMounted(() => {
      fetchStats()
    })

    return {
      serverCount,
      reservationCount,
      userCount,
      isRoot,
      refreshStats
    }
  }
}
</script>

<style scoped>
.home-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0;
}

/* Hero Section */
.hero-section {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1) 0%, rgba(147, 51, 234, 0.1) 100%);
  border-radius: 20px;
  padding: 60px 40px;
  margin-bottom: 40px;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(96, 165, 250, 0.2);
  text-align: center;
}

.hero-content {
  max-width: 800px;
  margin: 0 auto;
}

.hero-title {
  font-size: 3.5rem;
  font-weight: 800;
  margin-bottom: 20px;
  background: linear-gradient(135deg, #60a5fa, #a855f7);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
}

.hero-icon {
  font-size: 3.5rem;
  color: #60a5fa;
}

.hero-subtitle {
  font-size: 1.25rem;
  color: #cbd5e1;
  margin-bottom: 40px;
  line-height: 1.6;
}

.hero-stats {
  display: flex;
  justify-content: center;
  gap: 24px;
  flex-wrap: wrap;
}

.stat-card {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(51, 65, 85, 0.5);
  border-radius: 16px;
  padding: 24px;
  min-width: 180px;
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 10px 25px rgba(96, 165, 250, 0.2);
  border-color: rgba(96, 165, 250, 0.5);
}

.stat-icon {
  font-size: 2rem;
  color: #60a5fa;
}

.stat-content {
  text-align: left;
}

.stat-number {
  font-size: 2rem;
  font-weight: 700;
  color: #f1f5f9;
  line-height: 1;
}

.stat-label {
  font-size: 0.875rem;
  color: #94a3b8;
  margin-top: 4px;
}

/* Features Section */
.features-section {
  margin-bottom: 40px;
}

.section-title {
  text-align: center;
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 40px;
  background: linear-gradient(135deg, #f1f5f9, #cbd5e1);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.feature-card {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(51, 65, 85, 0.5);
  border-radius: 20px;
  padding: 32px;
  text-align: center;
  transition: all 0.3s ease;
  cursor: pointer;
  backdrop-filter: blur(10px);
  height: 100%;
  display: flex;
  flex-direction: column;
  margin-bottom: 24px;
}

.feature-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
  border-color: rgba(96, 165, 250, 0.5);
}

.feature-icon-wrapper {
  width: 80px;
  height: 80px;
  margin: 0 auto 24px;
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}

.feature-icon-wrapper.servers {
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
}

.feature-icon-wrapper.reservations {
  background: linear-gradient(135deg, #10b981, #059669);
}

.feature-icon-wrapper.users {
  background: linear-gradient(135deg, #f59e0b, #d97706);
}

.feature-icon-wrapper.analytics {
  background: linear-gradient(135deg, #8b5cf6, #7c3aed);
}

.feature-icon {
  font-size: 2.5rem;
  color: white;
  z-index: 1;
}

.feature-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 16px;
  color: #f1f5f9;
}

.feature-description {
  color: #cbd5e1;
  line-height: 1.6;
  margin-bottom: 24px;
  flex-grow: 1;
}

.feature-actions {
  margin-top: auto;
}

/* Admin Panel */
.admin-panel {
  margin-bottom: 40px;
}

.admin-panel-card {
  border-radius: 16px !important;
  overflow: hidden;
  border: 2px solid rgba(239, 68, 68, 0.3) !important;
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.05) 0%, rgba(147, 51, 234, 0.05) 100%);
}

.admin-panel-header {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 1.25rem;
  font-weight: 600;
  color: #f1f5f9;
}

.admin-icon {
  color: #f59e0b;
  font-size: 1.5rem;
}

.root-badge {
  margin-left: auto;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.admin-content {
  padding: 0;
}

.admin-description {
  color: #cbd5e1;
  font-size: 1rem;
  margin-bottom: 24px;
  text-align: center;
  line-height: 1.6;
}

.admin-stats {
  display: flex;
  justify-content: space-around;
  margin-bottom: 32px;
  flex-wrap: wrap;
  gap: 16px;
}

.admin-stat-item {
  display: flex;
  align-items: center;
  gap: 12px;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(51, 65, 85, 0.5);
  border-radius: 12px;
  padding: 16px 20px;
  min-width: 140px;
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
}

.admin-stat-item:hover {
  transform: translateY(-2px);
  border-color: rgba(239, 68, 68, 0.5);
  box-shadow: 0 8px 16px rgba(239, 68, 68, 0.1);
}

.admin-stat-icon {
  font-size: 1.5rem;
  color: #ef4444;
}

.admin-stat-info {
  text-align: left;
}

.admin-stat-number {
  font-size: 1.5rem;
  font-weight: 700;
  color: #f1f5f9;
  line-height: 1;
}

.admin-stat-label {
  font-size: 0.75rem;
  color: #94a3b8;
  margin-top: 2px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.admin-actions {
  display: flex;
  justify-content: center;
  gap: 16px;
  flex-wrap: wrap;
}

.admin-action-btn {
  min-width: 180px;
  font-weight: 600;
  border-radius: 10px !important;
  transition: all 0.3s ease;
}

.admin-action-btn:hover {
  transform: translateY(-2px);
}

/* Quick Actions */
.quick-actions {
  margin-bottom: 40px;
}

.quick-actions-card {
  border-radius: 16px !important;
  overflow: hidden;
}

.quick-actions-header {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 1.25rem;
  font-weight: 600;
  color: #f1f5f9;
}

.quick-action-buttons {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 16px;
}

.quick-action-buttons .el-button {
  min-width: 160px;
}

/* Responsive Design */
@media (max-width: 1024px) {
  .hero-title {
    font-size: 2.5rem;
  }
  
  .hero-icon {
    font-size: 2.5rem;
  }
  
  .hero-stats {
    flex-direction: column;
    align-items: center;
  }
  
  .stat-card {
    width: 100%;
    max-width: 300px;
  }
}

@media (max-width: 768px) {
  .hero-section {
    padding: 40px 20px;
  }
  
  .hero-title {
    font-size: 2rem;
    flex-direction: column;
    gap: 8px;
  }
  
  .hero-icon {
    font-size: 2rem;
  }
  
  .section-title {
    font-size: 2rem;
  }
  
  .feature-card {
    margin-bottom: 16px;
  }
  
  .admin-stats {
    flex-direction: column;
    align-items: center;
  }
  
  .admin-stat-item {
    width: 100%;
    max-width: 280px;
    justify-content: center;
  }
  
  .admin-actions {
    flex-direction: column;
  }
  
  .admin-action-btn {
    width: 100%;
  }
  
  .quick-action-buttons {
    flex-direction: column;
  }
  
  .quick-action-buttons .el-button {
    width: 100%;
  }
}

@media (max-width: 640px) {
  .hero-title {
    font-size: 1.75rem;
  }
  
  .hero-subtitle {
    font-size: 1rem;
  }
  
  .stat-card {
    flex-direction: column;
    text-align: center;
    gap: 8px;
  }
  
  .stat-content {
    text-align: center;
  }
}
</style> 