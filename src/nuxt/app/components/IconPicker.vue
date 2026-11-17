<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'

interface Props {
  modelValue: string
  open?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  open: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'update:open': [value: boolean]
}>()

const searchInput = ref<HTMLInputElement | null>(null)
const { searchQuery, searchResults, isSearching, searchError, reset } = useIconSearch()

const displayOpen = computed({
  get: () => props.open,
  set: (value) => emit('update:open', value)
})

watch(displayOpen, async (newValue) => {
  if (newValue) {
    await nextTick()
    searchInput.value?.focus()
  } else {
    // Clear search when closing
    reset()
  }
})

function selectIcon(iconResult: { full: string }) {
  emit('update:modelValue', iconResult.full)
  displayOpen.value = false
  reset()
}

function close() {
  displayOpen.value = false
  reset()
}
</script>

<template>
  <div class="relative inline-block">
    <!-- Öppen modal bakgrund -->
    <div v-if="displayOpen" class="fixed inset-0 bg-black/50 z-40" @click="close"></div>

    <!-- Modal -->
    <div v-if="displayOpen" class="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-50 w-full max-w-2xl mx-auto">
      <div class="bg-surface-2 rounded-lg shadow-lg border border-line p-4">
        <!-- Header -->
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-xl font-bold text-ink">Välj ikon</h2>
          <button @click="close" class="text-ink-muted hover:text-ink p-1 rounded hover:bg-elevated">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Sökfält -->
        <input ref="searchInput" v-model="searchQuery" type="text" placeholder="Sök ikoner..."
          class="w-full px-4 py-2 bg-surface border border-line rounded-lg text-ink placeholder-ink-dim focus:outline-none focus:ring-2 focus:ring-accent mb-4" />

        <!-- Instruktioner när tomt -->
        <div v-if="!searchQuery || searchQuery.trim().length < 2" class="text-center py-8 text-ink-muted">
          Skriv minst 2 tecken för att söka bland alla ikoner
        </div>

        <!-- Loading status -->
        <div v-else-if="isSearching" class="text-center py-8 text-ink-muted">
          Söker ikoner...
        </div>

        <!-- Ikon grid -->
        <div v-else-if="searchResults.length > 0" class="grid grid-cols-6 gap-2 max-h-96 overflow-y-auto">
          <button v-for="icon in searchResults" :key="icon.full" @click="selectIcon(icon)"
            class="aspect-square flex flex-col items-center justify-center rounded-lg p-2 hover:bg-elevated transition-colors group relative"
            :class="modelValue === icon.full ? 'bg-accent-strong' : 'bg-elevated'">
            <Icon :name="icon.full" class="w-8 h-8 text-white" />
            <div class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-surface rounded text-xs text-ink whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none max-w-xs">
              <div class="font-mono">{{ icon.full }}</div>
              <div class="text-ink-dim text-xs">{{ icon.collection }}</div>
            </div>
          </button>
        </div>

        <!-- Meddelande när sökningen misslyckades -->
        <div v-else-if="searchError" class="text-center py-8 text-red-400">
          {{ searchError }}
        </div>

        <!-- Meddelande när inga resultat -->
        <div v-else-if="searchQuery && searchQuery.trim().length >= 2" class="text-center py-8 text-ink-muted">
          Ingen ikon matchade "{{ searchQuery }}"
        </div>

        <!-- Results counter -->
        <div v-if="searchResults.length > 0" class="text-center py-2 text-xs text-ink-dim">
          Visar {{ searchResults.length }} resultat
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
::-webkit-scrollbar {
  width: 8px;
}

::-webkit-scrollbar-track {
  background: #1f2937;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb {
  background: #4b5563;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #6b7280;
}
</style>
