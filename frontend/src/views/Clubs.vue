<template>
  <div class="clubs-page">
    <el-card>
      <template #header>
        <span>俱乐部列表</span>
      </template>

      <el-row :gutter="20">
        <el-col v-for="club in clubs" :key="club.id" :span="8" style="margin-bottom: 20px">
          <el-card shadow="hover" class="club-card">
            <h3>{{ club.name }}</h3>
            <p class="description">{{ club.description }}</p>
            <div class="owner" v-if="club.owner">
              <el-tag type="success">拥有者: {{ club.owner.nickname }}</el-tag>
            </div>
            <div class="owner" v-else>
              <el-tag type="info">暂无拥有者</el-tag>
            </div>
            <div class="policy" v-if="club.policy">
              <el-divider />
              <h4>国策</h4>
              <p>{{ club.policy }}</p>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-empty v-if="!clubs.length && !loading" description="暂无俱乐部数据" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { assetApi } from '../api'

const clubs = ref([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const response = await assetApi.getAllClubs()
    clubs.value = response.data
  } catch (error) {
    console.error('Failed to load clubs:', error)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.clubs-page {
  max-width: 1200px;
  margin: 0 auto;
}

.club-card {
  height: 100%;
}

.club-card h3 {
  color: #667eea;
  margin-bottom: 12px;
}

.club-card .description {
  color: #666;
  font-size: 14px;
  min-height: 40px;
}

.club-card .owner {
  margin-top: 12px;
}

.club-card .policy {
  font-size: 13px;
}

.club-card .policy h4 {
  font-size: 14px;
  color: #333;
  margin-bottom: 8px;
}

.club-card .policy p {
  color: #666;
}
</style>
