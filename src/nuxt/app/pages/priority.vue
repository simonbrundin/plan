<script setup lang="ts">
import type { Goal } from '~/types/goal'
import { canBeFinished } from '~/utils/goalRules'
import { ETERNAL_ACCENT, ETERNAL_BORDER, ETERNAL_TEXT, ETERNAL_TEXT_HOVER } from '~/utils/eternalStyle'

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
  // Eviga mål pågår hela livet och kan inte slutföras.
  if (!canBeFinished(goal)) {
    return
  }

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
  if (weight <= 2) return 'text-ink-muted'
  if (weight <= 9) return 'text-ink-dim'
  if (weight <= 14) return 'text-ink-dim'
  return 'text-accent'
}

function getWeightBgColor(weight: number): string {
  if (weight <= 2) return 'bg-elevated/30'
  if (weight <= 9) return 'bg-elevated/50'
  if (weight <= 14) return 'bg-elevated/70'
  return 'bg-accent/30'
}

const filterOptions = [
  { value: 'all', label: 'Alla' },
  { value: 'started', label: 'Påbörjade' },
  { value: 'not_started', label: 'Ej påbörjade' },
] as const
</script>

<template>
  <div v-if="user" class="max-w-4xl mx-auto px-6 py-8">
    <div class="flex flex-wrap items-start justify-between gap-4 mb-6">
      <div class="flex items-center gap-3">
        <h1 class="text-3xl font-bold tracking-tight text-ink">Prioriteringslista</h1>
        <span
          v-if="isPriorityMode"
          class="px-3 py-1 rounded-full text-xs bg-fuchsia-600/20 text-fuchsia-300 border border-fuchsia-500/30"
        >
          Prioriteringsläge (håll P + K/J)
        </span>
      </div>
      <NuxtLink to="/" class="rounded-lg border border-line bg-surface-2 px-4 py-2 text-ink-muted transition-colors hover:border-line-strong hover:text-ink inline-flex items-center gap-2">
        <Icon name="lucide:arrow-left" class="w-4 h-4" />
        Tillbaka
      </NuxtLink>
    </div>

    <div class="text-xs text-ink-dim mb-6">
      <span class="text-ink-muted">j/k</span> navigera &nbsp;|&nbsp;
      <span class="text-ink-muted">l</span> öppna mål &nbsp;|&nbsp;
      <span class="text-ink-muted">Håll P</span> + <span class="text-ink-muted">K/J</span> ändra vikt &nbsp;|&nbsp;
      <span class="text-ink-muted">s</span> markera påbörjad
    </div>

    <div class="flex flex-wrap items-center gap-2 mb-6">
      <button
        v-for="option in filterOptions"
        :key="option.value"
        type="button"
        @click="setStartedFilter(option.value)"
        class="px-3 py-1.5 rounded-full text-sm transition-colors border"
        :class="
          startedFilter === option.value
            ? 'bg-accent-strong text-white border-accent-strong'
            : 'bg-surface-2 text-ink-muted border-line hover:border-line-strong hover:text-ink'
        "
      >
        {{ option.label }}
      </button>
      <span class="text-xs text-ink-dim ml-2">
        {{ visibleGoals.length }} av {{ prioritizedGoals.length }} mål visas
      </span>
    </div>

    <div v-if="isLoading" class="text-center py-12">
      <p class="text-ink-muted">Laddar...</p>
    </div>

    <div v-else-if="prioritizedGoals.length === 0" class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 p-10 text-center">
      <p class="text-ink-muted">Inga mål att prioritera</p>
    </div>

    <div v-else-if="visibleGoals.length === 0" class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 p-10 text-center">
      <p class="text-ink-muted">
        Inga mål matchar filtret &quot;{{
          filterOptions.find((o) => o.value === startedFilter)?.label
        }}&quot;
      </p>
    </div>

    <div v-else class="space-y-2">
      <div
        v-for="(goal, index) in visibleGoals"
        :key="goal.id"
        :data-index="index"
        class="rounded-2xl border border-line transition-colors hover:border-line-strong flex items-center gap-4 px-4 py-3.5 cursor-pointer"
        :class="[
          getWeightBgColor(goal.weight),
          selectedGoalId === goal.id ? 'ring-1 ring-accent' : '',
          isPriorityMode && selectedGoalId === goal.id ? 'ring-1 ring-fuchsia-500' : ''
        ]"
        @click="selectGoal(goal.id)"
      >
        <div class="flex-shrink-0 w-8 text-center">
          <span class="text-lg font-bold tabular-nums" :class="getWeightColor(goal.weight)">
            {{ index + 1 }}
          </span>
        </div>

        <div class="flex-shrink-0">
          <button
            @click.stop="toggleStarted(goal)"
            class="rounded-lg p-1.5 transition-colors hover:bg-elevated"
            :class="goal.started ? 'text-yellow-400' : 'text-ink-dim hover:text-ink-muted'"
            :title="goal.started ? 'Avbörja (s)' : 'Påbörja (s)'"
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
            class="w-6 h-6 rounded-md border-2 flex items-center justify-center transition-colors"
            :class="goal.finished ? 'bg-green-600 border-green-600' : goal.eternal ? ETERNAL_BORDER + ' cursor-not-allowed' : 'border-line-strong hover:border-ink-muted'"
            :title="goal.eternal ? 'Ett evigt mål kan inte slutföras' : goal.finished ? 'Avmarkera som klar' : 'Markera som klar'"
            :disabled="goal.eternal"
          >
            <svg v-if="goal.finished" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
          </button>
        </div>

        <div class="flex items-center gap-2 flex-1 min-w-0">
          <Icon :name="goal.icon || 'heroicons:star'" class="w-5 h-5 flex-shrink-0 text-ink-muted"
            :class="goal.eternal ? ETERNAL_TEXT : ''" />
          <NuxtLink
            :to="`/goal/${goal.id}`"
            class="truncate transition-colors"
            :class="goal.eternal ? ETERNAL_TEXT + ' ' + ETERNAL_TEXT_HOVER : 'text-ink hover:text-white'"
            @click.stop
          >
            {{ goal.title }}
          </NuxtLink>
          <span v-if="goal.eternal" :class="ETERNAL_ACCENT + ' text-xs flex-shrink-0'">∞</span>
        </div>

        <div class="flex-shrink-0 text-right">
          <div
            class="text-sm font-mono px-2 py-1 rounded-lg min-w-[3rem] text-center tabular-nums"
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
      <h1 class="text-3xl font-bold tracking-tight text-ink mb-6">Prioriteringslista</h1>
      <NuxtLink to="/login">
        <UButton size="lg" icon="material-symbols:login-rounded">
          Logga in
        </UButton>
      </NuxtLink>
    </div>
  </div>
</template>
