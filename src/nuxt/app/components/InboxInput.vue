<script setup lang="ts">
import { useGoalsStore } from '~/stores/goals'

const emit = defineEmits<{
  add: [goal: { title: string; inbox: number }]
}>()

const goalsStore = useGoalsStore()
const newGoalTitle = ref("")

const addGoal = () => {
  const title = newGoalTitle.value.trim()
  if (title) {
    emit("add", { title, inbox: 1 })
    newGoalTitle.value = ""
  }
}

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === "Enter") {
    addGoal()
  }
}
</script>

<template>
  <div class="flex gap-2 mb-6">
    <input
      v-model="newGoalTitle"
      @keydown="handleKeydown"
      type="text"
      placeholder="Lägg till ett nytt mål i inbox..."
      class="rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30 flex-1"
    />
    <button
      @click="addGoal"
      :disabled="!newGoalTitle.trim()"
      class="rounded-lg bg-accent-strong px-4 py-2 font-medium text-white transition-colors hover:bg-accent-deep disabled:opacity-40 disabled:cursor-not-allowed disabled:opacity-40 disabled:cursor-not-allowed"
    >
      Lägg till
    </button>
  </div>
</template>