<script setup lang="ts">
import type { Goal } from '~/types/goal'

defineProps<{
  showChildSearch: boolean
  childSearchQuery: string
  childSearchResults: Goal[]
}>()

const emit = defineEmits<{
  'update:childSearchQuery': [value: string]
  'add-existing-child': [childId: number]
  'handle-search-keydown': [event: KeyboardEvent]
  'close': []
}>()

const childSearchInput = ref<HTMLInputElement | null>(null)

// Focus the input when search is shown
watch(
  () => childSearchInput.value,
  (el) => {
    if (el) {
      el.focus()
    }
  }
)
</script>

<template>
  <!-- Sökfält för undermål -->
  <div v-if="showChildSearch" class="border border-line rounded-lg p-4 bg-surface-2 mb-4">
    <div class="mb-2">
      <input
        ref="childSearchInput"
        :value="childSearchQuery"
        @input="$emit('update:childSearchQuery', ($event.target as HTMLInputElement).value)"
        @keydown.enter.prevent="$emit('handle-search-keydown', $event)"
        @keydown.escape.prevent="$emit('close')"
        type="text"
        placeholder="Sök efter mål eller skapa nytt (tryck Enter)"
        class="w-full px-4 py-2 bg-surface border border-line rounded-lg text-ink placeholder-ink-dim focus:outline-none focus:ring-2 focus:ring-accent"
      />
    </div>

    <!-- Sökresultat -->
    <div v-if="childSearchResults.length > 0" class="space-y-1 mt-2">
      <button
        v-for="result in childSearchResults"
        :key="result.id"
        @click="$emit('add-existing-child', result.id)"
        class="w-full text-left px-4 py-2 hover:bg-elevated rounded-lg transition-colors"
      >
        <div class="font-medium text-ink">{{ result.title }}</div>
        <div class="text-xs text-ink-dim">
          {{ new Date(result.created).toLocaleDateString("sv-SE") }}
        </div>
      </button>
    </div>

    <!-- Meddelande när inga resultat -->
    <div v-else-if="childSearchQuery.trim()" class="text-sm text-ink-dim px-4 py-2">
      Tryck Enter för att skapa "{{ childSearchQuery }}" som nytt undermål
    </div>
  </div>
</template>
