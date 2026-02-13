import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('../layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue')
      },
      {
        path: 'roster',
        name: 'MyRoster',
        component: () => import('../views/MyRoster.vue')
      },
      {
        path: 'generals',
        name: 'Generals',
        component: () => import('../views/Generals.vue')
      },
      {
        path: 'treasures',
        name: 'Treasures',
        component: () => import('../views/Treasures.vue')
      },
      {
        path: 'clubs',
        name: 'Clubs',
        component: () => import('../views/Clubs.vue')
      },
      {
        path: 'cities',
        name: 'Cities',
        component: () => import('../views/Cities.vue')
      },
      {
        path: 'rules',
        name: 'Rules',
        component: () => import('../views/Rules.vue')
      },
      {
        path: 'players',
        name: 'Players',
        component: () => import('../views/Players.vue')
      },
      {
        path: 'players/:id',
        name: 'PlayerDetail',
        component: () => import('../views/PlayerDetail.vue')
      },
      {
        path: 'draw',
        name: 'Draw',
        component: () => import('../views/Draw.vue')
      },
      {
        path: 'draft',
        name: 'Draft',
        component: () => import('../views/Draft.vue')
      },
      {
        path: 'trade',
        name: 'Trade',
        component: () => import('../views/Trade.vue')
      },
      {
        path: 'admin',
        name: 'Admin',
        component: () => import('../views/Admin.vue'),
        meta: { requiresAdmin: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Navigation guard
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  const isAuthenticated = userStore.isAuthenticated
  const isAdmin = userStore.isAdmin

  if (to.meta.requiresAuth && !isAuthenticated) {
    next('/login')
  } else if (to.meta.requiresAdmin && !isAdmin) {
    next('/')
  } else if ((to.path === '/login' || to.path === '/register') && isAuthenticated) {
    next('/')
  } else {
    next()
  }
})

export default router
