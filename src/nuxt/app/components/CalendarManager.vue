<script setup lang="ts">
/**
 * Hanterare för kalendrar.
 *
 * Användaren kan skapa hur många kalendrar som helst, ge dem eget namn
 * och egen färg, dölja dem i tidslinjen och ta bort dem. Händelserna i en
 * borttagen kalender finns kvar, men ligger inte längre i någon kalender.
 */

const props = defineProps<{ open: boolean }>();
const emit = defineEmits<{ "update:open": [boolean] }>();

const { calendars, load, add, save, remove, isHidden, toggleVisibility } = useCalendars();
import { normalizeHexColor } from '~/utils/calendarColors'

/** Färger som snabbt går att välja när en ny kalender skapas. */
const PALETTE = [
  "#6366f1",
  "#0ea5e9",
  "#10b981",
  "#f59e0b",
  "#ef4444",
  "#ec4899",
  "#8b5cf6",
  "#64748b",
];

const newName = ref("");
/** Färgkoden som skrivs, så att det går att ange vilken som helst färg. */
const newColorText = ref(PALETTE[0] ?? "#6366f1");
const saving = ref(false);
const errorMessage = ref("");

/** Den giltiga färgen, eller null om färgkoden är felaktig. */
const normalizedColor = computed(() => normalizeHexColor(newColorText.value));

const colorError = computed(() => {
  if (newColorText.value.trim() === "") return "Skriv en färgkod, till exempel #6366f1.";
  if (!normalizedColor.value) return "Färgkoden måste se ut som #6366f1.";
  return "";
});

/** Namn som redigeras just nu, så att fältet inte hoppar medan man skriver. */
const draftNames = reactive<Record<number, string>>({});

/** Färgkoder som redigeras just nu, på samma sätt som namnen. */
const draftColors = reactive<Record<number, string>>({});

// Namn- och färgfältet redigeras i en egen draft, som fylls på från
// kalenderns nuvarande värden. Utan det skulle fälten visa tomma strängar.
watch(
  calendars,
  (list) => {
    for (const calendar of list) {
      if (draftNames[calendar.id] === undefined) {
        draftNames[calendar.id] = calendar.name;
      }
      if (draftColors[calendar.id] === undefined) {
        draftColors[calendar.id] = calendar.color;
      }
    }
  },
  { immediate: true, deep: true },
);

const close = () => emit("update:open", false);

const create = async () => {
  const name = newName.value.trim();
  const color = normalizedColor.value;
  if (!name || !color) return;

  saving.value = true;
  errorMessage.value = "";
  try {
    // Den första kalendern blir standardkalender, så att nya händelser
    // hamnar någonstans från början.
    await add(name, color, calendars.value.length === 0);
    newName.value = "";
    // Nästa kalender får en ny färg, så att två kalendrar inte blir
    // identiska av en slump.
    const index = PALETTE.indexOf(color);
    newColorText.value = PALETTE[(index + 1) % PALETTE.length] ?? PALETTE[0];
  } catch (error) {
    console.error("Kunde inte skapa kalendern:", error);
    errorMessage.value = "Kunde inte skapa kalendern. Försök igen.";
  } finally {
    saving.value = false;
  }
};

const rename = async (calendarId: number) => {
  const calendar = calendars.value.find((candidate) => candidate.id === calendarId);
  if (!calendar) return;

  const name = (draftNames[calendarId] ?? calendar.name).trim();
  if (!name || name === calendar.name) {
    draftNames[calendarId] = calendar.name;
    return;
  }

  try {
    await save(calendarId, { name });
    draftNames[calendarId] = name;
  } catch (error) {
    console.error("Kunde inte byta namn på kalendern:", error);
    draftNames[calendarId] = calendar.name;
  }
};

const recolor = async (calendarId: number, color: string) => {
  try {
    await save(calendarId, { color });
  } catch (error) {
    console.error("Kunde inte byta färg på kalendern:", error);
  }
};

/**
 * Sätter kalenderns färg från den färgkod som skrivits i fältet.
 *
 * En ogiltig kod avvisas och fältet återställs till kalenderns nuvarande
 * färg, så att man aldrig ser en trasig färg i tidslinjen.
 */
const applyHex = async (calendarId: number) => {
  const calendar = calendars.value.find((candidate) => candidate.id === calendarId);
  if (!calendar) return;

  const normalized = normalizeHexColor(draftColors[calendarId] ?? "");
  if (!normalized) {
    draftColors[calendarId] = calendar.color;
    errorMessage.value = "Färgkoden måste se ut som #6366f1.";
    return;
  }
  if (normalized === calendar.color) {
    draftColors[calendarId] = calendar.color;
    return;
  }

  try {
    await save(calendarId, { color: normalized });
    draftColors[calendarId] = normalized;
  } catch (error) {
    console.error("Kunde inte byta färg på kalendern:", error);
    draftColors[calendarId] = calendar.color;
    errorMessage.value = "Kunde inte byta färg på kalendern. Försök igen.";
  }
};

/** Gör kalendern till standardkalender. */
const setDefault = async (calendarId: number) => {
  try {
    await save(calendarId, { is_default: true });
  } catch (error) {
    console.error("Kunde inte sätta standardkalendern:", error);
    errorMessage.value = "Kunde inte sätta standardkalendern. Försök igen.";
  }
};

/** Tar bort standardmarkeringen, så att nya händelser blir kalenderlösa. */
const clearDefault = async (calendarId: number) => {
  try {
    await save(calendarId, { is_default: false });
  } catch (error) {
    console.error("Kunde inte ta bort standardkalendern:", error);
    errorMessage.value = "Kunde inte ta bort standardkalendern. Försök igen.";
  }
};

const confirmRemove = async (calendarId: number, name: string) => {
  const confirmed = window.confirm(
    `Ta bort kalendern "${name}"? Händelserna i den finns kvar, men utan kalender.`,
  );
  if (!confirmed) return;

  try {
    await remove(calendarId);
    delete draftNames[calendarId];
    delete draftColors[calendarId];
  } catch (error) {
    console.error("Kunde inte ta bort kalendern:", error);
    errorMessage.value = "Kunde inte ta bort kalendern. Försök igen.";
  }
};

watch(
  () => props.open,
  async (open) => {
    if (open) await load();
  },
);

onMounted(() => load());
</script>

<template>
  <Transition name="manager">
    <aside v-if="open" class="calendar-manager" aria-label="Kalendrar">
      <header class="manager-header">
        <strong>Kalendrar</strong>
        <button type="button" class="close-button" aria-label="Stäng" @click="close">✕</button>
      </header>

      <p v-if="errorMessage" class="error-message" role="status">{{ errorMessage }}</p>

      <ul v-if="calendars.length" class="calendar-list">
        <li v-for="calendar in calendars" :key="calendar.id" class="calendar-row">
          <input
            type="color"
            class="color-input"
            :value="calendar.color"
            :aria-label="`Färg för ${calendar.name}`"
            @change="recolor(calendar.id, ($event.target as HTMLInputElement).value)"
          />
          <input
            v-model="draftColors[calendar.id]"
            class="hex-input"
            spellcheck="false"
            autocomplete="off"
            maxlength="7"
            :aria-label="`Färgkod för ${calendar.name}`"
            @blur="applyHex(calendar.id)"
            @keyup.enter="applyHex(calendar.id)"
          />
          <input
            v-model="draftNames[calendar.id]"
            class="name-input"
            maxlength="64"
            :aria-label="`Namn på ${calendar.name}`"
            @blur="rename(calendar.id)"
            @keyup.enter="rename(calendar.id)"
          />
          <button
            type="button"
            class="icon-button star"
            :class="{ active: calendar.is_default }"
            :aria-pressed="calendar.is_default"
            :aria-label="
              calendar.is_default
                ? `Ta bort ${calendar.name} som standardkalender`
                : `Sätt ${calendar.name} som standardkalender`
            "
            :title="
              calendar.is_default
                ? 'Ingen standardkalender'
                : 'Använd som standardkalender för nya händelser'
            "
            @click="calendar.is_default ? clearDefault(calendar.id) : setDefault(calendar.id)"
          >
            {{ calendar.is_default ? "★" : "☆" }}
          </button>
          <button
            type="button"
            class="icon-button"
            :aria-pressed="isHidden(calendar.id)"
            :aria-label="isHidden(calendar.id) ? 'Visa i tidslinjen' : 'Dölj i tidslinjen'"
            :title="isHidden(calendar.id) ? 'Visa i tidslinjen' : 'Dölj i tidslinjen'"
            @click="toggleVisibility(calendar.id)"
          >
            {{ isHidden(calendar.id) ? "🙈" : "👁" }}
          </button>
          <button
            type="button"
            class="icon-button danger"
            :aria-label="`Ta bort ${calendar.name}`"
            @click="confirmRemove(calendar.id, calendar.name)"
          >
            🗑
          </button>
        </li>
      </ul>

      <p v-else class="empty-hint">
        Du har inga kalendrar än. Lägg till en för att kunna skilja till exempel jobb från fritid.
      </p>

      <form class="create-form" @submit.prevent="create">
        <input
          v-model="newName"
          maxlength="64"
          placeholder="Ny kalender, t.ex. Jobb"
          aria-label="Kalendernamn"
        />
        <!-- Färgkoden kan skrivas valfri, och paletten under är genvägar. -->
        <div class="color-field">
          <span
            class="color-preview"
            :style="{ background: normalizedColor ?? 'transparent' }"
            aria-hidden="true"
          />
          <input
            v-model="newColorText"
            class="hex-input"
            spellcheck="false"
            autocomplete="off"
            maxlength="7"
            placeholder="#6366f1"
            aria-label="Färgkod"
          />
        </div>
        <p v-if="colorError" class="error-message" role="status">{{ colorError }}</p>
        <div class="palette" role="group" aria-label="Färg">
          <button
            v-for="color in PALETTE"
            :key="color"
            type="button"
            class="swatch"
            :class="{ selected: normalizedColor === color }"
            :style="{ background: color }"
            :aria-label="`Färg ${color}`"
            @click="newColorText = color"
          />
        </div>
        <button
          type="submit"
          class="save-button"
          :disabled="saving || !newName.trim() || !normalizedColor"
        >
          {{ saving ? "Sparar…" : "Lägg till" }}
        </button>
      </form>
    </aside>
  </Transition>
</template>

<style scoped>
/* Listan med kalendrar. Den öppnas under knappen i verktygsraden och
   ligger därför inte längre uppe i hörnet som en klump. */
.calendar-manager {
  position: absolute;
  top: 3rem;
  right: 1.5rem;
  z-index: 30;
  width: min(24rem, calc(100% - 2rem));
  max-height: calc(100% - 3.5rem);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.85rem;
  border: 1px solid var(--color-line-strong);
  border-radius: var(--radius-xl);
  background: var(--color-surface);
  box-shadow: 0 18px 40px rgb(0 0 0 / 0.45);
}

.manager-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--color-ink);
}

.close-button {
  border: none;
  background: none;
  color: var(--color-ink-muted);
  cursor: pointer;
  font-size: 0.9rem;
}

.create-form {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding-top: 0.65rem;
  border-top: 1px solid var(--color-line);
}

.create-form input[type="text"],
.create-form input:not([type]) {
  width: 100%;
}

.create-form input {
  padding: 0.4rem 0.55rem;
  border: 1px solid var(--color-line-strong);
  border-radius: var(--radius-lg);
  background: var(--color-surface-2);
  color: var(--color-ink);
}

.palette {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

/* Färgkoden och en förhandsvisning av den. */
.color-field {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.color-preview {
  width: 1.35rem;
  height: 1.35rem;
  flex-shrink: 0;
  border: 1px solid var(--color-line-strong);
  border-radius: 9999px;
  background: transparent;
}

.hex-input {
  width: 100%;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.78rem;
  letter-spacing: 0.02em;
}

.calendar-row .hex-input {
  width: 5.2rem;
  flex-shrink: 0;
  padding: 0.3rem 0.4rem;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  background: transparent;
  color: var(--color-ink-muted);
}

.calendar-row .hex-input:hover,
.calendar-row .hex-input:focus {
  border-color: var(--color-line-strong);
  background: var(--color-surface-2);
  color: var(--color-ink);
}

.swatch {
  width: 1.35rem;
  height: 1.35rem;
  border: 2px solid transparent;
  border-radius: 9999px;
  cursor: pointer;
}

.swatch.selected {
  border-color: var(--color-ink);
}

.save-button {
  padding: 0.4rem 0.6rem;
  border: 1px solid var(--color-accent-strong);
  border-radius: var(--radius-lg);
  background: var(--color-accent-deep);
  color: white;
  cursor: pointer;
}

.save-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.calendar-list {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  list-style: none;
  margin: 0;
  padding: 0;
}

.calendar-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.color-input {
  width: 1.75rem;
  height: 1.75rem;
  padding: 0;
  border: 1px solid var(--color-line-strong);
  border-radius: var(--radius-lg);
  background: var(--color-surface-2);
  cursor: pointer;
}

.name-input {
  flex: 1;
  min-width: 0;
  padding: 0.35rem 0.5rem;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  background: transparent;
  color: var(--color-ink);
}

.name-input:hover,
.name-input:focus {
  border-color: var(--color-line-strong);
  background: var(--color-surface-2);
}

.icon-button {
  border: 1px solid var(--color-line);
  border-radius: var(--radius-lg);
  background: var(--color-surface-2);
  color: var(--color-ink-muted);
  cursor: pointer;
  font-size: 0.85rem;
  padding: 0.25rem 0.4rem;
}

/* En dold kalender visas som nedtryckt i listan, så att det går att se
   vilka kalendrar som inte syns i tidslinjen. */
.icon-button[aria-pressed="true"] {
  opacity: 0.55;
}

.icon-button.danger:hover {
  border-color: #ef4444;
  color: #ef4444;
}

.icon-button.star {
  color: var(--color-ink-muted);
}

.icon-button.star.active {
  color: #fbbf24;
  border-color: #fbbf24;
}

.error-message {
  color: #f87171;
  font-size: 0.8rem;
}

.empty-hint {
  color: var(--color-ink-muted);
  font-size: 0.8rem;
  line-height: 1.4;
}

.manager-enter-active,
.manager-leave-active {
  transition: opacity 0.16s ease, transform 0.16s ease;
}

.manager-enter-from,
.manager-leave-to {
  opacity: 0;
  transform: translateX(0.75rem);
}

@media (max-width: 640px) {
  .calendar-manager {
    top: 2.8rem;
    right: 0.45rem;
    width: min(24rem, calc(100% - 0.9rem));
    max-height: calc(100% - 3.2rem);
  }
}
</style>
