<script setup lang="ts">
interface Props {
  open: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  "update:open": [value: boolean];
}>();

const route = useRoute();
const isCalendar = computed(() => route.path === "/calendar");

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
      class="bg-surface border border-line rounded-2xl shadow-2xl p-6 sm:p-8 w-[min(94vw,960px)] max-h-[88vh] overflow-y-auto"
    >
      <div class="flex items-center justify-between mb-4">
        <div>
          <h2 class="text-2xl font-bold text-white">⌨️ Kortkommandon</h2>
          <p class="mt-1 text-sm text-ink-muted">Snabbkommandon för den vy du använder</p>
        </div>
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

      <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
        <!-- Normal Mode -->
        <div class="rounded-xl border border-line bg-surface-2/40 p-5">
          <h3
            class="text-sm font-semibold text-accent uppercase tracking-widest mb-4"
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
                >c</kbd
              >
              <span class="text-ink-muted">Sätt deadline (due-datum)</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >b</kbd
              >
              <span class="text-ink-muted">Sätt startdatum (syns i /priority först när datumet passerat)</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-accent font-mono font-bold min-w-[2rem] text-center"
                >m</kbd
              >
              <span class="text-ink-muted">Sätt uppskattad tidsåtgång</span>
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
                >z</kbd
              >
              <span class="text-ink-muted">Ångra senaste handlingen</span>
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
        <div class="rounded-xl border border-line bg-surface-2/40 p-5">
          <h3
            class="text-sm font-semibold text-green-400 uppercase tracking-widest mb-4"
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
                >g</kbd
              >
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >p</kbd
              >
              <span class="text-ink-muted">Gå till prioritetslistan</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >g</kbd
              >
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >c</kbd
              >
              <span class="text-ink-muted">Gå till kalendern</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >g</kbd
              >
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >d</kbd
              >
              <span class="text-ink-muted">Gå till avslutade mål</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >g</kbd
              >
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >s</kbd
              >
              <span class="text-ink-muted">Gå till inställningar</span>
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
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-green-400 font-mono font-bold min-w-[2rem] text-center"
                >d</kbd
              >
              <span class="text-ink-muted">Sätt både startdatum och deadline för markerat mål</span>
            </div>
          </div>
        </div>

        <!-- Calendar -->
        <div v-if="isCalendar" class="rounded-xl border border-cyan-400/20 bg-cyan-400/5 p-5">
          <h3
            class="text-sm font-semibold text-cyan-400 uppercase tracking-widest mb-4"
          >
            Kalender
          </h3>
          <div class="space-y-2">
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-cyan-400 font-mono font-bold min-w-[2rem] text-center"
                >t</kbd
              >
              <span class="text-ink-muted">Hoppa till nu (visar närmaste 24 timmarna)</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-cyan-400 font-mono font-bold min-w-[2rem] text-center"
                >r</kbd
              >
              <span class="text-ink-muted">Återställ kalenderns vy och zoom</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-cyan-400 font-mono font-bold min-w-[2rem] text-center"
                >j</kbd
              >
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-cyan-400 font-mono font-bold min-w-[2rem] text-center"
                >↓</kbd
              >
              <span class="text-ink-muted">Zooma ut (lodrät scrollning: ner zoomar ut)</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-cyan-400 font-mono font-bold min-w-[2rem] text-center"
                >k</kbd
              >
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-cyan-400 font-mono font-bold min-w-[2rem] text-center"
                >↑</kbd
              >
              <span class="text-ink-muted">Zooma in (lodrät scrollning: upp zoomar in)</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-cyan-400 font-mono font-bold min-w-[2rem] text-center"
                >h</kbd
              >
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-cyan-400 font-mono font-bold min-w-[2rem] text-center"
                >l</kbd
              >
              <span class="text-ink-muted">Som horisontell scrollning: h flyttar tiden bakåt, l framåt</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-cyan-400 font-mono font-bold min-w-[2rem] text-center"
                >+</kbd
              >
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-cyan-400 font-mono font-bold min-w-[2rem] text-center"
                >−</kbd
              >
              <span class="text-ink-muted">Zooma in eller ut</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-cyan-400 font-mono font-bold min-w-[2rem] text-center"
                >Drag</kbd
              >
              <span class="text-ink-muted">Dra lodrätt för att zooma och vågrätt för att flytta tiden</span>
            </div>
            <div class="flex items-center gap-4">
              <kbd
                class="px-2 py-1 bg-surface-2 border border-line-strong rounded text-cyan-400 font-mono font-bold min-w-[2rem] text-center"
                >Scroll</kbd
              >
              <span class="text-ink-muted">Scrolla lodrätt för att zooma och vågrätt för att flytta tiden</span>
            </div>
          </div>
        </div>

        <!-- General -->
        <div class="rounded-xl border border-line bg-surface-2/40 p-5">
          <h3
            class="text-sm font-semibold text-fuchsia-400 uppercase tracking-widest mb-4"
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
                >n</kbd
              >
              <span class="text-ink-muted"
                >Snabbt lägg till (mål eller händelse med naturlig text)</span
              >
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
