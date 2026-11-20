import { matchLeaderKey } from "~/utils/leaderKeys";

export interface LeaderCommand {
	/** Tangentsekvens efter leader, t.ex. "gp" */
	keys: string;
	/** Visas i modalen, t.ex. "G+P" */
	label: string;
	/** Beskrivning av vad kommandot gör */
	title: string;
	run: () => void;
}

// Leader-tangenten och kommandona är globala - de ska fungera på alla sidor.
const isLeaderOpen = ref(false);
const pendingKeys = ref("");
const pageCommandGroups = ref<Record<string, LeaderCommand[]>>({});
let registrationCounter = 0;

// Händelser som leader-läget redan hanterat. Sidor som lyssnar på
// tangenttryck ska kunna hoppa över dem, så att t.ex. "i" inte både
// byter ikon och triggar något annat.
const consumedEvents = new WeakSet<Event>();

export function useLeaderKey() {
	const router = useRouter();
	const { toggleSidebar } = useSidebarState();

	const globalCommands: LeaderCommand[] = [
		{
			keys: "gi",
			label: "G+I",
			title: "Gå till Inkorgen",
			run: () => router.push("/"),
		},
		{
			keys: "gg",
			label: "G+G",
			title: "Gå till Mål",
			run: () => router.push("/goals"),
		},
		{
			keys: "gp",
			label: "G+P",
			title: "Gå till Prioritet",
			run: () => router.push("/priority"),
		},
		{
			keys: "gc",
			label: "G+C",
			title: "Gå till Kalendern",
			run: () => router.push("/calendar"),
		},
		{
			keys: "gd",
			label: "G+D",
			title: "Gå till Avslutade mål",
			run: () => router.push("/completed"),
		},
		{
			keys: "gs",
			label: "G+S",
			title: "Gå till Inställningar",
			run: () => router.push("/settings"),
		},
		{
			keys: "m",
			label: "G+M",
			title: "Visa eller dölj menyn",
			run: toggleSidebar,
		},
	];

	const allCommands = computed<LeaderCommand[]>(() => [
		...globalCommands,
		...Object.values(pageCommandGroups.value).flat(),
	]);

	const commandMap = computed<Record<string, LeaderCommand>>(() =>
		Object.fromEntries(allCommands.value.map((command) => [command.keys, command])),
	);

	function openLeader() {
		isLeaderOpen.value = true;
		pendingKeys.value = "";
	}

	function closeLeader() {
		isLeaderOpen.value = false;
		pendingKeys.value = "";
	}

	/**
	 * Kommandon som bara finns på en viss sida (t.ex. målsidan).
	 * Returnerar en funktion som avregistrerar dem igen.
	 */
	function registerPageCommands(commands: LeaderCommand[]): () => void {
		const id = `page-${++registrationCounter}`;
		pageCommandGroups.value = { ...pageCommandGroups.value, [id]: commands };
		return () => {
			const groups = { ...pageCommandGroups.value };
			delete groups[id];
			pageCommandGroups.value = groups;
		};
	}

	/** Kör ett kommando direkt, t.ex. från en knapp i modalen. */
	function executeCommand(keys: string): boolean {
		const command = commandMap.value[keys];
		if (!command) return false;
		closeLeader();
		command.run();
		return true;
	}

	function handleLeaderKey(key: string) {
		const action = matchLeaderKey(
			allCommands.value.map((command) => command.keys),
			pendingKeys.value,
			key,
		);

		if (action.type === "run") {
			closeLeader();
			commandMap.value[action.keys].run();
			return;
		}

		if (action.type === "pending") {
			// Kommandot är inte komplett än - vänta på nästa tangent.
			pendingKeys.value = action.keys;
			return;
		}

		closeLeader();
	}

	/**
	 * Hanterar tangenttryck för leader-läget.
	 * Returnerar true om händelsen är konsumerad av leader-läget.
	 */
	function handleKeydown(event: KeyboardEvent): boolean {
		const target = event.target as HTMLElement | null;
		const isFormField =
			!!target &&
			(["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName) ||
				target.isContentEditable);
		if (isFormField) return false;

		const { isSearchOpen } = useSearchState();
		if (isSearchOpen.value) return false;

		if (!isLeaderOpen.value) {
			// Mellanslag startar leader-läget. En fokuserad knapp ska
			// aktiveras med mellanslag istället.
			if (event.key === " " && target?.tagName !== "BUTTON") {
				event.preventDefault();
				// Stoppa övriga tangentbevakare från att reagera på
				// leader-tangenten.
				event.stopPropagation();
				openLeader();
				consumedEvents.add(event);
				return true;
			}
			return false;
		}

		consumedEvents.add(event);
		event.stopPropagation();

		if (event.key === "Escape") {
			event.preventDefault();
			closeLeader();
			return true;
		}

		// Modifiers (Shift, Control, ...) och annan specialtangent
		// avbryter inte sekvensen.
		if (event.key.length !== 1) return true;

		event.preventDefault();
		handleLeaderKey(event.key.toLowerCase());
		return true;
	}

	return {
		isLeaderOpen: readonly(isLeaderOpen),
		pendingKeys: readonly(pendingKeys),
		commands: allCommands,
		openLeader,
		closeLeader,
		executeCommand,
		registerPageCommands,
		handleKeydown,
	};
}

/** Sant om händelsen redan hanterats av leader-läget. */
export function isLeaderConsumed(event: Event): boolean {
	return consumedEvents.has(event);
}
