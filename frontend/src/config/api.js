// API configuration
const getApiBaseUrl = () => {
  // In production (when served by the backend), use relative URLs
  if (process.env.NODE_ENV === 'production') {
    return ''
  }
  
  // In development, use the backend URL
  return process.env.VUE_APP_API_URL || 'http://localhost:8080'
}

export const API_BASE_URL = getApiBaseUrl()

// Create axios instance with base configuration
import axios from 'axios'

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Add request interceptor to include auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Add response interceptor to handle auth errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Clear token and redirect to login
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default apiClient 