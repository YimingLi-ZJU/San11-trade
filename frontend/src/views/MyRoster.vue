<template>
  <div class="roster-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>我的阵容</span>
          <div class="space-info">
            <span>空间：{{ userStore.user?.used_space || 0 }} / {{ userStore.user?.space || 350 }}</span>
            <el-progress
              :percentage="spacePercentage"
              :stroke-width="8"
              style="width: 150px; margin-left: 16px"
            />
          </div>
        </div>
      </template>

      <el-tabs v-model="activeTab">
        <el-tab-pane label="武将" name="generals">
          <el-table :data="roster?.generals || []" stripe>
            <el-table-column prop="name" label="姓名" width="100" />
            <el-table-column prop="command" label="统率" width="80" sortable />
            <el-table-column prop="force" label="武力" width="80" sortable />
            <el-table-column prop="intelligence" label="智力" width="80" sortable />
            <el-table-column prop="politics" label="政治" width="80" sortable />
            <el-table-column prop="charm" label="魅力" width="80" sortable />
            <el-table-column prop="salary" label="薪资" width="80" sortable />
            <el-table-column prop="skills" label="特技" min-width="150" />
            <el-table-column label="档次" width="80">
              <template #default="{ row }">
                <el-tag :type="getTierType(row.tier)">T{{ row.tier }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="!roster?.generals?.length" description="暂无武将" />
        </el-tab-pane>

        <el-tab-pane label="宝物" name="treasures">
          <el-table :data="roster?.treasures || []" stripe>
            <el-table-column prop="name" label="名称" width="120" />
            <el-table-column prop="type" label="类型" width="100" />
            <el-table-column prop="effect" label="效果" />
            <el-table-column prop="skill" label="特技" width="120" />
          </el-table>
          <el-empty v-if="!roster?.treasures?.length" description="暂无宝物" />
        </el-tab-pane>

        <el-tab-pane label="俱乐部" name="club">
          <div v-if="roster?.club" class="club-info">
            <h3>{{ roster.club.name }}</h3>
            <p class="description">{{ roster.club.description }}</p>
            <div class="policy" v-if="roster.club.policy">
              <h4>国策</h4>
              <p>{{ roster.club.policy }}</p>
            </div>
          </div>
          <el-empty v-else description="暂未选择俱乐部" />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '../stores/user'
import { authApi } from '../api'

const userStore = useUserStore()
const activeTab = ref('generals')
const roster = ref(null)

const spacePercentage = computed(() => {
  if (!userStore.user) return 0
  return Math.round((userStore.user.used_space / userStore.user.space) * 100)
})

onMounted(async () => {
  try {
    const response = await authApi.getMyRoster()
    roster.value = response.data
  } catch (error) {
    console.error('Failed to load roster:', error)
  }
})

function getTierType(tier) {
  const types = {
    1: 'danger',
    2: 'warning',
    3: 'primary',
    4: 'success',
    5: 'info'
  }
  return types[tier] || 'info'
}
</script>

<style scoped>
.roster-page {
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.space-info {
  display: flex;
  align-items: center;
  font-size: 14px;
  color: #666;
}

.club-info {
  padding: 20px;
}

.club-info h3 {
  color: #667eea;
  margin-bottom: 16px;
}

.club-info .description {
  color: #666;
  margin-bottom: 20px;
}

.club-info .policy {
  background: #f5f7fa;
  padding: 16px;
  border-radius: 8px;
}

.club-info .policy h4 {
  margin-bottom: 8px;
  color: #333;
}
</style>
