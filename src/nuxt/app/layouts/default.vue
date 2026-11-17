<script setup lang="ts">
import type { NavigationMenuItem } from "@nuxt/ui";
import type { Goal } from "~/types/goal";

const route = useRoute();
const { loggedIn } = useUserSession();

const open = ref(false);
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
  handleKeyup: handleNavigationKeyup,
  isLeaderMode,
} = useGlobalNavigation();

const { isOpen: isQuickAddOpen, open: openQuickAdd, close: closeQuickAdd } = useQuickAdd();

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
      label: "Prioritet",
      icon: "i-lucide-sort-desc",
      to: "/priority",
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
      label: "Login",
      icon: "i-lucide-log-in",
      to: "/login",
      onSelect: () => {
        open.value = false;
      },
    },
  ],
] satisfies NavigationMenuItem[][];

const goalsData = ref<Goal[] | null>(null);
const { loadAllGoals } = useGoalApi();

const fetchGoals = async () => {
  try {
    goalsData.value = await loadAllGoals();
  } catch (error) {
    // Silently fail for navigation - goals are optional there
    goalsData.value = null;
  }
};

onMounted(() => {
  fetchGoals();
});

const groups = computed(() => {
  const goalItems =
    goalsData.value?.map((goal) => ({
      id: `goal-${goal.id}`,
      label: goal.title,
      icon: goal.finished ? "i-lucide-check-circle" : "i-lucide-circle",
      to: `/goal/${goal.id}`,
      suffix: goal.finished ? "Klar" : undefined,
    })) ?? [];

  return [
    {
      id: "goals",
      label: "Mål",
      items: goalItems,
    },
    {
      id: "links",
      label: "Navigering",
      items: links.flat(),
    },
  ];
});

function handleGlobalKeydown(event: KeyboardEvent) {
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
    !isLeaderMode.value &&
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
  handleNavigationKeyup(event);
}

function updateSearchOpen(open: boolean) {
  if (open) {
    openSearch();
  } else {
    closeSearch();
  }
}

onMounted(() => {
  window.addEventListener("keydown", handleGlobalKeydown);
  window.addEventListener("keyup", handleGlobalKeyup);
  loadPrioritizedGoals();
  enableNavigation();
});

onUnmounted(() => {
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
      collapsible
      resizable
      class="bg-surface/80 backdrop-blur-xl border-r border-line"
      :ui="{ footer: 'lg:border-t lg:border-line' }"
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

    <UDashboardSearch :open="isSearchOpen" :groups="groups" @update:open="updateSearchOpen" />

    <HelpModal v-model:open="showHelpModal" />

    <QuickAdd @created="closeQuickAdd" @closed="closeQuickAdd" />

    <Transition name="fade">
      <div
        v-if="showPriorityIndicator && prioritizedGoals.length > 0"
        class="fixed top-4 right-4 z-50 px-4 py-2.5 bg-fuchsia-600/90 backdrop-blur-sm rounded-xl shadow-lg border border-fuchsia-400/30"
      >
        <div class="flex items-center gap-3">
          <span class="text-fuchsia-100 text-sm font-medium">Prioriteringsläge</span>
          <div class="flex items-center gap-2 text-xs">
            <span class="text-fuchsia-200">j/k: navigera</span>
            <span class="text-fuchsia-200">P+K/J: ändra vikt</span>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Kalendern ska ligga direkt på sidans bakgrund, så panelens egen
         yta sätts av på den sidan. -->
    <UDashboardPanel
      class="!overflow-y-auto !h-screen"
      :class="route.path === '/calendar' ? 'bg-transparent' : 'bg-canvas'"
    >
      <UDashboardNavbar v-if="route.path !== '/calendar'" class="hidden lg:flex">
        <template #toggle>
          <UDashboardSidebarToggle />
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
