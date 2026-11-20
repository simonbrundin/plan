<script setup lang="ts">
import {
	describeParsedEntry,
	goalTitle,
	parseNaturalLanguage,
	resolveQuickAddMode,
	type ParsedEntry,
} from "~/utils/naturalLanguage";
import { describeGoalRecurrence } from "~/utils/recurrence";
import type { Goal } from "~/types/goal";

const emit = defineEmits<{
	created: [];
	closed: [];
}>();

const { createCalendarEvent } = useCalendarApi();
const { createGoal, addParentRelation } = useGoalApi();
const { isOpen, mode, close, markCreated } = useQuickAdd();
const { calendars, load: loadCalendars, defaultCalendarId } = useCalendars();
const goalsStore = useGoalsStore();

/**
 * Vilken kalender den nya händelsen ska ligga i. Tomt betyder "ingen
 * kalender", och händelsen får då standardfärgen.
 */
const calendarId = ref("");

const text = ref("");
const saving = ref(false);
const errorMessage = ref("");
const inputRef = ref<HTMLInputElement | null>(null);
const parentSearchInput = ref<HTMLInputElement | null>(null);
const panelRef = ref<HTMLDivElement | null>(null);
const parentSearchQuery = ref("");
const selectedParent = ref<Goal | null>(null);

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
	if (effectiveMode.value === "goal") {
		return goalTitle(trimmedText.value, parsed.value);
	}
	return parsed.value.title;
});

const parentSearchResults = computed(() => {
	const query = parentSearchQuery.value.trim().toLowerCase();
	if (!query || selectedParent.value) return [];

	return goalsStore.goals
		.filter((goal) => goal.id !== selectedParent.value?.id)
		.filter((goal) => goal.title.toLowerCase().includes(query))
		.slice(0, 8);
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
	return describeGoalRecurrence(rule);
});

const focusInput = () => {
	nextTick(() => inputRef.value?.focus());
};

const focusParentSearch = () => {
	nextTick(() => parentSearchInput.value?.focus());
};

const handleMainInputBlur = () => {
	if (effectiveMode.value === "goal" && trimmedText.value) {
		focusParentSearch();
	}
};

const closePopup = () => {
	reset();
	close();
	emit("closed");
};

const handleEscape = (event: KeyboardEvent) => {
	event.preventDefault();
	event.stopPropagation();
	closePopup();
};

const handleParentSearchKeydown = async (event: KeyboardEvent) => {
	if (event.key === "Enter") {
		event.preventDefault();
		event.stopPropagation();
		const firstResult = parentSearchResults.value[0];
		if (firstResult) {
			selectParent(firstResult);
			await submit();
		}
		return;
	}
	if (event.key === "Escape") handleEscape(event);
};

const reset = () => {
	text.value = "";
	mode.value = "auto";
	errorMessage.value = "";
	calendarId.value = "";
	parentSearchQuery.value = "";
	selectedParent.value = null;
};

const selectParent = (goal: Goal) => {
	selectedParent.value = goal;
	parentSearchQuery.value = "";
	// Förhindra att Enter-eventet aktiverar knappen när sökfältet försvinner.
	// Fokusera popupens panel i stället; målet sparas först med ett musklick.
	nextTick(() => panelRef.value?.focus());
};

const clearParent = () => {
	selectedParent.value = null;
	parentSearchQuery.value = "";
};

const submit = async () => {
	if (!canSubmit.value) return;

	saving.value = true;
	errorMessage.value = "";

	try {
		if (effectiveMode.value === "goal") {
			const newGoal = await createGoal(
				title.value,
				undefined,
				undefined,
				parsed.value.recurrenceRule ?? undefined,
			);
			goalsStore.addGoal({
				...newGoal,
				has_parent: Boolean(selectedParent.value),
			});
			if (selectedParent.value) {
				await addParentRelation(newGoal.id, selectedParent.value.id);
				goalsStore.addRelation(newGoal.id, selectedParent.value.id);
			}
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
				// Nya händelser hamnar i standardkalendern om ingen kalender
				// väljs uttryckligen. 0 betyder "ingen kalender".
				calendar_id: calendarId.value ? Number(calendarId.value) : 0,
				...(entry.recurrenceRule
					? { recurrence_rule: entry.recurrenceRule }
					: {}),
			});
		}

		markCreated();
		reset();
		close();
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
		if (effectiveMode.value === "goal") {
			if (!selectedParent.value && trimmedText.value) focusParentSearch();
			return;
		}
		submit();
		return;
	}
	if (event.key === "Escape") handleEscape(event);
};

watch(isOpen, async (open) => {
	if (!open) {
		reset();
		return;
	}

	if (open) {
		focusInput();
		// Kalendrarna behövs för att välja vilken kalender händelsen
		// ska ligga i.
		await Promise.all([
			loadCalendars(),
			goalsStore.loadGraph(true).catch((error) => {
				// Händelser ska fortfarande kunna skapas om målgraphen tillfälligt
				// inte går att hämta.
				console.error("Kunde inte hämta mål för föräldrasökning:", error);
			}),
		]);
		// Standardkalendern är förvald, så att det bara är att skriva.
		calendarId.value = defaultCalendarId.value
			? String(defaultCalendarId.value)
			: "";
	}
});

const handleWindowKeydown = (event: KeyboardEvent) => {
	if (isOpen.value && event.key === "Escape") handleEscape(event);
};

onMounted(() => {
	window.addEventListener("keydown", handleWindowKeydown, true);
	if (isOpen.value) focusInput();
});

onBeforeUnmount(() => {
	window.removeEventListener("keydown", handleWindowKeydown, true);
});

defineExpose({ focusInput });
</script>

<template>
	<Teleport to="body">
		<div v-if="isOpen" class="quick-add-backdrop" @click.self="closePopup">
			<div
				ref="panelRef"
				class="quick-add-panel"
				role="dialog"
				aria-label="Lägg till"
				tabindex="-1"
				@keydown.esc="handleEscape"
				@keydown.enter.stop.prevent
			>
				<div class="quick-add-header">
					<div>
						<div class="quick-add-eyebrow">Snabbtillägg</div>
						<h2>Vad vill du få gjort?</h2>
					</div>
					<button type="button" class="quick-add-close" aria-label="Stäng" @click="closePopup">
						<span aria-hidden="true">×</span>
					</button>
				</div>
				<input
					ref="inputRef"
					v-model="text"
					type="text"
					class="quick-add-input"
					placeholder="Lunch med Anna imorgon kl 12"
					aria-label="Skriv naturlig text"
					@keydown.stop="handleKeydown"
					@blur="handleMainInputBlur"
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

					<div v-if="effectiveMode === 'goal'" class="quick-add-parent">
						<div v-if="selectedParent" class="quick-add-parent-selected">
							<span>Förälder: {{ selectedParent.title }}</span>
							<button type="button" @click="clearParent">Byt</button>
						</div>
						<template v-else>
							<input
								ref="parentSearchInput"
								v-model="parentSearchQuery"
								type="search"
								class="quick-add-parent-input"
								placeholder="Sök föräldramål · Enter väljer och lägger till"
								aria-label="Sök föräldramål"
								@keydown.stop="handleParentSearchKeydown"
							/>
							<div v-if="parentSearchResults.length" class="quick-add-parent-results">
								<button
									v-for="goal in parentSearchResults"
									:key="goal.id"
									type="button"
									@click="selectParent(goal)"
								>
									{{ goal.title }}
								</button>
							</div>
						</template>
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
	align-items: center;
	justify-content: center;
	padding: 1rem;
	background: rgb(3 5 15 / 0.68);
	backdrop-filter: blur(12px);
	animation: quick-add-fade 0.18s ease-out;
}

.quick-add-panel {
	width: min(40rem, 100%);
	overflow: visible;
	border: 1px solid rgb(148 163 184 / 0.2);
	border-radius: 1.25rem;
	padding: 1.25rem;
	background:
		radial-gradient(circle at 100% 0%, rgb(99 102 241 / 0.16), transparent 34%),
		linear-gradient(145deg, var(--color-surface), var(--color-surface-2));
	box-shadow: 0 1.5rem 5rem rgb(0 0 0 / 0.5), 0 0 0 1px rgb(255 255 255 / 0.03) inset;
	animation: quick-add-rise 0.22s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.quick-add-header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	gap: 1rem;
	margin-bottom: 0.8rem;
}

.quick-add-eyebrow {
	margin-bottom: 0.2rem;
	color: #a5b4fc;
	font-size: 0.65rem;
	font-weight: 700;
	letter-spacing: 0.12em;
	text-transform: uppercase;
}

.quick-add-header h2 {
	margin: 0;
	color: var(--color-ink);
	font-size: 1.35rem;
	font-weight: 650;
	letter-spacing: -0.025em;
}

.quick-add-close {
	display: grid;
	width: 2rem;
	height: 2rem;
	place-items: center;
	border: 1px solid rgb(148 163 184 / 0.18);
	border-radius: 0.65rem;
	color: var(--color-ink-dim);
	background: rgb(15 23 42 / 0.28);
	font-size: 1.35rem;
	line-height: 1;
	cursor: pointer;
	transition: color 0.15s ease, background 0.15s ease, transform 0.15s ease;
}

.quick-add-close:hover {
	color: var(--color-ink);
	background: rgb(129 140 248 / 0.18);
	transform: rotate(4deg);
}

.quick-add-input {
	width: 100%;
	border: 1px solid rgb(148 163 184 / 0.15);
	border-radius: 0.85rem;
	padding: 0.95rem 1rem;
	color: var(--color-ink);
	background: rgb(2 6 23 / 0.28);
	font-size: 1.08rem;
	outline: none;
	transition: border-color 0.15s ease, box-shadow 0.15s ease, background 0.15s ease;
}

.quick-add-input:focus {
	border-color: rgb(129 140 248 / 0.75);
	background: rgb(2 6 23 / 0.42);
	box-shadow: 0 0 0 4px rgb(99 102 241 / 0.12);
}

.quick-add-input::placeholder {
	color: var(--color-ink-dim);
}

.quick-add-footer {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: 0.75rem;
	margin-top: 0.85rem;
}

.quick-add-modes {
	display: flex;
	border: 1px solid rgb(148 163 184 / 0.18);
	border-radius: 0.65rem;
	overflow: hidden;
	background: rgb(15 23 42 / 0.26);
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
	padding: 0.42rem 0.7rem;
	color: var(--color-ink-dim);
	background: transparent;
	font-size: 0.7rem;
	cursor: pointer;
	transition: color 0.15s ease, background 0.15s ease;
}

.quick-add-mode:hover {
	color: var(--color-ink);
}

.quick-add-mode.active {
	color: #eef2ff;
	background: linear-gradient(135deg, rgb(99 102 241 / 0.45), rgb(129 140 248 / 0.2));
	box-shadow: 0 0 0 1px rgb(165 180 252 / 0.18) inset;
}

.quick-add-parent {
	position: relative;
	flex: 1 1 100%;
	min-width: 12rem;
}

.quick-add-parent-input {
	width: 100%;
	border: 1px solid rgb(129 140 248 / 0.42);
	border-radius: 0.7rem;
	padding: 0.62rem 0.75rem;
	color: var(--color-ink);
	background: var(--color-surface-2);
	font-size: 0.78rem;
}

.quick-add-parent-input:focus {
	border-color: rgb(165 180 252 / 0.9);
	outline: none;
	box-shadow: 0 0 0 4px rgb(99 102 241 / 0.12);
}

.quick-add-parent-results {
	position: absolute;
	z-index: 1;
	top: calc(100% + 0.25rem);
	left: 0;
	right: 0;
	max-height: 10rem;
	overflow-y: auto;
	border: 1px solid var(--color-line-strong);
	border-radius: 0.5rem;
	padding: 0.2rem;
	background: var(--color-surface);
	box-shadow: 0 0.75rem 1.5rem rgb(0 0 0 / 0.35);
}

.quick-add-parent-results button {
	display: block;
	width: 100%;
	border: 0;
	border-radius: 0.3rem;
	padding: 0.35rem 0.45rem;
	color: var(--color-ink);
	background: transparent;
	text-align: left;
	font-size: 0.75rem;
	cursor: pointer;
}

.quick-add-parent-results button:hover {
	background: var(--color-surface-2);
}

.quick-add-parent-selected {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.5rem;
	border: 1px solid rgb(74 222 128 / 0.32);
	border-radius: 0.7rem;
	padding: 0.62rem 0.75rem;
	color: #bbf7d0;
	background: rgb(34 197 94 / 0.08);
	font-size: 0.78rem;
}

.quick-add-parent-selected button {
	border: 0;
	padding: 0;
	color: var(--color-accent);
	background: transparent;
	font-size: 0.72rem;
	cursor: pointer;
}

.quick-add-preview {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: 0.45rem;
	flex: 1;
	min-width: 12rem;
	margin: 0;
	padding: 0.55rem 0.7rem;
	border: 1px solid rgb(148 163 184 / 0.12);
	border-radius: 0.7rem;
	color: var(--color-ink-muted);
	background: rgb(15 23 42 / 0.2);
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
	border-radius: 0.7rem;
	padding: 0.62rem 1rem;
	color: #fff;
	background: linear-gradient(135deg, #6366f1, #8b5cf6);
	font-size: 0.78rem;
	font-weight: 600;
	cursor: pointer;
	box-shadow: 0 0.45rem 1rem rgb(99 102 241 / 0.25);
	cursor: pointer;
	transition: opacity 0.15s ease, transform 0.15s ease, box-shadow 0.15s ease;
}

.quick-add-submit:hover:not(:disabled) {
	box-shadow: 0 0.65rem 1.25rem rgb(99 102 241 / 0.35);
	transform: translateY(-1px);
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
	.quick-add-panel {
		padding: 1rem;
	}

	.quick-add-header h2 {
		font-size: 1.15rem;
	}

	.quick-add-input {
		font-size: 1rem;
	}
}

@keyframes quick-add-fade {
	from { opacity: 0; }
	to { opacity: 1; }
}

@keyframes quick-add-rise {
	from { opacity: 0; transform: translateY(0.7rem) scale(0.98); }
	to { opacity: 1; transform: translateY(0) scale(1); }
}
</style>
