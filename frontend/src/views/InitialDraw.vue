<template>
  <div class="initial-draw-page">
    <!-- My Draw Status -->
    <el-card class="status-card">
      <template #header>
        <div class="card-header">
          <span>我的初抽</span>
          <el-tag :type="statusTagType" size="large">
            {{ gameStore.getPhaseName(phase?.current_phase) }}
          </el-tag>
        </div>
      </template>

      <el-row :gutter="20">
        <el-col :span="8">
          <el-statistic title="保底剩余" :value="status?.guarantee_remaining || 0">
            <template #suffix>
              <span class="stat-total"> / 3</span>
            </template>
          </el-statistic>
        </el-col>
        <el-col :span="8">
          <el-statistic title="普通剩余" :value="status?.normal_remaining || 0">
            <template #suffix>
              <span class="stat-total"> / 7</span>
            </template>
          </el-statistic>
        </el-col>
        <el-col :span="8">
          <el-statistic title="总计剩余" :value="status?.total_remaining || 0">
            <template #suffix>
              <span class="stat-total"> / 10</span>
            </template>
          </el-statistic>
        </el-col>
      </el-row>

      <div class="draw-action">
        <el-button 
          type="primary" 
          size="large"
          :disabled="!canDraw"
          :loading="drawing"
          @click="handleDraw"
        >
          {{ drawButtonText }}
        </el-button>
        <div class="draw-tip">
          <span v-if="status?.guarantee_remaining > 0">
            下一抽：<el-tag type="danger">保底池</el-tag>
          </span>
          <span v-else-if="status?.normal_remaining > 0">
            下一抽：<el-tag type="warning">普通池</el-tag>
          </span>
        </div>
      </div>
    </el-card>

    <!-- Draw Result Dialog -->
    <el-dialog v-model="showResult" title="抽将结果" width="500px" center>
      <div v-if="lastDrawResult" class="draw-result">
        <div class="general-avatar">
          {{ lastDrawResult.name[0] }}
        </div>
        <h2>{{ lastDrawResult.name }}</h2>
        <el-tag :type="lastDrawType === 'initial_guarantee' ? 'danger' : 'warning'" size="large">
          {{ lastDrawType === 'initial_guarantee' ? '保底' : '普通' }}
        </el-tag>
        <div class="general-stats">
          <el-row :gutter="10">
            <el-col :span="8">统率：{{ lastDrawResult.command }}</el-col>
            <el-col :span="8">武力：{{ lastDrawResult.force }}</el-col>
            <el-col :span="8">智力：{{ lastDrawResult.intelligence }}</el-col>
          </el-row>
          <el-row :gutter="10" style="margin-top: 8px">
            <el-col :span="8">政治：{{ lastDrawResult.politics }}</el-col>
            <el-col :span="8">魅力：{{ lastDrawResult.charm }}</el-col>
            <el-col :span="8">薪资：{{ lastDrawResult.salary }}</el-col>
          </el-row>
        </div>
        <div class="general-skills" v-if="lastDrawResult.skills">
          <el-tag v-for="skill in lastDrawResult.skills.split(' ')" :key="skill" size="small">
            {{ skill }}
          </el-tag>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="showResult = false">确定</el-button>
        <el-button 
          v-if="status?.total_remaining > 0" 
          type="success" 
          @click="handleDrawAgain"
          :loading="drawing"
        >
          继续抽取
        </el-button>
      </template>
    </el-dialog>

    <!-- All Players Results -->
    <el-card style="margin-top: 20px">
      <template #header>
        <div class="card-header">
          <span>所有玩家初抽结果</span>
          <el-button text type="primary" @click="loadResults">
            <el-icon><Refresh /></el-icon> 刷新
          </el-button>
        </div>
      </template>

      <el-table :data="results" stripe v-loading="loadingResults" @row-click="expandRow">
        <el-table-column type="expand">
          <template #default="{ row }">
            <div class="expand-content">
              <el-tag 
                v-for="general in row.generals" 
                :key="general.id"
                class="general-tag"
              >
                {{ general.name }} (薪资:{{ general.salary }})
              </el-tag>
              <el-empty v-if="!row.generals?.length" description="暂无武将" />
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="nickname" label="玩家" width="120" />
        <el-table-column label="已抽取" width="100">
          <template #default="{ row }">
            {{ row.generals?.length || 0 }} / 10
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.draw_complete ? 'success' : 'warning'" size="small">
              {{ row.draw_complete ? '已完成' : '进行中' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="总薪资" width="100">
          <template #default="{ row }">
            {{ row.total_salary }}
          </template>
        </el-table-column>
        <el-table-column label="武将预览">
          <template #default="{ row }">
            <div class="general-preview">
              <el-tag 
                v-for="general in row.generals?.slice(0, 5)" 
                :key="general.id"
                size="small"
                style="margin-right: 4px"
              >
                {{ general.name }}
              </el-tag>
              <span v-if="row.generals?.length > 5" class="more-hint">
                +{{ row.generals.length - 5 }}
              </span>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Pool Info (collapsible) -->
    <el-card style="margin-top: 20px">
      <template #header>
        <div class="card-header">
          <span>初抽池信息</span>
          <el-button text @click="showPool = !showPool">
            {{ showPool ? '收起' : '展开' }}
          </el-button>
        </div>
      </template>

      <div v-show="showPool">
        <el-tabs v-model="poolTab">
          <el-tab-pane label="保底池" name="guarantee">
            <el-tag type="info" style="margin-bottom: 10px">
              剩余 {{ pool?.guarantee?.length || 0 }} 名武将
            </el-tag>
            <el-table :data="pool?.guarantee || []" stripe max-height="400">
              <el-table-column prop="name" label="姓名" width="100" />
              <el-table-column prop="command" label="统率" width="80" />
              <el-table-column prop="force" label="武力" width="80" />
              <el-table-column prop="intelligence" label="智力" width="80" />
              <el-table-column prop="politics" label="政治" width="80" />
              <el-table-column prop="charm" label="魅力" width="80" />
              <el-table-column prop="salary" label="薪资" width="80" />
              <el-table-column prop="skills" label="特技" />
            </el-table>
          </el-tab-pane>
          <el-tab-pane label="普通池" name="normal">
            <el-tag type="info" style="margin-bottom: 10px">
              剩余 {{ pool?.normal?.length || 0 }} 名武将
            </el-tag>
            <el-table :data="pool?.normal || []" stripe max-height="400">
              <el-table-column prop="name" label="姓名" width="100" />
              <el-table-column prop="command" label="统率" width="80" />
              <el-table-column prop="force" label="武力" width="80" />
              <el-table-column prop="intelligence" label="智力" width="80" />
              <el-table-column prop="politics" label="政治" width="80" />
              <el-table-column prop="charm" label="魅力" width="80" />
              <el-table-column prop="salary" label="薪资" width="80" />
              <el-table-column prop="skills" label="特技" />
            </el-table>
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { useGameStore } from '../stores/game'
import { useUserStore } from '../stores/user'
import { initialDrawApi } from '../api'

const gameStore = useGameStore()
const userStore = useUserStore()

const phase = ref(null)
const status = ref(null)
const results = ref([])
const pool = ref(null)
const lastDrawResult = ref(null)
const lastDrawType = ref('')

const drawing = ref(false)
const loadingResults = ref(false)
const showResult = ref(false)
const showPool = ref(false)
const poolTab = ref('guarantee')

const canDraw = computed(() => {
  return phase.value?.current_phase === 'initial_draw' && 
         status.value?.total_remaining > 0 &&
         userStore.user?.is_registered
})

const statusTagType = computed(() => {
  if (phase.value?.current_phase === 'initial_draw') return 'success'
  return 'info'
})

const drawButtonText = computed(() => {
  if (!userStore.user?.is_registered) return '请先报名'
  if (phase.value?.current_phase !== 'initial_draw') return '非初抽阶段'
  if (status.value?.total_remaining <= 0) return '已完成初抽'
  return '抽取武将'
})

onMounted(async () => {
  await Promise.all([
    loadPhase(),
    loadStatus(),
    loadResults(),
    loadPool()
  ])
})

async function loadPhase() {
  try {
    await gameStore.fetchPhase()
    phase.value = gameStore.phase
  } catch (error) {
    console.error('Failed to load phase:', error)
  }
}

async function loadStatus() {
  try {
    const response = await initialDrawApi.getStatus()
    status.value = response.data
  } catch (error) {
    console.error('Failed to load status:', error)
  }
}

async function loadResults() {
  loadingResults.value = true
  try {
    const response = await initialDrawApi.getResults()
    results.value = response.data || []
  } catch (error) {
    console.error('Failed to load results:', error)
  } finally {
    loadingResults.value = false
  }
}

async function loadPool() {
  try {
    const response = await initialDrawApi.getPool()
    pool.value = response.data
  } catch (error) {
    console.error('Failed to load pool:', error)
  }
}

async function handleDraw() {
  drawing.value = true
  try {
    const response = await initialDrawApi.draw()
    lastDrawResult.value = response.data.general
    lastDrawType.value = response.data.draw_type
    showResult.value = true
    
    // Refresh data
    await Promise.all([
      loadStatus(),
      loadResults(),
      loadPool(),
      userStore.fetchUser()
    ])
    
    ElMessage.success('抽取成功！')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '抽取失败')
  } finally {
    drawing.value = false
  }
}

async function handleDrawAgain() {
  showResult.value = false
  await handleDraw()
}

function expandRow(row) {
  // Auto expand on click
}
</script>

<style scoped>
.initial-draw-page {
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-card .el-statistic {
  text-align: center;
}

.stat-total {
  color: #909399;
  font-size: 14px;
}

.draw-action {
  margin-top: 30px;
  text-align: center;
}

.draw-tip {
  margin-top: 10px;
  color: #666;
}

.draw-result {
  text-align: center;
  padding: 20px;
}

.draw-result .general-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 36px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
}

.draw-result h2 {
  margin: 16px 0;
  color: #333;
}

.general-stats {
  margin: 20px 0;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.general-skills {
  margin-top: 16px;
}

.general-skills .el-tag {
  margin: 2px;
}

.expand-content {
  padding: 20px;
}

.general-tag {
  margin: 4px;
}

.general-preview {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.more-hint {
  color: #909399;
  font-size: 12px;
}
</style>
