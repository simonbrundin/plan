<script setup lang="ts">
defineProps<{
  open: boolean
  goalTitle: string
}>()

defineEmits<{
  'update:open': [value: boolean]
  'execute-delete': []
}>()

const selection = ref<'cancel' | 'confirm'>('confirm')

watch(
  () => open,
  (newVal) => {
    if (newVal) {
      selection.value = 'confirm'
    }
  }
)

function handleKeydown(event: KeyboardEvent) {
  if (!open) return

  if (event.key === 'h') {
    event.preventDefault()
    event.stopPropagation()
    selection.value = 'cancel'
  } else if (event.key === 'l') {
    event.preventDefault()
    event.stopPropagation()
    selection.value = 'confirm'
  } else if (event.key === 'Enter') {
    event.preventDefault()
    event.stopPropagation()
    if (selection.value === 'confirm') {
      $emit('execute-delete')
    } else {
      $emit('update:open', false)
    }
  } else if (event.key === 'Escape') {
    event.preventDefault()
    event.stopPropagation()
    $emit('update:open', false)
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <UModal :model-value="open" title="Ta bort mål?" @update:model-value="(val) => $emit('update:open', val)">
    <template #footer="{ close }">
      <div class="flex justify-end gap-2">
        <UButton
          color="gray"
          variant="ghost"
          @click="close"
          :class="selection === 'cancel' ? 'ring-2 ring-blue-500' : ''"
        >
          Avbryt
        </UButton>
        <UButton
          color="red"
          @click="$emit('execute-delete')"
          :class="selection === 'confirm' ? 'ring-2 ring-blue-400' : ''"
        >
          Ta bort
        </UButton>
      </div>
    </template>
  </UModal>
</template>
