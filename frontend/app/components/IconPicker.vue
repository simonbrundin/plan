<script setup lang="ts">
import { ref, computed } from 'vue'

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
const displayOpen = computed({
  get: () => props.open,
  set: (value) => emit('update:open', value)
})

// Lista över vanligt använda Material Symbols-ikoner
const commonIcons = [
  'material-symbols:circle',
  'material-symbols:star',
  'material-symbols:check-circle',
  'material-symbols:close-circle',
  'material-symbols:bookmark',
  'material-symbols:flag',
  'material-symbols:heart',
  'material-symbols:lightbulb',
  'material-symbols:rocket',
  'material-symbols:target',
  'material-symbols:trophy',
  'material-symbols:trending-up',
  'material-symbols:calendar-today',
  'material-symbols:schedule',
  'material-symbols:time-to-leave',
  'material-symbols:alarm',
  'material-symbols:shopping-cart',
  'material-symbols:wallet',
  'material-symbols:credit-card',
  'material-symbols:home',
  'material-symbols:apartment',
  'material-symbols:work',
  'material-symbols:school',
  'material-symbols:library-books',
  'material-symbols:sports-basketball',
  'material-symbols:sports-soccer',
  'material-symbols:sports-tennis',
  'material-symbols:fitness-center',
  'material-symbols:favorite',
  'material-symbols:pets',
  'material-symbols:local-dining',
  'material-symbols:local-cafe',
  'material-symbols:local-bar',
  'material-symbols:restaurant',
  'material-symbols:music-note',
  'material-symbols:movie',
  'material-symbols:camera',
  'material-symbols:video-camera',
  'material-symbols:photo',
  'material-symbols:brush',
  'material-symbols:palette',
  'material-symbols:edit',
  'material-symbols:note',
  'material-symbols:newspaper',
  'material-symbols:article',
  'material-symbols:dashboard',
  'material-symbols:build',
  'material-symbols:settings',
  'material-symbols:computer',
  'material-symbols:smartphone',
  'material-symbols:laptop',
  'material-symbols:headphones',
  'material-symbols:airplanes',
  'material-symbols:train',
  'material-symbols:directions-car',
  'material-symbols:two-wheeler',
  'material-symbols:directions-bike',
  'material-symbols:directions-walk',
  'material-symbols:directions-run',
  'material-symbols:flight',
  'material-symbols:hotel',
  'material-symbols:beach-access',
  'material-symbols:terrain',
  'material-symbols:language',
  'material-symbols:public',
  'material-symbols:location-on',
  'material-symbols:map',
  'material-symbols:navigation',
  'material-symbols:explore',
  'material-symbols:info',
  'material-symbols:help',
  'material-symbols:warning',
  'material-symbols:error',
  'material-symbols:done',
  'material-symbols:cloud',
  'material-symbols:cloud-upload',
  'material-symbols:cloud-download',
  'material-symbols:backup',
  'material-symbols:storage',
  'material-symbols:delete',
  'material-symbols:restore',
  'material-symbols:copy',
  'material-symbols:cut',
  'material-symbols:menu',
  'material-symbols:search',
  'material-symbols:more-vert',
  'material-symbols:add',
  'material-symbols:remove',
  'material-symbols:arrow-forward',
  'material-symbols:arrow-back',
  'material-symbols:arrow-upward',
  'material-symbols:arrow-downward',
]

const filteredIcons = computed(() => {
  if (!searchQuery.value.trim()) return commonIcons

  const query = searchQuery.value.toLowerCase()
  return commonIcons.filter(icon =>
    icon.toLowerCase().includes(query)
  )
})

function selectIcon(icon: string) {
  emit('update:modelValue', icon)
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
        <input v-model="searchQuery" type="text" placeholder="Sök ikoner..."
          class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-gray-200 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 mb-4" autofocus />

        <!-- Ikon grid -->
        <div class="grid grid-cols-6 gap-2 max-h-96 overflow-y-auto">
          <button v-for="icon in filteredIcons" :key="icon" @click="selectIcon(icon)"
            class="aspect-square flex items-center justify-center rounded-lg p-2 hover:bg-gray-700 transition-colors group relative"
            :class="modelValue === icon ? 'bg-blue-600' : 'bg-gray-700'">
            <Icon :name="icon" class="w-8 h-8" />
            <div class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-2 py-1 bg-gray-900 rounded text-xs text-gray-200 whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none">
              {{ icon }}
            </div>
          </button>
        </div>

        <!-- Meddelande när inga resultat -->
        <div v-if="filteredIcons.length === 0" class="text-center py-8 text-gray-400">
          Ingen ikon matchade dina sökkriterier
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
