<script setup lang="ts">
import type { Goal } from '~/types/goal'

const { data: goals, pending, error } = await useFetch<Goal[]>('/api/goals')

const rootGoals = computed(() => {
  if (!goals.value) return []
  // Root goals are those without parent relations
  // For now, show all goals - the user can navigate to any
  return goals.value
})

definePageMeta({
  title: 'Mina mål',
})
</script>

<template>
  <div class="p-6 max-w-4xl mx-auto">
    <h1 class="text-3xl font-bold mb-6">Mina mål</h1>

    <div v-if="pending" class="text-gray-400">
      Laddar...
    </div>

    <div v-else-if="error" class="text-red-400">
      Fel vid laddning av mål: {{ error.message }}
    </div>

    <div v-else-if="rootGoals.length === 0" class="text-gray-400">
      Inga mål hittades.
    </div>

    <div v-else class="space-y-2">
      <NuxtLink
        v-for="goal in rootGoals"
        :key="goal.id"
        :to="`/goal/${goal.id}`"
        class="flex items-center gap-3 p-4 bg-gray-800 rounded-lg hover:bg-gray-700 transition-colors"
      >
        <Icon :name="goal.icon || 'heroicons:star'" class="w-6 h-6" />
        <span class="flex-1">{{ goal.title }}</span>
        <span v-if="goal.finished" class="text-green-400 text-sm">Klar</span>
      </NuxtLink>
    </div>
  </div>
</template>
