/**
 * Inställningar för kalendern.
 *
 * Inställningarna sparas i webbläsaren, så de gäller på den enhet de gjordes
 * på. De är avsedda att ändras sällan, och en inställning per enhet är
 * rimligare än att bygga ett konto-underlag för dem.
 */

export interface CalendarSettings {
  /**
   * Hur många förekomster av en återkommande serie som får synas i vyn.
   *
   * Har serien fler, döljs hela serien i tidslinjen.
   */
  maxInstances: number;
  /**
   * Om återkommande serier får döljas när de inte får plats.
   *
   * Med avstängd döljning visas alla serier, och kolumnen kan bli full.
   */
  hideRecurring: boolean;
}

export const DEFAULT_CALENDAR_SETTINGS: CalendarSettings = {
  maxInstances: 7,
  hideRecurring: true,
};

const STORAGE_KEY = "plan:calendar-settings";

/** Läser sparade inställningar. En trasig sparad inställning ignoreras. */
const readStored = (): CalendarSettings | null => {
  if (!import.meta.client) return null;

  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;

    const parsed = JSON.parse(raw) as Partial<CalendarSettings> & {
      // Äldre sparade inställningar hette dailyVisibleLimit.
      dailyVisibleLimit?: number;
    };

    // Den gamla inställningen är samma sak i nytt namn, så den läses in
    // när den finns kvar.
    const rawLimit = parsed.maxInstances ?? parsed.dailyVisibleLimit;
    const limit = Number(rawLimit);

    return {
      maxInstances: Number.isFinite(limit)
        ? Math.min(100, Math.max(1, Math.round(limit)))
        : DEFAULT_CALENDAR_SETTINGS.maxInstances,
      hideRecurring:
        typeof parsed.hideRecurring === "boolean"
          ? parsed.hideRecurring
          : DEFAULT_CALENDAR_SETTINGS.hideRecurring,
    };
  } catch {
    return null;
  }
};

// Skrivningen registreras bara en gång, oavsett hur många gånger
// inställningarna används.
let persistRegistered = false;

export function useCalendarSettings() {
  const settings = useState<CalendarSettings>("calendar-settings", () => ({
    ...DEFAULT_CALENDAR_SETTINGS,
  }));
  const loaded = useState<boolean>("calendar-settings:loaded", () => false);

  if (import.meta.client && !loaded.value) {
    loaded.value = true;
    const stored = readStored();
    if (stored) settings.value = stored;
  }

  if (import.meta.client && !persistRegistered) {
    persistRegistered = true;

    watch(
      settings,
      (value) => {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(value));
      },
      { deep: true },
    );
  }

  const update = (patch: Partial<CalendarSettings>) => {
    settings.value = { ...settings.value, ...patch };
  };

  return { settings, update };
}
