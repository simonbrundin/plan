<script setup lang="ts">
const isOpen = defineModel<boolean>('open', { default: false })

const emit = defineEmits<{
  'execute-command': [key: string]
}>()

const handleCommand = (key: string) => {
  console.log('Leader command executed:', key)
  emit('execute-command', key)
  isOpen.value = false
}
</script>

<template>
  <UModal v-model:open="isOpen" title="Leader Commands">
    <!-- Leader Key Display -->
    <div
      style="display: flex; align-items: center; gap: 0.5rem; padding-bottom: 0.75rem; border-bottom: 1px solid rgb(55, 65, 81); margin-bottom: 1rem;">
      <span style="color: rgb(156, 163, 175); font-size: 0.875rem; font-family: monospace;">Leader Key:</span>
      <kbd
        style="padding: 0.25rem 0.75rem; background-color: rgb(17, 24, 39); border: 1px solid rgb(75, 85, 99); border-radius: 0.375rem; color: rgb(96, 165, 250); font-family: monospace; font-weight: bold;">SPACE</kbd>
    </div>

    <!-- Goals section header -->
    <h3
      style="font-size: 0.75rem; font-weight: 600; color: rgb(96, 165, 250); text-transform: uppercase; letter-spacing: 0.05em; margin: 1rem 0 0.5rem 0;">
      Goals</h3>

    <!-- Toggle completed goals command button -->
    <button @click="handleCommand('d')"
      style="width: 100%; text-align: left; font-size: 0.875rem; background-color: rgb(55, 65, 81); padding: 0.75rem; border-radius: 0.375rem; color: rgb(243, 244, 246); border: 1px solid rgb(75, 85, 99); cursor: pointer; font-family: inherit; display: block; margin-bottom: 1rem;">
      <span style="font-family: monospace; font-weight: bold; color: rgb(96, 165, 250);">D</span>
      <span style="color: rgb(243, 244, 246); margin-left: 0.5rem;">Toggle show/hide completed goals</span>
    </button>

    <!-- Modal footer -->
    <template #footer>
      <div
        style="text-align: center; font-size: 0.75rem; color: rgb(107, 114, 128); padding-top: 0.75rem; border-top: 1px solid rgb(55, 65, 81);">
        Press
        <kbd
          style="padding: 0.125rem 0.25rem; background-color: rgb(17, 24, 39); border: 1px solid rgb(75, 85, 99); border-radius: 0.375rem; color: rgb(156, 163, 175); font-family: monospace; font-size: 0.75rem; display: inline-block; margin: 0 0.125rem;">ESC</kbd>
        to close
      </div>
    </template>
  </UModal>
</template>

<style scoped>
:deep(.modal-content) {
  background-color: rgb(17, 24, 39) !important;
  border: 1px solid rgb(55, 65, 81) !important;
}
</style>
