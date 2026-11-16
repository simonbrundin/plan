<script setup lang="ts">
interface Props {
  open: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  "update:open": [value: boolean];
}>();

const closeModal = () => {
  emit("update:open", false);
};

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === "Escape") {
    event.preventDefault();
    closeModal();
  }
};

onMounted(() => {
  window.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleKeydown);
});
</script>

<template>
  <div
    v-if="open"
    class="fixed inset-0 z-[100] flex items-center justify-center bg-black bg-opacity-70"
    @click.self="closeModal"
  >
    <div
      class="bg-surface border border-line rounded-lg shadow-xl p-6 w-[600px] max-h-[80vh] overflow-y-auto"
    >
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-bold text-white">⌨️ Keyboard Shortcuts</h2>
        <button
          @click="closeModal"
          class="text-ink-muted hover:text-white transition-colors"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="w-6 h-6"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>

      <div class="space-y-6">
        <!-- Normal Mode -->
        <div>
          <h3
            class="text-sm font-semibold text-accent uppercase tracking-widest mb-3"
          >
            Normal Mode
          </h3>
          <div class="space-y-2">
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >p</kbd
              >
              <span class="text-ink-muted">Växla till/från prioritetssidan</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >?</kbd
              >
              <span class="text-ink-muted">Visa denna hjälpmodal</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >j</kbd
              >
              <span class="text-ink-muted"
                >Markera nästa mål / gå till parent mode</span
              >
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >k</kbd
              >
              <span class="text-ink-muted"
                >Markera föregående mål / gå till goal mode</span
              >
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >l</kbd
              >
              <span class="text-ink-muted">Öppna markerat mål</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >h</kbd
              >
              <span class="text-ink-muted">Gå till första föräldern</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >i</kbd
              >
              <span class="text-ink-muted">Redigera titel (vid markering)</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >a</kbd
              >
              <span class="text-ink-muted">Redigera titel vid slutet</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >d</kbd
              >
              <span class="text-ink-muted">Toggle slutfört</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >s</kbd
              >
              <span class="text-ink-muted">Toggle påbörjat</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >t</kbd
              >
              <span class="text-ink-muted">Redigera beskrivning (vid markering)</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >x</kbd
              >
              <span class="text-ink-muted">Ta bort mål</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >Enter</kbd
              >
              <span class="text-ink-muted">Sök/lägg till undermål</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >↑</kbd
              >
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >↓</kbd
              >
              <span class="text-ink-muted">Flytta markerat mål upp/ner</span>
            </div>
          </div>
        </div>

        <!-- Leader Mode -->
        <div>
          <h3
            class="text-sm font-semibold text-green-400 uppercase tracking-widest mb-3"
          >
            Leader Mode
            <span class="text-ink-dim font-normal"
              >(Tryck
              <kbd
                class="px-1 py-0.5 bg-surface-2 border border-line-strong rounded text-green-400 font-mono text-xs"
                >SPACE</kbd
              >
              först)</span
            >
          </h3>
          <div class="space-y-2">
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >g</kbd
              >
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >i</kbd
              >
              <span class="text-ink-muted">Gå till inbox</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >g</kbd
              >
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >g</kbd
              >
              <span class="text-ink-muted">Gå till mål-lista</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >f</kbd
              >
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >d</kbd
              >
              <span class="text-ink-muted">Toggle visa/dölj avklarade mål</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >i</kbd
              >
              <span class="text-ink-muted">Byt ikon för markerat mål</span>
            </div>
          </div>
        </div>

        <!-- General -->
        <div>
          <h3
            class="text-sm font-semibold text-fuchsia-400 uppercase tracking-widest mb-3"
          >
            General
          </h3>
          <div class="space-y-2">
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-fuchsia-400 font-mono font-bold min-w-[2rem] text-center"
                >Ctrl</kbd
              >
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-fuchsia-400 font-mono font-bold min-w-[2rem] text-center"
                >K</kbd
              >
              <span class="text-ink-muted">Sök bland mål</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-fuchsia-400 font-mono font-bold min-w-[2rem] text-center"
                >Esc</kbd
              >
              <span class="text-ink-muted">Stäng modal / avbryt</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-fuchsia-400 font-mono font-bold min-w-[2rem] text-center"
                >/</kbd
              >
              <span class="text-ink-muted">Öppna sökning</span>
            </div>
          </div>
        </div>
      </div>

      <div
        class="text-center text-xs text-ink-dim pt-4 border-t border-line mt-6"
      >
        Tryck
        <kbd
          class="px-1 py-0.5 bg-surface-2 border border-line rounded text-ink-muted font-mono text-xs"
          >ESC</kbd
        >
        eller klicka utanför för att stänga
      </div>
    </div>
  </div>
</template>
