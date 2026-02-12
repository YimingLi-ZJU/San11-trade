<template>
  <div class="players-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>玩家列表</span>
          <el-tag type="info">共 {{ players.length }} 名玩家</el-tag>
        </div>
      </template>

      <el-table :data="players" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="nickname" label="昵称" width="120" />
        <el-table-column prop="username" label="用户名" width="120" />
        <el-table-column label="空间使用" width="150">
          <template #default="{ row }">
            <el-progress
              :percentage="Math.round((row.used_space / row.space) * 100)"
              :stroke-width="8"
            />
            <span style="font-size: 12px; color: #999">
              {{ row.used_space }} / {{ row.space }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="俱乐部" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.club" type="warning">{{ row.club.name }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button text type="primary" @click="viewPlayer(row.id)">
              查看阵容
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { gameApi } from '../api'

const router = useRouter()
const players = ref([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const response = await gameApi.getPlayers()
    players.value = response.data
  } catch (error) {
    console.error('Failed to load players:', error)
  } finally {
    loading.value = false
  }
})

function viewPlayer(id) {
  router.push(`/players/${id}`)
}
</script>

<style scoped>
.players-page {
  max-width: 1000px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
