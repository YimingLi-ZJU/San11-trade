<template>
  <div class="player-detail">
    <el-page-header @back="$router.back()" :content="roster?.user?.nickname || '玩家详情'" />
    
    <el-card style="margin-top: 20px" v-if="roster">
      <template #header>
        <div class="card-header">
          <span>{{ roster.user.nickname }} 的阵容</span>
          <div class="space-info">
            <span>空间：{{ roster.user.used_space }} / {{ roster.user.space }}</span>
          </div>
        </div>
      </template>

      <el-tabs v-model="activeTab">
        <el-tab-pane label="武将" name="generals">
          <el-table :data="roster.generals || []" stripe>
            <el-table-column prop="name" label="姓名" width="100" />
            <el-table-column prop="command" label="统率" width="80" />
            <el-table-column prop="force" label="武力" width="80" />
            <el-table-column prop="intelligence" label="智力" width="80" />
            <el-table-column prop="politics" label="政治" width="80" />
            <el-table-column prop="charm" label="魅力" width="80" />
            <el-table-column prop="salary" label="薪资" width="80" />
            <el-table-column prop="skills" label="特技" />
          </el-table>
          <el-empty v-if="!roster.generals?.length" description="暂无武将" />
        </el-tab-pane>

        <el-tab-pane label="宝物" name="treasures">
          <el-table :data="roster.treasures || []" stripe>
            <el-table-column prop="name" label="名称" width="120" />
            <el-table-column prop="type" label="类型" width="100" />
            <el-table-column prop="effect" label="效果" />
          </el-table>
          <el-empty v-if="!roster.treasures?.length" description="暂无宝物" />
        </el-tab-pane>

        <el-tab-pane label="俱乐部" name="club">
          <div v-if="roster.club" class="club-info">
            <h3>{{ roster.club.name }}</h3>
            <p>{{ roster.club.description }}</p>
          </div>
          <el-empty v-else description="暂未选择俱乐部" />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-card v-else-if="!loading">
      <el-empty description="玩家不存在" />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { gameApi } from '../api'

const route = useRoute()
const roster = ref(null)
const loading = ref(false)
const activeTab = ref('generals')

onMounted(async () => {
  loading.value = true
  try {
    const response = await gameApi.getPlayerRoster(route.params.id)
    roster.value = response.data
  } catch (error) {
    console.error('Failed to load player roster:', error)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.player-detail {
  max-width: 1200px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.space-info {
  color: #666;
  font-size: 14px;
}

.club-info {
  padding: 20px;
}

.club-info h3 {
  color: #667eea;
  margin-bottom: 12px;
}
</style>
