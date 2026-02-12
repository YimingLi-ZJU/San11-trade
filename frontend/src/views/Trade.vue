<template>
  <div class="trade-page">
    <el-row :gutter="20">
      <!-- Left: Trade List -->
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>交易中心</span>
              <el-button type="primary" @click="showCreateDialog = true" :disabled="!canTrade">
                发起交易
              </el-button>
            </div>
          </template>

          <el-tabs v-model="activeTab">
            <el-tab-pane label="待处理" name="pending">
              <el-table :data="pendingTrades" stripe v-loading="loading">
                <el-table-column label="类型" width="80">
                  <template #default="{ row }">
                    <el-tag :type="row.proposer_id === userStore.user?.id ? 'primary' : 'success'">
                      {{ row.proposer_id === userStore.user?.id ? '发出' : '收到' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="对方" width="100">
                  <template #default="{ row }">
                    {{ row.proposer_id === userStore.user?.id ? row.receiver.nickname : row.proposer.nickname }}
                  </template>
                </el-table-column>
                <el-table-column label="我方出" min-width="150">
                  <template #default="{ row }">
                    {{ formatTradeItems(row, row.proposer_id === userStore.user?.id ? 'offer' : 'request') }}
                  </template>
                </el-table-column>
                <el-table-column label="我方得" min-width="150">
                  <template #default="{ row }">
                    {{ formatTradeItems(row, row.proposer_id === userStore.user?.id ? 'request' : 'offer') }}
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="180">
                  <template #default="{ row }">
                    <template v-if="row.receiver_id === userStore.user?.id">
                      <el-button type="success" size="small" @click="handleAccept(row.id)">
                        接受
                      </el-button>
                      <el-button type="danger" size="small" @click="handleReject(row.id)">
                        拒绝
                      </el-button>
                    </template>
                    <template v-else>
                      <el-button type="warning" size="small" @click="handleCancel(row.id)">
                        撤回
                      </el-button>
                    </template>
                  </template>
                </el-table-column>
              </el-table>
              <el-empty v-if="!pendingTrades.length" description="暂无待处理交易" />
            </el-tab-pane>

            <el-tab-pane label="历史记录" name="history">
              <el-table :data="historyTrades" stripe>
                <el-table-column label="对方" width="100">
                  <template #default="{ row }">
                    {{ row.proposer_id === userStore.user?.id ? row.receiver.nickname : row.proposer.nickname }}
                  </template>
                </el-table-column>
                <el-table-column label="内容" min-width="200">
                  <template #default="{ row }">
                    {{ formatTradeContent(row) }}
                  </template>
                </el-table-column>
                <el-table-column label="状态" width="100">
                  <template #default="{ row }">
                    <el-tag :type="getStatusType(row.status)">
                      {{ getStatusText(row.status) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="created_at" label="时间" width="180">
                  <template #default="{ row }">
                    {{ formatTime(row.created_at) }}
                  </template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </el-col>

      <!-- Right: Players -->
      <el-col :span="8">
        <el-card>
          <template #header>选择交易对象</template>
          <el-table :data="players" stripe size="small" max-height="500">
            <el-table-column prop="nickname" label="玩家" />
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button
                  v-if="row.id !== userStore.user?.id"
                  type="primary"
                  size="small"
                  text
                  @click="selectTradePartner(row)"
                >
                  交易
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <!-- Create Trade Dialog -->
    <el-dialog v-model="showCreateDialog" title="发起交易" width="700px">
      <el-form label-width="100px">
        <el-form-item label="交易对象">
          <el-select v-model="tradeForm.receiver_id" placeholder="选择玩家" style="width: 100%">
            <el-option
              v-for="p in players.filter(p => p.id !== userStore.user?.id)"
              :key="p.id"
              :label="p.nickname"
              :value="p.id"
            />
          </el-select>
        </el-form-item>

        <el-divider content-position="left">我方出</el-divider>
        
        <el-form-item label="出让武将">
          <el-select v-model="tradeForm.offer_generals" multiple placeholder="选择武将" style="width: 100%">
            <el-option
              v-for="g in myGenerals"
              :key="g.id"
              :label="`${g.name} (薪资:${g.salary})`"
              :value="g.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="出让空间">
          <el-input-number v-model="tradeForm.offer_space" :min="0" :max="userStore.remainingSpace" />
        </el-form-item>

        <el-divider content-position="left">我方得</el-divider>

        <el-form-item label="索要武将">
          <el-select v-model="tradeForm.request_generals" multiple placeholder="选择武将" style="width: 100%">
            <el-option
              v-for="g in partnerGenerals"
              :key="g.id"
              :label="`${g.name} (薪资:${g.salary})`"
              :value="g.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="索要空间">
          <el-input-number v-model="tradeForm.request_space" :min="0" />
        </el-form-item>

        <el-form-item label="留言">
          <el-input v-model="tradeForm.message" type="textarea" placeholder="可选留言" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">
          发起交易
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '../stores/user'
import { useGameStore } from '../stores/game'
import { tradeApi, gameApi, authApi } from '../api'

const userStore = useUserStore()
const gameStore = useGameStore()

const activeTab = ref('pending')
const loading = ref(false)
const pendingTrades = ref([])
const historyTrades = ref([])
const players = ref([])
const myGenerals = ref([])
const partnerGenerals = ref([])
const showCreateDialog = ref(false)
const creating = ref(false)

const tradeForm = ref({
  receiver_id: null,
  offer_generals: [],
  offer_treasures: [],
  offer_space: 0,
  request_generals: [],
  request_treasures: [],
  request_space: 0,
  message: ''
})

const canTrade = computed(() =>
  ['trading', 'draft'].includes(gameStore.phase?.current_phase)
)

watch(() => tradeForm.value.receiver_id, async (newId) => {
  if (newId) {
    try {
      const response = await gameApi.getPlayerRoster(newId)
      partnerGenerals.value = response.data.generals || []
    } catch (error) {
      console.error('Failed to load partner roster:', error)
    }
  } else {
    partnerGenerals.value = []
  }
})

onMounted(async () => {
  await gameStore.fetchPhase()
  await loadData()
})

async function loadData() {
  loading.value = true
  try {
    const [pending, history, playerList, roster] = await Promise.all([
      tradeApi.getPendingTrades(),
      tradeApi.getTradeHistory(),
      gameApi.getPlayers(),
      authApi.getMyRoster()
    ])
    pendingTrades.value = pending.data || []
    historyTrades.value = history.data || []
    players.value = playerList.data || []
    myGenerals.value = roster.data?.generals || []
  } catch (error) {
    console.error('Failed to load trade data:', error)
  } finally {
    loading.value = false
  }
}

function selectTradePartner(player) {
  tradeForm.value.receiver_id = player.id
  showCreateDialog.value = true
}

async function handleCreate() {
  if (!tradeForm.value.receiver_id) {
    ElMessage.warning('请选择交易对象')
    return
  }

  creating.value = true
  try {
    await tradeApi.createTrade(tradeForm.value)
    ElMessage.success('交易请求已发送')
    showCreateDialog.value = false
    resetForm()
    await loadData()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '发起交易失败')
  } finally {
    creating.value = false
  }
}

async function handleAccept(id) {
  try {
    await ElMessageBox.confirm('确定接受这笔交易吗？', '确认')
    await tradeApi.acceptTrade(id)
    ElMessage.success('交易已接受')
    await loadData()
    await userStore.fetchUser()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '操作失败')
    }
  }
}

async function handleReject(id) {
  try {
    await ElMessageBox.confirm('确定拒绝这笔交易吗？', '确认')
    await tradeApi.rejectTrade(id)
    ElMessage.success('交易已拒绝')
    await loadData()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '操作失败')
    }
  }
}

async function handleCancel(id) {
  try {
    await ElMessageBox.confirm('确定撤回这笔交易吗？', '确认')
    await tradeApi.cancelTrade(id)
    ElMessage.success('交易已撤回')
    await loadData()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '操作失败')
    }
  }
}

function resetForm() {
  tradeForm.value = {
    receiver_id: null,
    offer_generals: [],
    offer_treasures: [],
    offer_space: 0,
    request_generals: [],
    request_treasures: [],
    request_space: 0,
    message: ''
  }
}

function formatTradeItems(trade, type) {
  const generals = JSON.parse(trade[`${type}_generals`] || '[]')
  const space = trade[`${type}_space`] || 0
  
  const parts = []
  if (generals.length) parts.push(`${generals.length}名武将`)
  if (space) parts.push(`${space}空间`)
  
  return parts.join(', ') || '无'
}

function formatTradeContent(trade) {
  const offer = formatTradeItems(trade, 'offer')
  const request = formatTradeItems(trade, 'request')
  return `出: ${offer} ↔ 得: ${request}`
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
.trade-page {
  max-width: 1400px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
