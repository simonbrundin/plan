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
    class="flex-shrink-0 w-5 h-5 rounded flex items-center justify-center text-xs font-bold transition-colors"
    :class="props.active ? 'text-blue-400 bg-blue-500/15' : 'text-ink-dim hover:text-ink-muted'"
    :title="props.active ? 'Ta bort agent-markeringen' : 'Markera som agent-mål'"
    :aria-pressed="props.active"
  >
    A
  </button>
</template>
