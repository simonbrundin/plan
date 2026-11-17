/**
 * Globalt tillstånd för snabbt läggande till.
 *
 * Fältet ska kunna öppnas från valfri sida med tangenten och fokus ska
 * ligga i textfältet direkt, som i Fantastical.
 */

export type QuickAddMode = "auto" | "event" | "goal";

export function useQuickAdd() {
	const isOpen = useState("quick-add:open", () => false);
	const mode = useState<QuickAddMode>("quick-add:mode", () => "auto");
	// Ökas varje gång något skapats, så att sidor som visar innehåll
	// kan ladda om det utan att vara kopplade till komponenten.
	const revision = useState("quick-add:revision", () => 0);

	const open = (nextMode?: QuickAddMode) => {
		if (nextMode) mode.value = nextMode;
		isOpen.value = true;
	};

	const close = () => {
		isOpen.value = false;
	};

	const toggle = (nextMode?: QuickAddMode) => {
		if (isOpen.value) {
			close();
			return;
		}
		open(nextMode);
	};

	const markCreated = () => {
		revision.value += 1;
	};

	return { isOpen, mode, revision, open, close, toggle, markCreated };
}
