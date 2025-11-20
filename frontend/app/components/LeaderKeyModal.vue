<script setup lang="ts">
interface Props {
  open: boolean
}

defineProps<Props>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  'execute-command': [key: string]
}>()

const handleCommand = (key: string) => {
  emit('execute-command', key)
  emit('update:open', false)
}

const closeModal = () => {
  emit('update:open', false)
}
</script>

<template>
  <!-- Leader Key Modal - Custom Implementation -->
  <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50"
    @click.self="closeModal" @keydown.esc="closeModal">
    <div class="bg-gray-900 border border-gray-700 rounded-lg shadow-lg p-6 w-96">
      <h2 class="text-lg font-bold text-white mb-4">Leader Commands</h2>

      <div class="space-y-4">
        <!-- <div class="flex items-center gap-2 pb-3 border-b border-gray-700"> -->
        <!--   <span class="text-gray-400 text-sm font-mono">Leader Key:</span> -->
        <!--   <kbd -->
        <!--     class="px-3 py-1 bg-gray-800 border border-gray-600 rounded text-blue-400 font-mono font-bold">SPACE</kbd> -->
        <!-- </div> -->

        <div class="space-y-2">
          <h3 class="text-xs font-semibold text-blue-400 uppercase tracking-widest">Goals</h3>
          <button @click="handleCommand('f')"
            class="w-full text-left text-sm bg-gray-700 hover:bg-gray-600 px-3 py-2 rounded border border-gray-600 text-gray-100 cursor-pointer">
            <span class="font-mono font-bold text-blue-400">F</span>
            <span class="font-mono font-bold text-blue-400 ml-1">+</span>
            <span class="font-mono font-bold text-blue-400">D</span>
            <span class="text-gray-100 ml-2">- Toggle show/hide completed goals</span>
          </button>
          <button @click="handleCommand('i')"
            class="w-full text-left text-sm bg-gray-700 hover:bg-gray-600 px-3 py-2 rounded border border-gray-600 text-gray-100 cursor-pointer">
            <span class="font-mono font-bold text-blue-400">I</span>
            <span class="text-gray-100 ml-2">- Change icon for selected goal</span>
          </button>
        </div>
      </div>

      <div class="text-center text-xs text-gray-500 pt-4 border-t border-gray-700 mt-4">
        Press <kbd
          class="px-1 py-0.5 bg-gray-800 border border-gray-700 rounded text-gray-400 font-mono text-xs inline-block mx-0.5">ESC</kbd>
        to close
      </div>
    </div>
  </div>
</template>
