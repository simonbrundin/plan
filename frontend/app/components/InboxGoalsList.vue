<script setup lang="ts">
import { useGoalsStore } from '~/stores/goals'

const goalsStore = useGoalsStore()
const inboxGoals = computed(() => goalsStore.getInboxGoals)

const moveFromInbox = async (goalId: number) => {
  try {
    // TODO: Call GraphQL mutation to move goal from inbox
    // For now, just update the store
    goalsStore.moveGoalFromInbox(goalId)
  } catch (error) {
    console.error('Failed to move goal from inbox:', error)
  }
}
</script>

<template>
  <div class="space-y-4">
    <h2 class="text-xl font-semibold text-gray-900">Inbox</h2>

    <div v-if="inboxGoals.length === 0" class="text-gray-500 text-center py-8">
      Din inbox är tom. Lägg till nya mål ovan!
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="goal in inboxGoals"
        :key="goal.id"
        class="flex items-center justify-between p-4 bg-white border border-gray-200 rounded-lg shadow-sm"
      >
        <div class="flex items-center space-x-3">
          <Icon :name="goal.icon || 'heroicons:star'" class="w-6 h-6 text-gray-600" />
          <span class="text-gray-900">{{ goal.title }}</span>
        </div>

        <button
          @click="moveFromInbox(goal.id)"
          class="px-3 py-1 text-sm bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
        >
          Bearbeta
        </button>
      </div>
    </div>
  </div>
</template>