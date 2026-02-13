<template>
  <div class="generals-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>武将列表 ({{ filteredGenerals.length }})</span>
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
        :default-sort="{ prop: 'excel_id', order: 'ascending' }"
        v-loading="loading"
        max-height="700"
      >
        <el-table-column prop="excel_id" label="序号" width="70" sortable fixed />
        <el-table-column prop="name" label="姓名" width="90" fixed />
        <el-table-column prop="salary" label="价值" width="65" sortable />
        <el-table-column prop="command" label="统" width="55" sortable />
        <el-table-column prop="force" label="武" width="55" sortable />
        <el-table-column prop="intelligence" label="智" width="55" sortable />
        <el-table-column prop="politics" label="政" width="55" sortable />
        <el-table-column prop="charm" label="魅" width="55" sortable />
        <el-table-column prop="affinity" label="相性" width="65" sortable />
        <el-table-column label="兵种适性" width="180">
          <template #default="{ row }">
            <span class="aptitude">
              <span title="枪">{{ row.spear || '-' }}</span>
              <span title="戟">{{ row.halberd || '-' }}</span>
              <span title="弩">{{ row.crossbow || '-' }}</span>
              <span title="骑">{{ row.cavalry || '-' }}</span>
              <span title="兵">{{ row.soldier || '-' }}</span>
              <span title="水">{{ row.water || '-' }}</span>
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="skills" label="特技" min-width="120" show-overflow-tooltip />
        <el-table-column prop="morality" label="义理" width="80" />
        <el-table-column prop="ambition" label="野望" width="65" />
        <el-table-column prop="personality" label="性格" width="65" />
        <el-table-column label="归属" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.owner" type="success" size="small">{{ row.owner.nickname }}</el-tag>
            <span v-else class="no-owner">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="note" label="改动说明" width="150" show-overflow-tooltip />
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
    g.skills?.toLowerCase().includes(text) ||
    g.note?.toLowerCase().includes(text)
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
</script>

<style scoped>
.generals-page {
  max-width: 1600px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.aptitude {
  display: flex;
  gap: 8px;
  font-family: monospace;
}

.aptitude span {
  min-width: 16px;
  text-align: center;
}

.no-owner {
  color: #999;
}
</style>
