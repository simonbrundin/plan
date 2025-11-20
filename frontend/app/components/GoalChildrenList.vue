<script setup lang="ts">
import type { Goal } from '~/types/goal'

interface SwipeState {
  startX: number
  currentX: number
  isSwiping: boolean
  childId: number | null
}

defineProps<{
  filteredChildren: Goal[]
  selectedChildIndex: number
  mode: 'normal' | 'insert'
  editingGoalId: number | null
  editTitle: string
  draggedChildIndex: number | null
  dragOverChildIndex: number | null
  swipeState: SwipeState
  showCompleted: boolean
}>()

defineEmits<{
  'toggle-finished': [child: Goal]
  'edit-goal': [childId: number, title: string, atBeginning: boolean]
  'save-edit': []
  'cancel-edit': []
  'update:editTitle': [value: string]
  'navigate-to-child': [childId: number]
  'dragstart': [event: DragEvent, index: number]
  'dragover': [event: DragEvent, index: number]
  'dragleave': []
  'drop': [event: DragEvent, index: number]
  'touchstart': [event: TouchEvent, childId: number]
  'touchmove': [event: TouchEvent]
  'touchend': [child: Goal]
  'toggle-completed': []
  'open-child-search': []
  'open-icon-picker': [childId: number]
}>()

function getSwipeOffset(childId: number): number {
  if (swipeState.isSwiping && swipeState.childId === childId) {
    const delta = swipeState.currentX - swipeState.startX
    return Math.max(0, Math.min(delta, 100))
  }
  return 0
}
</script>

<template>
  <div class="flex-1 overflow-y-auto min-h-0">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-2xl font-semibold text-gray-300"></h2>
      <div class="flex items-center gap-2">
        <button
          @click="$emit('toggle-completed')"
          class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800"
          :title="showCompleted ? 'Dölj avklarade mål' : 'Visa avklarade mål'"
        >
          <svg
            v-if="showCompleted"
            xmlns="http://www.w3.org/2000/svg"
            class="h-5 w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
            />
          </svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"
            />
          </svg>
        </button>
        <button
          @click="$emit('open-child-search')"
          class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800"
          title="Lägg till undermål"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
        </button>
      </div>
    </div>

    <ul v-if="filteredChildren.length > 0" class="space-y-3">
      <li
        v-for="(child, index) in filteredChildren"
        :key="child.id"
        :data-child-index="index"
        draggable="true"
        @dragstart="$emit('dragstart', $event, index)"
        @dragover="$emit('dragover', $event, index)"
        @dragleave="$emit('dragleave')"
        @drop="$emit('drop', $event, index)"
        class="relative overflow-hidden rounded-lg transition-opacity"
        :class="dragOverChildIndex === index ? 'opacity-50' : 'opacity-100'"
      >
        <!-- Swipe bakgrund -->
        <div
          class="absolute inset-0 flex items-center justify-start px-6"
          :class="child.finished ? 'bg-red-900/50' : 'bg-green-900/50'"
        >
          <span class="text-2xl">{{ child.finished ? "↩️" : "✓" }}</span>
        </div>

        <!-- Huvudinnehåll -->
        <div
          class="relative rounded-lg transition-all bg-gray-900"
          :class="selectedChildIndex === index ? 'border border-blue-500' : ''"
          :style="{
            transform: `translateX(${getSwipeOffset(child.id)}px)`,
            transition: swipeState.isSwiping ? 'none' : 'transform 0.3s ease',
          }"
          @touchstart="$emit('touchstart', $event, child.id)"
          @touchmove="$emit('touchmove', $event)"
          @touchend="$emit('touchend', child)"
        >
          <!-- Insert mode - visa input -->
          <div v-if="mode === 'insert' && editingGoalId === child.id" class="p-4">
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

          <!-- Normal mode - visa länk -->
          <div v-else class="flex items-center gap-2">
            <button
              v-if="!(mode === 'insert' && editingGoalId === child.id)"
              @click.stop="$emit('open-icon-picker', child.id)"
              class="flex-shrink-0 text-gray-400 hover:text-gray-200 transition-colors rounded p-1 hover:bg-gray-600"
              title="Ändra ikon"
            >
              <Icon :name="child.icon || 'roentgen:default'" class="w-6 h-6 text-white" />
            </button>
            <NuxtLink :to="`/goal/${child.id}`" class="flex-1 p-4 block">
              <h3 class="text-lg font-medium" :class="child.finished ? 'text-gray-500' : 'text-gray-200'">
                {{ child.title }}
              </h3>
            </NuxtLink>
          </div>
        </div>
      </li>
    </ul>

    <div v-else class="text-gray-500 p-6 border border-gray-700 rounded-lg text-center">
      Inga undermål ännu. Skapa ett för att dela upp detta mål i mindre delar.
    </div>
  </div>
</template>
