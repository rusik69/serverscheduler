<template>
  <el-container class="app-container dark">
    <el-header class="app-header">
      <div class="nav-brand">
        <el-icon class="brand-icon"><Cpu /></el-icon>
        <span class="brand-text">Server Scheduler</span>
      </div>
      <el-menu 
        mode="horizontal" 
        router 
        class="nav-menu"
        background-color="transparent"
        text-color="#e5e7eb"
        active-text-color="#60a5fa"
      >
        <el-menu-item index="/" class="nav-item">
          <el-icon><House /></el-icon>
          <span>Dashboard</span>
        </el-menu-item>
        <el-menu-item index="/servers" class="nav-item">
          <el-icon><Monitor /></el-icon>
          <span>Servers</span>
        </el-menu-item>
        <el-menu-item index="/reservations" class="nav-item">
          <el-icon><Calendar /></el-icon>
          <span>Reservations</span>
        </el-menu-item>
        <el-menu-item v-if="isRoot" index="/users" class="nav-item admin-nav-item">
          <el-icon><Management /></el-icon>
          <span>User Management</span>
        </el-menu-item>
        <div class="flex-grow" />
        <el-menu-item v-if="!isAuthenticated" index="/login" class="nav-item">
          <el-icon><User /></el-icon>
          <span>Login</span>
        </el-menu-item>
        <el-menu-item v-if="!isAuthenticated" index="/register" class="nav-item">
          <el-icon><Plus /></el-icon>
          <span>Register</span>
        </el-menu-item>
        <el-dropdown v-if="isAuthenticated" class="user-dropdown">
          <span class="user-info" :class="{ 'root-user': isRoot }">
            <el-icon><User /></el-icon>
            <div class="user-details">
              <span class="username">{{ username }}</span>
              <span v-if="isRoot" class="user-role">ROOT</span>
            </div>
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-if="isRoot" @click="$router.push('/users')" class="admin-dropdown-item">
                <el-icon><Management /></el-icon>
                User Management
              </el-dropdown-item>
              <el-dropdown-item @click="logout">
                <el-icon><SwitchButton /></el-icon>
                Logout
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-menu>
    </el-header>
    <el-main class="app-main">
      <div class="main-content">
        <router-view />
      </div>
    </el-main>
  </el-container>
</template>

<script>
import { computed } from 'vue'
import { useStore } from 'vuex'
import { useRouter } from 'vue-router'
import { 
  Cpu, 
  House, 
  Monitor, 
  Calendar, 
  User, 
  Plus, 
  ArrowDown, 
  SwitchButton,
  Management 
} from '@element-plus/icons-vue'

export default {
  name: 'App',
  components: {
    Cpu,
    House,
    Monitor,
    Calendar,
    User,
    Plus,
    ArrowDown,
    SwitchButton,
    Management
  },
  setup() {
    const store = useStore()
    const router = useRouter()

    const isAuthenticated = computed(() => store.getters['auth/isAuthenticated'])
    const username = computed(() => store.getters['auth/user']?.username || 'User')
    const isRoot = computed(() => store.getters['auth/user']?.role === 'root')

    // Debug logging
    const user = computed(() => store.getters['auth/user'])
    console.log('Debug - Current user:', user.value)
    console.log('Debug - Is authenticated:', isAuthenticated.value)
    console.log('Debug - Is root:', isRoot.value)

    const logout = async () => {
      await store.dispatch('auth/logout')
      router.push('/login')
    }

    return {
      isAuthenticated,
      username,
      isRoot,
      logout,
      $router: router
    }
  }
}
</script>

<style>
/* Dark theme global styles */
html.dark {
  color-scheme: dark;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #334155 100%);
  color: #e2e8f0;
}

.app-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #334155 100%);
}

/* Header styling */
.app-header {
  padding: 0;
  background: rgba(15, 23, 42, 0.95);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(30, 41, 59, 0.5);
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  z-index: 1000;
}

.nav-brand {
  display: flex;
  align-items: center;
  padding: 0 24px;
  gap: 12px;
  min-width: 250px;
}

.brand-icon {
  color: #60a5fa;
  font-size: 24px;
}

.brand-text {
  font-size: 20px;
  font-weight: 700;
  background: linear-gradient(135deg, #60a5fa, #3b82f6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.nav-menu {
  flex: 1;
  border: none !important;
  background: transparent !important;
}

.nav-item {
  border-radius: 8px !important;
  margin: 0 4px !important;
  transition: all 0.3s ease !important;
}

.nav-item:hover {
  background-color: rgba(96, 165, 250, 0.1) !important;
  transform: translateY(-1px);
}

.nav-item.is-active {
  background-color: rgba(96, 165, 250, 0.2) !important;
  border-bottom: 2px solid #60a5fa !important;
}

/* Admin nav item styling */
.admin-nav-item {
  position: relative;
}

.admin-nav-item::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 4px;
  height: 4px;
  background: #f59e0b;
  border-radius: 50%;
  transform: translate(-4px, 4px);
}

.admin-nav-item:hover {
  background-color: rgba(245, 158, 11, 0.1) !important;
  color: #f59e0b !important;
}

.admin-nav-item.is-active {
  background-color: rgba(245, 158, 11, 0.2) !important;
  border-bottom: 2px solid #f59e0b !important;
  color: #f59e0b !important;
}

.admin-nav-item .el-icon {
  color: #f59e0b;
}

.user-dropdown {
  margin-right: 24px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  color: #e5e7eb;
}

.user-info:hover {
  background-color: rgba(96, 165, 250, 0.1);
  color: #60a5fa;
}

.user-info.root-user {
  border: 1px solid rgba(245, 158, 11, 0.3);
  background-color: rgba(245, 158, 11, 0.1);
}

.user-info.root-user:hover {
  background-color: rgba(245, 158, 11, 0.2);
  border-color: rgba(245, 158, 11, 0.5);
}

.user-details {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}

.username {
  font-size: 14px;
  font-weight: 500;
  line-height: 1;
}

.user-role {
  font-size: 10px;
  font-weight: 700;
  color: #f59e0b;
  background-color: rgba(245, 158, 11, 0.2);
  padding: 1px 6px;
  border-radius: 4px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  line-height: 1;
}

.admin-dropdown-item {
  color: #f59e0b !important;
}

.admin-dropdown-item:hover {
  background-color: rgba(245, 158, 11, 0.1) !important;
}

.admin-dropdown-item .el-icon {
  color: #f59e0b;
}

.flex-grow {
  flex-grow: 1;
}

/* Main content area */
.app-main {
  padding: 0;
  min-height: calc(100vh - 60px);
  background: transparent;
}

.main-content {
  padding: 32px;
  max-width: 1400px;
  margin: 0 auto;
}

/* Global dark theme overrides for Element Plus */
.dark .el-card {
  background-color: rgba(30, 41, 59, 0.8) !important;
  border: 1px solid rgba(51, 65, 85, 0.5) !important;
  backdrop-filter: blur(10px);
  border-radius: 12px !important;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05) !important;
}

.dark .el-card__header {
  background-color: rgba(15, 23, 42, 0.5) !important;
  border-bottom: 1px solid rgba(51, 65, 85, 0.3) !important;
  padding: 20px 24px !important;
}

.dark .el-card__body {
  padding: 24px !important;
}

.dark .el-table {
  background-color: transparent !important;
  border-radius: 8px !important;
  overflow: hidden;
}

.dark .el-table th.el-table__cell {
  background-color: rgba(15, 23, 42, 0.8) !important;
  color: #f1f5f9 !important;
  border-bottom: 1px solid rgba(51, 65, 85, 0.5) !important;
  font-weight: 600;
}

.dark .el-table td.el-table__cell {
  background-color: rgba(30, 41, 59, 0.6) !important;
  border-bottom: 1px solid rgba(51, 65, 85, 0.3) !important;
  color: #e2e8f0 !important;
}

.dark .el-table tr:hover td {
  background-color: rgba(51, 65, 85, 0.5) !important;
}

.dark .el-button {
  border-radius: 8px !important;
  font-weight: 500;
  transition: all 0.3s ease !important;
}

.dark .el-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.dark .el-button--primary {
  background: linear-gradient(135deg, #3b82f6, #2563eb) !important;
  border: none !important;
}

.dark .el-button--primary:hover {
  background: linear-gradient(135deg, #2563eb, #1d4ed8) !important;
}

.dark .el-dialog {
  background-color: rgba(30, 41, 59, 0.95) !important;
  border: 1px solid rgba(51, 65, 85, 0.5) !important;
  border-radius: 12px !important;
  backdrop-filter: blur(10px);
}

.dark .el-dialog__header {
  background-color: rgba(15, 23, 42, 0.8) !important;
  border-bottom: 1px solid rgba(51, 65, 85, 0.3) !important;
  border-radius: 12px 12px 0 0 !important;
  padding: 20px 24px !important;
}

.dark .el-dialog__title {
  color: #f1f5f9 !important;
  font-weight: 600;
}

.dark .el-dialog__body {
  padding: 24px !important;
}

.dark .el-form-item__label {
  color: #cbd5e1 !important;
  font-weight: 500;
}

.dark .el-input__wrapper {
  background-color: rgba(15, 23, 42, 0.8) !important;
  border: 1px solid rgba(51, 65, 85, 0.5) !important;
  border-radius: 8px !important;
  box-shadow: none !important;
}

.dark .el-input__wrapper:hover {
  border-color: rgba(96, 165, 250, 0.5) !important;
}

.dark .el-input__wrapper.is-focus {
  border-color: #60a5fa !important;
  box-shadow: 0 0 0 2px rgba(96, 165, 250, 0.2) !important;
}

.dark .el-input__inner {
  color: #e2e8f0 !important;
}

.dark .el-select .el-input__wrapper {
  background-color: rgba(15, 23, 42, 0.8) !important;
}

.dark .el-tag {
  border-radius: 6px !important;
  font-weight: 500;
}

.dark .el-tag--success {
  background: rgba(34, 197, 94, 0.2) !important;
  border-color: rgba(34, 197, 94, 0.3) !important;
  color: #4ade80 !important;
}

.dark .el-tag--danger {
  background: rgba(239, 68, 68, 0.2) !important;
  border-color: rgba(239, 68, 68, 0.3) !important;
  color: #f87171 !important;
}

/* Dropdown menu styling */
.dark .el-dropdown-menu {
  background-color: rgba(30, 41, 59, 0.95) !important;
  border: 1px solid rgba(51, 65, 85, 0.5) !important;
  border-radius: 8px !important;
  backdrop-filter: blur(10px);
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05) !important;
}

.dark .el-dropdown-menu__item {
  color: #e2e8f0 !important;
  transition: all 0.2s ease;
}

.dark .el-dropdown-menu__item:hover {
  background-color: rgba(51, 65, 85, 0.5) !important;
  color: #60a5fa !important;
}

/* Responsive design */
@media (max-width: 768px) {
  .nav-brand {
    min-width: auto;
    padding: 0 16px;
  }
  
  .brand-text {
    display: none;
  }
  
  .main-content {
    padding: 16px;
  }
  
  .user-info span:not(.el-icon) {
    display: none;
  }
}

/* Loading and transitions */
.el-loading-mask {
  background-color: rgba(15, 23, 42, 0.8) !important;
  backdrop-filter: blur(4px);
}

/* Scrollbar styling */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: rgba(30, 41, 59, 0.5);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb {
  background: rgba(96, 165, 250, 0.5);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(96, 165, 250, 0.7);
}
</style> 