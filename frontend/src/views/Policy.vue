<template>
  <div class="policy-container">
    <el-card class="header-card">
      <template #header>
        <div class="card-header">
          <span>🏆 国策选择</span>
          <el-tag :type="statusTagType" size="large">{{ statusText }}</el-tag>
        </div>
      </template>
      
      <!-- Status Info -->
      <div class="status-info">
        <el-descriptions :column="3" border>
          <el-descriptions-item label="当前状态">
            <el-tag :type="statusTagType">{{ statusText }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="开始时间" v-if="config?.start_time">
            {{ formatTime(config.start_time) }}
          </el-descriptions-item>
          <el-descriptions-item label="超时时间" v-if="config?.timeout_minutes">
            {{ config.timeout_minutes }} 分钟/人
          </el-descriptions-item>
          <el-descriptions-item label="当前选择者" v-if="currentUser">
            <el-tag type="warning">{{ currentUser.nickname || currentUser.username }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="截止时间" v-if="config?.current_deadline">
            <el-countdown :value="new Date(config.current_deadline).getTime()" format="HH:mm:ss" />
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-card>

    <!-- Bidding Phase -->
    <el-card v-if="config?.status === 'bidding'" class="bidding-card">
      <template #header>
        <span>💰 出价竞拍</span>
      </template>
      
      <el-alert type="info" :closable="false" style="margin-bottom: 20px">
        <p>请输入您愿意花费的空间数量进行暗拍。出价越高，选择顺序越靠前。</p>
        <p>您当前可用空间：<strong>{{ availableSpace }}</strong></p>
      </el-alert>
      
      <el-form :model="bidForm" label-width="100px">
        <el-form-item label="出价空间">
          <el-input-number 
            v-model="bidForm.amount" 
            :min="0" 
            :max="availableSpace"
            :step="10"
          />
          <span style="margin-left: 10px; color: #909399">（0 表示不出价，仍可参与选择）</span>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submitBid" :loading="submitting">
            {{ myBid ? '更新出价' : '提交出价' }}
          </el-button>
        </el-form-item>
      </el-form>
      
      <el-divider>偏好设置（可选）</el-divider>
      
      <el-alert type="warning" :closable="false" style="margin-bottom: 20px">
        设置您的国策偏好顺序。如果您未能在规定时间内选择，系统将按此顺序自动为您分配。
      </el-alert>
      
      <div class="preference-section">
        <div class="clubs-available">
          <h4>可选国策</h4>
          <div class="filter-bar">
            <el-select v-model="filterLeague" placeholder="按联赛筛选" clearable style="width: 120px">
              <el-option v-for="league in filters.leagues" :key="league" :label="league" :value="league" />
            </el-select>
            <el-select v-model="filterTag" placeholder="按标签筛选" clearable style="width: 120px">
              <el-option v-for="tag in filters.tags" :key="tag" :label="tag" :value="tag" />
            </el-select>
          </div>
          <div class="club-list">
            <div 
              v-for="club in filteredAvailableClubs" 
              :key="club.id" 
              class="club-item"
              @click="addToPreference(club)"
            >
              <div class="club-name">{{ club.name }}</div>
              <div class="club-tags">
                <el-tag v-if="club.league" size="small" type="info">{{ club.league }}</el-tag>
                <el-tag v-for="tag in club.tags" :key="tag.id" size="small">{{ tag.tag }}</el-tag>
              </div>
            </div>
          </div>
        </div>
        
        <div class="preference-list">
          <h4>偏好顺序（拖拽排序）</h4>
          <draggable v-model="preferenceList" item-key="id" class="drag-list">
            <template #item="{ element, index }">
              <div class="preference-item">
                <span class="priority">{{ index + 1 }}</span>
                <span class="name">{{ element.name }}</span>
                <el-button type="danger" size="small" @click="removeFromPreference(index)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </template>
          </draggable>
          <el-button 
            type="success" 
            @click="savePreferences" 
            :disabled="preferenceList.length === 0"
            style="margin-top: 10px"
          >
            保存偏好
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- Selection Phase - Your Turn -->
    <el-card v-if="config?.status === 'selecting' && isMyTurn" class="selection-card">
      <template #header>
        <span>🎯 轮到您选择国策！</span>
      </template>
      
      <el-alert type="success" :closable="false" style="margin-bottom: 20px">
        <p>现在轮到您选择国策，请在截止时间前做出选择。</p>
        <p v-if="config?.current_deadline">
          截止时间：<strong>{{ formatTime(config.current_deadline) }}</strong>
        </p>
      </el-alert>
      
      <div class="filter-bar" style="margin-bottom: 20px">
        <el-select v-model="filterLeague" placeholder="按联赛筛选" clearable style="width: 120px">
          <el-option v-for="league in filters.leagues" :key="league" :label="league" :value="league" />
        </el-select>
        <el-select v-model="filterTag" placeholder="按标签筛选" clearable style="width: 120px">
          <el-option v-for="tag in filters.tags" :key="tag" :label="tag" :value="tag" />
        </el-select>
      </div>
      
      <div class="clubs-grid">
        <el-card 
          v-for="club in filteredAvailableClubsForSelection" 
          :key="club.id" 
          class="club-card"
          shadow="hover"
          @click="confirmSelectClub(club)"
        >
          <div class="club-header">
            <h3>{{ club.name }}</h3>
            <div class="club-tags">
              <el-tag v-if="club.league" type="info">{{ club.league }}</el-tag>
              <el-tag v-for="tag in club.tags" :key="tag.id">{{ tag.tag }}</el-tag>
            </div>
          </div>
          <div class="club-policies">
            <div v-for="policy in club.policies" :key="policy.id" class="policy-item">
              <span v-if="policy.condition" class="condition">{{ policy.condition }}</span>
              <span class="effect">{{ policy.effect }}</span>
            </div>
          </div>
        </el-card>
      </div>
    </el-card>

    <!-- Selection Phase - Waiting -->
    <el-card v-if="config?.status === 'selecting' && !isMyTurn" class="waiting-card">
      <template #header>
        <span>⏳ 等待其他玩家选择</span>
      </template>
      
      <el-alert type="info" :closable="false">
        <p v-if="currentUser">当前正在选择：<strong>{{ currentUser.nickname || currentUser.username }}</strong></p>
        <p v-if="myRank">您的选择顺序：第 <strong>{{ myRank }}</strong> 位</p>
      </el-alert>
    </el-card>

    <!-- Selection Order -->
    <el-card class="order-card">
      <template #header>
        <span>📋 选择顺序</span>
      </template>
      
      <el-table :data="bids" stripe>
        <el-table-column prop="rank" label="顺序" width="80" />
        <el-table-column label="玩家">
          <template #default="{ row }">
            {{ row.user?.nickname || row.user?.username }}
          </template>
        </el-table-column>
        <el-table-column label="出价" v-if="config?.status !== 'bidding'">
          <template #default="{ row }">
            {{ row.bid_amount }}
          </template>
        </el-table-column>
        <el-table-column label="状态">
          <template #default="{ row }">
            <el-tag v-if="getSelectionForUser(row.user_id)" type="success">
              已选择: {{ getSelectionForUser(row.user_id).club?.name }}
            </el-tag>
            <el-tag v-else-if="config?.current_selector === row.user_id" type="warning">
              正在选择
            </el-tag>
            <el-tag v-else type="info">等待中</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Selection Results -->
    <el-card v-if="selections.length > 0" class="results-card">
      <template #header>
        <span>✅ 选择结果</span>
      </template>
      
      <el-table :data="selections" stripe>
        <el-table-column prop="select_order" label="顺序" width="80" />
        <el-table-column label="玩家">
          <template #default="{ row }">
            {{ row.user?.nickname || row.user?.username }}
          </template>
        </el-table-column>
        <el-table-column label="选择国策">
          <template #default="{ row }">
            <div>
              <strong>{{ row.club?.name }}</strong>
              <el-tag v-if="row.club?.league" size="small" type="info" style="margin-left: 8px">
                {{ row.club.league }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="bid_cost" label="花费空间" width="100" />
        <el-table-column label="方式" width="100">
          <template #default="{ row }">
            <el-tag :type="row.auto_assigned ? 'warning' : 'success'">
              {{ row.auto_assigned ? '自动分配' : '手动选择' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete } from '@element-plus/icons-vue'
import draggable from 'vuedraggable'
import { policyApi } from '../api'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()

// Data
const config = ref(null)
const bids = ref([])
const selections = ref([])
const availableClubs = ref([])
const currentUser = ref(null)
const myBid = ref(null)
const preferenceList = ref([])
const filters = ref({ leagues: [], tags: [] })

// Form
const bidForm = ref({ amount: 0 })
const submitting = ref(false)

// Filters
const filterLeague = ref('')
const filterTag = ref('')

// Polling
let pollInterval = null

// Computed
const statusText = computed(() => {
  const statusMap = {
    bidding: '出价阶段',
    closed: '出价已截止',
    selecting: '选择进行中',
    completed: '选择完成'
  }
  return statusMap[config.value?.status] || '未知状态'
})

const statusTagType = computed(() => {
  const typeMap = {
    bidding: 'warning',
    closed: 'info',
    selecting: 'success',
    completed: ''
  }
  return typeMap[config.value?.status] || 'info'
})

const availableSpace = computed(() => {
  const user = userStore.user
  if (!user) return 0
  return user.space - user.used_space
})

const isMyTurn = computed(() => {
  return config.value?.current_selector === userStore.user?.id
})

const myRank = computed(() => {
  const bid = bids.value.find(b => b.user_id === userStore.user?.id)
  return bid?.rank
})

const filteredAvailableClubs = computed(() => {
  let clubs = availableClubs.value.filter(club => {
    // Filter out clubs already in preference list
    return !preferenceList.value.find(p => p.id === club.id)
  })
  
  if (filterLeague.value) {
    clubs = clubs.filter(c => c.league === filterLeague.value)
  }
  if (filterTag.value) {
    clubs = clubs.filter(c => c.tags?.some(t => t.tag === filterTag.value))
  }
  
  return clubs
})

const filteredAvailableClubsForSelection = computed(() => {
  let clubs = availableClubs.value
  
  if (filterLeague.value) {
    clubs = clubs.filter(c => c.league === filterLeague.value)
  }
  if (filterTag.value) {
    clubs = clubs.filter(c => c.tags?.some(t => t.tag === filterTag.value))
  }
  
  return clubs
})

// Methods
async function fetchStatus() {
  try {
    const response = await policyApi.getStatus()
    config.value = response.data.config
    bids.value = response.data.bids || []
    selections.value = response.data.selections || []
    availableClubs.value = response.data.available_clubs || []
    currentUser.value = response.data.current_user
  } catch (error) {
    console.error('Failed to fetch policy status:', error)
  }
}

async function fetchMyBid() {
  try {
    const response = await policyApi.getMyBid()
    myBid.value = response.data.bid
    if (myBid.value) {
      bidForm.value.amount = myBid.value.bid_amount
    }
    // Load saved preferences
    const savedPrefs = response.data.preferences || []
    preferenceList.value = savedPrefs.map(p => p.club).filter(Boolean)
  } catch (error) {
    console.error('Failed to fetch my bid:', error)
  }
}

async function fetchFilters() {
  try {
    const response = await policyApi.getFilters()
    filters.value = response.data
  } catch (error) {
    console.error('Failed to fetch filters:', error)
  }
}

async function submitBid() {
  submitting.value = true
  try {
    await policyApi.placeBid(bidForm.value.amount)
    ElMessage.success('出价成功')
    await fetchMyBid()
    await fetchStatus()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '出价失败')
  } finally {
    submitting.value = false
  }
}

function addToPreference(club) {
  if (!preferenceList.value.find(p => p.id === club.id)) {
    preferenceList.value.push(club)
  }
}

function removeFromPreference(index) {
  preferenceList.value.splice(index, 1)
}

async function savePreferences() {
  try {
    const clubIds = preferenceList.value.map(c => c.id)
    await policyApi.setPreferences(clubIds)
    ElMessage.success('偏好保存成功')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存偏好失败')
  }
}

async function confirmSelectClub(club) {
  try {
    await ElMessageBox.confirm(
      `确定选择「${club.name}」作为您的国策吗？`,
      '确认选择',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    await selectClub(club.id)
  } catch {
    // Cancelled
  }
}

async function selectClub(clubId) {
  try {
    await policyApi.selectClub(clubId)
    ElMessage.success('选择成功！')
    await fetchStatus()
    await userStore.fetchUser()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '选择失败')
  }
}

function getSelectionForUser(userId) {
  return selections.value.find(s => s.user_id === userId)
}

function formatTime(time) {
  return new Date(time).toLocaleString('zh-CN')
}

// Lifecycle
onMounted(async () => {
  await Promise.all([fetchStatus(), fetchMyBid(), fetchFilters()])
  
  // Poll for updates every 5 seconds during active phases
  pollInterval = setInterval(async () => {
    if (config.value?.status === 'selecting' || config.value?.status === 'bidding') {
      await fetchStatus()
    }
  }, 5000)
})

onUnmounted(() => {
  if (pollInterval) {
    clearInterval(pollInterval)
  }
})
</script>

<style scoped>
.policy-container {
  max-width: 1400px;
  margin: 0 auto;
}

.header-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-info {
  margin-top: 10px;
}

.bidding-card,
.selection-card,
.waiting-card,
.order-card,
.results-card {
  margin-bottom: 20px;
}

.preference-section {
  display: flex;
  gap: 40px;
}

.clubs-available,
.preference-list {
  flex: 1;
}

.filter-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

.club-list {
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  padding: 10px;
}

.club-item {
  padding: 10px;
  border-bottom: 1px solid #ebeef5;
  cursor: pointer;
  transition: background-color 0.3s;
}

.club-item:hover {
  background-color: #f5f7fa;
}

.club-item:last-child {
  border-bottom: none;
}

.club-name {
  font-weight: bold;
  margin-bottom: 5px;
}

.club-tags {
  display: flex;
  gap: 5px;
  flex-wrap: wrap;
}

.drag-list {
  min-height: 200px;
  border: 2px dashed #dcdfe6;
  border-radius: 4px;
  padding: 10px;
}

.preference-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  background: #f5f7fa;
  border-radius: 4px;
  margin-bottom: 8px;
  cursor: move;
}

.preference-item .priority {
  width: 30px;
  height: 30px;
  background: #409eff;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
}

.preference-item .name {
  flex: 1;
}

.clubs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
}

.club-card {
  cursor: pointer;
  transition: transform 0.3s, box-shadow 0.3s;
}

.club-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 20px rgba(0, 0, 0, 0.1);
}

.club-header h3 {
  margin: 0 0 10px 0;
}

.club-policies {
  margin-top: 15px;
}

.policy-item {
  padding: 8px 0;
  border-bottom: 1px solid #ebeef5;
}

.policy-item:last-child {
  border-bottom: none;
}

.policy-item .condition {
  color: #909399;
  display: block;
  margin-bottom: 4px;
}

.policy-item .effect {
  color: #303133;
}
</style>
