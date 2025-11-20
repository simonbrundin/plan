<script setup lang="ts">
import type { Goal } from '~/types/goal'

defineProps<{
  showChildSearch: boolean
  childSearchQuery: string
  childSearchResults: Goal[]
  childSearchInput: HTMLInputElement | null
}>()

defineEmits<{
  'update:childSearchQuery': [value: string]
  'add-existing-child': [childId: number]
  'handle-search-keydown': [event: KeyboardEvent]
}>()

const childSearchInputRef = ref<HTMLInputElement | null>(null)

watch(
  () => childSearchInputRef.value,
  (newVal) => {
    if (newVal) {
      newVal.focus()
    }
  }
)
</script>

<template>
  <!-- Sökfält för undermål -->
  <div v-if="showChildSearch" class="border border-gray-700 rounded-lg p-4 bg-gray-800 mb-4">
    <div class="mb-2">
      <input
        ref="childSearchInputRef"
        :value="childSearchQuery"
        @input="$emit('update:childSearchQuery', ($event.target as HTMLInputElement).value)"
        @keydown="$emit('handle-search-keydown', $event)"
        type="text"
        placeholder="Sök efter mål eller skapa nytt (tryck Enter)"
        class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-gray-200 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
        autofocus
      />
    </div>

    <!-- Sökresultat -->
    <div v-if="childSearchResults.length > 0" class="space-y-1 mt-2">
      <button
        v-for="result in childSearchResults"
        :key="result.id"
        @click="$emit('add-existing-child', result.id)"
        class="w-full text-left px-4 py-2 hover:bg-gray-700 rounded-lg transition-colors"
      >
        <div class="font-medium text-gray-200">{{ result.title }}</div>
        <div class="text-xs text-gray-500">
          {{ new Date(result.created).toLocaleDateString("sv-SE") }}
        </div>
      </button>
    </div>

    <!-- Meddelande när inga resultat -->
    <div v-else-if="childSearchQuery.trim()" class="text-sm text-gray-500 px-4 py-2">
      Tryck Enter för att skapa "{{ childSearchQuery }}" som nytt undermål
    </div>
  </div>
</template>
