<script setup lang="ts">
import type { Goal } from "~/types/goal";
import { formatEstimate, parseEstimateMinutes } from "~/utils/estimate";

const props = defineProps<{
	/** Målet som uppskattningen gäller. */
	goal: Goal;
	/** Sant när fältet är öppet för just detta mål. */
	editorOpen: boolean;
	/** Storlek på ikonen och texten. */
	size?: "sm" | "md";
}>();

const emit = defineEmits<{
	/** Användaren vill öppna fältet. */
	open: [];
	/** Användaren sparar ett värde (i minuter), eller null för att rensa. */
	save: [minutes: number | null];
	/** Användaren avbryter. */
	cancel: [];
}>();

const draft = ref("");
const inputRef = ref<HTMLInputElement | null>(null);

// När fältet öppnas fylls det med målets nuvarande uppskattning.
watch(
	() => props.editorOpen,
	(isOpen) => {
		if (!isOpen) return;
		draft.value = props.goal.estimated_minutes
			? String(props.goal.estimated_minutes)
			: "";
		nextTick(() => inputRef.value?.focus());
	},
);

function save() {
	const text = draft.value.trim();
	if (!text) {
		emit("save", null);
		return;
	}
	const minutes = parseEstimateMinutes(text);
	if (minutes === null) {
		// Ogiltigt värde - låt fältet stå kvar så att det kan rättas.
		return;
	}
	emit("save", minutes === 0 ? null : minutes);
}

function clear() {
	emit("save", null);
}

const iconClass = computed(() => (props.size === "sm" ? "w-4 h-4" : "w-6 h-6"));
const textClass = computed(() => (props.size === "sm" ? "text-xs" : "text-sm"));
</script>

<template>
	<div class="flex items-center gap-1.5 flex-shrink-0" @click.stop>
		<template v-if="editorOpen">
			<input
				ref="inputRef"
				:data-estimate-input="goal.id"
				v-model="draft"
				type="text"
				placeholder="1h30"
				class="rounded-lg border border-line bg-surface-2 px-2 py-1 text-ink text-sm focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
				@keydown.enter.prevent="save"
				@keydown.esc.prevent="emit('cancel')" />
			<button
				@click.stop="clear"
				class="rounded-lg p-1 text-ink-dim transition-colors hover:bg-elevated hover:text-ink"
				title="Rensa tidsåtgång">
				<Icon name="lucide:x" class="w-4 h-4" />
			</button>
		</template>
		<template v-else>
			<button
				@click.stop="emit('open')"
				class="rounded-lg p-1.5 transition-colors hover:bg-elevated"
				:class="goal.estimated_minutes ? 'text-accent' : 'text-ink-dim hover:text-ink-muted'"
				:title="goal.estimated_minutes ? `Tidsåtgång: ${formatEstimate(goal.estimated_minutes)}` : 'Sätt tidsåtgång (m)'">
				<Icon name="lucide:timer" :class="iconClass"
					:style="{ opacity: goal.estimated_minutes ? 1 : 0.25 }" />
			</button>
			<span v-if="goal.estimated_minutes" class="tabular-nums text-accent" :class="textClass">
				{{ formatEstimate(goal.estimated_minutes) }}
			</span>
		</template>
	</div>
</template>
