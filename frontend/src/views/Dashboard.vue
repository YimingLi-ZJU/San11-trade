<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <!-- User Status Card -->
      <el-col :span="8">
        <el-card class="status-card">
          <template #header>
            <div class="card-header">
              <el-icon><User /></el-icon>
              <span>我的状态</span>
            </div>
          </template>
          <div class="user-status" v-if="userStore.user">
            <div class="status-item">
              <span class="label">昵称</span>
              <span class="value">{{ userStore.user.nickname }}</span>
            </div>
            <div class="status-item">
              <span class="label">报名状态</span>
              <el-tag :type="userStore.isRegistered ? 'success' : 'info'">
                {{ userStore.isRegistered ? '已报名' : '未报名' }}
              </el-tag>
            </div>
            <div class="status-item">
              <span class="label">剩余空间</span>
              <span class="value">{{ userStore.remainingSpace }} / {{ userStore.user.space }}</span>
            </div>
            <el-button
              v-if="!userStore.isRegistered && phase?.current_phase === 'signup'"
              type="primary"
              @click="handleSignUp"
              :loading="signingUp"
              style="width: 100%; margin-top: 16px"
            >
              立即报名
            </el-button>
          </div>
        </el-card>
      </el-col>

      <!-- Game Phase Card -->
      <el-col :span="8">
        <el-card class="status-card">
          <template #header>
            <div class="card-header">
              <el-icon><Timer /></el-icon>
              <span>游戏阶段</span>
            </div>
          </template>
          <div class="phase-info" v-if="phase">
            <div class="phase-name">
              {{ gameStore.getPhaseName(phase.current_phase) }}
            </div>
            <el-progress
              :percentage="getPhaseProgress(phase.current_phase)"
              :stroke-width="12"
              status="success"
            />
            <div class="phase-hint">
              {{ getPhaseHint(phase.current_phase) }}
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- Statistics Card -->
      <el-col :span="8">
        <el-card class="status-card">
          <template #header>
            <div class="card-header">
              <el-icon><DataAnalysis /></el-icon>
              <span>统计数据</span>
            </div>
          </template>
          <div class="stats" v-if="stats">
            <div class="stat-item">
              <span class="stat-value">{{ stats.registered_players }}</span>
              <span class="stat-label">已报名玩家</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ stats.owned_generals }}</span>
              <span class="stat-label">已归属武将</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ stats.accepted_trades }}</span>
              <span class="stat-label">完成交易</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Quick Actions -->
    <el-card class="quick-actions" style="margin-top: 20px">
      <template #header>
        <div class="card-header">
          <el-icon><Operation /></el-icon>
          <span>快捷操作</span>
        </div>
      </template>
      <el-row :gutter="20">
        <el-col :span="6">
          <el-button
            size="large"
            style="width: 100%; height: 80px"
            @click="$router.push('/draw')"
            :disabled="!canDraw"
          >
            <div class="action-btn">
              <el-icon size="24"><MagicStick /></el-icon>
              <span>抽将</span>
            </div>
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button
            size="large"
            style="width: 100%; height: 80px"
            @click="$router.push('/draft')"
            :disabled="phase?.current_phase !== 'draft'"
          >
            <div class="action-btn">
              <el-icon size="24"><Select /></el-icon>
              <span>选秀</span>
            </div>
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button
            size="large"
            style="width: 100%; height: 80px"
            @click="$router.push('/trade')"
            :disabled="!canTrade"
          >
            <div class="action-btn">
              <el-icon size="24"><Switch /></el-icon>
              <span>交易</span>
            </div>
          </el-button>
        </el-col>
        <el-col :span="6">
          <el-button
            size="large"
            style="width: 100%; height: 80px"
            @click="$router.push('/roster')"
          >
            <div class="action-btn">
              <el-icon size="24"><List /></el-icon>
              <span>我的阵容</span>
            </div>
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <!-- My Generals Preview -->
    <el-card style="margin-top: 20px" v-if="roster">
      <template #header>
        <div class="card-header">
          <el-icon><Avatar /></el-icon>
          <span>我的武将 ({{ roster.generals?.length || 0 }})</span>
          <el-button text type="primary" @click="$router.push('/roster')">
            查看全部
          </el-button>
        </div>
      </template>
      <el-empty v-if="!roster.generals?.length" description="暂无武将" />
      <el-row :gutter="10" v-else>
        <el-col
          v-for="general in roster.generals.slice(0, 8)"
          :key="general.id"
          :span="3"
        >
          <div class="general-preview">
            <div class="general-avatar">{{ general.name[0] }}</div>
            <div class="general-name">{{ general.name }}</div>
          </div>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../stores/user'
import { useGameStore } from '../stores/game'
import { authApi, gameApi } from '../api'

const userStore = useUserStore()
const gameStore = useGameStore()

const phase = ref(null)
const stats = ref(null)
const roster = ref(null)
const signingUp = ref(false)

const canDraw = computed(() => {
  return ['guarantee_draw', 'normal_draw'].includes(phase.value?.current_phase)
})

const canTrade = computed(() => {
  return ['trading', 'draft'].includes(phase.value?.current_phase)
})

onMounted(async () => {
  try {
    const [phaseRes, statsRes, rosterRes] = await Promise.all([
      gameApi.getPhase(),
      gameApi.getStatistics(),
      authApi.getMyRoster().catch(() => ({ data: null }))
    ])
    phase.value = phaseRes.data
    stats.value = statsRes.data
    roster.value = rosterRes.data
  } catch (error) {
    console.error('Failed to load dashboard data:', error)
  }
})

async function handleSignUp() {
  signingUp.value = true
  try {
    await userStore.signUp()
    ElMessage.success('报名成功！')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '报名失败')
  } finally {
    signingUp.value = false
  }
}

function getPhaseProgress(phaseName) {
  const phases = ['signup', 'guarantee_draw', 'normal_draw', 'draft', 'trading', 'match', 'finished']
  const index = phases.indexOf(phaseName)
  return Math.round(((index + 1) / phases.length) * 100)
}

function getPhaseHint(phaseName) {
  const hints = {
    signup: '报名参加本赛季联赛',
    guarantee_draw: '使用保底抽将机会',
    normal_draw: '使用普通抽将机会',
    draft: '选秀阶段，按顺序选择武将',
    trading: '自由交易阶段，与其他玩家交换武将',
    auction: '拍卖阶段',
    match: '比赛进行中',
    finished: '本赛季已结束'
  }
  return hints[phaseName] || ''
}
</script>

<style scoped>
.dashboard {
  max-width: 1400px;
  margin: 0 auto;
}

.status-card {
  height: 100%;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.card-header .el-button {
  margin-left: auto;
}

.user-status .status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.user-status .status-item:last-child {
  border-bottom: none;
}

.user-status .label {
  color: #666;
}

.user-status .value {
  font-weight: 600;
  color: #333;
}

.phase-info {
  text-align: center;
}

.phase-name {
  font-size: 24px;
  font-weight: bold;
  color: #667eea;
  margin-bottom: 20px;
}

.phase-hint {
  margin-top: 16px;
  color: #999;
  font-size: 14px;
}

.stats {
  display: flex;
  justify-content: space-around;
}

.stat-item {
  text-align: center;
}

.stat-value {
  display: block;
  font-size: 32px;
  font-weight: bold;
  color: #667eea;
}

.stat-label {
  color: #999;
  font-size: 14px;
}

.action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.general-preview {
  text-align: center;
  padding: 10px;
}

.general-avatar {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: bold;
  margin: 0 auto 8px;
}

.general-name {
  font-size: 12px;
  color: #666;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
