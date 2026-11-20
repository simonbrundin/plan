import { computed, readonly, ref } from "vue";
// Relativ import - composable ska kunna testas utan Nuxt's alias.
import {
	popUndoEntry,
	pushUndoEntry,
	type UndoEntry,
} from "../utils/undoStack";

/**
 * Global undo-stapel.
 *
 * Stapeln ligger utanför komponenterna så att en handling på en sida kan
 * ångras från en annan sida - tangenten "z" är global.
 */

// Appen körs utan SSR, så modulnivå är säkert här.
const entries = ref<UndoEntry[]>([]);
// Ökas efter varje ångrande. Sidor som visar egna listor kan lyssna på
// den för att hämta om data när något har ångras.
const revision = ref(0);
const isUndoing = ref(false);

export function useUndoStack() {
	const canUndo = computed(
		() => entries.value.length > 0 && !isUndoing.value,
	);

	// Beskrivningen av den handling som "z" skulle ångra.
	const nextLabel = computed(
		() => entries.value[entries.value.length - 1]?.label ?? null,
	);

	/** Registrerar en handling som kan ångras. */
	function push(entry: UndoEntry) {
		entries.value = pushUndoEntry(entries.value, entry);
	}

	/**
	 * Ångrar den senaste handlingen.
	 * Returnerar beskrivningen av det som ångrades, eller null om det
	 * inte fanns något att ångra.
	 */
	async function undoLast(): Promise<string | null> {
		if (isUndoing.value) return null;

		const { entries: rest, entry } = popUndoEntry(entries.value);
		if (!entry) return null;

		isUndoing.value = true;
		try {
			await entry.undo();
		} catch (error) {
			// Ångras inte - posten läggs tillbaka så att man kan försöka igen.
			entries.value = pushUndoEntry(entries.value, entry);
			throw error;
		} finally {
			isUndoing.value = false;
		}

		entries.value = rest;
		revision.value += 1;
		return entry.label;
	}

	/** Tömmer stapeln - t.ex. när man byter session. */
	function clear() {
		entries.value = [];
	}

	return {
		entries: readonly(entries),
		canUndo,
		nextLabel,
		revision: readonly(revision),
		isUndoing: readonly(isUndoing),
		push,
		undoLast,
		clear,
	};
}
