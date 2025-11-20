<script setup lang="ts">
import type { Goal } from '~/types/goal'

defineProps<{
  goal: Goal | undefined
  isGoalSelected: boolean
  editingGoalId: number | null
  editTitle: string
  mode: 'normal' | 'insert'
}>()

defineEmits<{
  'update:editTitle': [value: string]
  'update:mode': [value: 'normal' | 'insert']
  'update:editingGoalId': [value: number | null]
  'save-edit': []
  'cancel-edit': []
  'open-icon-picker': []
}>()
</script>

<template>
  <!-- Edit mode -->
  <div v-if="mode === 'insert' && editingGoalId === goal?.id" class="mb-4">
    <input
      :value="editTitle"
      @input="$emit('update:editTitle', ($event.target as HTMLInputElement).value)"
      type="text"
      class="w-full px-3 py-2 bg-gray-800 border border-blue-500 rounded text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-400"
      @keydown.enter.prevent="$emit('save-edit')"
      @keydown.esc.prevent="$emit('cancel-edit')"
    />
    <div class="text-xs text-gray-500 mt-2">
      Enter för att spara, Escape för att avbryta
    </div>
  </div>

  <!-- Normal mode -->
  <div v-else class="flex items-center gap-4 flex-shrink-0">
    <h1 :class="['text-4xl font-bold transition-colors px-3 py-2 rounded flex-1',
      isGoalSelected ? 'border border-blue-500 text-gray-100' : 'text-gray-100'
    ]">
      {{ goal?.title }}
    </h1>
    <button
      v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)"
      @click.stop="$emit('open-icon-picker')"
      class="text-gray-400 hover:text-gray-200 transition-colors p-2 rounded hover:bg-gray-800"
      title="Ändra ikon"
    >
      <Icon :name="goal.icon || 'heroicons:star'" class="w-8 h-8 text-white" />
    </button>
  </div>
</template>
