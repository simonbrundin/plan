import { fuzzyScore } from "~/utils/fuzzy";

/**
 * Ett objekt i kommandopaletten.
 *
 * `searchText` är det som faktiskt söks i. Det kan innehålla mer än det
 * som visas - en händelses datum, en måls beskrivning - så att sökningen
 * hittar saker på innehåll och inte bara på titlar.
 */
export interface PaletteItem {
	id: string;
	label: string;
	suffix?: string;
	icon?: string;
	to?: string;
	/** Namn på en slot i paletten, så en egen ikon kan renderas. */
	slot?: string;
	/** Eventens ikon (emoji eller Iconify-namn), för egen slot. */
	eventIcon?: string;
	searchText?: string;
	onSelect?: () => void;
}

/** En navigeringslänk som ska gå att söka på. */
export interface PaletteLink {
	label: string;
	icon?: string;
	to?: string;
}

/** Antalet träffar som visas per grupp. */
const MAX_ITEMS_PER_GROUP = 12;

const dateWithTime = new Intl.DateTimeFormat("sv-SE", {
	dateStyle: "medium",
	timeStyle: "short",
});

const dateOnly = new Intl.DateTimeFormat("sv-SE", { dateStyle: "medium" });

/**
 * Kommandopaletten: mål, händelser och navigering.
 *
 * Sökningen är fzf-liknande - egna fuzzy-träffar, inte substring - så
 * att "mt" hittar "Mötet i rummet". Därför filtreras grupperna här, och
 * paletten får dem med `ignoreFilter` så att den inte filtrerar om.
 */
export function useCommandPalette(links: PaletteLink[]) {
	const searchTerm = ref("");

	const goalsStore = useGoalsStore();
	const { events, load: loadEvents } = useCalendarSearch();
	const { focusEvent } = useCalendarFocus();
	const { calendars, load: loadCalendars } = useCalendars();

	/** Namnet på kalendern en händelse ligger i. */
	function calendarName(calendarId?: number): string {
		if (!calendarId) return "";
		return calendars.value.find((calendar) => calendar.id === calendarId)?.name ?? "";
	}

	/** Tiden en händelse visas med i paletten. */
	function eventSuffix(event: CalendarEvent): string {
		const start = new Date(event.start_at);
		const time = event.all_day
			? dateOnly.format(start)
			: dateWithTime.format(start);

		return event.recurrence_rule ? `${time} · Återkommande` : time;
	}

	const goalItems = computed<PaletteItem[]>(() =>
		goalsStore.goals.map((goal) => {
			const children = goalsStore.childStats.get(goal.id)?.total ?? 0;
			const suffix = [
				goal.finished ? "Klar" : undefined,
				children > 0 ? `${children} undermål` : undefined,
			]
				.filter(Boolean)
				.join(" · ");

			return {
				id: `goal-${goal.id}`,
				label: goal.title,
				icon: goal.finished ? "i-lucide-check-circle" : "i-lucide-circle",
				to: `/goal/${goal.id}`,
				suffix: suffix || undefined,
				searchText: [goal.title, goal.description, ...(goal.tags ?? [])]
					.filter(Boolean)
					.join(" "),
			};
		}),
	);

	const eventItems = computed<PaletteItem[]>(() =>
		events.value.map((event) => ({
			id: `event-${event.id}`,
			label: event.title,
			suffix: eventSuffix(event),
			icon: "i-lucide-calendar-clock",
			slot: "event",
			eventIcon: event.icon,
			searchText: [
				event.title,
				event.description,
				calendarName(event.calendar_id),
			]
				.filter(Boolean)
				.join(" "),
			onSelect: () => {
				focusEvent(event);

				// På mobil ligger kalendern på en egen sida, så paletten
				// måste ta användaren dit.
				if (
					typeof window !== "undefined" &&
					window.matchMedia("(min-width: 1024px)").matches === false
				) {
					void navigateTo("/calendar");
				}
			},
		})),
	);

	const linkItems = computed<PaletteItem[]>(() =>
		links.map((link) => ({
			id: `link-${link.to ?? link.label}`,
			label: link.label,
			icon: link.icon,
			to: link.to,
			searchText: link.label,
		})),
	);

	/**
	 * Sorterar och begränsar en grupp utifrån sökordet.
	 *
	 * Träffar som börjar på sökordet vinner över träffar mitt i texten,
	 * och kortare titlar vinner när poängen är lika.
	 */
	function rank(items: PaletteItem[]): PaletteItem[] {
		const query = searchTerm.value.trim();

		if (!query) return items.slice(0, MAX_ITEMS_PER_GROUP);

		return items
			.map((item) => ({
				item,
				score: fuzzyScore(query, item.searchText ?? item.label),
			}))
			.filter((entry) => entry.score !== null)
			.sort(
				(left, right) =>
					(right.score ?? 0) - (left.score ?? 0) ||
					left.item.label.length - right.item.label.length,
			)
			.slice(0, MAX_ITEMS_PER_GROUP)
			.map((entry) => entry.item);
	}

	const groups = computed(() => [
		{
			id: "goals",
			label: "Mål",
			ignoreFilter: true,
			items: rank(goalItems.value),
		},
		{
			id: "events",
			label: "Händelser",
			ignoreFilter: true,
			items: rank(eventItems.value),
		},
		{
			id: "links",
			label: "Navigering",
			ignoreFilter: true,
			items: rank(linkItems.value),
		},
	]);

	/**
	 * Hämtar det som sökningen behöver: kalenderhändelserna och
	 * kalendrarnas namn.
	 */
	async function load(): Promise<void> {
		await Promise.all([loadEvents(), loadCalendars()]);
	}

	return {
		searchTerm,
		groups,
		load,
	};
}
