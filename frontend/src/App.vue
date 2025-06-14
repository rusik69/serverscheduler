<template>
  <el-container class="app-container">
    <el-header>
      <el-menu mode="horizontal" router>
        <el-menu-item index="/">Server Scheduler</el-menu-item>
        <el-menu-item index="/servers">Servers</el-menu-item>
        <el-menu-item index="/reservations">Reservations</el-menu-item>
        <div class="flex-grow" />
        <el-menu-item v-if="!isAuthenticated" index="/login">Login</el-menu-item>
        <el-menu-item v-if="!isAuthenticated" index="/register">Register</el-menu-item>
        <el-menu-item v-if="isAuthenticated" @click="logout">Logout</el-menu-item>
      </el-menu>
    </el-header>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
</template>

<script>
import { computed } from 'vue'
import { useStore } from 'vuex'
import { useRouter } from 'vue-router'

export default {
  name: 'App',
  setup() {
    const store = useStore()
    const router = useRouter()

    const isAuthenticated = computed(() => store.getters['auth/isAuthenticated'])

    const logout = async () => {
      await store.dispatch('auth/logout')
      router.push('/login')
    }

    return {
      isAuthenticated,
      logout
    }
  }
}
</script>

<style>
.app-container {
  min-height: 100vh;
}

.el-header {
  padding: 0;
}

.flex-grow {
  flex-grow: 1;
}

.el-main {
  padding: 20px;
}
</style> 