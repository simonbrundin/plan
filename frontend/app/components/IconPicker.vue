<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'

interface IconResult {
  name: string
  collection: string
  full: string
}

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

const searchQuery = ref('')
const searchInput = ref<HTMLInputElement | null>(null)
const searchResults = ref<IconResult[]>([])
const isSearching = ref(false)
let searchTimeout: ReturnType<typeof setTimeout> | null = null

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
    searchQuery.value = ''
    searchResults.value = []
  }
})

// Watch search query and debounce API calls
watch(searchQuery, (newQuery) => {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }

  // Don't search with less than 2 characters
  if (!newQuery || newQuery.trim().length < 2) {
    searchResults.value = []
    return
  }

  isSearching.value = true
  searchTimeout = setTimeout(async () => {
    await performSearch(newQuery)
  }, 300) // Debounce for 300ms
})

async function performSearch(query: string) {
  try {
    const response = await fetch(`/api/icons?q=${encodeURIComponent(query)}`)
    const data = await response.json()
    searchResults.value = data.results || []
  } catch (error) {
    console.error('Failed to search icons:', error)
    searchResults.value = []
  } finally {
    isSearching.value = false
  }
}

const filteredIcons = computed(() => {
  return searchResults.value
})

function selectIcon(iconResult: IconResult) {
  emit('update:modelValue', iconResult.full)
  displayOpen.value = false
  searchQuery.value = ''
}

function close() {
  displayOpen.value = false
  searchQuery.value = ''
}
</script>

<template>
  <div class="relative inline-block">
    <!-- Öppen modal bakgrund -->
    <div v-if="displayOpen" class="fixed inset-0 bg-black/50 z-40" @click="close"></div>

    <!-- Modal -->
    <div v-if="displayOpen" class="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-50 w-full max-w-2xl mx-auto">
      <div class="bg-gray-800 rounded-lg shadow-lg border border-gray-700 p-4">
        <!-- Header -->
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-xl font-bold text-gray-100">Välj ikon</h2>
          <button @click="close" class="text-gray-400 hover:text-gray-200 p-1 rounded hover:bg-gray-700">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Sökfält -->
        <input ref="searchInput" v-model="searchQuery" type="text" placeholder="Sök ikoner..."
          class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-gray-200 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 mb-4" />

        <!-- Instruktioner när tomt -->
        <div v-if="!searchQuery || searchQuery.trim().length < 2" class="text-center py-8 text-gray-400">
          Skriv minst 2 tecken för att söka bland alla ikoner
        </div>

        <!-- Loading status -->
        <div v-else-if="isSearching" class="text-center py-8 text-gray-400">
          Söker ikoner...
        </div>

        <!-- Ikon grid -->
        <div v-else-if="filteredIcons.length > 0" class="grid grid-cols-6 gap-2 max-h-96 overflow-y-auto">
          <button v-for="icon in filteredIcons" :key="icon.full" @click="selectIcon(icon)"
            class="aspect-square flex flex-col items-center justify-center rounded-lg p-2 hover:bg-gray-700 transition-colors group relative"
            :class="modelValue === icon.full ? 'bg-blue-600' : 'bg-gray-700'">
            <Icon :name="icon.full" class="w-8 h-8" />
            <div class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-gray-900 rounded text-xs text-gray-200 whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none max-w-xs">
              <div class="font-mono">{{ icon.full }}</div>
              <div class="text-gray-500 text-xs">{{ icon.collection }}</div>
            </div>
          </button>
        </div>

        <!-- Meddelande när inga resultat -->
        <div v-else-if="searchQuery && searchQuery.trim().length >= 2" class="text-center py-8 text-gray-400">
          Ingen ikon matchade "{{ searchQuery }}"
        </div>

        <!-- Results counter -->
        <div v-if="filteredIcons.length > 0" class="text-center py-2 text-xs text-gray-500">
          Visar {{ filteredIcons.length }} resultat
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
