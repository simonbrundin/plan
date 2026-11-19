<script setup lang="ts">
import type { Place } from "~/types/place";

const props = defineProps<{
	/** Platsen som redigeras. Null = ny plats. */
	place?: Place | null;
	/** Koordinat för en ny plats (när den skapats genom klick i kartan). */
	latitude?: number;
	longitude?: number;
}>();

const emit = defineEmits<{
	save: [payload: { name: string; latitude: number; longitude: number; radius_meters: number }];
	delete: [];
	cancel: [];
}>();

const name = ref(props.place?.name ?? "");
const radius = ref(props.place?.radius_meters ?? 100);
const latitude = ref(props.place?.latitude ?? props.latitude ?? 59.3293);
const longitude = ref(props.place?.longitude ?? props.longitude ?? 18.0686);

watch(
	() => props.place,
	(place) => {
		if (!place) return;
		name.value = place.name;
		radius.value = place.radius_meters;
		latitude.value = place.latitude;
		longitude.value = place.longitude;
	},
);

watch(
	() => [props.latitude, props.longitude] as const,
	([lat, lon]) => {
		if (lat === undefined || lon === undefined) return;
		latitude.value = lat;
		longitude.value = lon;
	},
);

function submit() {
	emit("save", {
		name: name.value.trim(),
		latitude: latitude.value,
		longitude: longitude.value,
		radius_meters: radius.value,
	});
}
</script>

<template>
	<form class="space-y-3" @submit.prevent="submit">
		<div>
			<label class="block text-xs text-ink-dim mb-1">Namn</label>
			<input
				v-model="name"
				type="text"
				required
				placeholder="Hem"
				class="w-full px-3 py-2 bg-surface-2 border border-accent rounded text-ink focus:outline-none focus:ring-2 focus:ring-accent"
			/>
		</div>

		<div class="grid grid-cols-2 gap-2">
			<div>
				<label class="block text-xs text-ink-dim mb-1">Latitud</label>
				<input
					v-model.number="latitude"
					type="number"
					step="0.000001"
					required
					class="w-full px-3 py-2 bg-surface-2 border border-accent rounded text-ink focus:outline-none focus:ring-2 focus:ring-accent"
				/>
			</div>
			<div>
				<label class="block text-xs text-ink-dim mb-1">Longitud</label>
				<input
					v-model.number="longitude"
					type="number"
					step="0.000001"
					required
					class="w-full px-3 py-2 bg-surface-2 border border-accent rounded text-ink focus:outline-none focus:ring-2 focus:ring-accent"
				/>
			</div>
		</div>

		<div>
			<label class="block text-xs text-ink-dim mb-1">
				Radie: {{ radius }} meter
			</label>
			<input
				v-model.number="radius"
				type="range"
				min="10"
				max="1000"
				step="10"
				class="w-full accent-accent"
			/>
			<input
				v-model.number="radius"
				type="number"
				min="1"
				max="100000"
				step="1"
				class="w-full px-3 py-2 mt-1 bg-surface-2 border border-accent rounded text-ink focus:outline-none focus:ring-2 focus:ring-accent"
			/>
		</div>

		<div class="flex gap-2">
			<button
				type="submit"
				class="px-3 py-1 text-sm rounded bg-accent text-white hover:opacity-90"
			>
				{{ place ? "Spara" : "Skapa plats" }}
			</button>
			<button
				type="button"
				class="px-3 py-1 text-sm rounded text-ink-dim hover:text-ink"
				@click="emit('cancel')"
			>
				Avbryt
			</button>
			<button
				v-if="place"
				type="button"
				class="px-3 py-1 text-sm rounded text-red-400 hover:text-red-300 ml-auto"
				@click="emit('delete')"
			>
				Ta bort
			</button>
		</div>
	</form>
</template>
