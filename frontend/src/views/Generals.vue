<template>
  <div class="generals-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>武将列表</span>
          <el-input
            v-model="searchText"
            placeholder="搜索武将..."
            prefix-icon="Search"
            style="width: 200px"
            clearable
          />
        </div>
      </template>

      <el-table
        :data="filteredGenerals"
        stripe
        :default-sort="{ prop: 'tier', order: 'ascending' }"
        v-loading="loading"
      >
        <el-table-column prop="name" label="姓名" width="100" fixed />
        <el-table-column label="档次" width="80" sortable prop="tier">
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
        <el-table-column prop="skills" label="特技" min-width="150" />
        <el-table-column label="归属" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.owner" type="success">{{ row.owner.nickname }}</el-tag>
            <el-tag v-else type="info">未归属</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="池类型" width="100">
          <template #default="{ row }">
            <el-tag type="warning">{{ getPoolName(row.pool_type) }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { assetApi } from '../api'

const generals = ref([])
const loading = ref(false)
const searchText = ref('')

const filteredGenerals = computed(() => {
  if (!searchText.value) return generals.value
  const text = searchText.value.toLowerCase()
  return generals.value.filter(g =>
    g.name.toLowerCase().includes(text) ||
    g.skills?.toLowerCase().includes(text)
  )
})

onMounted(async () => {
  loading.value = true
  try {
    const response = await assetApi.getAllGenerals()
    generals.value = response.data
  } catch (error) {
    console.error('Failed to load generals:', error)
  } finally {
    loading.value = false
  }
})

function getTierType(tier) {
  const types = { 1: 'danger', 2: 'warning', 3: 'primary', 4: 'success', 5: 'info' }
  return types[tier] || 'info'
}

function getPoolName(poolType) {
  const names = {
    guarantee: '保底',
    normal: '普通',
    draft: '选秀',
    second: '二抽',
    bigcore: '大核'
  }
  return names[poolType] || poolType
}
</script>

<style scoped>
.generals-page {
  max-width: 1400px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
