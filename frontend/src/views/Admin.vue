<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <span>管理后台</span>
      </template>

      <el-tabs v-model="activeTab">
        <!-- Phase Control -->
        <el-tab-pane label="阶段控制" name="phase">
          <el-form label-width="120px" style="max-width: 500px">
            <el-form-item label="当前阶段">
              <el-tag type="warning" size="large">
                {{ gameStore.getPhaseName(phase?.current_phase) }}
              </el-tag>
            </el-form-item>

            <el-form-item label="切换阶段">
              <el-select v-model="newPhase" placeholder="选择阶段" style="width: 100%">
                <el-option label="报名阶段" value="signup" />
                <el-option label="保底抽将" value="guarantee_draw" />
                <el-option label="普通抽将" value="normal_draw" />
                <el-option label="选秀阶段" value="draft" />
                <el-option label="自由交易" value="trading" />
                <el-option label="拍卖阶段" value="auction" />
                <el-option label="比赛阶段" value="match" />
                <el-option label="赛季结束" value="finished" />
              </el-select>
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="handleChangePhase" :loading="changingPhase">
                更新阶段
              </el-button>
            </el-form-item>
          </el-form>

          <el-divider />

          <el-popconfirm
            title="确定要重置赛季吗？这将清除所有玩家数据！"
            confirm-button-text="确定"
            cancel-button-text="取消"
            @confirm="handleReset"
          >
            <template #reference>
              <el-button type="danger" :loading="resetting">
                重置赛季
              </el-button>
            </template>
          </el-popconfirm>
        </el-tab-pane>

        <!-- Data Import -->
        <el-tab-pane label="数据导入" name="import">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            accept=".xlsx,.xls"
            :on-change="handleFileChange"
          >
            <template #trigger>
              <el-button type="primary">选择Excel文件</el-button>
            </template>
            <template #tip>
              <div class="el-upload__tip">
                支持 .xlsx/.xls 格式，请确保包含"武将"、"宝物"、"俱乐部"等sheet
              </div>
            </template>
          </el-upload>

          <el-button
            type="success"
            @click="handleImport"
            :loading="importing"
            :disabled="!selectedFile"
            style="margin-top: 20px"
          >
            开始导入
          </el-button>

          <div v-if="importResult" class="import-result">
            <el-alert type="success" :closable="false">
              <p>导入成功！</p>
              <p>武将：{{ importResult.generals }} 条</p>
              <p>宝物：{{ importResult.treasures }} 条</p>
              <p>俱乐部：{{ importResult.clubs }} 条</p>
            </el-alert>
          </div>
        </el-tab-pane>

        <!-- Statistics -->
        <el-tab-pane label="统计数据" name="stats">
          <el-descriptions :column="2" border v-if="stats">
            <el-descriptions-item label="总用户数">{{ stats.total_players }}</el-descriptions-item>
            <el-descriptions-item label="已报名玩家">{{ stats.registered_players }}</el-descriptions-item>
            <el-descriptions-item label="总武将数">{{ stats.total_generals }}</el-descriptions-item>
            <el-descriptions-item label="已归属武将">{{ stats.owned_generals }}</el-descriptions-item>
            <el-descriptions-item label="总宝物数">{{ stats.total_treasures }}</el-descriptions-item>
            <el-descriptions-item label="已归属宝物">{{ stats.owned_treasures }}</el-descriptions-item>
            <el-descriptions-item label="总交易数">{{ stats.total_trades }}</el-descriptions-item>
            <el-descriptions-item label="完成交易">{{ stats.accepted_trades }}</el-descriptions-item>
          </el-descriptions>
        </el-tab-pane>

        <!-- All Trades -->
        <el-tab-pane label="所有交易" name="trades">
          <el-table :data="allTrades" stripe v-loading="loadingTrades">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column label="发起方" width="100">
              <template #default="{ row }">{{ row.proposer?.nickname }}</template>
            </el-table-column>
            <el-table-column label="接收方" width="100">
              <template #default="{ row }">{{ row.receiver?.nickname }}</template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">{{ getStatusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="时间">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useGameStore } from '../stores/game'
import { adminApi, gameApi } from '../api'

const gameStore = useGameStore()

const activeTab = ref('phase')
const phase = ref(null)
const newPhase = ref('')
const stats = ref(null)
const allTrades = ref([])
const selectedFile = ref(null)
const importResult = ref(null)

const changingPhase = ref(false)
const resetting = ref(false)
const importing = ref(false)
const loadingTrades = ref(false)

onMounted(async () => {
  await loadData()
})

async function loadData() {
  try {
    const [phaseRes, statsRes] = await Promise.all([
      gameApi.getPhase(),
      gameApi.getStatistics()
    ])
    phase.value = phaseRes.data
    newPhase.value = phase.value.current_phase
    stats.value = statsRes.data
  } catch (error) {
    console.error('Failed to load admin data:', error)
  }

  loadTrades()
}

async function loadTrades() {
  loadingTrades.value = true
  try {
    const response = await adminApi.getAllTrades()
    allTrades.value = response.data || []
  } catch (error) {
    console.error('Failed to load trades:', error)
  } finally {
    loadingTrades.value = false
  }
}

async function handleChangePhase() {
  changingPhase.value = true
  try {
    await adminApi.setPhase({ phase: newPhase.value })
    ElMessage.success('阶段已更新')
    await loadData()
    await gameStore.fetchPhase()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '更新失败')
  } finally {
    changingPhase.value = false
  }
}

async function handleReset() {
  resetting.value = true
  try {
    await adminApi.resetSeason()
    ElMessage.success('赛季已重置')
    await loadData()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '重置失败')
  } finally {
    resetting.value = false
  }
}

function handleFileChange(file) {
  selectedFile.value = file.raw
}

async function handleImport() {
  if (!selectedFile.value) return

  const formData = new FormData()
  formData.append('file', selectedFile.value)

  importing.value = true
  try {
    const response = await adminApi.importData(formData)
    importResult.value = response.data
    ElMessage.success('数据导入成功')
    await loadData()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '导入失败')
  } finally {
    importing.value = false
  }
}

function getStatusType(status) {
  const types = { pending: 'warning', accepted: 'success', rejected: 'danger', cancelled: 'info' }
  return types[status] || 'info'
}

function getStatusText(status) {
  const texts = { pending: '待处理', accepted: '已完成', rejected: '已拒绝', cancelled: '已撤回' }
  return texts[status] || status
}

function formatTime(time) {
  return new Date(time).toLocaleString('zh-CN')
}
</script>

<style scoped>
.admin-page {
  max-width: 1200px;
  margin: 0 auto;
}

.import-result {
  margin-top: 20px;
}
</style>
