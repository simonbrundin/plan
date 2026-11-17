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
const newColor = ref(PALETTE[0] ?? "#6366f1");
const saving = ref(false);
const errorMessage = ref("");

/** Namn som redigeras just nu, så att fältet inte hoppar medan man skriver. */
const draftNames = reactive<Record<number, string>>({});

// Namnfältet redigeras i en egen draft, som fylls på från kalenderns
// nuvarande namn. Utan det skulle fältet visa en tom sträng.
watch(
  calendars,
  (list) => {
    for (const calendar of list) {
      if (draftNames[calendar.id] === undefined) {
        draftNames[calendar.id] = calendar.name;
      }
    }
  },
  { immediate: true, deep: true },
);

const close = () => emit("update:open", false);

const create = async () => {
  const name = newName.value.trim();
  if (!name) return;

  saving.value = true;
  errorMessage.value = "";
  try {
    await add(name, newColor.value);
    newName.value = "";
    // Nästa kalender får en ny färg, så att två kalendrar inte blir
    // identiska av en slump.
    const index = PALETTE.indexOf(newColor.value);
    newColor.value = PALETTE[(index + 1) % PALETTE.length] ?? PALETTE[0];
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

const confirmRemove = async (calendarId: number, name: string) => {
  const confirmed = window.confirm(
    `Ta bort kalendern "${name}"? Händelserna i den finns kvar, men utan kalender.`,
  );
  if (!confirmed) return;

  try {
    await remove(calendarId);
    delete draftNames[calendarId];
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

      <form class="create-form" @submit.prevent="create">
        <input
          v-model="newName"
          maxlength="64"
          placeholder="Ny kalender, t.ex. Jobb"
          aria-label="Kalendernamn"
        />
        <div class="palette" role="group" aria-label="Färg">
          <button
            v-for="color in PALETTE"
            :key="color"
            type="button"
            class="swatch"
            :class="{ selected: newColor === color }"
            :style="{ background: color }"
            :aria-label="`Färg ${color}`"
            @click="newColor = color"
          />
        </div>
        <button type="submit" class="save-button" :disabled="saving || !newName.trim()">
          {{ saving ? "Sparar…" : "Lägg till" }}
        </button>
      </form>

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
            v-model="draftNames[calendar.id]"
            class="name-input"
            maxlength="64"
            :aria-label="`Namn på ${calendar.name}`"
            @blur="rename(calendar.id)"
            @keyup.enter="rename(calendar.id)"
          />
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
    </aside>
  </Transition>
</template>

<style scoped>
.calendar-manager {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  z-index: 30;
  width: min(22rem, calc(100% - 1rem));
  max-height: calc(100% - 1rem);
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

.icon-button[aria-pressed="true"] {
  opacity: 0.55;
}

.icon-button.danger:hover {
  border-color: #ef4444;
  color: #ef4444;
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
</style>
