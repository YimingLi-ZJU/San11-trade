<template>
  <div class="draft-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>选秀</span>
          <el-tag :type="canDraft ? 'success' : 'info'">
            {{ gameStore.getPhaseName(gameStore.phase?.current_phase) }}
          </el-tag>
        </div>
      </template>

      <el-alert
        v-if="!canDraft"
        type="warning"
        title="当前不在选秀阶段"
        description="请等待管理员开启选秀阶段"
        show-icon
        :closable="false"
        style="margin-bottom: 20px"
      />

      <el-input
        v-model="searchText"
        placeholder="搜索武将..."
        prefix-icon="Search"
        style="width: 300px; margin-bottom: 20px"
        clearable
      />

      <el-table
        :data="filteredPool"
        stripe
        v-loading="loading"
        @row-click="handleSelect"
        :row-class-name="getRowClassName"
        style="cursor: pointer"
      >
        <el-table-column prop="name" label="姓名" width="100" />
        <el-table-column label="档次" width="80">
          <template #default="{ row }">
            <el-tag :type="getTierType(row.tier)">T{{ row.tier }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="command" label="统率" width="80" sortable />
        <el-table-column prop="force" label="武力" width="80" sortable />
        <el-table-column prop="intelligence" label="智力" width="80" sortable />
        <el-table-column prop="politics" label="政治" width="80" sortable />
        <el-table-column prop="charm" label="魅力" width="80" sortable />
        <el-table-column prop="salary" label="薪资" width="80" sortable />
        <el-table-column prop="skills" label="特技" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              :disabled="!canDraft || row.salary > userStore.remainingSpace"
              @click.stop="confirmPick(row)"
            >
              选择
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Confirm Dialog -->
    <el-dialog v-model="showConfirm" title="确认选秀" width="400px">
      <div v-if="selectedGeneral" class="confirm-content">
        <p>确定要选择 <strong>{{ selectedGeneral.name }}</strong> 吗？</p>
        <p>薪资消耗：{{ selectedGeneral.salary }}</p>
        <p>剩余空间：{{ userStore.remainingSpace }} → {{ userStore.remainingSpace - selectedGeneral.salary }}</p>
      </div>
      <template #footer>
        <el-button @click="showConfirm = false">取消</el-button>
        <el-button type="primary" :loading="picking" @click="handlePick">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useGameStore } from '../stores/game'
import { useUserStore } from '../stores/user'
import { drawApi } from '../api'

const gameStore = useGameStore()
const userStore = useUserStore()

const pool = ref([])
const loading = ref(false)
const searchText = ref('')
const selectedGeneral = ref(null)
const showConfirm = ref(false)
const picking = ref(false)

const canDraft = computed(() => gameStore.phase?.current_phase === 'draft')

const filteredPool = computed(() => {
  if (!searchText.value) return pool.value
  const text = searchText.value.toLowerCase()
  return pool.value.filter(g =>
    g.name.toLowerCase().includes(text) ||
    g.skills?.toLowerCase().includes(text)
  )
})

onMounted(async () => {
  await gameStore.fetchPhase()
  await loadPool()
})

async function loadPool() {
  loading.value = true
  try {
    const response = await drawApi.getDraftPool()
    pool.value = response.data || []
  } catch (error) {
    console.error('Failed to load draft pool:', error)
  } finally {
    loading.value = false
  }
}

function handleSelect(row) {
  if (canDraft.value && row.salary <= userStore.remainingSpace) {
    confirmPick(row)
  }
}

function confirmPick(general) {
  selectedGeneral.value = general
  showConfirm.value = true
}

async function handlePick() {
  if (!selectedGeneral.value) return
  
  picking.value = true
  try {
    await drawApi.draftPick(selectedGeneral.value.id)
    ElMessage.success(`成功选择 ${selectedGeneral.value.name}！`)
    showConfirm.value = false
    await loadPool()
    await userStore.fetchUser()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '选秀失败')
  } finally {
    picking.value = false
  }
}

function getTierType(tier) {
  const types = { 1: 'danger', 2: 'warning', 3: 'primary', 4: 'success', 5: 'info' }
  return types[tier] || 'info'
}

function getRowClassName({ row }) {
  if (row.salary > userStore.remainingSpace) {
    return 'disabled-row'
  }
  return ''
}
</script>

<style scoped>
.draft-page {
  max-width: 1400px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.confirm-content {
  text-align: center;
  padding: 20px;
}

.confirm-content p {
  margin: 8px 0;
}

:deep(.disabled-row) {
  opacity: 0.5;
  cursor: not-allowed !important;
}
</style>
