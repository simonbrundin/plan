<script setup lang="ts">
import type { Place } from "~/types/place";
// Leaflet är ett CommonJS-paket. En statisk import ger rätt interop -
// dynamisk import ger ett namespace utan "default" och kartan renderas inte.
import L from "leaflet";

const props = withDefaults(
	defineProps<{
		/** Platser som ska visas på kartan. */
		places?: Place[];
		/** Platsen som ska markeras (t.ex. den som redigeras). */
		selectedId?: number | null;
		/** Höjd på kartcontainer i pixel. */
		height?: number;
		/** Startkoordinat ([lat, lng]) när det finns inga platser att visa. */
		fallbackCenter?: [number, number];
	}>(),
	{
		places: () => [],
		selectedId: null,
		height: 420,
		fallbackCenter: () => [59.3293, 18.0686],
	},
);

const emit = defineEmits<{
	/** Klick i kartan - används för att skapa en ny plats. */
	"map-click": [latitude: number, longitude: number];
	/** Klick på en befintlig plats. */
	"place-click": [place: Place];
}>();

const mapElement = ref<HTMLDivElement | null>(null);
const map = shallowRef<any>(null);
const leaflet = shallowRef<any>(null);
const markers = new Map<number, any>();

/** Renderar en plats som en pin med en cirkel för radien. */
function renderPlace(place: Place) {
	if (!map.value || !leaflet.value) return;

	const existing = markers.get(place.id);
	if (existing) {
		existing.marker.setLatLng([place.latitude, place.longitude]);
		existing.circle.setLatLng([place.latitude, place.longitude]);
		existing.circle.setRadius(place.radius_meters);
		return;
	}

	const marker = leaflet.value
		.marker([place.latitude, place.longitude], {
			icon: leaflet.value.divIcon({
				className: "place-pin",
				html: '<span class="place-pin-glyph">📍</span>',
				iconSize: [24, 24],
				iconAnchor: [12, 24],
			}),
		})
		.addTo(map.value);

	marker.on("click", () => emit("place-click", place));

	const circle = leaflet.value
		.circle([place.latitude, place.longitude], {
			radius: place.radius_meters,
			color: "#6366f1",
			weight: 1,
			fillColor: "#6366f1",
			fillOpacity: 0.12,
		})
		.addTo(map.value);

	markers.set(place.id, { marker, circle });
}

/** Markerar den valda platsen tydligare. */
function highlightSelected() {
	if (!map.value || !leaflet.value) return;

	for (const [id, entry] of markers) {
		const selected = id === props.selectedId;
		entry.circle.setStyle({
			color: selected ? "#f59e0b" : "#6366f1",
			weight: selected ? 3 : 1,
			fillOpacity: selected ? 0.25 : 0.12,
		});
		if (selected) {
			const place = props.places.find((p) => p.id === id);
			if (place) {
				map.value.setView([place.latitude, place.longitude], 15);
			}
		}
	}
}

/**
 * Initierar kartan när DOM-elementet finns.
 *
 * Elementet binds via template-ref, och refen sätts efter att komponenten
 * mounted (komponenten ligger inuti ClientOnly). Därför initieras kartan
 * från en watcher på refen i stället för i onMounted.
 */
function initMap(element: HTMLDivElement) {
	if (map.value) return;

	const api = L as any;
	leaflet.value = api;

	map.value = api.map(element);

	api.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
		attribution: "© OpenStreetMap",
		maxZoom: 19,
	}).addTo(map.value);

	map.value.on("click", (event: any) => {
		emit("map-click", event.latlng.lat, event.latlng.lng);
	});

	for (const place of props.places) {
		renderPlace(place);
	}

	if (props.places.length === 0) {
		map.value.setView(props.fallbackCenter, 12);
	} else {
		const bounds = api.latLngBounds(
			props.places.map((p) => [p.latitude, p.longitude] as [number, number]),
		);
		map.value.fitBounds(bounds.pad(0.3));
	}

	highlightSelected();
}

watch(
	mapElement,
	(element) => {
		if (element) initMap(element);
	},
	{ immediate: true, flush: "post" },
);

watch(
	() => props.places,
	(places) => {
		for (const place of places) {
			renderPlace(place);
		}
	},
	{ deep: true },
);

watch(() => props.selectedId, highlightSelected);

onBeforeUnmount(() => {
	if (map.value) {
		map.value.remove();
		map.value = null;
	}
});
</script>

<template>
	<div
		ref="mapElement"
		class="w-full rounded overflow-hidden border border-surface-3"
		:style="{ height: `${height}px` }"
	/>
</template>

<style>
.place-pin-glyph {
	font-size: 20px;
	line-height: 24px;
	text-align: center;
	filter: drop-shadow(0 1px 2px rgb(0 0 0 / 0.4));
}
</style>
