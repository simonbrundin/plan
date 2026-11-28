/**
 * Toggle för "Deadline först" i "Gör nu"-listan.
 *
 * Inställningen sparas i webbläsaren, så den gäller på den enhet den gjordes
 * på. Den är av som standard - den manuella ordningen är listans huvudregel,
 * och deadline-först är ett val för den som vill se frister överst.
 */

const STORAGE_KEY = "plan:now-due-first";

// Skrivningen registreras bara en gång, oavsett hur många gånger
// inställningen används.
let persistRegistered = false;

export function useDueFirstToggle() {
  const dueFirst = useState<boolean>("now-due-first", () => false);
  const loaded = useState<boolean>("now-due-first:loaded", () => false);

  if (import.meta.client && !loaded.value) {
    loaded.value = true;
    dueFirst.value = localStorage.getItem(STORAGE_KEY) === "1";
  }

  if (import.meta.client && !persistRegistered) {
    persistRegistered = true;

    watch(dueFirst, (value) => {
      localStorage.setItem(STORAGE_KEY, value ? "1" : "0");
    });
  }

  function toggleDueFirst() {
    dueFirst.value = !dueFirst.value;
  }

  function setDueFirst(value: boolean) {
    dueFirst.value = value;
  }

  return {
    dueFirst: readonly(dueFirst),
    toggleDueFirst,
    setDueFirst,
  };
}
