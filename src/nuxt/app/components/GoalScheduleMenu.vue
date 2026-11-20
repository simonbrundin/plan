<script setup lang="ts">
import type { DurationOption } from "~/utils/calendarScheduling";

/**
 * Meny som visas när ett mål släpps i kalendern.
 *
 * Menyn innehåller de tidsalternativ som kan väljas, och en egen
 * inmatning för ett värde som inte finns i listan. Den ligger kvar tills
 * användaren väljer en tid eller avbryter.
 */
const props = defineProps<{
	/** Målets titel, som visas i menyn. */
	title: string;
	/** Tiden som händelsen ska starta, i millisekunder. */
	startAt: number;
	/** Tidsalternativen som kan väljas. */
	options: DurationOption[];
}>();

const emit = defineEmits<{
	/** Användaren har valt en tid, i minuter. */
	select: [minutes: number];
	/** Användaren avbryter. */
	cancel: [];
}>();

const customDraft = ref("");

const timeLabel = computed(() =>
	new Intl.DateTimeFormat("sv-SE", {
		weekday: "short",
		hour: "2-digit",
		minute: "2-digit",
	}).format(props.startAt),
);

// Den egna inmatningen godkänns först när den är ett heltal över noll.
const customValid = computed(() => {
	const value = Number(customDraft.value.trim());
	return Number.isFinite(value) && value > 0;
});

function chooseCustom() {
	if (!customValid.value) return;

	emit("select", Math.round(Number(customDraft.value.trim())));
}
</script>

<template>
	<div
		class="schedule-menu"
		role="dialog"
		aria-label="Välj tid för målet"
		@click.stop
		@keydown.esc.stop.prevent="emit('cancel')"
	>
		<div class="schedule-menu-header">
			<strong>{{ title }}</strong>
			<span>{{ timeLabel }}</span>
		</div>

		<ul class="schedule-menu-options">
			<li v-for="option in options" :key="option.minutes">
				<button
					type="button"
					class="schedule-option"
					:class="{ suggested: option.suggested }"
					@click="emit('select', option.minutes)"
				>
					{{ option.label }}
					<small v-if="option.suggested">uppskattning</small>
				</button>
			</li>
		</ul>

		<div class="schedule-menu-custom">
			<input
				v-model="customDraft"
				type="number"
				min="1"
				step="5"
				placeholder="Minuter"
				aria-label="Egen tid i minuter"
				@keydown.enter.prevent="chooseCustom"
				@keydown.esc.stop.prevent="emit('cancel')"
			/>
			<button type="button" :disabled="!customValid" @click="chooseCustom">
				Sätt
			</button>
		</div>

		<button type="button" class="schedule-cancel" @click="emit('cancel')">
			Avbryt
		</button>
	</div>
</template>

<style scoped>
.schedule-menu {
	position: absolute;
	z-index: 100000;
	display: flex;
	flex-direction: column;
	gap: 0.4rem;
	min-width: 12rem;
	padding: 0.55rem 0.6rem;
	border: 1px solid var(--color-line-strong);
	border-radius: 0.6rem;
	background: var(--color-surface);
	box-shadow: 0 12px 32px rgb(0 0 0 / 0.45);
	color: var(--color-ink);
	font-size: 0.75rem;
}

.schedule-menu-header {
	display: flex;
	flex-direction: column;
	gap: 0.1rem;
}

.schedule-menu-header strong {
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.schedule-menu-header span {
	color: var(--color-ink-muted);
	font-size: 0.7rem;
	font-variant-numeric: tabular-nums;
}

.schedule-menu-options {
	display: flex;
	flex-direction: column;
	gap: 0.15rem;
	margin: 0;
	padding: 0;
	list-style: none;
}

.schedule-option {
	display: flex;
	align-items: baseline;
	justify-content: space-between;
	gap: 0.5rem;
	width: 100%;
	border: none;
	border-radius: 0.4rem;
	padding: 0.3rem 0.45rem;
	background: none;
	color: var(--color-ink);
	font-size: 0.75rem;
	text-align: left;
	cursor: pointer;
}

.schedule-option:hover {
	background: var(--color-surface-2);
	color: var(--color-accent);
}

.schedule-option.suggested {
	color: var(--color-accent);
	font-weight: 600;
}

.schedule-option small {
	color: var(--color-ink-dim);
	font-size: 0.62rem;
	font-weight: 400;
}

.schedule-menu-custom {
	display: flex;
	gap: 0.3rem;
}

.schedule-menu-custom input {
	min-width: 0;
	flex: 1 1 auto;
	border: 1px solid var(--color-line-strong);
	border-radius: 0.4rem;
	padding: 0.25rem 0.4rem;
	background: var(--color-surface-2);
	color: var(--color-ink);
	font-size: 0.72rem;
}

.schedule-menu-custom button,
.schedule-cancel {
	border: 1px solid var(--color-line-strong);
	border-radius: 0.4rem;
	padding: 0.25rem 0.5rem;
	background: var(--color-surface-2);
	color: var(--color-ink-muted);
	font-size: 0.7rem;
	cursor: pointer;
}

.schedule-menu-custom button:hover:not(:disabled),
.schedule-cancel:hover {
	border-color: var(--color-accent);
	color: var(--color-ink);
}

.schedule-menu-custom button:disabled {
	opacity: 0.4;
	cursor: not-allowed;
}
</style>
