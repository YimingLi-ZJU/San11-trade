<template>
  <el-container class="main-layout">
    <el-header class="header">
      <div class="logo">
        <el-icon size="24"><Trophy /></el-icon>
        <span>三国志11交易系统</span>
      </div>
      <div class="header-right">
        <el-tag v-if="phase" type="warning" size="large">
          {{ gameStore.getPhaseName(phase.current_phase) }}
        </el-tag>
        <el-dropdown @command="handleCommand">
          <span class="user-info">
            <el-avatar :size="32">{{ userStore.user?.nickname?.[0] || 'U' }}</el-avatar>
            <span class="username">{{ userStore.user?.nickname || '用户' }}</span>
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">
                <el-icon><User /></el-icon>个人信息
              </el-dropdown-item>
              <el-dropdown-item command="roster">
                <el-icon><List /></el-icon>我的阵容
              </el-dropdown-item>
              <el-dropdown-item v-if="userStore.isAdmin" command="admin" divided>
                <el-icon><Setting /></el-icon>管理后台
              </el-dropdown-item>
              <el-dropdown-item command="logout" divided>
                <el-icon><SwitchButton /></el-icon>退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>
    
    <el-container>
      <el-aside width="220px" class="aside">
        <el-menu
          :default-active="$route.path"
          router
          class="menu"
        >
          <el-menu-item index="/">
            <el-icon><House /></el-icon>
            <span>首页</span>
          </el-menu-item>
          
          <el-sub-menu index="game">
            <template #title>
              <el-icon><Opportunity /></el-icon>
              <span>游戏</span>
            </template>
            <el-menu-item index="/initial-draw">
              <el-icon><Aim /></el-icon>初抽
            </el-menu-item>
            <el-menu-item index="/draw">
              <el-icon><MagicStick /></el-icon>抽将
            </el-menu-item>
            <el-menu-item index="/draft">
              <el-icon><Select /></el-icon>选秀
            </el-menu-item>
            <el-menu-item index="/trade">
              <el-icon><Switch /></el-icon>交易
            </el-menu-item>
          </el-sub-menu>
          
          <el-sub-menu index="assets">
            <template #title>
              <el-icon><Folder /></el-icon>
              <span>资产</span>
            </template>
            <el-menu-item index="/roster">
              <el-icon><UserFilled /></el-icon>我的阵容
            </el-menu-item>
            <el-menu-item index="/generals">
              <el-icon><Avatar /></el-icon>武将列表
            </el-menu-item>
            <el-menu-item index="/treasures">
              <el-icon><Present /></el-icon>宝物列表
            </el-menu-item>
            <el-menu-item index="/clubs">
              <el-icon><Flag /></el-icon>俱乐部
            </el-menu-item>
            <el-menu-item index="/cities">
              <el-icon><OfficeBuilding /></el-icon>城市列表
            </el-menu-item>
            <el-menu-item index="/rules">
              <el-icon><Document /></el-icon>游戏规则
            </el-menu-item>
          </el-sub-menu>
          
          <el-menu-item index="/players">
            <el-icon><User /></el-icon>
            <span>玩家列表</span>
          </el-menu-item>
          
          <el-menu-item v-if="userStore.isAdmin" index="/admin">
            <el-icon><Setting /></el-icon>
            <span>管理后台</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { useGameStore } from '../stores/game'

const router = useRouter()
const userStore = useUserStore()
const gameStore = useGameStore()

const phase = ref(null)

onMounted(async () => {
  try {
    await userStore.fetchUser()
    phase.value = await gameStore.fetchPhase()
  } catch (error) {
    console.error('Failed to load data:', error)
  }
})

function handleCommand(command) {
  switch (command) {
    case 'profile':
      router.push('/roster')
      break
    case 'roster':
      router.push('/roster')
      break
    case 'admin':
      router.push('/admin')
      break
    case 'logout':
      userStore.logout()
      router.push('/login')
      break
  }
}
</script>

<style scoped>
.main-layout {
  min-height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 0 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 20px;
  font-weight: bold;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: white;
}

.username {
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.aside {
  background-color: #fff;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.05);
}

.menu {
  border-right: none;
  height: 100%;
}

.main {
  background-color: #f5f7fa;
  padding: 20px;
}
</style>
