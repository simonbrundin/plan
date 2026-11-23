<script setup lang="ts">
interface Props {
  open: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  "update:open": [value: boolean];
}>();

interface Shortcut {
  /** Tangenterna, t.ex. ["Ctrl", "K"] eller ["Shift+P", "J"] */
  keys: string[];
  description: string;
}

interface Section {
  id: string;
  title: string;
  /** Textfärgen används för rubrik och tangentboxar */
  accent: string;
  /** Extra klasser på kortet */
  container?: string;
  shortcuts: Shortcut[];
}

const route = useRoute();
const isCalendar = computed(() => route.path === "/calendar");

// Söktexten i filtret. Fältet får fokus när modalen öppnas, så man kan
// börja skriva direkt.
const query = ref("");
const searchInput = ref<HTMLInputElement | null>(null);

const closeModal = () => {
  emit("update:open", false);
};

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === "Escape") {
    event.preventDefault();
    closeModal();
  }
};

// Mål-kommandona som är gemensamma för målsidan, mållistan och
// prioritetslistan. Var tangent har sin egen rad, så att det är tydligt
// vad varje tangent gör.
const goalActionShortcuts: Shortcut[] = [
  { keys: ["d"], description: "Toggle slutfört" },
  { keys: ["s"], description: "Toggle påbörjat" },
  { keys: ["e"], description: "Toggle evigt-märke" },
  { keys: ["t"], description: "Redigera beskrivning" },
  { keys: ["c"], description: "Sätt deadline (due-datum)" },
  { keys: ["b"], description: "Sätt startdatum" },
  { keys: ["m"], description: "Sätt uppskattad tidsåtgång" },
  { keys: ["r"], description: "Sätt upprepning" },
  { keys: ["x"], description: "Ta bort mål" },
];

const sections = computed<Section[]>(() => {
  const list: Section[] = [
    {
      id: "global",
      title: "Globalt",
      accent: "text-accent",
      shortcuts: [
        { keys: ["?"], description: "Visa denna hjälp" },
        { keys: ["Ctrl", "K"], description: "Sök bland mål (Cmd+K på Mac)" },
        { keys: ["n"], description: "Snabbt lägg till (mål eller händelse)" },
        { keys: ["z"], description: "Ångra senaste handlingen" },
        {
          keys: ["p"],
          description:
            "Växla till/från Gör nu (på målsidan används p för plats)",
        },
        { keys: ["Space"], description: "Leader-läge (se nedan)" },
        { keys: ["Esc"], description: "Stäng modal / avbryt" },
      ],
    },
    {
      id: "goal",
      title: "Målsidan",
      accent: "text-accent",
      shortcuts: [
        {
          keys: ["j"],
          description: "Nästa undermål (från måltiteln: till första undermålet)",
        },
        {
          keys: ["k"],
          description:
            "Föregående undermål (från måltiteln: till förälder-läge)",
        },
        { keys: ["h"], description: "Gå till första föräldern" },
        {
          keys: ["l"],
          description: "Öppna markerat mål / gå till första undermålet",
        },
        {
          keys: ["Enter"],
          description:
            "Lägg till undermål (i förälder-läge: förälder) - skapar ett nytt mål",
        },
        {
          keys: ["↑", "↓"],
          description:
            "Markera ett befintligt mål i sökfältet (tryck sedan Enter för att använda det)",
        },
        { keys: ["↑"], description: "Flytta markerat undermål upp" },
        { keys: ["↓"], description: "Flytta markerat undermål ner" },
        { keys: ["i"], description: "Redigera titel (i början)" },
        { keys: ["a"], description: "Redigera titel (i slutet)" },
        { keys: ["p"], description: "Sätt plats (var målet kan slutföras)" },
        {
          keys: ["w"],
          description: "Gör markerat mål beroende av ett annat mål (sök bland alla mål)",
        },
        ...goalActionShortcuts,
      ],
    },
    {
      id: "goals",
      title: "Mållistan",
      accent: "text-accent",
      shortcuts: [
        { keys: ["j"], description: "Nästa mål" },
        { keys: ["k"], description: "Föregående mål" },
        { keys: ["g"], description: "Hoppa till första målet" },
        { keys: ["G"], description: "Hoppa till sista målet" },
        { keys: ["l"], description: "Öppna markerat mål" },
        { keys: ["h"], description: "Tillbaka till inkorgen" },
        { keys: ["Enter"], description: "Lägg till nytt grundmål" },
        { keys: ["/"], description: "Sök i mållistan" },
        { keys: ["↑"], description: "Flytta markerat mål upp" },
        { keys: ["↓"], description: "Flytta markerat mål ner" },
        {
          keys: ["w"],
          description: "Gör markerat mål beroende av ett annat mål (sök bland alla mål)",
        },
        ...goalActionShortcuts,
      ],
    },
    {
      id: "now",
      title: "Gör nu",
      accent: "text-accent",
      shortcuts: [
        { keys: ["j"], description: "Nästa mål" },
        { keys: ["k"], description: "Föregående mål" },
        { keys: ["l"], description: "Öppna markerat mål" },
        { keys: ["↑"], description: "Flytta upp i prioritetsordningen" },
        { keys: ["↓"], description: "Flytta ner i prioritetsordningen" },
        {
          keys: ["Shift+P", "J"],
          description: "Minska vikten (håll Shift+P nedtryckt)",
        },
        {
          keys: ["Shift+P", "K"],
          description: "Öka vikten (håll Shift+P nedtryckt)",
        },
        ...goalActionShortcuts,
      ],
    },
    {
      id: "leader",
      title: "Leader-läge",
      accent: "text-green-400",
      container: "border-line bg-surface-2/40",
      shortcuts: [
        { keys: ["Space"], description: "Öppna leader-läget" },
        { keys: ["g", "i"], description: "Gå till inkorgen" },
        { keys: ["g", "g"], description: "Gå till mållistan" },
        { keys: ["g", "n"], description: "Gå till Gör nu" },
        { keys: ["g", "c"], description: "Gå till kalendern" },
        { keys: ["g", "d"], description: "Gå till avslutade mål" },
        { keys: ["g", "s"], description: "Gå till inställningar" },
        { keys: ["m"], description: "Visa/dölj sidomenyn" },
        {
          keys: ["f", "d"],
          description: "Visa/dölj avslutade mål (målsidan)",
        },
        {
          keys: ["i"],
          description: "Byt ikon för markerat mål (målsidan)",
        },
        {
          keys: ["a"],
          description: "Växla agent-märke för markerat mål",
        },
        {
          keys: ["n"],
          description: "Växla nu-märke för markerat mål (syns i Gör nu)",
        },
        {
          keys: ["d"],
          description: "Sätt startdatum och deadline (målsidan, /goals, /now)",
        },
        {
          keys: ["t"],
          description: "Sätt taggar på markerat mål (målsidan, /goals, /now)",
        },
        { keys: ["Esc"], description: "Avbryt leader-läget" },
      ],
    },
  ];

  if (isCalendar.value) {
    list.push({
      id: "calendar",
      title: "Kalender",
      accent: "text-cyan-400",
      container: "border-cyan-400/20 bg-cyan-400/5",
      shortcuts: [
        { keys: ["t"], description: "Hoppa till nu (visar närmaste 24 timmarna)" },
        { keys: ["r"], description: "Återställ kalenderns vy och zoom" },
        { keys: ["j"], description: "Zooma ut" },
        { keys: ["↓"], description: "Zooma ut" },
        { keys: ["k"], description: "Zooma in" },
        { keys: ["↑"], description: "Zooma in" },
        { keys: ["h"], description: "Flytta tiden bakåt" },
        { keys: ["l"], description: "Flytta tiden framåt" },
        { keys: ["+"], description: "Zooma in" },
        { keys: ["−"], description: "Zooma ut" },
        {
          keys: ["Drag", "Scroll"],
          description:
            "Dra/scrolla lodrätt för att zooma, vågrätt för att flytta tiden",
        },
        { keys: ["Esc"], description: "Stäng öppen händelse eller dialog" },
      ],
    });
  }

  list.push({
    id: "editing",
    title: "Redigering & dialoger",
    accent: "text-fuchsia-400",
    container: "border-line bg-surface-2/40",
    shortcuts: [
      { keys: ["Enter"], description: "Bekräfta val i sökfält och redigering" },
      {
        keys: ["↑", "↓"],
        description:
          "Markera ett befintligt mål i sökfält (Enter skapar annars ett nytt mål)",
      },
      {
        keys: ["Ctrl", "Enter"],
        description: "Spara beskrivningen (Cmd+Enter på Mac)",
      },
      { keys: ["Esc"], description: "Avbryt redigering eller stäng dialog" },
      { keys: ["h"], description: "Avbryt i dialogen för att ta bort" },
      { keys: ["l"], description: "Bekräfta i dialogen för att ta bort" },
    ],
  });

  return list;
});

// Sökningen matchar både tangentnamnet och beskrivningen, så både "ctrl"
// och "deadline" hittar rätt rad. Flera ord måste alla stämma (AND).
const filteredSections = computed<Section[]>(() => {
  const terms = query.value.trim().toLowerCase().split(/\s+/).filter(Boolean);
  if (terms.length === 0) return sections.value;

  return sections.value
    .map((section) => ({
      ...section,
      shortcuts: section.shortcuts.filter((shortcut) => {
        const haystack = [...shortcut.keys, shortcut.description]
          .join(" ")
          .toLowerCase();
        return terms.every((term) => haystack.includes(term));
      }),
    }))
    .filter((section) => section.shortcuts.length > 0);
});

const totalShortcuts = computed(() =>
  sections.value.reduce((sum, section) => sum + section.shortcuts.length, 0),
);

const visibleShortcuts = computed(() =>
  filteredSections.value.reduce(
    (sum, section) => sum + section.shortcuts.length,
    0,
  ),
);

// När modalen öppnas ska filtret vara tomt och ha fokus, så man kan
// börja skriva direkt.
watch(
  () => props.open,
  async (isOpen) => {
    if (!isOpen) return;
    query.value = "";
    await nextTick();
    searchInput.value?.focus();
  },
);

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
          <p class="mt-1 text-sm text-ink-muted">
            Snabbkommandon för den vy du använder
          </p>
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

      <!-- Filtret får fokus när modalen öppnas, så man kan söka direkt. -->
      <div class="mb-5">
        <div
          class="flex items-center gap-3 px-4 py-2.5 bg-surface-2 border border-line-strong rounded-xl focus-within:border-accent transition-colors"
        >
          <Icon
            name="lucide:search"
            class="w-4 h-4 text-ink-dim shrink-0"
          />
          <input
            ref="searchInput"
            v-model="query"
            type="text"
            placeholder="Filtrera kommandon…"
            spellcheck="false"
            autocomplete="off"
            class="flex-1 bg-transparent outline-none text-sm text-white placeholder:text-ink-dim"
          />
          <button
            v-if="query"
            type="button"
            @click="query = ''"
            class="text-ink-dim hover:text-white transition-colors"
            aria-label="Rensa filtret"
          >
            <Icon name="lucide:x" class="w-4 h-4" />
          </button>
        </div>
        <p class="mt-2 text-xs text-ink-dim">
          <template v-if="query">
            {{ visibleShortcuts }} av {{ totalShortcuts }} kommandon matchar
          </template>
          <template v-else>
            Skriv för att filtrera - t.ex. ”deadline”, ”ctrl” eller ”kalender”.
          </template>
        </p>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
        <div
          v-for="section in filteredSections"
          :key="section.id"
          class="rounded-xl border p-5"
          :class="section.container ?? 'border-line bg-surface-2/40'"
        >
          <h3
            class="text-sm font-semibold uppercase tracking-widest mb-4"
            :class="section.accent"
          >
            {{ section.title }}
          </h3>
          <div class="space-y-2">
            <div
              v-for="(shortcut, index) in section.shortcuts"
              :key="index"
              class="flex items-center gap-4"
            >
              <template v-for="(key, keyIndex) in shortcut.keys" :key="keyIndex">
                <kbd
                  class="px-2 py-1 bg-surface-2 border border-line-strong rounded font-mono font-bold min-w-[2rem] text-center"
                  :class="section.accent"
                  >{{ key }}</kbd
                >
              </template>
              <span class="text-ink-muted">{{ shortcut.description }}</span>
            </div>
          </div>
        </div>

        <p
          v-if="filteredSections.length === 0"
          class="col-span-full text-center text-sm text-ink-muted py-8"
        >
          Inga kommandon matchar sökningen.
        </p>
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
