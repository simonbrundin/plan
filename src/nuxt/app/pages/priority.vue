<script setup lang="ts">
import type { Goal } from '~/types/goal'

definePageMeta({
  ssr: false,
})

const { user } = useUserSession()
const { toggleGoalFinished, toggleGoalStarted } = useGoalApi()

type PrioritizedGoalRow = Goal & {
  weight: number
  parentTitle: string | null
}

const {
  isPriorityMode,
  selectedGoalId,
  prioritizedGoals,
  visibleGoals,
  isLoading,
  startedFilter,
  loadPrioritizedGoals,
  selectGoal,
  setStartedFilter,
} = usePriorityMode()

onMounted(() => {
  loadPrioritizedGoals()

  const onKeydown = (e: KeyboardEvent) => {
    const tag = (e.target as HTMLElement).tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA' || (e.target as HTMLElement).isContentEditable) return
    if (e.key === 's' && !e.ctrlKey && !e.metaKey) {
      e.preventDefault()
      const goal = visibleGoals.value.find(g => g.id === selectedGoalId.value)
      if (goal) toggleStarted(goal)
    }
  }
  window.addEventListener('keydown', onKeydown)
  onUnmounted(() => window.removeEventListener('keydown', onKeydown))
})

async function toggleFinished(goal: PrioritizedGoalRow) {
  try {
    const newFinishedValue = goal.finished ? null : new Date().toISOString()
    await toggleGoalFinished(goal.id, newFinishedValue)
    await loadPrioritizedGoals()
  } catch (err) {
    console.error('Failed to toggle goal finished:', err)
  }
}

async function toggleStarted(goal: PrioritizedGoalRow) {
  try {
    const newStartedValue = goal.started ? null : new Date().toISOString()
    await toggleGoalStarted(goal.id, newStartedValue)
    await loadPrioritizedGoals()
  } catch (err) {
    console.error('Failed to toggle goal started:', err)
  }
}

function getWeightColor(weight: number): string {
  if (weight <= 2) return 'text-gray-400'
  if (weight <= 9) return 'text-gray-500'
  if (weight <= 14) return 'text-gray-600'
  return 'text-blue-400'
}

function getWeightBgColor(weight: number): string {
  if (weight <= 2) return 'bg-gray-700/30'
  if (weight <= 9) return 'bg-gray-700/50'
  if (weight <= 14) return 'bg-gray-700/70'
  return 'bg-blue-900/30'
}

const filterOptions = [
  { value: 'all', label: 'Alla' },
  { value: 'started', label: 'Påbörjade' },
  { value: 'not_started', label: 'Ej påbörjade' },
] as const
</script>

<template>
  <div v-if="user" class="max-w-4xl mx-auto p-6">
    <div class="flex items-center justify-between mb-8">
      <div class="flex items-center gap-4">
        <h1 class="text-3xl font-bold text-gray-100">Prioriteringslista</h1>
        <div v-if="isPriorityMode" class="px-3 py-1 bg-purple-600/30 rounded-full text-sm text-purple-300">
          Prioriteringsläge (håll P + K/J)
        </div>
      </div>
      <NuxtLink to="/" class="text-gray-400 hover:text-gray-200 transition-colors">
        ← Tillbaka
      </NuxtLink>
    </div>

    <div class="text-sm text-gray-500 mb-4">
      <span class="text-gray-400">j/k</span> navigera &nbsp;|&nbsp;
      <span class="text-gray-400">l</span> öppna mål &nbsp;|&nbsp;
      <span class="text-gray-400">Håll P</span> + <span class="text-gray-400">K/J</span> ändra vikt &nbsp;|&nbsp;
      <span class="text-gray-400">s</span> markera påbörjad
    </div>

    <div class="flex flex-wrap items-center gap-2 mb-6">
      <button
        v-for="option in filterOptions"
        :key="option.value"
        type="button"
        @click="setStartedFilter(option.value)"
        class="px-3 py-1.5 rounded-full text-sm transition-colors"
        :class="
          startedFilter === option.value
            ? 'bg-blue-600 text-white'
            : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
        "
      >
        {{ option.label }}
      </button>
      <span class="text-xs text-gray-500 ml-2">
        {{ visibleGoals.length }} av {{ prioritizedGoals.length }} mål visas
      </span>
    </div>

    <div v-if="isLoading" class="text-center py-12">
      <p class="text-gray-500">Laddar...</p>
    </div>

    <div v-else-if="prioritizedGoals.length === 0" class="text-center py-12">
      <p class="text-gray-500">Inga mål att prioritera</p>
    </div>

    <div v-else-if="visibleGoals.length === 0" class="text-center py-12">
      <p class="text-gray-500">
        Inga mål matchar filtret &quot;{{
          filterOptions.find((o) => o.value === startedFilter)?.label
        }}&quot;
      </p>
    </div>

    <div v-else class="space-y-3">
      <div
        v-for="(goal, index) in visibleGoals"
        :key="goal.id"
        :data-index="index"
        class="flex items-center gap-4 p-4 rounded-lg transition-all cursor-pointer"
        :class="[
          getWeightBgColor(goal.weight),
          selectedGoalId === goal.id ? 'ring-2 ring-blue-500' : '',
          isPriorityMode && selectedGoalId === goal.id ? 'ring-2 ring-purple-500' : ''
        ]"
        @click="selectGoal(goal.id)"
      >
        <div class="flex-shrink-0 w-8 text-center">
          <span class="text-lg font-bold" :class="getWeightColor(goal.weight)">
            {{ index + 1 }}
          </span>
        </div>

        <div class="flex-shrink-0">
          <button
            @click.stop="toggleStarted(goal)"
            class="p-1.5 rounded transition-colors"
            :class="goal.started
              ? 'text-yellow-400 hover:text-yellow-300'
              : 'text-gray-600 hover:text-gray-400'"
            :title="goal.started ? 'Avmarkerad (s)' : 'Påbörja (s)'"
          >
            <Icon
              name="lucide:circle-play"
              class="w-5 h-5"
              :style="{ opacity: goal.started ? 1 : 0.25 }"
            />
          </button>
        </div>

        <div class="flex-shrink-0">
          <button
            @click.stop="toggleFinished(goal)"
            class="w-6 h-6 rounded border-2 border-gray-600 flex items-center justify-center transition-colors hover:border-gray-400"
            :class="goal.finished ? 'bg-green-600 border-green-600' : ''"
            :title="goal.finished ? 'Avmarkera som klar' : 'Markera som klar'"
          >
            <svg v-if="goal.finished" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
          </button>
        </div>

        <div class="flex items-center gap-2 flex-1 min-w-0">
          <Icon :name="goal.icon || 'heroicons:star'" class="w-5 h-5 text-gray-400 flex-shrink-0" />
          <NuxtLink
            :to="`/goal/${goal.id}`"
            class="text-gray-200 hover:text-white transition-colors truncate"
            @click.stop
          >
            {{ goal.title }}
          </NuxtLink>
        </div>

        <div class="flex-shrink-0 text-right">
          <div
            class="text-sm font-mono px-2 py-1 rounded min-w-[3rem] text-center"
            :class="getWeightBgColor(goal.weight) + ' ' + getWeightColor(goal.weight)"
          >
            {{ goal.weight }}
          </div>
        </div>
      </div>
    </div>
  </div>

  <div v-else class="flex items-center justify-center min-h-[50vh]">
    <div class="text-center">
      <h1 class="text-3xl font-bold text-gray-900 mb-6">Prioriteringslista</h1>
      <NuxtLink to="/login">
        <UButton size="lg" icon="material-symbols:login-rounded">
          Logga in
        </UButton>
      </NuxtLink>
    </div>
  </div>
</template>