<script setup lang="ts">
defineProps<{
  open: boolean
  parentTitle: string | undefined
}>()

defineEmits<{
  'update:open': [value: boolean]
  'confirm': []
}>()
</script>

<template>
  <UModal :model-value="open" title="Ta bort föräldrarelation?" @update:model-value="(val) => $emit('update:open', val)">
    <p class="text-gray-400">
      Vill du ta bort relationen till
      <strong>{{ parentTitle }}</strong
      >? Själva målet kommer inte tas bort.
    </p>

    <template #footer="{ close }">
      <div class="flex justify-end gap-2">
        <UButton color="gray" variant="ghost" @click="close">
          Avbryt
        </UButton>
        <UButton color="red" @click="$emit('confirm')">
          Ta bort relation
        </UButton>
      </div>
    </template>
  </UModal>
</template>
