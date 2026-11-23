<script setup lang="ts">
import type { Goal } from '~/types/goal'

defineProps<{
  goal: Goal | undefined
  isGoalSelected: boolean
  editingGoalId: number | null
  editTitle: string
  mode: 'normal' | 'insert'
  showDependencies: boolean
}>()

const emit = defineEmits<{
  'update:editTitle': [value: string]
  'save-edit': []
  'cancel-edit': []
  'open-icon-picker': []
  'toggle-started': []
  'toggle-dependencies': []
}>()
</script>

<template>
  <!-- Edit mode -->
  <div v-if="mode === 'insert' && editingGoalId === goal?.id" class="mb-4">
    <input
      :value="editTitle"
      @input="$emit('update:editTitle', ($event.target as HTMLInputElement).value)"
      type="text"
      class="w-full px-3 py-2 bg-surface-2 border border-accent rounded text-ink focus:outline-none focus:ring-2 focus:ring-accent"
      @keydown.enter.prevent="$emit('save-edit')"
      @keydown.esc.prevent="$emit('cancel-edit')"
    />
    <div class="text-xs text-ink-dim mt-2">
      Enter för att spara, Escape för att avbryta
    </div>
  </div>

  <!-- Normal mode -->
  <div v-else class="flex items-center gap-4 flex-shrink-0">
    <h1
      :class="[
        'text-4xl font-bold transition-colors px-3 py-2 rounded flex-1',
        isGoalSelected ? 'border border-accent text-ink' : 'text-ink'
      ]"
    >
      {{ goal?.title }}
      <ChildCountBadge v-if="goal" :goal-id="goal.id" size="md" class="ml-2 align-middle" />
    </h1>

    <!-- Started-toggle (leader + s) -->
    <button
      v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)"
      @click.stop="$emit('toggle-started')"
      class="p-2 rounded transition-colors"
      :class="
        goal.started
          ? 'text-yellow-400 hover:text-yellow-300'
          : 'text-ink-dim hover:text-ink-muted'
      "
      :title="goal.started ? 'Avbörja (leader + s)' : 'Påbörja (leader + s)'"
    >
      <Icon
        name="lucide:circle-play"
        class="w-6 h-6"
        :style="{ opacity: goal.started ? 1 : 0.25 }"
      />
    </button>

    <!-- Icon button -->
    <button
      v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)"
      @click.stop="$emit('open-icon-picker')"
      class="text-ink-muted hover:text-ink transition-colors p-2 rounded hover:bg-surface-2"
      title="Ändra ikon"
    >
      <Icon :name="goal.icon || 'heroicons:star'" class="w-8 h-8 text-white" />
    </button>

    <!-- Toggle dependencies visibility -->
    <button
      @click.stop="$emit('toggle-dependencies')"
      class="p-2 rounded transition-colors"
      :class="
        showDependencies
          ? 'text-orange-400 hover:text-orange-300'
          : 'text-ink-dim hover:text-ink-muted'
      "
      title="Beroenden (blockerade mål)"
    >
      <Icon name="lucide:git-branch" class="w-6 h-6" />
    </button>
  </div>
</template>
