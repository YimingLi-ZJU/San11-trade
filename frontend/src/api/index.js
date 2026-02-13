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
  getStatistics: () => api.get('/statistics'),
  getRegistrationConfig: () => api.get('/config/registration')
}

// Draw APIs (unified)
export const drawApi = {
  draw: () => api.post('/draw'),
  getStatus: () => api.get('/draw/status'),
  getResults: () => api.get('/draw/results'),
  getPool: (type) => api.get(`/draw/pool${type ? '?type=' + type : ''}`),
  // Draft
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

// Auction APIs
export const auctionApi = {
  getPool: () => api.get('/auction/pool'),
  getResults: () => api.get('/auction/results'),
  getStats: () => api.get('/auction/stats')
}

// Policy APIs (国策拍卖)
export const policyApi = {
  // Player APIs
  getStatus: () => api.get('/policy/status'),
  getMyBid: () => api.get('/policy/my-bid'),
  placeBid: (bidAmount) => api.post('/policy/bid', { bid_amount: bidAmount }),
  setPreferences: (clubIds) => api.post('/policy/preferences', { club_ids: clubIds }),
  selectClub: (clubId) => api.post('/policy/select', { club_id: clubId }),
  getResults: () => api.get('/policy/results'),
  getClubs: (params) => api.get('/policy/clubs', { params }),
  getFilters: () => api.get('/policy/filters')
}

// Admin APIs
export const adminApi = {
  setPhase: (data) => api.post('/admin/phase', data),
  resetSeason: () => api.post('/admin/reset'),
  getAllTrades: () => api.get('/admin/trades'),
  importData: (formData) => api.post('/admin/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  // Invite code management
  generateInviteCodes: (data) => api.post('/admin/invite-codes', data),
  getInviteCodes: (page = 1, pageSize = 20) => api.get(`/admin/invite-codes?page=${page}&page_size=${pageSize}`),
  getInviteCodeStats: () => api.get('/admin/invite-codes/stats'),
  deleteInviteCode: (id) => api.delete(`/admin/invite-codes/${id}`),
  getInviteCodeUsages: (id) => api.get(`/admin/invite-codes/${id}/usages`),
  // Draw management
  resetUserDraw: (userId) => api.post(`/admin/draw/reset/${userId}`),
  resetAllDraw: () => api.post('/admin/draw/reset-all'),
  drawForUser: (userId) => api.post(`/admin/draw/for/${userId}`),
  drawForAll: () => api.post('/admin/draw/for-all'),
  // Auction management
  getAuctionStats: () => api.get('/admin/auction/stats'),
  assignAuction: (data) => api.post('/admin/auction/assign', data),
  resetAuction: (generalId) => api.post(`/admin/auction/reset/${generalId}`),
  // Policy management (国策管理)
  closePolicyBidding: () => api.post('/admin/policy/close-bidding'),
  startPolicySelection: (data) => api.post('/admin/policy/start-selection', data),
  getPolicyBids: () => api.get('/admin/policy/bids'),
  resetPolicyPhase: () => api.post('/admin/policy/reset'),
  resetUserPolicySelection: (userId) => api.post(`/admin/policy/reset-user/${userId}`),
  selectClubForUser: (userId, clubId) => api.post(`/admin/policy/select-for/${userId}`, { club_id: clubId }),
  checkPolicyTimeout: () => api.post('/admin/policy/check-timeout'),
  forceNextSelector: () => api.post('/admin/policy/force-next')
}

// Invite code APIs (public)
export const inviteCodeApi = {
  validate: (code) => api.get(`/invite-codes/validate?code=${code}`)
}

export default api
