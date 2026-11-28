<script setup lang="ts">
import type { NavigationMenuItem } from "@nuxt/ui";

const route = useRoute();
const { loggedIn } = useUserSession();

/**
 * Skalet runt sidinnehållet. Kalendern till höger visas bara om den
 * efterfrågas - dokumentationslayouten lämnar ut den platsen så att
 * texterna får hela sidbredden.
 */
const props = withDefaults(defineProps<{ calendar?: boolean }>(), {
  calendar: true,
});

const { isSidebarOpen: open } = useSidebarState();
const showHelpModal = ref(false);
const showPriorityIndicator = ref(false);
const { isSearchOpen, openSearch, closeSearch } = useSearchState();


const {
  isPriorityMode,
  selectedGoalId,
  prioritizedGoals,
  loadPrioritizedGoals,
  handleKeydown: handlePriorityKeydown,
  handleKeyup: handlePriorityKeyup,
} = usePriorityMode();

const {
  handleKeydown: handlePriorityToggleKeydown,
} = usePriorityToggle();

const {
  enableNavigation,
  disableNavigation,
  handleKeydown: handleNavigationKeydown,
} = useGlobalNavigation();

const {
  isLeaderOpen,
  pendingKeys: leaderPendingKeys,
  commands: leaderCommands,
  handleKeydown: handleLeaderKeydownBase,
  executeCommand: executeLeaderCommand,
} = useLeaderKey();

// Popupen som sätter startdatum och due-datum (leader + d).
const goalDates = useGoalDatesModal();

// Taggväljaren (leader + t). Den renderas här så att den kan öppnas
// från alla sidor där ett mål är markerat.
const tagPicker = useGoalTagPicker();

// Leader-listenern körs i capture-fasen och kan annars konsumera Escape
// innan datum-popupens egen listener hinner få tangenten.
function handleLeaderKeydown(event: KeyboardEvent): boolean {
  if (goalDates.isOpen.value && event.key === "Escape") {
    event.preventDefault();
    event.stopPropagation();
    goalDates.close();
    return true;
  }
  if (isQuickAddOpen.value && event.key === "Escape") {
    event.preventDefault();
    event.stopPropagation();
    closeQuickAdd();
    return true;
  }
  return handleLeaderKeydownBase(event);
}

const { isOpen: isQuickAddOpen, open: openQuickAdd, close: closeQuickAdd } = useQuickAdd();

const { undoLast } = useUndoStack();
const toast = useToast();

// Ångra den senaste handlingen. Den kan påverka alla sidor, så den
// hanteras här och inte i sidornas egna tangentbevakare.
async function undoLastAction() {
  try {
    const label = await undoLast();
    if (!label) {
      toast.add({
        title: "Inget att ångra",
        description: "Det finns ingen ångrbar handling.",
        color: "neutral",
        icon: "i-lucide-undo-2",
      });
      return;
    }

    toast.add({
      title: "Ångrade",
      description: label,
      color: "success",
      icon: "i-lucide-undo-2",
    });
  } catch (error) {
    console.error("Failed to undo last action:", error);
    toast.add({
      title: "Kunde inte ångra",
      description: "Handlingen kunde inte återställas.",
      color: "error",
      icon: "i-lucide-undo-2",
    });
  }
}

watch(isPriorityMode, (val) => {
  showPriorityIndicator.value = val;
});

const links = [
  [
    {
      label: "Inbox",
      icon: "i-lucide-inbox",
      to: "/",
      onSelect: () => {
        open.value = false;
      },
    },
    {
      label: "Goals",
      icon: "i-lucide-target",
      to: "/goals",
      onSelect: () => {
        open.value = false;
      },
    },
    {
      label: "Gör nu",
      icon: "i-lucide-sort-desc",
      to: "/now",
      onSelect: () => {
        open.value = false;
      },
    },
    {
      label: "Kalender",
      icon: "i-lucide-calendar-days",
      to: "/calendar",
      onSelect: () => {
        open.value = false;
      },
    },
    {
      label: "Avslutade",
      icon: "i-lucide-check-circle-2",
      to: "/completed",
      onSelect: () => {
        open.value = false;
      },
    },
    {
      label: "Platser",
      icon: "i-lucide-map-pin",
      to: "/platser",
      onSelect: () => {
        open.value = false;
      },
    },
    {
      label: "Dokumentation",
      icon: "i-lucide-book-open",
      to: "/dokumentation",
      onSelect: () => {
        open.value = false;
      },
    },
    {
      label: "Inställningar",
      icon: "i-lucide-settings",
      to: "/settings",
      onSelect: () => {
        open.value = false;
      },
    },
    {
      label: "Login",
      icon: "i-lucide-log-in",
      to: "/login",
      onSelect: () => {
        open.value = false;
      },
    },
  ],
] satisfies NavigationMenuItem[][];

const goalsStore = useGoalsStore();

// Sökpaletten söker bland mål, händelser och navigering. Länkarna är
// samma som i sidomenyn, men utan deras egen logik - paletten navigerar
// bara.
const paletteLinks = links.flat().map((link) => ({
	label: String(link.label),
	icon: typeof link.icon === "string" ? link.icon : undefined,
	to: typeof link.to === "string" ? link.to : undefined,
}));

const { searchTerm, groups, load: loadSearchData } = useCommandPalette(paletteLinks);
const { clearFocus } = useCalendarFocus();

// Händelserna hämtas när paletten öppnas första gången. Sökningen ska
// inte ladda data i onödan, men den ska hitta händelser direkt när
// användaren söker.
watch(isSearchOpen, (open) => {
	if (!open) {
		// Sökningen ska inte ligga kvar nästa gång paletten öppnas.
		searchTerm.value = "";
		return;
	}

	// En gammal fokusbegäran ska inte slå till när paletten öppnas igen.
	clearFocus();

	void loadSearchData();
});

onMounted(() => {
  // Grafen laddas direkt vid start, så att byten mellan målsidor inte
  // behöver vänta på servern. Ett fel ignoreras - menyn fungerar ändå.
  goalsStore.loadGraph().catch(() => {});
});

function handleGlobalKeydown(event: KeyboardEvent) {
  // Leader-läget hanteras i en capture-listener (se onMounted) så att det
  // hinner före sidornas egna tangentbevakare.
  if (isLeaderConsumed(event)) return;

  const target = event.target as HTMLElement;
  const isInInput =
    target.tagName === "INPUT" ||
    target.tagName === "TEXTAREA" ||
    target.isContentEditable;

  // Snabbt läggande till öppnas med "n", som i Fantastical. Fältet får
  // fokus direkt så att man kan börja skriva.
  if (
    event.key === "n" &&
    !event.ctrlKey &&
    !event.metaKey &&
    !event.altKey &&
    !event.defaultPrevented &&
    !isInInput &&
    !isSearchOpen.value &&
    !isLeaderOpen.value &&
    !isQuickAddOpen.value
  ) {
    event.preventDefault();
    openQuickAdd();
    return;
  }

  // Escape stänger snabbfältet när det är öppet. Det hanteras här och inte
  // bara i textfältet, så att Escape även fungerar när fokus ligger på
  // en knapp i overlayen.
  if (event.key === "Escape" && isQuickAddOpen.value) {
    event.preventDefault();
    closeQuickAdd();
    return;
  }

  // Handle Escape FIRST - always let it propagate for Vim
  if (event.key === "Escape") {
    if (isSearchOpen.value) {
      closeSearch();
    }
    // Don't return - let Escape propagate for Vim insert mode exit
  }

  if (isSearchOpen.value) {
    // When search is open, block navigation keys but let Escape and Enter through
    const navKeys = ["h", "j", "k", "l", "i", "a", "d", "x", " "];
    
    if (navKeys.includes(event.key)) {
      event.stopPropagation();
      return;
    }
    
    // Let Escape and Enter pass through to Vim handlers
    if (event.key === "Escape" || event.key === "Enter") {
      // Don't stop propagation - let it reach Vim
    } else {
      return;
    }
  }

  // "z" ångra den senaste handlingen
  if (
    event.key.toLowerCase() === "z" &&
    !event.ctrlKey &&
    !event.metaKey &&
    !event.altKey &&
    !isInInput &&
    !isSearchOpen.value &&
    !isLeaderOpen.value &&
    !isQuickAddOpen.value &&
    !event.defaultPrevented
  ) {
    event.preventDefault();
    void undoLastAction();
    return;
  }

  if (event.key === "?" && !event.ctrlKey && !event.metaKey && !event.altKey && !isInInput) {
    event.preventDefault();
    showHelpModal.value = true;
  }

  // Open search on Ctrl+K or Cmd+K
  if ((event.ctrlKey || event.metaKey) && event.key === "k") {
    event.preventDefault();
    openSearch();
  }

  handlePriorityKeydown(event);
  handleNavigationKeydown(event);
  handlePriorityToggleKeydown(event);
}

function handleGlobalKeyup(event: KeyboardEvent) {
  handlePriorityKeyup(event);
}

function updateSearchOpen(open: boolean) {
  if (open) {
    openSearch();
  } else {
    closeSearch();
  }
}

onMounted(() => {
  // Capture-fasen gör att leader-läget vinner över sidornas egna
  // tangentbevakare, oavsett i vilken ordning lyssnarna registrerades.
  window.addEventListener("keydown", handleLeaderKeydown, true);
  window.addEventListener("keydown", handleGlobalKeydown);
  window.addEventListener("keyup", handleGlobalKeyup);
  loadPrioritizedGoals();
  enableNavigation();
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleLeaderKeydown, true);
  window.removeEventListener("keydown", handleGlobalKeydown);
  window.removeEventListener("keyup", handleGlobalKeyup);
  disableNavigation();
});
</script>

<template>
  <UDashboardGroup unit="rem">
    <UDashboardSidebar
      id="default"
      v-model:open="open"
      mode="slideover"
      collapsible
      resizable
      class="bg-surface/80 backdrop-blur-xl border-r border-line"
      :ui="{
        root: '!hidden',
        content: '!flex',
        overlay: '!block',
        footer: 'lg:border-t lg:border-line',
      }"
    >
      <template #default="{ collapsed }">
        <UNavigationMenu
          :collapsed="collapsed"
          :items="links[0]"
          orientation="vertical"
          tooltip
          popover
        />
      </template>

      <template #footer="{ collapsed }">
        <UserMenu :collapsed="collapsed" />
      </template>
    </UDashboardSidebar>

    <UDashboardSearch
      v-model:search-term="searchTerm"
      :open="isSearchOpen"
      :groups="groups"
      @update:open="updateSearchOpen"
    >
      <!-- Händelsernas ikon kan vara en emoji, som inte går att visa med
           UIcon. Egen slot används därför för händelser. -->
      <template #event-leading="{ item }">
        <EventIcon
          v-if="item.eventIcon"
          :icon="item.eventIcon"
          class="size-4 shrink-0"
        />
        <UIcon v-else name="i-lucide-calendar-clock" class="size-4 shrink-0" />
      </template>
    </UDashboardSearch>

    <HelpModal v-model:open="showHelpModal" />

    <LeaderKeyModal
      :open="isLeaderOpen"
      :commands="leaderCommands"
      :pending="leaderPendingKeys"
      @execute-command="executeLeaderCommand($event)"
    />

    <!-- Popupen för startdatum och due-datum (leader + d). Den renderas
         här så att den kan öppnas från alla sidor. -->
    <GoalDatesModal
      :open="goalDates.isOpen.value"
      :goal-title="goalDates.target.value?.title"
      :start-date="goalDates.target.value?.start_date ?? null"
      :due-date="goalDates.target.value?.due ?? null"
      @save="goalDates.save"
      @close="goalDates.close"
    />

    <!-- Taggväljaren (leader + t). -->
    <GoalTagPicker
      :open="tagPicker.isOpen.value"
      :target-title="tagPicker.target.value?.title ?? ''"
      :selected="tagPicker.draft.value"
      :available="tagPicker.available.value"
      @update:selected="tagPicker.save($event)"
      @close="tagPicker.close()" />

    <QuickAdd @created="closeQuickAdd" @closed="closeQuickAdd" />

    <Transition name="fade">
      <div
        v-if="showPriorityIndicator && prioritizedGoals.length > 0"
        class="fixed top-4 right-4 z-50 px-4 py-2.5 bg-fuchsia-600/90 backdrop-blur-sm rounded-xl shadow-lg border border-fuchsia-400/30"
      >
        <div class="flex items-center gap-3">
          <span class="text-fuchsia-100 text-sm font-medium">Ordna framsteg</span>
          <div class="flex items-center gap-2 text-xs">
            <span class="text-fuchsia-200">j/k: navigera</span>
            <span class="text-fuchsia-200">Shift+P + J/K: justera ordning</span>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Kalendern ska ligga direkt på sidans bakgrund, så panelens egen
         yta sätts av på den sidan. -->
    <UDashboardPanel
      class="!overflow-y-auto !h-screen min-w-0"
      :class="route.path === '/calendar' ? 'bg-transparent' : 'bg-canvas'"
    >
      <UDashboardNavbar class="hidden bg-transparent lg:flex">
        <template #toggle>
          <UDashboardSidebarToggle class="!flex" />
        </template>
      </UDashboardNavbar>

      <div class="pb-20 lg:pb-0">
        <slot />
      </div>

      <div
        class="fixed bottom-0 left-0 right-0 z-50 flex items-center justify-center border-t border-line bg-surface/90 p-1.5 backdrop-blur-xl lg:hidden"
      >
        <UDashboardSidebarToggle />
      </div>
    </UDashboardPanel>

    <!-- På större skärmar är kalendern en permanent högerspalt. Den hålls
         smal nog för att kännas som en mobilkalender, medan huvudinnehållet
         får resten av bredden. På små skärmar visas kalendern via /calendar
         i stället, så att layouten inte blir för trång. -->
    <aside
      v-if="props.calendar && loggedIn && route.path !== '/login' && route.path !== '/signup'"
      class="hidden min-w-0 shrink-0 flex-col border-l border-line bg-canvas lg:flex lg:w-[24rem] xl:w-[26rem]"
      aria-label="Kalender"
    >
      <div class="min-h-0 w-full flex-1">
        <VerticalCalendar />
      </div>
    </aside>
  </UDashboardGroup>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
