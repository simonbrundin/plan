<script setup lang="ts">
import {
	describeParsedEntry,
	parseNaturalLanguage,
	resolveQuickAddMode,
	type ParsedEntry,
} from "~/utils/naturalLanguage";

const emit = defineEmits<{
	created: [];
	closed: [];
}>();

const { createCalendarEvent } = useCalendarApi();
const { createGoal } = useGoalApi();
const { isOpen, mode, close, markCreated } = useQuickAdd();
const { calendars, load: loadCalendars } = useCalendars();

/**
 * Vilken kalender den nya händelsen ska ligga i. Tomt betyder "ingen
 * kalender", och händelsen får då standardfärgen.
 */
const calendarId = ref("");

const text = ref("");
const saving = ref(false);
const errorMessage = ref("");
const inputRef = ref<HTMLInputElement | null>(null);

const parsed = computed<ParsedEntry>(() => parseNaturalLanguage(text.value));
const trimmedText = computed(() => text.value.trim());

/** Emoji som plockats ut ur texten, visas i förhandsvisningen. */
const icon = computed(() => parsed.value.icon);

/**
 * I automatiskt läge blir det en händelse om texten innehåller datum,
 * klockslag eller upprepning, annars ett mål.
 */
const effectiveMode = computed<"event" | "goal">(() =>
	resolveQuickAddMode(parsed.value, mode.value),
);

const title = computed(() => {
	if (!trimmedText.value) return "";
	// Ett mål behåller hela texten. Mål har inget datumfält, så att
	// tappa bort ord skulle bara förlora information.
	if (effectiveMode.value === "goal") return trimmedText.value;
	return parsed.value.title;
});

const canSubmit = computed(() => title.value.length > 0 && !saving.value);

const preview = computed(() => {
	if (!trimmedText.value) return "";
	if (effectiveMode.value === "goal") return "";
	return describeParsedEntry(parsed.value);
});

const recurrenceLabel = computed(() => {
	const rule = parsed.value.recurrenceRule;
	if (!rule) return "";
	if (rule.includes("DAILY")) return "varje dag";
	if (rule.includes("WEEKLY")) {
		if (rule.includes("INTERVAL=2")) return "varannan vecka";
		const day = rule.match(/BYDAY=([A-Z]{2})/);
		const names: Record<string, string> = {
			MO: "måndag",
			TU: "tisdag",
			WE: "onsdag",
			TH: "torsdag",
			FR: "fredag",
			SA: "lördag",
			SU: "söndag",
		};
		return day ? `varje ${names[day[1]] ?? "vecka"}` : "varje vecka";
	}
	if (rule.includes("MONTHLY")) return "varje månad";
	if (rule.includes("YEARLY")) return "varje år";
	return "";
});

const focusInput = () => {
	nextTick(() => inputRef.value?.focus());
};

const reset = () => {
	text.value = "";
	mode.value = "auto";
	errorMessage.value = "";
	calendarId.value = "";
};

const submit = async () => {
	if (!canSubmit.value) return;

	saving.value = true;
	errorMessage.value = "";

	try {
		if (effectiveMode.value === "goal") {
			await createGoal(title.value);
		} else {
			const entry = parsed.value;
			const start = entry.startAt ?? new Date();
			const end =
				entry.endAt ?? new Date(start.getTime() + 60 * 60 * 1000);

			await createCalendarEvent({
				title: title.value,
				start_at: start.toISOString(),
				end_at: end.toISOString(),
				time_zone: Intl.DateTimeFormat().resolvedOptions().timeZone,
				all_day: entry.allDay,
				...(entry.icon ? { icon: entry.icon } : {}),
				...(calendarId.value ? { calendar_id: Number(calendarId.value) } : {}),
				...(entry.recurrenceRule
					? { recurrence_rule: entry.recurrenceRule }
					: {}),
			});
		}

		reset();
		close();
		markCreated();
		emit("created");
	} catch (error) {
		console.error("Kunde inte skapa innehållet:", error);
		errorMessage.value = "Kunde inte skapa. Försök igen.";
	} finally {
		saving.value = false;
	}
};

const handleKeydown = (event: KeyboardEvent) => {
	if (event.key === "Enter") {
		event.preventDefault();
		submit();
		return;
	}
	if (event.key === "Escape") {
		event.preventDefault();
		event.stopPropagation();
		reset();
		close();
		emit("closed");
	}
};

watch(isOpen, (open) => {
	if (open) {
		focusInput();
		// Kalendrarna behövs för att välja vilken kalender händelsen
		// ska ligga i.
		loadCalendars();
	}
});

onMounted(() => {
	if (isOpen.value) focusInput();
});

defineExpose({ focusInput });
</script>

<template>
	<Teleport to="body">
		<div v-if="isOpen" class="quick-add-backdrop" @click.self="close(); emit('closed')">
			<div class="quick-add-panel" role="dialog" aria-label="Lägg till">
				<input
					ref="inputRef"
					v-model="text"
					type="text"
					class="quick-add-input"
					placeholder="Lunch med Anna imorgon kl 12"
					aria-label="Skriv naturlig text"
					@keydown="handleKeydown"
				/>

				<div class="quick-add-footer">
					<div class="quick-add-modes" role="group" aria-label="Typ av innehåll">
						<button
							type="button"
							class="quick-add-mode"
							:class="{ active: mode === 'auto' }"
							@click="mode = 'auto'"
						>
							Auto
						</button>
						<button
							type="button"
							class="quick-add-mode"
							:class="{ active: mode === 'event' }"
							@click="mode = 'event'"
						>
							Händelse
						</button>
						<button
							type="button"
							class="quick-add-mode"
							:class="{ active: mode === 'goal' }"
							@click="mode = 'goal'"
						>
							Mål
						</button>
					</div>

					<p v-if="trimmedText" class="quick-add-preview">
						<span class="quick-add-kind">
							{{ effectiveMode === "goal" ? "Mål" : "Händelse" }}
						</span>
						<EventIcon v-if="icon" :icon="icon" class="quick-add-icon" />
						<strong>{{ title }}</strong>
						<span v-if="preview" class="quick-add-when">{{ preview }}</span>
						<span v-if="recurrenceLabel" class="quick-add-when">
							{{ recurrenceLabel }}
						</span>
					</p>

					<select
						v-if="effectiveMode === 'event' && calendars.length"
						v-model="calendarId"
						class="quick-add-calendar"
						aria-label="Kalender"
					>
						<option value="">Utan kalender</option>
						<option v-for="calendar in calendars" :key="calendar.id" :value="String(calendar.id)">
							{{ calendar.name }}
						</option>
					</select>

					<button
						type="button"
						class="quick-add-submit"
						:disabled="!canSubmit"
						@click="submit"
					>
						{{ saving ? "Sparar" : "Lägg till" }}
					</button>
				</div>

				<p v-if="errorMessage" class="quick-add-error">{{ errorMessage }}</p>
			</div>
		</div>
	</Teleport>
</template>

<style scoped>
.quick-add-backdrop {
	position: fixed;
	inset: 0;
	z-index: 100;
	display: flex;
	justify-content: center;
	padding: 12vh 1rem 1rem;
	background: rgb(4 6 14 / 0.55);
	backdrop-filter: blur(3px);
}

.quick-add-panel {
	width: min(34rem, 100%);
	border: 1px solid var(--color-line-strong);
	border-radius: 0.9rem;
	padding: 0.75rem;
	background: var(--color-surface);
	box-shadow: 0 1.5rem 4rem rgb(0 0 0 / 0.45);
}

.quick-add-input {
	width: 100%;
	border: 0;
	padding: 0.65rem 0.4rem;
	color: var(--color-ink);
	background: transparent;
	font-size: 1.05rem;
	outline: none;
}

.quick-add-input::placeholder {
	color: var(--color-ink-dim);
}

.quick-add-footer {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: 0.6rem;
	border-top: 1px solid var(--color-line);
	padding-top: 0.6rem;
}

.quick-add-modes {
	display: flex;
	border: 1px solid var(--color-line);
	border-radius: 0.5rem;
	overflow: hidden;
}

/* Kalenderväljaren i snabbfältet. Den visas bara för händelser, eftersom
   mål inte hör till någon kalender. */
.quick-add-calendar {
	border: 1px solid var(--color-line);
	border-radius: 0.5rem;
	padding: 0.28rem 0.5rem;
	color: var(--color-ink-muted);
	background: var(--color-surface-2);
	font-size: 0.85rem;
}

.quick-add-mode {
	border: 0;
	padding: 0.28rem 0.5rem;
	color: var(--color-ink-dim);
	background: transparent;
	font-size: 0.68rem;
	cursor: pointer;
	transition: color 0.15s ease, background 0.15s ease;
}

.quick-add-mode:hover {
	color: var(--color-ink);
}

.quick-add-mode.active {
	color: var(--color-ink);
	background: rgb(129 140 248 / 0.2);
}

.quick-add-preview {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: 0.4rem;
	flex: 1;
	margin: 0;
	color: var(--color-ink-muted);
	font-size: 0.75rem;
}

.quick-add-preview strong {
	color: var(--color-ink);
	font-weight: 600;
}

.quick-add-kind {
	border: 1px solid var(--color-line-strong);
	border-radius: 999px;
	padding: 0.1rem 0.45rem;
	color: var(--color-ink-dim);
	font-size: 0.62rem;
	text-transform: uppercase;
	letter-spacing: 0.04em;
}

.quick-add-icon {
	font-size: 1rem;
	line-height: 1;
}

.quick-add-when {
	color: #a5b4fc;
}

.quick-add-submit {
	border: 0;
	border-radius: 0.5rem;
	padding: 0.4rem 0.8rem;
	color: #fff;
	background: var(--color-accent-strong);
	font-size: 0.78rem;
	font-weight: 600;
	cursor: pointer;
	transition: opacity 0.15s ease;
}

.quick-add-submit:disabled {
	opacity: 0.4;
	cursor: not-allowed;
}

.quick-add-error {
	margin: 0.5rem 0 0;
	color: #f87171;
	font-size: 0.75rem;
}

@media (max-width: 640px) {
	.quick-add-backdrop {
		padding: 6vh 0.6rem 0.6rem;
	}

	.quick-add-input {
		font-size: 1rem;
	}
}
</style>
