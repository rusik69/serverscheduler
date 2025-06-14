<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <h2>Login</h2>
      </template>
      <el-form :model="form" :rules="rules" ref="loginForm" label-width="80px">
        <el-form-item label="Username" prop="username">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label="Password" prop="password">
          <el-input v-model="form.password" type="password" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleLogin" :loading="loading">Login</el-button>
          <el-button @click="$router.push('/register')">Register</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { useStore } from 'vuex'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

export default {
  name: 'Login',
  setup() {
    const store = useStore()
    const router = useRouter()
    const loginForm = ref(null)
    const loading = ref(false)

    const form = reactive({
      username: '',
      password: ''
    })

    const rules = {
      username: [
        { required: true, message: 'Please input username', trigger: 'blur' }
      ],
      password: [
        { required: true, message: 'Please input password', trigger: 'blur' }
      ]
    }

    const handleLogin = async () => {
      if (!loginForm.value) return

      try {
        await loginForm.value.validate()
        loading.value = true
        const success = await store.dispatch('auth/login', form)
        if (success) {
          ElMessage.success('Login successful')
          router.push('/')
        } else {
          ElMessage.error('Invalid credentials')
        }
      } catch (error) {
        console.error('Login error:', error)
        ElMessage.error('Login failed')
      } finally {
        loading.value = false
      }
    }

    return {
      form,
      rules,
      loginForm,
      loading,
      handleLogin
    }
  }
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 60px);
}

.login-card {
  width: 100%;
  max-width: 400px;
}
</style> 