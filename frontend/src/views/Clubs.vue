<template>
  <div class="clubs-page">
    <el-card>
      <template #header>
        <span>俱乐部列表 ({{ clubs.length }})</span>
      </template>

      <el-row :gutter="20" v-loading="loading">
        <el-col v-for="club in clubs" :key="club.id" :span="8" style="margin-bottom: 20px">
          <el-card shadow="hover" class="club-card" @click="showClubDetail(club)">
            <h3>
              <span class="club-index">{{ club.excel_id }}.</span>
              {{ club.name }}
            </h3>
            <p class="description">{{ club.description || '暂无基础效果' }}</p>
            <div class="owner" v-if="club.owner">
              <el-tag type="success">拥有者: {{ club.owner.nickname }}</el-tag>
            </div>
            <div class="owner" v-else>
              <el-tag type="info">暂无拥有者</el-tag>
            </div>
            <div class="view-detail">
              <el-button type="primary" link size="small">查看国策详情 →</el-button>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-empty v-if="!clubs.length && !loading" description="暂无俱乐部数据" />
    </el-card>

    <!-- Club Detail Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="selectedClub?.name + ' - 国策详情'"
      width="700px"
      destroy-on-close
    >
      <div v-if="selectedClub" class="club-detail">
        <div class="base-effect">
          <h4>基础效果</h4>
          <p>{{ selectedClub.description || '无' }}</p>
        </div>
        
        <el-divider />
        
        <div class="policies">
          <h4>国策列表</h4>
          <el-table :data="selectedClub.policies || []" stripe>
            <el-table-column type="index" label="#" width="50" />
            <el-table-column label="条件" min-width="200">
              <template #default="{ row }">
                <span v-if="row.condition">{{ row.condition }}</span>
                <el-tag v-else type="success" size="small">无条件(基础效果)</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="effect" label="效果" min-width="250" />
          </el-table>
          <el-empty v-if="!selectedClub.policies?.length" description="暂无国策数据" />
        </div>
        
        <div class="owner-info" v-if="selectedClub.owner">
          <el-divider />
          <el-tag type="success" size="large">当前拥有者: {{ selectedClub.owner.nickname }}</el-tag>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { assetApi } from '../api'
import { ElMessage } from 'element-plus'

const clubs = ref([])
const loading = ref(false)
const dialogVisible = ref(false)
const selectedClub = ref(null)

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

async function showClubDetail(club) {
  try {
    const response = await assetApi.getClubDetail(club.id)
    selectedClub.value = response.data
    dialogVisible.value = true
  } catch (error) {
    console.error('Failed to load club detail:', error)
    ElMessage.error('加载国策详情失败')
  }
}
</script>

<style scoped>
.clubs-page {
  max-width: 1200px;
  margin: 0 auto;
}

.club-card {
  height: 100%;
  cursor: pointer;
  transition: transform 0.2s;
}

.club-card:hover {
  transform: translateY(-4px);
}

.club-card h3 {
  color: #667eea;
  margin-bottom: 12px;
}

.club-index {
  color: #999;
  font-weight: normal;
}

.club-card .description {
  color: #666;
  font-size: 14px;
  min-height: 40px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.club-card .owner {
  margin-top: 12px;
}

.club-card .view-detail {
  margin-top: 12px;
  text-align: right;
}

.club-detail .base-effect h4,
.club-detail .policies h4 {
  font-size: 16px;
  color: #333;
  margin-bottom: 12px;
}

.club-detail .base-effect p {
  color: #666;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
}

.club-detail .owner-info {
  text-align: center;
}
</style>
