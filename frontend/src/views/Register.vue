<template>
  <div class="register-container">
    <el-card class="register-card">
      <template #header>
        <h2>Register</h2>
      </template>
      <el-form :model="form" :rules="rules" ref="registerForm" label-width="80px">
        <el-form-item label="Username" prop="username">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label="Password" prop="password">
          <el-input v-model="form.password" type="password" />
        </el-form-item>
        <el-form-item label="Confirm" prop="confirmPassword">
          <el-input v-model="form.confirmPassword" type="password" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleRegister" :loading="loading">Register</el-button>
          <el-button @click="$router.push('/login')">Login</el-button>
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
  name: 'Register',
  setup() {
    const store = useStore()
    const router = useRouter()
    const registerForm = ref(null)
    const loading = ref(false)

    const form = reactive({
      username: '',
      password: '',
      confirmPassword: ''
    })

    const validatePass = (rule, value, callback) => {
      if (value === '') {
        callback(new Error('Please input the password'))
      } else {
        if (form.confirmPassword !== '') {
          registerForm.value?.validateField('confirmPassword')
        }
        callback()
      }
    }

    const validatePass2 = (rule, value, callback) => {
      if (value === '') {
        callback(new Error('Please input the password again'))
      } else if (value !== form.password) {
        callback(new Error('Two inputs don\'t match!'))
      } else {
        callback()
      }
    }

    const rules = {
      username: [
        { required: true, message: 'Please input username', trigger: 'blur' }
      ],
      password: [
        { required: true, validator: validatePass, trigger: 'blur' }
      ],
      confirmPassword: [
        { required: true, validator: validatePass2, trigger: 'blur' }
      ]
    }

    const handleRegister = async () => {
      if (!registerForm.value) return

      try {
        await registerForm.value.validate()
        loading.value = true
        const success = await store.dispatch('auth/register', {
          username: form.username,
          password: form.password
        })
        if (success) {
          ElMessage.success('Registration successful')
          router.push('/')
        } else {
          ElMessage.error('Registration failed')
        }
      } catch (error) {
        console.error('Registration error:', error)
        ElMessage.error('Registration failed')
      } finally {
        loading.value = false
      }
    }

    return {
      form,
      rules,
      registerForm,
      loading,
      handleRegister
    }
  }
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 60px);
}

.register-card {
  width: 100%;
  max-width: 400px;
}
</style> 