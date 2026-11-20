<script setup lang="ts">
// Agentmärke på ett mål.
//
// Visar att målet ligger för en agent att arbeta med. Blått = påslaget,
// grått = inte påslaget. Klicka på märket för att växla.
const props = defineProps<{
  active: boolean;
}>();

const emit = defineEmits<{
  toggle: [];
}>();

const toggle = () => {
  emit('toggle');
};
</script>

<template>
  <button
    type="button"
    @click.stop="toggle"
    class="flex-shrink-0 w-6 h-6 rounded flex items-center justify-center transition-colors"
    :class="props.active ? 'text-blue-500 hover:text-blue-400' : 'text-ink-dim hover:text-ink-muted'"
    :title="props.active ? 'Ta bort agent-markeringen' : 'Markera som agent-mål'"
    :aria-pressed="props.active"
  >
    <Icon name="lucide:bot" class="w-4 h-4" />
  </button>
</template>
