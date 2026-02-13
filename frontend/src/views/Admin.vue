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
                <el-option label="抽将阶段" value="draw" />
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

        <!-- Invite Code Management -->
        <el-tab-pane label="邀请码管理" name="invite">
          <!-- Stats -->
          <el-row :gutter="20" class="invite-stats">
            <el-col :span="6">
              <el-statistic title="总邀请码" :value="inviteStats.total || 0" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="可用" :value="inviteStats.available || 0">
                <template #suffix>
                  <el-icon color="#67c23a"><CircleCheck /></el-icon>
                </template>
              </el-statistic>
            </el-col>
            <el-col :span="6">
              <el-statistic title="已使用" :value="inviteStats.used || 0" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="已过期" :value="inviteStats.expired || 0" />
            </el-col>
          </el-row>

          <el-divider />

          <!-- Generate Form -->
          <el-form :model="generateForm" label-width="100px" style="max-width: 600px">
            <el-form-item label="生成数量">
              <el-input-number v-model="generateForm.count" :min="1" :max="100" />
            </el-form-item>
            <el-form-item label="使用次数">
              <el-input-number v-model="generateForm.max_uses" :min="1" :max="100" />
              <span class="form-tip">每个邀请码可被使用的次数</span>
            </el-form-item>
            <el-form-item label="有效期(天)">
              <el-input-number v-model="generateForm.expire_days" :min="0" :max="365" />
              <span class="form-tip">0 表示永不过期</span>
            </el-form-item>
            <el-form-item label="备注">
              <el-input v-model="generateForm.remark" placeholder="可选，如：第一批玩家" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleGenerateCodes" :loading="generating">
                生成邀请码
              </el-button>
            </el-form-item>
          </el-form>

          <!-- Generated codes display -->
          <el-card v-if="generatedCodes.length > 0" class="generated-codes-card">
            <template #header>
              <div class="card-header">
                <span>新生成的邀请码 ({{ generatedCodes.length }} 个)</span>
                <el-button type="primary" link @click="copyAllCodes">
                  <el-icon><CopyDocument /></el-icon> 复制全部
                </el-button>
              </div>
            </template>
            <div class="code-list">
              <el-tag 
                v-for="code in generatedCodes" 
                :key="code"
                class="code-tag"
                @click="copyCode(code)"
              >
                {{ code }}
              </el-tag>
            </div>
          </el-card>

          <el-divider />

          <!-- Invite codes table -->
          <el-table :data="inviteCodes" stripe v-loading="loadingCodes">
            <el-table-column prop="code" label="邀请码" width="180">
              <template #default="{ row }">
                <el-text class="code-text" @click="copyCode(row.code)">{{ row.code }}</el-text>
              </template>
            </el-table-column>
            <el-table-column label="使用情况" width="100">
              <template #default="{ row }">
                {{ row.used_count }} / {{ row.max_uses }}
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getCodeStatusType(row)" size="small">
                  {{ getCodeStatusText(row) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="remark" label="备注" />
            <el-table-column label="过期时间" width="160">
              <template #default="{ row }">
                {{ row.expired_at ? formatTime(row.expired_at) : '永不过期' }}
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="160">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-popconfirm
                  title="确定删除此邀请码？"
                  @confirm="handleDeleteCode(row.id)"
                >
                  <template #reference>
                    <el-button type="danger" link size="small">删除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>

          <!-- Pagination -->
          <el-pagination
            v-model:current-page="currentPage"
            :page-size="pageSize"
            :total="totalCodes"
            layout="prev, pager, next"
            @current-change="loadInviteCodes"
            style="margin-top: 16px; justify-content: center;"
          />
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

        <!-- Draw Management -->
        <el-tab-pane label="抽将管理" name="draw">
          <el-alert type="info" :closable="false" style="margin-bottom: 20px">
            <template #title>抽将管理说明</template>
            <p>• 重置抽将：清除玩家的抽将记录，让其可以重新抽将</p>
            <p>• 代替抽将：管理员代替指定玩家完成所有抽将</p>
            <p>• 批量操作：对所有已报名玩家进行批量处理</p>
          </el-alert>

          <el-row :gutter="20">
            <!-- Single user operation -->
            <el-col :span="12">
              <el-card shadow="hover">
                <template #header><span>单个玩家操作</span></template>
                <el-form label-width="100px">
                  <el-form-item label="选择玩家">
                    <el-select v-model="selectedUserId" placeholder="选择玩家" style="width: 100%" filterable>
                      <el-option
                        v-for="player in registeredPlayers"
                        :key="player.id"
                        :label="player.nickname"
                        :value="player.id"
                      />
                    </el-select>
                  </el-form-item>
                  <el-form-item>
                    <el-space>
                      <el-popconfirm
                        title="确定要重置该玩家的抽将记录吗？"
                        @confirm="handleResetUserDraw"
                      >
                        <template #reference>
                          <el-button type="warning" :disabled="!selectedUserId" :loading="resettingDraw">
                            重置抽将
                          </el-button>
                        </template>
                      </el-popconfirm>
                      <el-popconfirm
                        title="确定要为该玩家完成所有抽将吗？"
                        @confirm="handleDrawForUser"
                      >
                        <template #reference>
                          <el-button type="primary" :disabled="!selectedUserId" :loading="drawingForUser">
                            代替抽将
                          </el-button>
                        </template>
                      </el-popconfirm>
                    </el-space>
                  </el-form-item>
                </el-form>
              </el-card>
            </el-col>

            <!-- Batch operation -->
            <el-col :span="12">
              <el-card shadow="hover">
                <template #header><span>批量操作</span></template>
                <el-space direction="vertical" fill style="width: 100%">
                  <el-popconfirm
                    title="确定要重置所有玩家的抽将记录吗？这将影响所有已报名玩家！"
                    @confirm="handleResetAllDraw"
                  >
                    <template #reference>
                      <el-button type="danger" :loading="resettingAllDraw" style="width: 100%">
                        重置所有玩家抽将
                      </el-button>
                    </template>
                  </el-popconfirm>
                  <el-popconfirm
                    title="确定要为所有玩家完成抽将吗？这将影响所有已报名玩家！"
                    @confirm="handleDrawForAll"
                  >
                    <template #reference>
                      <el-button type="success" :loading="drawingForAll" style="width: 100%">
                        为所有玩家抽将
                      </el-button>
                    </template>
                  </el-popconfirm>
                </el-space>
              </el-card>
            </el-col>
          </el-row>

          <!-- Draw result dialog -->
          <el-dialog v-model="showDrawResult" title="抽将结果" width="600px">
            <div v-if="drawResultData">
              <p><strong>已抽取 {{ drawResultData.count || drawResultData.generals?.length }} 名武将</strong></p>
              <el-table :data="drawResultData.generals" stripe max-height="400">
                <el-table-column prop="name" label="姓名" width="100" />
                <el-table-column prop="salary" label="薪资" width="80" />
                <el-table-column prop="command" label="统率" width="60" />
                <el-table-column prop="force" label="武力" width="60" />
                <el-table-column prop="intelligence" label="智力" width="60" />
                <el-table-column prop="skills" label="特技" />
              </el-table>
            </div>
          </el-dialog>
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { CircleCheck, CopyDocument } from '@element-plus/icons-vue'
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

// Draw management state
const registeredPlayers = ref([])
const selectedUserId = ref(null)
const resettingDraw = ref(false)
const drawingForUser = ref(false)
const resettingAllDraw = ref(false)
const drawingForAll = ref(false)
const showDrawResult = ref(false)
const drawResultData = ref(null)

// Invite code state
const inviteCodes = ref([])
const inviteStats = ref({})
const loadingCodes = ref(false)
const generating = ref(false)
const generatedCodes = ref([])
const currentPage = ref(1)
const pageSize = 20
const totalCodes = ref(0)

const generateForm = reactive({
  count: 10,
  max_uses: 1,
  expire_days: 30,
  remark: ''
})

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
  loadInviteCodes()
  loadInviteStats()
  loadRegisteredPlayers()
}

async function loadRegisteredPlayers() {
  try {
    const response = await gameApi.getPlayers()
    registeredPlayers.value = response.data || []
  } catch (error) {
    console.error('Failed to load players:', error)
  }
}

async function handleResetUserDraw() {
  if (!selectedUserId.value) return
  resettingDraw.value = true
  try {
    await adminApi.resetUserDraw(selectedUserId.value)
    ElMessage.success('玩家抽将记录已重置')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '重置失败')
  } finally {
    resettingDraw.value = false
  }
}

async function handleDrawForUser() {
  if (!selectedUserId.value) return
  drawingForUser.value = true
  try {
    const response = await adminApi.drawForUser(selectedUserId.value)
    drawResultData.value = response.data
    showDrawResult.value = true
    ElMessage.success(`代抽完成，共抽取 ${response.data.count} 名武将`)
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '代抽失败')
  } finally {
    drawingForUser.value = false
  }
}

async function handleResetAllDraw() {
  resettingAllDraw.value = true
  try {
    const response = await adminApi.resetAllDraw()
    ElMessage.success(`已重置 ${response.data.reset_count} 名玩家的抽将记录`)
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '批量重置失败')
  } finally {
    resettingAllDraw.value = false
  }
}

async function handleDrawForAll() {
  drawingForAll.value = true
  try {
    const response = await adminApi.drawForAll()
    ElMessage.success(`批量抽将完成，共为 ${response.data.user_count} 名玩家抽取 ${response.data.total_count} 名武将`)
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '批量抽将失败')
  } finally {
    drawingForAll.value = false
  }
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

async function loadInviteCodes() {
  loadingCodes.value = true
  try {
    const response = await adminApi.getInviteCodes(currentPage.value, pageSize)
    inviteCodes.value = response.data.codes || []
    totalCodes.value = response.data.total || 0
  } catch (error) {
    console.error('Failed to load invite codes:', error)
  } finally {
    loadingCodes.value = false
  }
}

async function loadInviteStats() {
  try {
    const response = await adminApi.getInviteCodeStats()
    inviteStats.value = response.data || {}
  } catch (error) {
    console.error('Failed to load invite stats:', error)
  }
}

async function handleGenerateCodes() {
  generating.value = true
  try {
    const response = await adminApi.generateInviteCodes(generateForm)
    generatedCodes.value = response.data.code_strings || []
    ElMessage.success(`成功生成 ${generatedCodes.value.length} 个邀请码`)
    await loadInviteCodes()
    await loadInviteStats()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '生成失败')
  } finally {
    generating.value = false
  }
}

async function handleDeleteCode(id) {
  try {
    await adminApi.deleteInviteCode(id)
    ElMessage.success('邀请码已删除')
    await loadInviteCodes()
    await loadInviteStats()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '删除失败')
  }
}

function copyCode(code) {
  navigator.clipboard.writeText(code)
  ElMessage.success('已复制到剪贴板')
}

function copyAllCodes() {
  const text = generatedCodes.value.join('\n')
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制全部邀请码')
}

function getCodeStatusType(code) {
  const now = new Date()
  if (code.expired_at && new Date(code.expired_at) < now) return 'info'
  if (code.used_count >= code.max_uses) return 'warning'
  return 'success'
}

function getCodeStatusText(code) {
  const now = new Date()
  if (code.expired_at && new Date(code.expired_at) < now) return '已过期'
  if (code.used_count >= code.max_uses) return '已用完'
  return '可用'
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

.invite-stats {
  margin-bottom: 20px;
}

.form-tip {
  color: #909399;
  font-size: 12px;
  margin-left: 10px;
}

.generated-codes-card {
  margin: 20px 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.code-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.code-tag {
  cursor: pointer;
  font-family: monospace;
}

.code-tag:hover {
  background-color: #ecf5ff;
}

.code-text {
  font-family: monospace;
  cursor: pointer;
}

.code-text:hover {
  color: #409eff;
}
</style>
