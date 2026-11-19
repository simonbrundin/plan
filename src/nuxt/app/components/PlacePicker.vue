<script setup lang="ts">
import type { Place } from "~/types/place";

const props = defineProps<{
	/** Platser som redan är kopplade. */
	selected: Place[];
	/** Alla platser som går att välja. */
	available: Place[];
	/** Spara-knappens text. */
	saveLabel?: string;
}>();

const emit = defineEmits<{
	/** Användaren har ändrat urvalet. */
	"update:selected": [places: Place[]];
	/** Användaren vill spara. */
	save: [];
	/** Användaren avbryter. */
	cancel: [];
}>();

const draft = ref<Place[]>([...props.selected]);

watch(
	() => props.selected,
	(places) => {
		draft.value = [...places];
	},
);

function isSelected(placeId: number): boolean {
	return draft.value.some((p) => p.id === placeId);
}

function toggle(place: Place) {
	if (isSelected(place.id)) {
		draft.value = draft.value.filter((p) => p.id !== place.id);
	} else {
		draft.value = [...draft.value, place];
	}
	emit("update:selected", draft.value);
}
</script>

<template>
	<div
		class="bg-surface-2 border border-accent rounded p-3 shadow-lg min-w-64 max-h-96 overflow-y-auto"
	>
		<div v-if="available.length === 0" class="text-sm text-ink-dim py-2">
			Inga platser ännu. Skapa en plats på sidan Platser.
		</div>

		<ul v-else class="space-y-1">
			<li v-for="place in available" :key="place.id">
				<label
					class="flex items-center gap-2 px-2 py-1 rounded cursor-pointer hover:bg-surface-3"
				>
					<input
						type="checkbox"
						:checked="isSelected(place.id)"
						@change="toggle(place)"
						class="accent-accent"
					/>
					<span class="text-sm">{{ place.name }}</span>
					<span class="text-xs text-ink-dim ml-auto">
						{{ place.radius_meters }} m
					</span>
				</label>
			</li>
		</ul>

		<div class="flex gap-2 mt-3">
			<button
				class="px-3 py-1 text-sm rounded bg-accent text-white hover:opacity-90"
				@click="emit('save')"
			>
				{{ saveLabel ?? "Spara" }}
			</button>
			<button
				class="px-3 py-1 text-sm rounded text-ink-dim hover:text-ink"
				@click="emit('cancel')"
			>
				Avbryt
			</button>
		</div>
	</div>
</template>
