import MockAdapter from 'axios-mock-adapter'
import apiClient from '@/config/api'

describe('API Client Configuration', () => {
  let mockAxios

  beforeEach(() => {
    mockAxios = new MockAdapter(apiClient)
    localStorage.clear()
  })

  afterEach(() => {
    mockAxios.restore()
  })

  describe('Base Configuration', () => {
    it('should have correct base URL', () => {
      expect(apiClient.defaults.baseURL).toBe('http://localhost:8080')
    })

    it('should have correct timeout', () => {
      expect(apiClient.defaults.timeout).toBe(10000)
    })

    it('should have correct default headers', () => {
      expect(apiClient.defaults.headers.common['Content-Type']).toBe('application/json')
    })
  })

  describe('Request Interceptor', () => {
    it('should add authorization header when token exists', async () => {
      const token = 'test-jwt-token'
      localStorage.setItem('token', token)

      mockAxios.onGet('/test').reply(config => {
        expect(config.headers.Authorization).toBe(`Bearer ${token}`)
        return [200, { success: true }]
      })

      await apiClient.get('/test')
    })

    it('should not add authorization header when no token', async () => {
      mockAxios.onGet('/test').reply(config => {
        expect(config.headers.Authorization).toBeUndefined()
        return [200, { success: true }]
      })

      await apiClient.get('/test')
    })

    it('should preserve existing headers', async () => {
      const token = 'test-token'
      localStorage.setItem('token', token)

      const customHeaders = {
        'X-Custom-Header': 'custom-value'
      }

      mockAxios.onGet('/test').reply(config => {
        expect(config.headers.Authorization).toBe(`Bearer ${token}`)
        expect(config.headers['X-Custom-Header']).toBe('custom-value')
        return [200, { success: true }]
      })

      await apiClient.get('/test', { headers: customHeaders })
    })
  })

  describe('Response Interceptor', () => {
    it('should return response data on success', async () => {
      const responseData = { message: 'Success', data: [1, 2, 3] }
      mockAxios.onGet('/test').reply(200, responseData)

      const response = await apiClient.get('/test')
      expect(response.data).toEqual(responseData)
    })

    it('should handle 401 unauthorized errors', async () => {
      mockAxios.onGet('/test').reply(401, { error: 'Unauthorized' })

      try {
        await apiClient.get('/test')
      } catch (error) {
        expect(error.response.status).toBe(401)
        expect(localStorage.removeItem).toHaveBeenCalledWith('token')
      }
    })

    it('should handle 403 forbidden errors', async () => {
      mockAxios.onGet('/test').reply(403, { error: 'Forbidden' })

      try {
        await apiClient.get('/test')
      } catch (error) {
        expect(error.response.status).toBe(403)
        expect(error.response.data.error).toBe('Forbidden')
      }
    })

    it('should handle network errors', async () => {
      mockAxios.onGet('/test').networkError()

      try {
        await apiClient.get('/test')
      } catch (error) {
        expect(error.message).toContain('Network Error')
      }
    })

    it('should handle timeout errors', async () => {
      mockAxios.onGet('/test').timeout()

      try {
        await apiClient.get('/test')
      } catch (error) {
        expect(error.code).toBe('ECONNABORTED')
      }
    })

    it('should handle server errors (5xx)', async () => {
      mockAxios.onGet('/test').reply(500, { error: 'Internal Server Error' })

      try {
        await apiClient.get('/test')
      } catch (error) {
        expect(error.response.status).toBe(500)
        expect(error.response.data.error).toBe('Internal Server Error')
      }
    })

    it('should handle client errors (4xx)', async () => {
      mockAxios.onGet('/test').reply(400, { error: 'Bad Request' })

      try {
        await apiClient.get('/test')
      } catch (error) {
        expect(error.response.status).toBe(400)
        expect(error.response.data.error).toBe('Bad Request')
      }
    })
  })

  describe('HTTP Methods', () => {
    it('should support GET requests', async () => {
      const responseData = { data: 'test' }
      mockAxios.onGet('/users').reply(200, responseData)

      const response = await apiClient.get('/users')
      expect(response.data).toEqual(responseData)
    })

    it('should support POST requests', async () => {
      const requestData = { name: 'Test User' }
      const responseData = { id: 1, ...requestData }
      
      mockAxios.onPost('/users', requestData).reply(201, responseData)

      const response = await apiClient.post('/users', requestData)
      expect(response.data).toEqual(responseData)
    })

    it('should support PUT requests', async () => {
      const requestData = { id: 1, name: 'Updated User' }
      const responseData = { ...requestData }
      
      mockAxios.onPut('/users/1', requestData).reply(200, responseData)

      const response = await apiClient.put('/users/1', requestData)
      expect(response.data).toEqual(responseData)
    })

    it('should support DELETE requests', async () => {
      mockAxios.onDelete('/users/1').reply(204)

      const response = await apiClient.delete('/users/1')
      expect(response.status).toBe(204)
    })

    it('should support PATCH requests', async () => {
      const requestData = { name: 'Patched User' }
      const responseData = { id: 1, ...requestData }
      
      mockAxios.onPatch('/users/1', requestData).reply(200, responseData)

      const response = await apiClient.patch('/users/1', requestData)
      expect(response.data).toEqual(responseData)
    })
  })

  describe('Request Configuration', () => {
    it('should support custom headers per request', async () => {
      const customHeaders = {
        'X-Custom-Header': 'test-value'
      }

      mockAxios.onGet('/test').reply(config => {
        expect(config.headers['X-Custom-Header']).toBe('test-value')
        return [200, { success: true }]
      })

      await apiClient.get('/test', { headers: customHeaders })
    })

    it('should support query parameters', async () => {
      const params = { page: 1, limit: 10 }

      mockAxios.onGet('/users').reply(config => {
        expect(config.params).toEqual(params)
        return [200, { data: [] }]
      })

      await apiClient.get('/users', { params })
    })

    it('should support request timeout override', async () => {
      mockAxios.onGet('/test').reply(config => {
        expect(config.timeout).toBe(5000)
        return [200, { success: true }]
      })

      await apiClient.get('/test', { timeout: 5000 })
    })
  })

  describe('Error Response Structure', () => {
    it('should preserve error response structure', async () => {
      const errorResponse = {
        error: 'Validation failed',
        details: {
          username: 'Username is required',
          password: 'Password too short'
        }
      }

      mockAxios.onPost('/auth/register').reply(400, errorResponse)

      try {
        await apiClient.post('/auth/register', {})
      } catch (error) {
        expect(error.response.data).toEqual(errorResponse)
        expect(error.response.status).toBe(400)
      }
    })
  })

  describe('Content Type Handling', () => {
    it('should handle JSON requests', async () => {
      const data = { name: 'Test' }

      mockAxios.onPost('/test').reply(config => {
        expect(config.headers['Content-Type']).toBe('application/json')
        expect(JSON.parse(config.data)).toEqual(data)
        return [200, { success: true }]
      })

      await apiClient.post('/test', data)
    })

    it('should handle form data requests', async () => {
      const formData = new FormData()
      formData.append('file', 'test-file')

      mockAxios.onPost('/upload').reply(config => {
        expect(config.data).toBeInstanceOf(FormData)
        return [200, { success: true }]
      })

      await apiClient.post('/upload', formData)
    })
  })
}) 