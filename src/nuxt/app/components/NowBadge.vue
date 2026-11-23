<script setup lang="ts">
// Nu-märke på ett mål.
//
// Visar att målet ska göras nu och därmed dyker upp på "Gör nu"-sidan.
// Grönt = påslaget, grått = inte påslaget. Klicka på märket för att växla.
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
    :class="props.active ? 'text-emerald-400 hover:text-emerald-300' : 'text-ink-dim hover:text-ink-muted'"
    :title="props.active ? 'Ta bort nu-markeringen' : 'Markera som ska göras nu'"
    :aria-pressed="props.active"
  >
    <Icon name="lucide:zap" class="w-4 h-4" />
  </button>
</template>
