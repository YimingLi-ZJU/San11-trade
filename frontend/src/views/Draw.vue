<template>
  <div class="draw-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>抽将</span>
          <el-tag :type="phaseType">{{ phaseText }}</el-tag>
        </div>
      </template>

      <el-alert
        v-if="!canDraw"
        type="warning"
        title="当前不在抽将阶段"
        description="请等待管理员开启抽将阶段"
        show-icon
        :closable="false"
        style="margin-bottom: 20px"
      />

      <el-row :gutter="40">
        <!-- Guarantee Draw -->
        <el-col :span="12">
          <div class="draw-section">
            <h3>保底抽将</h3>
            <p class="description">保底池武将，三抽必出高阶武将</p>
            <div class="draw-count">
              已抽次数：{{ guaranteeCount }} / 3
            </div>
            <el-button
              type="primary"
              size="large"
              :disabled="!canGuarantee"
              :loading="drawing === 'guarantee'"
              @click="handleDraw('guarantee')"
              style="width: 100%; margin-top: 16px"
            >
              保底抽将
            </el-button>
          </div>
        </el-col>

        <!-- Normal Draw -->
        <el-col :span="12">
          <div class="draw-section">
            <h3>普通抽将</h3>
            <p class="description">普通池武将，随机获得一名武将</p>
            <div class="draw-count">
              已抽次数：{{ normalCount }} / 7
            </div>
            <el-button
              type="success"
              size="large"
              :disabled="!canNormal"
              :loading="drawing === 'normal'"
              @click="handleDraw('normal')"
              style="width: 100%; margin-top: 16px"
            >
              普通抽将
            </el-button>
          </div>
        </el-col>
      </el-row>

      <!-- Draw Result -->
      <el-dialog v-model="showResult" title="抽将结果" width="400px" center>
        <div class="draw-result" v-if="drawResult">
          <div class="general-avatar">{{ drawResult.name[0] }}</div>
          <h2>{{ drawResult.name }}</h2>
          <el-tag :type="getTierType(drawResult.tier)" size="large">
            T{{ drawResult.tier }}
          </el-tag>
          <div class="stats">
            <div class="stat">
              <span class="label">统率</span>
              <span class="value">{{ drawResult.command }}</span>
            </div>
            <div class="stat">
              <span class="label">武力</span>
              <span class="value">{{ drawResult.force }}</span>
            </div>
            <div class="stat">
              <span class="label">智力</span>
              <span class="value">{{ drawResult.intelligence }}</span>
            </div>
            <div class="stat">
              <span class="label">政治</span>
              <span class="value">{{ drawResult.politics }}</span>
            </div>
            <div class="stat">
              <span class="label">魅力</span>
              <span class="value">{{ drawResult.charm }}</span>
            </div>
          </div>
          <p class="skills" v-if="drawResult.skills">特技：{{ drawResult.skills }}</p>
        </div>
      </el-dialog>

      <!-- Draw History -->
      <el-divider />
      <h4>抽将记录</h4>
      <el-table :data="drawRecords" stripe v-loading="loadingRecords">
        <el-table-column prop="general.name" label="武将" width="100" />
        <el-table-column prop="draw_type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.draw_type === 'guarantee' ? 'warning' : 'success'">
              {{ row.draw_type === 'guarantee' ? '保底' : '普通' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="general.salary" label="薪资" width="80" />
        <el-table-column prop="general.skills" label="特技" />
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useGameStore } from '../stores/game'
import { useUserStore } from '../stores/user'
import { drawApi, authApi } from '../api'

const gameStore = useGameStore()
const userStore = useUserStore()

const drawing = ref(null)
const showResult = ref(false)
const drawResult = ref(null)
const drawRecords = ref([])
const loadingRecords = ref(false)

const guaranteeCount = computed(() => 
  drawRecords.value.filter(r => r.draw_type === 'guarantee').length
)
const normalCount = computed(() =>
  drawRecords.value.filter(r => r.draw_type === 'normal').length
)

const canDraw = computed(() =>
  ['guarantee_draw', 'normal_draw'].includes(gameStore.phase?.current_phase)
)
const canGuarantee = computed(() =>
  gameStore.phase?.current_phase === 'guarantee_draw' && guaranteeCount.value < 3
)
const canNormal = computed(() =>
  gameStore.phase?.current_phase === 'normal_draw' && normalCount.value < 7
)

const phaseText = computed(() => gameStore.getPhaseName(gameStore.phase?.current_phase))
const phaseType = computed(() => canDraw.value ? 'success' : 'info')

onMounted(async () => {
  await gameStore.fetchPhase()
  await loadRecords()
})

async function loadRecords() {
  loadingRecords.value = true
  try {
    const response = await authApi.getMyDrawRecords()
    drawRecords.value = response.data || []
  } catch (error) {
    console.error('Failed to load draw records:', error)
  } finally {
    loadingRecords.value = false
  }
}

async function handleDraw(type) {
  drawing.value = type
  try {
    const api = type === 'guarantee' ? drawApi.guaranteeDraw : drawApi.normalDraw
    const response = await api()
    drawResult.value = response.data.general
    showResult.value = true
    await loadRecords()
    await userStore.fetchUser()
    ElMessage.success('抽将成功！')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '抽将失败')
  } finally {
    drawing.value = null
  }
}

function getTierType(tier) {
  const types = { 1: 'danger', 2: 'warning', 3: 'primary', 4: 'success', 5: 'info' }
  return types[tier] || 'info'
}

function formatTime(time) {
  return new Date(time).toLocaleString('zh-CN')
}
</script>

<style scoped>
.draw-page {
  max-width: 1000px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.draw-section {
  text-align: center;
  padding: 30px;
  background: #f5f7fa;
  border-radius: 12px;
}

.draw-section h3 {
  margin-bottom: 12px;
  color: #333;
}

.draw-section .description {
  color: #666;
  font-size: 14px;
}

.draw-count {
  margin-top: 20px;
  font-size: 18px;
  font-weight: bold;
  color: #667eea;
}

.draw-result {
  text-align: center;
}

.draw-result .general-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 36px;
  font-weight: bold;
  margin: 0 auto 16px;
}

.draw-result h2 {
  margin-bottom: 12px;
}

.draw-result .stats {
  display: flex;
  justify-content: center;
  gap: 16px;
  margin-top: 20px;
}

.draw-result .stat {
  text-align: center;
}

.draw-result .stat .label {
  display: block;
  font-size: 12px;
  color: #999;
}

.draw-result .stat .value {
  font-size: 20px;
  font-weight: bold;
  color: #333;
}

.draw-result .skills {
  margin-top: 16px;
  color: #666;
}
</style>
