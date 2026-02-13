<template>
  <div class="auction-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>拍卖武将</span>
          <el-tag v-if="stats" type="info">
            已拍卖 {{ stats.auctioned }}/{{ stats.total_generals }}
          </el-tag>
        </div>
      </template>

      <!-- Statistics -->
      <el-row :gutter="20" class="stats-row" v-if="stats">
        <el-col :span="6">
          <el-statistic title="总拍卖武将" :value="stats.total_generals" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="已成交" :value="stats.sold">
            <template #suffix>
              <el-icon color="#67c23a"><CircleCheck /></el-icon>
            </template>
          </el-statistic>
        </el-col>
        <el-col :span="6">
          <el-statistic title="流拍" :value="stats.unsold">
            <template #suffix>
              <el-icon color="#909399"><CircleClose /></el-icon>
            </template>
          </el-statistic>
        </el-col>
        <el-col :span="6">
          <el-statistic title="待拍卖" :value="stats.pending">
            <template #suffix>
              <el-icon color="#e6a23c"><Clock /></el-icon>
            </template>
          </el-statistic>
        </el-col>
      </el-row>

      <el-divider />

      <!-- Auction Results Table -->
      <el-table :data="results" stripe v-loading="loading" style="width: 100%">
        <el-table-column prop="general_name" label="武将" width="100" />
        <el-table-column prop="salary" label="底价" width="80">
          <template #default="{ row }">
            {{ row.salary }}
          </template>
        </el-table-column>
        <el-table-column label="属性" width="200">
          <template #default="{ row }">
            <div v-if="row.general" class="stats-mini">
              <span>统{{ row.general.command }}</span>
              <span>武{{ row.general.force }}</span>
              <span>智{{ row.general.intelligence }}</span>
              <span>政{{ row.general.politics }}</span>
              <span>魅{{ row.general.charm }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="特技" width="150">
          <template #default="{ row }">
            <span v-if="row.general">{{ row.general.skills || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="归属" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.is_unsold" type="info" size="small">流拍</el-tag>
            <el-tag v-else-if="row.user_id" type="success" size="small">
              {{ row.nickname }}
            </el-tag>
            <el-tag v-else type="warning" size="small">待拍卖</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="成交价" width="80">
          <template #default="{ row }">
            <span v-if="row.user_id">{{ row.price }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { CircleCheck, CircleClose, Clock } from '@element-plus/icons-vue'
import { auctionApi } from '../api'

const results = ref([])
const stats = ref(null)
const loading = ref(false)

onMounted(() => {
  loadData()
})

async function loadData() {
  loading.value = true
  try {
    const [resultsRes, statsRes] = await Promise.all([
      auctionApi.getResults(),
      auctionApi.getStats()
    ])
    results.value = resultsRes.data || []
    stats.value = statsRes.data
  } catch (error) {
    console.error('Failed to load auction data:', error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auction-page {
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stats-row {
  margin-bottom: 10px;
}

.stats-mini {
  display: flex;
  gap: 8px;
  font-size: 12px;
  color: #606266;
}

.stats-mini span {
  background: #f4f4f5;
  padding: 2px 6px;
  border-radius: 4px;
}
</style>
