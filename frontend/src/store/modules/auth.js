import apiClient from '@/config/api'

const state = {
  token: localStorage.getItem('token') || '',
  user: (() => {
    try {
      const userStr = localStorage.getItem('user')
      return userStr ? JSON.parse(userStr) : null
    } catch (e) {
      return null
    }
  })()
}

const getters = {
  isAuthenticated: state => !!state.token,
  currentUser: state => state.user,
  user: state => state.user
}

const actions = {
  async login({ commit }, credentials) {
    try {
      const response = await apiClient.post('/api/auth/login', credentials)
      const { token, user } = response.data
      localStorage.setItem('token', token)
      localStorage.setItem('user', JSON.stringify(user))
      commit('SET_AUTH', { token, user })
      return true
    } catch (error) {
      console.error('Login error:', error)
      return false
    }
  },

  async register({ commit }, userData) {
    try {
      const response = await apiClient.post('/api/auth/register', userData)
      const { token, user } = response.data
      localStorage.setItem('token', token)
      localStorage.setItem('user', JSON.stringify(user))
      commit('SET_AUTH', { token, user })
      return true
    } catch (error) {
      console.error('Registration error:', error)
      return false
    }
  },

  logout({ commit }) {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    commit('CLEAR_AUTH')
  }
}

const mutations = {
  SET_AUTH(state, { token, user }) {
    state.token = token
    state.user = user
  },
  CLEAR_AUTH(state) {
    state.token = ''
    state.user = null
  }
}

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
} 