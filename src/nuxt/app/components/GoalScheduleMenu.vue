<script setup lang="ts">
import { parseEstimateMinutes } from "~/utils/estimate";

/**
 * Meny som visas när ett mål släpps i kalendern.
 *
 * Menyn innehåller en enda inmatning för tidsåtgång. Den ligger kvar tills
 * användaren skriver ett värde och trycker Enter, eller avbryter.
 */
const emit = defineEmits<{
	/** Användaren har skrivit en tid och bekräftat den, i minuter. */
	select: [minutes: number];
	/** Användaren avbryter. */
	cancel: [];
}>();

const customDraft = ref("");
const inputRef = ref<HTMLInputElement | null>(null);

onMounted(() => {
	// Menyn skapas först efter drop-eventet. Fokusera därför uttryckligen
	// efter mount i stället för att förlita oss enbart på autofocus-attributet.
	nextTick(() => inputRef.value?.focus());
});

const customMinutes = computed(() => parseEstimateMinutes(customDraft.value));
const customValid = computed(() => customMinutes.value !== null && customMinutes.value > 0);

function chooseCustom() {
	if (customMinutes.value === null || customMinutes.value <= 0) return;

	emit("select", customMinutes.value);
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
		<input
			ref="inputRef"
			v-model="customDraft"
			type="text"
			inputmode="text"
			autofocus
			placeholder="t.ex. 1h eller 26m"
			aria-label="Tid för målet"
			@keydown.enter.prevent="chooseCustom"
			@keydown.esc.stop.prevent="emit('cancel')"
		/>
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

.schedule-menu input {
	width: 100%;
	border: 1px solid var(--color-line-strong);
	border-radius: 0.4rem;
	padding: 0.4rem 0.5rem;
	background: var(--color-surface-2);
	color: var(--color-ink);
	font-size: 0.8rem;
}

.schedule-menu input:focus {
	border-color: var(--color-accent);
	outline: none;
	box-shadow: 0 0 0 2px rgb(99 102 241 / 0.2);
}
</style>
