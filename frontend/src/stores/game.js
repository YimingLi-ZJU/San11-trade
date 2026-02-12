import { defineStore } from 'pinia'
import { ref } from 'vue'
import { gameApi } from '../api'

export const useGameStore = defineStore('game', () => {
  const phase = ref(null)
  const statistics = ref(null)
  const loading = ref(false)

  async function fetchPhase() {
    try {
      const response = await gameApi.getPhase()
      phase.value = response.data
      return response.data
    } catch (error) {
      console.error('Failed to fetch game phase:', error)
      throw error
    }
  }

  async function fetchStatistics() {
    try {
      const response = await gameApi.getStatistics()
      statistics.value = response.data
      return response.data
    } catch (error) {
      console.error('Failed to fetch statistics:', error)
      throw error
    }
  }

  // Phase display names
  const phaseNames = {
    signup: '报名阶段',
    guarantee_draw: '保底抽将',
    normal_draw: '普通抽将',
    draft: '选秀阶段',
    trading: '自由交易',
    auction: '拍卖阶段',
    match: '比赛阶段',
    finished: '赛季结束'
  }

  function getPhaseName(phaseName) {
    return phaseNames[phaseName] || phaseName
  }

  return {
    phase,
    statistics,
    loading,
    fetchPhase,
    fetchStatistics,
    getPhaseName
  }
})
