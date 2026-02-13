import axios from 'axios'
import { useUserStore } from '../stores/user'
import router from '../router'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor - add auth token
api.interceptors.request.use(
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

// Response interceptor - handle errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const userStore = useUserStore()
      userStore.logout()
      router.push('/login')
    }
    return Promise.reject(error)
  }
)

// Auth APIs
export const authApi = {
  register: (data) => api.post('/auth/register', data),
  login: (data) => api.post('/auth/login', data),
  getCurrentUser: () => api.get('/me'),
  updateProfile: (data) => api.put('/me', data),
  getMyRoster: () => api.get('/me/roster'),
  getMyDrawRecords: () => api.get('/me/draws'),
  getMyDraftRecords: () => api.get('/me/drafts')
}

// Game APIs
export const gameApi = {
  getPhase: () => api.get('/phase'),
  signUp: () => api.post('/signup'),
  getPlayers: () => api.get('/players'),
  getPlayerRoster: (id) => api.get(`/players/${id}/roster`),
  getStatistics: () => api.get('/statistics')
}

// Draw APIs
export const drawApi = {
  guaranteeDraw: () => api.post('/draw/guarantee'),
  normalDraw: () => api.post('/draw/normal'),
  getDrawPool: (type) => api.get(`/draw/pool/${type}`),
  getDraftPool: () => api.get('/draft/pool'),
  draftPick: (generalId) => api.post('/draft/pick', { general_id: generalId })
}

// Asset APIs
export const assetApi = {
  getAllGenerals: () => api.get('/generals'),
  getGeneral: (id) => api.get(`/generals/${id}`),
  getAllTreasures: () => api.get('/treasures'),
  getTreasure: (id) => api.get(`/treasures/${id}`),
  getAllClubs: () => api.get('/clubs'),
  getClub: (id) => api.get(`/clubs/${id}`),
  getClubDetail: (id) => api.get(`/clubs/${id}/detail`),
  getAllCities: () => api.get('/cities'),
  getGameRules: () => api.get('/rules')
}

// Trade APIs
export const tradeApi = {
  createTrade: (data) => api.post('/trades', data),
  getPendingTrades: () => api.get('/trades/pending'),
  getTradeHistory: () => api.get('/trades/history'),
  getTrade: (id) => api.get(`/trades/${id}`),
  acceptTrade: (id) => api.post(`/trades/${id}/accept`),
  rejectTrade: (id) => api.post(`/trades/${id}/reject`),
  cancelTrade: (id) => api.post(`/trades/${id}/cancel`)
}

// Admin APIs
export const adminApi = {
  setPhase: (data) => api.post('/admin/phase', data),
  resetSeason: () => api.post('/admin/reset'),
  getAllTrades: () => api.get('/admin/trades'),
  importData: (formData) => api.post('/admin/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export default api
