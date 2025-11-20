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
      class="flex-1 px-4 py-2 border text-gray-600 border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
    />
    <button
      @click="addGoal"
      :disabled="!newGoalTitle.trim()"
      class="px-6 py-2 bg-blue-600 text-gray-600 rounded-lg hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors"
    >
      Lägg till
    </button>
  </div>
</template>