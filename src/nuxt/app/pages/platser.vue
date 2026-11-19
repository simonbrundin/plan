<script setup lang="ts">
import type { Place } from "~/types/place";

definePageMeta({
	ssr: false,
});

const { places, fetchPlaces, createPlace, updatePlace, deletePlace } = usePlaces();

const selectedId = ref<number | null>(null);
const draftCoordinates = ref<{ latitude: number; longitude: number } | null>(null);
const errorMessage = ref("");

const selectedPlace = computed<Place | null>(
	() => places.value.find((p) => p.id === selectedId.value) ?? null,
);

const editing = computed(() => selectedPlace.value !== null || draftCoordinates.value !== null);

onMounted(async () => {
	try {
		await fetchPlaces();
	} catch (error) {
		errorMessage.value = "Kunde inte ladda platserna.";
	}
});

/** Klick i kartan utanför en befintlig plats startar en ny plats. */
function onMapClick(latitude: number, longitude: number) {
	selectedId.value = null;
	draftCoordinates.value = { latitude, longitude };
}

function onPlaceClick(place: Place) {
	draftCoordinates.value = null;
	selectedId.value = place.id;
}

function startNewPlace() {
	selectedId.value = null;
	draftCoordinates.value = {
		latitude: places.value[0]?.latitude ?? 59.3293,
		longitude: places.value[0]?.longitude ?? 18.0686,
	};
}

function cancelEdit() {
	selectedId.value = null;
	draftCoordinates.value = null;
}

async function savePlace(payload: {
	name: string;
	latitude: number;
	longitude: number;
	radius_meters: number;
}) {
	errorMessage.value = "";

	if (!payload.name) {
		errorMessage.value = "Platsen behöver ett namn.";
		return;
	}

	try {
		if (selectedPlace.value) {
			await updatePlace(selectedPlace.value.id, payload);
			selectedId.value = selectedPlace.value.id;
		} else {
			const created = await createPlace(payload);
			selectedId.value = created.id;
		}
		draftCoordinates.value = null;
	} catch (error: any) {
		errorMessage.value = error?.data?.message ?? "Kunde inte spara platsen.";
	}
}

async function removePlace() {
	if (!selectedPlace.value) return;

	try {
		await deletePlace(selectedPlace.value.id);
		cancelEdit();
	} catch (error: any) {
		errorMessage.value = error?.data?.message ?? "Kunde inte ta bort platsen.";
	}
}
</script>

<template>
	<div class="p-4 max-w-6xl mx-auto">
		<div class="flex items-center justify-between mb-4">
			<h1 class="text-2xl font-bold text-ink">Platser</h1>
			<button
				class="px-3 py-1 text-sm rounded bg-accent text-white hover:opacity-90"
				@click="startNewPlace"
			>
				Lägg till plats
			</button>
		</div>

		<p v-if="errorMessage" class="text-sm text-red-400 mb-3">
			{{ errorMessage }}
		</p>

		<div class="grid gap-4 md:grid-cols-[minmax(0,1fr)_320px]">
			<ClientOnly>
				<PlaceMap
					:places="places"
					:selected-id="selectedId"
					@map-click="onMapClick"
					@place-click="onPlaceClick"
				/>
				<template #fallback>
					<div class="h-420 rounded border border-surface-3 flex items-center justify-center text-ink-dim">
						Laddar kartan…
					</div>
				</template>
			</ClientOnly>

			<div class="space-y-4">
				<div
					v-if="editing"
					class="bg-surface-2 border border-accent rounded p-3"
				>
					<h2 class="text-sm font-semibold text-ink mb-2">
						{{ selectedPlace ? "Redigera plats" : "Ny plats" }}
					</h2>
					<PlaceEditor
						:place="selectedPlace"
						:latitude="draftCoordinates?.latitude"
						:longitude="draftCoordinates?.longitude"
						@save="savePlace"
						@delete="removePlace"
						@cancel="cancelEdit"
					/>
				</div>

				<div class="bg-surface-2 border border-surface-3 rounded p-3">
					<h2 class="text-sm font-semibold text-ink mb-2">
						Mina platser
					</h2>
					<p v-if="places.length === 0" class="text-sm text-ink-dim">
						Inga platser ännu. Klicka i kartan för att skapa en.
					</p>
					<ul v-else class="space-y-1">
						<li v-for="place in places" :key="place.id">
							<button
								class="w-full text-left px-2 py-1 rounded hover:bg-surface-3"
								:class="{ 'bg-surface-3': place.id === selectedId }"
								@click="onPlaceClick(place)"
							>
								<span class="text-sm text-ink">{{ place.name }}</span>
								<span class="text-xs text-ink-dim block">
									{{ place.latitude.toFixed(5) }},
									{{ place.longitude.toFixed(5) }} · {{ place.radius_meters }} m
								</span>
							</button>
						</li>
					</ul>
				</div>
			</div>
		</div>
	</div>
</template>
