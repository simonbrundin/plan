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
    <h2 class="text-sm font-semibold uppercase tracking-wider text-ink-dim">Inbox</h2>

    <div v-if="inboxGoals.length === 0" class="text-ink-dim text-center py-10 border border-dashed border-line rounded-2xl">
      Din inbox är tom. Lägg till nya mål ovan!
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="goal in inboxGoals"
        :key="goal.id"
        class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 flex items-center justify-between px-4 py-3"
      >
        <div class="flex items-center space-x-3 min-w-0">
          <Icon :name="goal.icon || 'heroicons:star'" class="w-6 h-6 text-ink-muted flex-shrink-0" />
          <span class="text-ink truncate">{{ goal.title }}</span>
        </div>

        <button
          @click="moveFromInbox(goal.id)"
          class="ml-3 px-3 py-1.5 text-sm bg-accent-strong text-white rounded-lg hover:bg-accent-deep transition-colors flex-shrink-0"
        >
          Bearbeta
        </button>
      </div>
    </div>
  </div>
</template>