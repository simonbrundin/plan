<script setup lang="ts">
import type { Goal } from '~/types/goal'

defineProps<{
  parents: Goal[]
  selectedParentIndex: number
  isParentMode: boolean
  mode: 'normal' | 'insert'
  editingGoalId: number | null
  editTitle: string
  showParentSearch: boolean
  parentSearchQuery: string
  searchResults: Goal[]
  parentSearchInput: HTMLInputElement | null
}>()

defineEmits<{
  'toggle-parent-search': []
  'update:parentSearchQuery': [value: string]
  'add-existing-parent': [parentId: number]
  'handle-parent-mousedown': [parentId: number]
  'handle-parent-mouseup': []
  'handle-parent-mouseleave': []
  'handle-search-keydown': [event: KeyboardEvent]
  'save-edit': []
  'cancel-edit': []
  'update:editTitle': [value: string]
  'navigate-to-parent': [parentId: number]
}>()

const router = useRouter()
</script>

<template>
  <!-- Breadcrumb / Föräldrar -->
  <div class="flex items-center justify-between flex-shrink-0">
    <div class="flex items-center gap-2 text-sm flex-wrap">
      <NuxtLink
        v-if="parents.length === 0"
        to="/goal/1"
        class="text-blue-400 hover:text-blue-300 transition-colors font-medium"
      >
        Root
      </NuxtLink>

      <!-- Insert mode för förälder -->
      <div v-for="(parent, index) in parents" :key="parent.id">
        <div v-if="mode === 'insert' && editingGoalId === parent.id" class="inline-block">
          <input
            :value="editTitle"
            @input="$emit('update:editTitle', ($event.target as HTMLInputElement).value)"
            type="text"
            class="px-3 py-1 bg-gray-800 border border-blue-500 rounded text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-400"
            @keydown.enter.prevent="$emit('save-edit')"
            @keydown.esc.prevent="$emit('cancel-edit')"
          />
        </div>

        <!-- Normal mode för förälder -->
        <NuxtLink
          v-else
          :to="`/goal/${parent.id}`"
          :data-parent-index="index"
          class="px-2 py-1 rounded transition-all select-none inline-block"
          :class="isParentMode && selectedParentIndex === index
            ? 'text-gray-100 bg-blue-500 font-medium'
            : 'text-gray-500 hover:text-gray-300'
          "
          @mousedown="$emit('handle-parent-mousedown', parent.id)"
          @mouseup="$emit('handle-parent-mouseup')"
          @mouseleave="$emit('handle-parent-mouseleave')"
          @touchstart="$emit('handle-parent-mousedown', parent.id)"
          @touchend="$emit('handle-parent-mouseup')"
          @touchcancel="$emit('handle-parent-mouseup')"
        >
          {{ parent.title }}
        </NuxtLink>
      </div>
    </div>

    <!-- + knapp för att lägga till förälder -->
    <button
      @click="$emit('toggle-parent-search')"
      class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800"
      title="Lägg till förälder"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
      </svg>
    </button>
  </div>

  <!-- Sökfält för föräldrar -->
  <div v-if="showParentSearch" class="border border-gray-700 rounded-lg p-4 bg-gray-800 flex-shrink-0">
    <div class="mb-2">
      <input
        :ref="el => { if (el) parentSearchInput = el as HTMLInputElement }"
        :value="parentSearchQuery"
        @input="$emit('update:parentSearchQuery', ($event.target as HTMLInputElement).value)"
        @keydown="$emit('handle-search-keydown', $event)"
        type="text"
        placeholder="Sök efter mål eller skapa nytt (tryck Enter)"
        class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-gray-200 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
        autofocus
      />
    </div>

    <!-- Sökresultat -->
    <div v-if="searchResults.length > 0" class="space-y-1 mt-2">
      <button
        v-for="result in searchResults"
        :key="result.id"
        @click="$emit('add-existing-parent', result.id)"
        class="w-full text-left px-4 py-2 hover:bg-gray-700 rounded-lg transition-colors"
      >
        <div class="font-medium text-gray-200">{{ result.title }}</div>
        <div class="text-xs text-gray-500">
          {{ new Date(result.created).toLocaleDateString("sv-SE") }}
        </div>
      </button>
    </div>

    <!-- Meddelande när inga resultat -->
    <div v-else-if="parentSearchQuery.trim()" class="text-sm text-gray-500 px-4 py-2">
      Tryck Enter för att skapa "{{ parentSearchQuery }}" som nytt mål
    </div>
  </div>
</template>
