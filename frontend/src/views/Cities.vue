<template>
  <div class="cities-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>城市列表 ({{ cities.length }})</span>
          <el-input
            v-model="searchText"
            placeholder="搜索城市..."
            prefix-icon="Search"
            style="width: 200px"
            clearable
          />
        </div>
      </template>

      <el-table
        :data="filteredCities"
        stripe
        :default-sort="{ prop: 'excel_id', order: 'ascending' }"
        v-loading="loading"
      >
        <el-table-column prop="excel_id" label="序号" width="80" sortable />
        <el-table-column prop="name" label="名称" width="100" />
        <el-table-column label="特产" width="80">
          <template #default="{ row }">
            <el-tag :type="getSpecialtyType(row.specialty)" size="small">
              {{ row.specialty || '-' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="max_soldiers" label="最大士兵" width="120" sortable>
          <template #default="{ row }">
            {{ formatNumber(row.max_soldiers) }}
          </template>
        </el-table-column>
        <el-table-column prop="gold_income" label="金收入" width="100" sortable />
        <el-table-column prop="food_income" label="粮收入" width="100" sortable />
        <el-table-column prop="durability" label="耐久" width="100" sortable />
        <el-table-column prop="tiles" label="地块" width="80" sortable />
      </el-table>

      <el-empty v-if="!cities.length && !loading" description="暂无城市数据" />
    </el-card>

    <!-- Summary Statistics -->
    <el-card style="margin-top: 20px" v-if="cities.length">
      <template #header>
        <span>特产统计</span>
      </template>
      <el-row :gutter="20">
        <el-col :span="4" v-for="(count, specialty) in specialtyStats" :key="specialty">
          <div class="stat-item">
            <el-tag :type="getSpecialtyType(specialty)" size="large">{{ specialty }}</el-tag>
            <span class="count">{{ count }} 个城市</span>
          </div>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { assetApi } from '../api'

const cities = ref([])
const loading = ref(false)
const searchText = ref('')

const filteredCities = computed(() => {
  if (!searchText.value) return cities.value
  const text = searchText.value.toLowerCase()
  return cities.value.filter(c =>
    c.name.toLowerCase().includes(text) ||
    c.specialty?.toLowerCase().includes(text)
  )
})

const specialtyStats = computed(() => {
  const stats = {}
  cities.value.forEach(city => {
    const sp = city.specialty || '无'
    stats[sp] = (stats[sp] || 0) + 1
  })
  return stats
})

onMounted(async () => {
  loading.value = true
  try {
    const response = await assetApi.getAllCities()
    cities.value = response.data
  } catch (error) {
    console.error('Failed to load cities:', error)
  } finally {
    loading.value = false
  }
})

function getSpecialtyType(specialty) {
  const types = {
    '马': 'danger',
    '工': 'warning',
    '弩': 'success',
    '铁': 'info',
    '盐': 'primary'
  }
  return types[specialty] || 'default'
}

function formatNumber(num) {
  return num ? num.toLocaleString() : '-'
}
</script>

<style scoped>
.cities-page {
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.stat-item .count {
  color: #666;
  font-size: 14px;
}
</style>
