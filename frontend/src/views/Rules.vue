<template>
  <div class="rules-page">
    <el-card>
      <template #header>
        <span>游戏规则</span>
      </template>

      <div v-loading="loading">
        <!-- Group rules by category -->
        <div v-for="category in categories" :key="category" class="rule-category">
          <h3 class="category-title">
            <el-icon><Document /></el-icon>
            {{ category }}
          </h3>
          <el-table :data="getRulesByCategory(category)" stripe>
            <el-table-column prop="title" label="项目" width="180" />
            <el-table-column prop="content" label="内容" />
          </el-table>
        </div>

        <el-empty v-if="!rules.length && !loading" description="暂无规则数据" />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { assetApi } from '../api'
import { Document } from '@element-plus/icons-vue'

const rules = ref([])
const loading = ref(false)

const categories = computed(() => {
  const cats = [...new Set(rules.value.map(r => r.category || '其他'))]
  return cats.filter(c => c) // Remove empty categories
})

function getRulesByCategory(category) {
  return rules.value.filter(r => (r.category || '其他') === category)
}

onMounted(async () => {
  loading.value = true
  try {
    const response = await assetApi.getGameRules()
    rules.value = response.data
  } catch (error) {
    console.error('Failed to load rules:', error)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.rules-page {
  max-width: 1000px;
  margin: 0 auto;
}

.rule-category {
  margin-bottom: 30px;
}

.category-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #667eea;
  border-bottom: 2px solid #667eea;
  padding-bottom: 8px;
  margin-bottom: 15px;
}
</style>
