import { createStore } from 'vuex'
import auth from '@/store/modules/auth'
import MockAdapter from 'axios-mock-adapter'
import apiClient from '@/config/api'

describe('Auth Store Module', () => {
  let store
  let mockAxios

  beforeEach(() => {
    mockAxios = new MockAdapter(apiClient)
    store = createStore({
      modules: {
        auth
      }
    })
    localStorage.clear()
  })

  afterEach(() => {
    mockAxios.restore()
  })

  describe('Getters', () => {
    it('should return correct isAuthenticated status when token exists', () => {
      store.commit('auth/SET_TOKEN', 'fake-token')
      expect(store.getters['auth/isAuthenticated']).toBe(true)
    })

    it('should return false for isAuthenticated when no token', () => {
      expect(store.getters['auth/isAuthenticated']).toBe(false)
    })

    it('should return user data', () => {
      const userData = { id: 1, username: 'testuser', role: 'user' }
      store.commit('auth/SET_USER', userData)
      expect(store.getters['auth/user']).toEqual(userData)
    })

    it('should return token', () => {
      const token = 'fake-token'
      store.commit('auth/SET_TOKEN', token)
      expect(store.getters['auth/token']).toBe(token)
    })
  })

  describe('Mutations', () => {
    it('should set token', () => {
      const token = 'test-token'
      store.commit('auth/SET_TOKEN', token)
      expect(store.state.auth.token).toBe(token)
    })

    it('should set user', () => {
      const user = { id: 1, username: 'testuser', role: 'admin' }
      store.commit('auth/SET_USER', user)
      expect(store.state.auth.user).toEqual(user)
    })

    it('should clear auth data', () => {
      store.commit('auth/SET_TOKEN', 'token')
      store.commit('auth/SET_USER', { id: 1 })
      store.commit('auth/CLEAR_AUTH')
      
      expect(store.state.auth.token).toBeNull()
      expect(store.state.auth.user).toBeNull()
    })
  })

  describe('Actions', () => {
    describe('login', () => {
      it('should login successfully with valid credentials', async () => {
        const loginData = { username: 'testuser', password: 'password' }
        const responseData = {
          token: 'fake-jwt-token',
          user: { id: 1, username: 'testuser', role: 'user' }
        }

        mockAxios.onPost('/api/auth/login').reply(200, responseData)

        await store.dispatch('auth/login', loginData)

        expect(store.state.auth.token).toBe(responseData.token)
        expect(store.state.auth.user).toEqual(responseData.user)
        expect(localStorage.setItem).toHaveBeenCalledWith('token', responseData.token)
      })

      it('should handle login failure', async () => {
        const loginData = { username: 'testuser', password: 'wrongpassword' }
        mockAxios.onPost('/api/auth/login').reply(401, { error: 'Invalid credentials' })

        await expect(store.dispatch('auth/login', loginData)).rejects.toThrow()
      })
    })

    describe('register', () => {
      it('should register successfully', async () => {
        const registerData = {
          username: 'newuser',
          password: 'password',
          email: 'test@example.com'
        }
        const responseData = {
          token: 'fake-jwt-token',
          user: { id: 2, username: 'newuser', role: 'user' }
        }

        mockAxios.onPost('/api/auth/register').reply(201, responseData)

        await store.dispatch('auth/register', registerData)

        expect(store.state.auth.token).toBe(responseData.token)
        expect(store.state.auth.user).toEqual(responseData.user)
      })

      it('should handle registration failure', async () => {
        const registerData = { username: 'existing', password: 'pass' }
        mockAxios.onPost('/api/auth/register').reply(400, { error: 'User already exists' })

        await expect(store.dispatch('auth/register', registerData)).rejects.toThrow()
      })
    })

    describe('logout', () => {
      it('should logout and clear auth data', async () => {
        store.commit('auth/SET_TOKEN', 'token')
        store.commit('auth/SET_USER', { id: 1 })

        await store.dispatch('auth/logout')

        expect(store.state.auth.token).toBeNull()
        expect(store.state.auth.user).toBeNull()
        expect(localStorage.removeItem).toHaveBeenCalledWith('token')
      })
    })

    describe('fetchUser', () => {
      it('should fetch user data successfully', async () => {
        const userData = { id: 1, username: 'testuser', role: 'admin' }
        mockAxios.onGet('/api/auth/user').reply(200, userData)

        await store.dispatch('auth/fetchUser')

        expect(store.state.auth.user).toEqual(userData)
      })

      it('should handle fetch user failure', async () => {
        mockAxios.onGet('/api/auth/user').reply(401, { error: 'Unauthorized' })

        await expect(store.dispatch('auth/fetchUser')).rejects.toThrow()
      })
    })

    describe('initializeAuth', () => {
      it('should initialize auth from localStorage token', async () => {
        const token = 'stored-token'
        const userData = { id: 1, username: 'testuser' }
        
        localStorage.getItem.mockReturnValue(token)
        mockAxios.onGet('/api/auth/user').reply(200, userData)

        await store.dispatch('auth/initializeAuth')

        expect(store.state.auth.token).toBe(token)
        expect(store.state.auth.user).toEqual(userData)
      })

      it('should not initialize if no token in localStorage', async () => {
        localStorage.getItem.mockReturnValue(null)

        await store.dispatch('auth/initializeAuth')

        expect(store.state.auth.token).toBeNull()
        expect(store.state.auth.user).toBeNull()
      })

      it('should clear auth if token is invalid', async () => {
        localStorage.getItem.mockReturnValue('invalid-token')
        mockAxios.onGet('/api/auth/user').reply(401)

        await store.dispatch('auth/initializeAuth')

        expect(store.state.auth.token).toBeNull()
        expect(store.state.auth.user).toBeNull()
        expect(localStorage.removeItem).toHaveBeenCalledWith('token')
      })
    })
  })
}) 