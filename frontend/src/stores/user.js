import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, gameApi } from '../api'

export const useUserStore = defineStore('user', () => {
  const user = ref(null)
  const token = ref(localStorage.getItem('token') || null)

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.is_admin || false)
  const isRegistered = computed(() => user.value?.is_registered || false)
  const remainingSpace = computed(() => {
    if (!user.value) return 0
    return user.value.space - user.value.used_space
  })

  async function login(credentials) {
    const response = await authApi.login(credentials)
    token.value = response.data.token
    user.value = response.data.user
    localStorage.setItem('token', response.data.token)
    return response.data
  }

  async function register(data) {
    const response = await authApi.register(data)
    token.value = response.data.token
    user.value = response.data.user
    localStorage.setItem('token', response.data.token)
    return response.data
  }

  async function fetchUser() {
    if (!token.value) return null
    try {
      const response = await authApi.getCurrentUser()
      user.value = response.data
      return response.data
    } catch (error) {
      logout()
      throw error
    }
  }

  async function signUp() {
    const response = await gameApi.signUp()
    await fetchUser() // Refresh user data
    return response.data
  }

  function logout() {
    user.value = null
    token.value = null
    localStorage.removeItem('token')
  }

  // Initialize user if token exists
  if (token.value) {
    fetchUser().catch(() => {})
  }

  return {
    user,
    token,
    isAuthenticated,
    isAdmin,
    isRegistered,
    remainingSpace,
    login,
    register,
    fetchUser,
    signUp,
    logout
  }
})
