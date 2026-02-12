<template>
  <div class="treasures-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>宝物列表</span>
          <el-input
            v-model="searchText"
            placeholder="搜索宝物..."
            prefix-icon="Search"
            style="width: 200px"
            clearable
          />
        </div>
      </template>

      <el-table :data="filteredTreasures" stripe v-loading="loading">
        <el-table-column prop="name" label="名称" width="120" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag>{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="value" label="价值" width="80" sortable />
        <el-table-column prop="effect" label="效果" />
        <el-table-column prop="skill" label="特技" width="120" />
        <el-table-column label="归属" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.owner" type="success">{{ row.owner.nickname }}</el-tag>
            <el-tag v-else type="info">未归属</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { assetApi } from '../api'

const treasures = ref([])
const loading = ref(false)
const searchText = ref('')

const filteredTreasures = computed(() => {
  if (!searchText.value) return treasures.value
  const text = searchText.value.toLowerCase()
  return treasures.value.filter(t =>
    t.name.toLowerCase().includes(text) ||
    t.type?.toLowerCase().includes(text) ||
    t.effect?.toLowerCase().includes(text)
  )
})

onMounted(async () => {
  loading.value = true
  try {
    const response = await assetApi.getAllTreasures()
    treasures.value = response.data
  } catch (error) {
    console.error('Failed to load treasures:', error)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.treasures-page {
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
