<script setup lang="ts">
import LeaderKeyModal from '~/components/LeaderKeyModal.vue'
import GoalHeader from '~/components/GoalHeader.vue'
import GoalParentSection from '~/components/GoalParentSection.vue'
import GoalChildrenList from '~/components/GoalChildrenList.vue'
import GoalChildSearch from '~/components/GoalChildSearch.vue'
import ConfirmDeleteGoalModal from '~/components/ConfirmDeleteGoalModal.vue'
import ConfirmRemoveParentModal from '~/components/ConfirmRemoveParentModal.vue'
import type { Goal, GoalWithWeight } from '~/types/goal'

// Make this page client-only to ensure GraphQL requests work properly and prevent Pinia hydration issues
definePageMeta({
  ssr: false,
});

const route = useRoute();
const router = useRouter();
const goalId = computed(() => parseInt(route.params.id as string));
const { user } = useUserSession();
const { fetchGoalData, updateGoalTitle, updateGoalIcon: updateGoalIconApi, toggleGoalFinished, toggleGoalStarted, deleteGoal, addParentRelation, removeParentRelation, addChildRelation, updateGoalOrder, updateGoalWeight, createGoal, addDependency, removeDependency } = useGoalApi();

// Hämta målet med dess relationer
const goalData = ref<GetGoalResponse | null>(null);
const isLoading = ref(true);
const error = ref<string | null>(null);

const refresh = async (forceRefresh = false) => {
  isLoading.value = true;
  error.value = null;

  try {
    goalData.value = await fetchGoalData(goalId.value, forceRefresh);
  } catch (err) {
    console.error("Fetch goal error:", err);
    error.value = err instanceof Error ? err.message : "Unknown error";
  } finally {
    isLoading.value = false;
  }
};

// Fetch initial data
await refresh();

// Re-fetch when navigating to a different goal
watch(
  () => route.params.id,
  async () => {
    await refresh();
  }
);

const goal = computed(() => goalData.value?.goal);

// Extrahera alla föräldrar (REST API format)
const parents = computed(
  () => goalData.value?.parents || []
);

// Barn med vikt och ordning (REST API returnerar detta direkt)
const children = computed(() => {
  return goalData.value?.children || [];
});

// Beräkna progress baserat på färdiga undermål
const progress = computed(() => {
  if (children.value.length === 0) return 0;
  const completed = children.value.filter((c) => c.finished).length;
  return Math.round((completed / children.value.length) * 100);
});

// Dependencies (mål som detta mål väntar på)
const dependsOn = computed(() => {
  return goalData.value?.dependsOn || [];
});

// Blocking (mål som detta mål blockerar)
const blocking = computed(() => {
  return goalData.value?.blocking || [];
});

// Kollar om detta mål är "klart" baserat på dependencies
const isBlocked = computed(() => {
  // Om inga dependencies, är det inte blockerat
  if (dependsOn.value.length === 0) return false;
  // Om alla dependencies är klara, är det inte blockerat
  return dependsOn.value.some((d: any) => !d.finished);
});

	// Bygg dependency-info för varje barn
const childDependencies = computed(() => {
	const deps: Record<number, { dependsOn: any[]; blocking: any[] }> = {};
	children.value.forEach(child => {
		deps[child.id] = {
			dependsOn: child.dependsOn || [],  // Array av Goals
			blocking: child.blocking || []     // Array av Goals
		};
	});
	return deps;
});

// Hjälpfunktioner för barn-status
function isChildBlocked(childId: number): boolean {
  const info = childDependencies.value[childId];
  if (!info || info.dependsOn.length === 0) return false;
  return info.dependsOn.some((d: any) => !d.finished);
}

function getChildWaitingForTitle(childId: number): string | null {
  const info = childDependencies.value[childId];
  if (!info || info.dependsOn.length === 0) return null;
  const waitingFor = info.dependsOn.find((d: any) => !d.finished);
  return waitingFor?.title || null;
}

// Visa/dölj dependencies-sektion
const showDependencies = ref(false);

// Sök för att lägga till dependency
const showDependencySearch = ref(false);
const dependencySearchQuery = ref("");
const dependencySearchInput = ref<HTMLInputElement | null>(null);

const filteredDependencySearchResults = computed(() => {
  if (!dependencySearchQuery.value.trim()) return [];

  const results = goalsStore.searchGoals(dependencySearchQuery.value);

  // Filtrera bort nuvarande målet, befintliga dependencies, och mål som redan är klara
  const currentDependsOnIds = dependsOn.value.map((d: any) => d.id);
  return results.filter(
    (g) => g.id !== goalId.value && !currentDependsOnIds.includes(g.id)
  );
});

async function addExistingDependency(dependencyGoalId: number) {
  try {
    await addDependency(goalId.value, dependencyGoalId);
    await refresh();
    showDependencySearch.value = false;
    dependencySearchQuery.value = "";
  } catch (error: any) {
    console.error("Failed to add dependency:", error);
    alert(error?.data?.message || error?.message || "Kunde inte lägga till beroende");
  }
}

async function removeExistingDependency(dependencyGoalId: number) {
  try {
    await removeDependency(goalId.value, dependencyGoalId);
    await refresh();
  } catch (error) {
    console.error("Failed to remove dependency:", error);
  }
}

async function toggleDependencySearch() {
  showDependencySearch.value = !showDependencySearch.value;
  if (showDependencySearch.value) {
    await nextTick();
    dependencySearchInput.value?.focus();
  }
}

function handleDependencySearchKeydown(event: KeyboardEvent) {
  if (event.key === "Escape") {
    showDependencySearch.value = false;
    dependencySearchQuery.value = "";
  } else if (event.key === "Enter" && filteredDependencySearchResults.value.length > 0) {
    addExistingDependency(filteredDependencySearchResults.value[0].id);
  }
}

// Dependency drag functions (mouse-based)
function handleDepMouseDown(event: MouseEvent, childId: number) {
  // Endast hantera vänster musknapp
  if (event.button !== 0) return;
  event.preventDefault();
  event.stopPropagation();
  depDragChildId.value = childId;
  isDepDragging.value = true;
}

function handleDepMouseMove(event: MouseEvent) {
  if (!isDepDragging.value || !depDragChildId.value) return;
  
  // Hitta elementet under muspekaren
  const target = document.elementFromPoint(event.clientX, event.clientY);
  if (!target) {
    depDragOverChildId.value = null;
    return;
  }
  
  // Hitta närmaste li-element (barnrad)
  const li = target.closest('li[data-child-index]') as HTMLElement;
  if (li) {
    const index = parseInt(li.dataset.childIndex || '0');
    const child = filteredChildren.value[index];
    if (child && child.id !== depDragChildId.value) {
      depDragOverChildId.value = child.id;
    } else {
      depDragOverChildId.value = null;
    }
  } else {
    depDragOverChildId.value = null;
  }
}

function handleDepMouseUp(_event: MouseEvent) {
  if (!isDepDragging.value) return;
  
  // Skapa eller ta bort dependency beroende på om den redan finns
  if (depDragOverChildId.value && depDragChildId.value !== depDragOverChildId.value) {
    const childInfo = childDependencies.value[depDragChildId.value];
    // dependsOn är nu array av Goals, inte {depends_on_id} objekter
    const existingDep = childInfo?.dependsOn?.find(
      (d: any) => d.id === depDragOverChildId.value
    );
    
    if (existingDep) {
      // Ta bort beroendet om det redan finns
      console.log('Removing dependency:', depDragChildId.value, '->', depDragOverChildId.value);
      removeDependency(depDragChildId.value, depDragOverChildId.value).then(() => {
        refresh();
      });
    } else {
      // Skapa beroendet om det inte finns
      console.log('Creating dependency:', depDragChildId.value, '->', depDragOverChildId.value);
      addDependency(depDragChildId.value, depDragOverChildId.value).then(() => {
        refresh();
      });
    }
  }
  
  // Nollställ state
  depDragChildId.value = null;
  depDragOverChildId.value = null;
  isDepDragging.value = false;
}

// Globala mus-event listeners för dependency drag
onMounted(() => {
  window.addEventListener('mousemove', handleDepMouseMove);
  window.addEventListener('mouseup', handleDepMouseUp);
});

onUnmounted(() => {
  window.removeEventListener('mousemove', handleDepMouseMove);
  window.removeEventListener('mouseup', handleDepMouseUp);
});

// Använd goals store
const goalsStore = useGoalsStore();

// Ladda goals om de inte är laddade
onMounted(async () => {
  // Populate store with goals from the allGoals data we already fetched
  if (goalData.value?.allGoals && goalsStore.goals.length === 0) {
    goalData.value.allGoals.forEach(g => goalsStore.addGoal(g));
    goalsStore.isLoaded = true;
  }
});

// Visa/dölj avklarade mål
const showCompleted = ref(false);
const showStarted = ref(false);

// Icon picker state
const showIconPicker = ref(false);
const editingIconGoalId = ref<number | null>(null);
const editingIconGoalIcon = ref<string>('');

// Weight editing state
const weightEditingChildId = ref<number | null>(null);
const tempWeight = ref(10);

// Dependency drag state (mouse-based for simpler handling)
const depDragChildId = ref<number | null>(null);
const depDragOverChildId = ref<number | null>(null);
const isDepDragging = ref(false);

// Filtrerade undermål baserat på showCompleted och showStarted
const filteredChildren = computed(() => {
  let result = children.value;
  if (!showCompleted.value) {
    result = result.filter((c) => !c.finished);
  }
  if (showStarted.value) {
    result = result.filter((c) => c.started !== null);
  }
  
  // Topologisk sortering: goals whose dependencies are done come first
  // Then goals in dependency chain order
  
  // Build a map of goal -> array of goal IDs it depends on (that are not finished)
  // Use dependsOn directly from each child
  const dependenciesMap = new Map<number, number[]>();
  result.forEach(child => {
    const dependsOn = child.dependsOn || [];
    // Filter out finished dependencies
    const unfinishedDeps = dependsOn
      .filter((d: any) => !d.finished)
      .map((d: any) => d.depends_on_id);
    dependenciesMap.set(child.id, unfinishedDeps);
  });
  
  // Debug log
  console.log('Topological sort debug:', {
    childrenIds: result.map(c => c.id),
    childrenCount: result.length,
    dependenciesMap: JSON.stringify([...dependenciesMap.entries()]),
  });
  
  // Kahn's algorithm - process one goal at a time to keep dependency chains together
  const mutableDeps = new Map<number, number[]>();
  
  result.forEach(child => {
    const depsNeeded = (dependenciesMap.get(child.id) || []).slice();
    mutableDeps.set(child.id, depsNeeded);
  });
  
  const sorted: any[] = [];
  const completed = new Set<number>();
  
  // Keep processing until all goals are sorted
  while (sorted.length < result.length) {
    // Find all goals that are now ready (no remaining dependencies)
    const ready: any[] = [];
    for (const child of result) {
      if (completed.has(child.id)) continue;
      const depsNeeded = mutableDeps.get(child.id) || [];
      if (depsNeeded.length === 0) {
        ready.push(child);
      }
    }
    
    if (ready.length === 0) {
      // No ready goals - shouldn't happen with valid dependencies
      for (const child of result) {
        if (!completed.has(child.id)) {
          sorted.push(child);
          completed.add(child.id);
        }
      }
      break;
    }
    
    // Sort ready: prefer goals that HAVE dependencies (part of a chain) over goals that don't
    // This keeps dependency chains together
    ready.sort((a, b) => {
      // Check if goal has original dependencies (is part of a chain)
      const aHasDeps = (dependenciesMap.get(a.id) || []).length > 0;
      const bHasDeps = (dependenciesMap.get(b.id) || []).length > 0;
      
      // If one has deps and one doesn't, prefer the one with deps
      if (aHasDeps && !bHasDeps) return -1;
      if (!aHasDeps && bHasDeps) return 1;
      
      // If both have deps or both don't, sort by original order
      const aOriginal = result.indexOf(a);
      const bOriginal = result.indexOf(b);
      return aOriginal - bOriginal;
    });
    const current = ready[0];
    sorted.push(current);
    completed.add(current.id);
    
    // Remove this goal from dependencies of remaining goals
    for (const child of result) {
      if (completed.has(child.id)) continue;
      const depsNeeded = mutableDeps.get(child.id) || [];
      const idx = depsNeeded.indexOf(current.id);
      if (idx !== -1) {
        depsNeeded.splice(idx, 1);
        mutableDeps.set(child.id, depsNeeded);
      }
    }
  }
  
  return sorted;
});

// Sökfunktion för att lägga till föräldrar
const showParentSearch = ref(false);
const parentSearchQuery = ref("");
const parentSearchInput = ref<HTMLInputElement | null>(null);

// Fokusera input när plus-knappen klickas
async function toggleParentSearch() {
  showParentSearch.value = !showParentSearch.value;
  if (showParentSearch.value) {
    await nextTick();
    parentSearchInput.value?.focus();
  }
}

// Sök i lokal state istället för GraphQL
const searchResults = computed(() => {
  if (!parentSearchQuery.value.trim()) return [];

  const results = goalsStore.searchGoals(parentSearchQuery.value);

  // Filtrera bort nuvarande målet och befintliga föräldrar
  const currentParentIds = parents.value.map((p) => p.id);
  return results.filter(
    (g) => g.id !== goalId.value && !currentParentIds.includes(g.id)
  );
});

// Lägg till befintligt mål som förälder
async function addExistingParent(parentId: number) {
  try {
    await addParentRelation(goalId.value, parentId);
    goalsStore.addRelation(goalId.value, parentId);
    await refresh();
    showParentSearch.value = false;
    parentSearchQuery.value = "";
  } catch (error) {
    console.error("Failed to add parent:", error);
  }
}

// Skapa nytt mål som förälder
async function createNewParent() {
  if (!parentSearchQuery.value.trim()) return;

  try {
    const newGoal = await createGoal(parentSearchQuery.value.trim());
    await addParentRelation(goalId.value, newGoal.id);
    goalsStore.addGoal(newGoal);
    goalsStore.addRelation(goalId.value, newGoal.id);
    await goalsStore.loadGoals();
    showParentSearch.value = false;
    parentSearchQuery.value = "";
  } catch (error) {
    console.error("Failed to create new parent:", error);
  }
}

// Hantera Enter-tangent
function handleSearchKeydown(event: KeyboardEvent) {
  if (event.key === "Enter") {
    if (searchResults.value.length === 0) {
      // Skapa nytt mål om inga resultat
      createNewParent();
    } else {
      // Välj första resultatet
      addExistingParent(searchResults.value[0].id);
    }
  } else if (event.key === "Escape") {
    showParentSearch.value = false;
    parentSearchQuery.value = "";
  }
}

// Sökfunktion för att lägga till barn (undermål)
const showChildSearch = ref(false);
const childSearchQuery = ref("");
const childSearchInput = ref<HTMLInputElement | null>(null);

  // Fokusera input när plus-knappen klickas
  async function toggleChildSearch() {
    showChildSearch.value = !showChildSearch.value;
    if (showChildSearch.value) {
      await nextTick();
      childSearchInput.value?.focus();
    }
  }

// Sök för undermål
const childSearchResults = computed(() => {
  if (!childSearchQuery.value.trim()) return [];

  const results = goalsStore.searchGoals(childSearchQuery.value);

  // Filtrera bort nuvarande målet och befintliga barn
  const currentChildIds = children.value.map((c) => c.id);
  return results.filter(
    (g) => g.id !== goalId.value && !currentChildIds.includes(g.id)
  );
});

// Lägg till befintligt mål som barn
async function addExistingChild(childId: number) {
  try {
    const nextOrder = goal.value?.childRelations?.length || 0;
    await addChildRelation(childId, goalId.value, nextOrder);
    goalsStore.addRelation(childId, goalId.value);
    await refresh();
    showChildSearch.value = false;
    childSearchQuery.value = "";
  } catch (error) {
    console.error("Failed to add child:", error);
  }
}

  // Skapa nytt mål som barn
  async function createNewChild() {
    if (!childSearchQuery.value.trim()) return;

    try {
      const newGoal = await createGoal(childSearchQuery.value.trim());
      const nextOrder = children.value.length;
      await addChildRelation(newGoal.id, goalId.value, nextOrder);
      goalsStore.addGoal(newGoal);
      goalsStore.addRelation(newGoal.id, goalId.value);
      await goalsStore.loadGoals();
      await refresh();
      // Markera det nya målet direkt
      await nextTick();
      const newChildIndex = filteredChildren.value.findIndex(c => c.id === newGoal.id);
      if (newChildIndex !== -1) {
        selectedChildIndex.value = newChildIndex;
        isGoalSelected.value = false;
        isParentMode.value = false;
      }
      showChildSearch.value = false;
      childSearchQuery.value = "";
    } catch (error) {
      console.error("Failed to create new child:", error);
    }
  }

  // Hantera Enter-tangent för undermål
  function handleChildSearchKeydown(event: KeyboardEvent) {
    if (childSearchResults.value.length === 0) {
      // Skapa nytt mål om inga resultat
      createNewChild();
    } else {
      // Välj första resultatet
      addExistingChild(childSearchResults.value[0].id);
    }
  }

// Long-press för att ta bort föräldrarelation
const pressTimer = ref<NodeJS.Timeout | null>(null);
const parentToRemove = ref<number | null>(null);
const showRemoveConfirmation = ref(false);

function handleParentMouseDown(parentId: number) {
  pressTimer.value = setTimeout(() => {
    parentToRemove.value = parentId;
    showRemoveConfirmation.value = true;
  }, 500); // 500ms = håll inne i 0.5 sek
}

function handleParentMouseUp() {
  if (pressTimer.value) {
    clearTimeout(pressTimer.value);
    pressTimer.value = null;
  }
}

function cancelRemoveParent() {
  showRemoveConfirmation.value = false;
  parentToRemove.value = null;
}

async function confirmRemoveParent() {
  if (!parentToRemove.value) return;

  try {
    await removeParentRelation(goalId.value, parentToRemove.value);
    goalsStore.removeRelation(goalId.value, parentToRemove.value);
    await refresh();
    showRemoveConfirmation.value = false;
    parentToRemove.value = null;
  } catch (error) {
    console.error("Failed to remove parent relation:", error);
  }
}

// Swipe funktionalitet för att markera undermål som färdiga
interface SwipeState {
  startX: number;
  currentX: number;
  isSwiping: boolean;
  childId: number | null;
}

const swipeState = ref<SwipeState>({
  startX: 0,
  currentX: 0,
  isSwiping: false,
  childId: null,
});

function handleTouchStart(event: TouchEvent, childId: number) {
  const touch = event.touches[0];
  swipeState.value = {
    startX: touch.clientX,
    currentX: touch.clientX,
    isSwiping: true,
    childId: childId,
  };
}

function handleTouchMove(event: TouchEvent) {
  if (!swipeState.value.isSwiping) return;

  const touch = event.touches[0];
  swipeState.value.currentX = touch.clientX;
}

async function handleTouchEnd(child: Goal) {
  if (!swipeState.value.isSwiping) return;

  const deltaX = swipeState.value.currentX - swipeState.value.startX;
  const threshold = 50; // Minimum swipe distance i pixels

  if (Math.abs(deltaX) > threshold && deltaX > 0) {
    // Swipade åt höger - toggla finished status
    await toggleFinished(child);
  }

  // Reset swipe state
  swipeState.value = {
    startX: 0,
    currentX: 0,
    isSwiping: false,
    childId: null,
  };
}

// Beräkna swipe offset för visuell feedback
function getSwipeOffset(childId: number): number {
  if (swipeState.value.isSwiping && swipeState.value.childId === childId) {
    const delta = swipeState.value.currentX - swipeState.value.startX;
    return Math.max(0, Math.min(delta, 100)); // Begränsa till 0-100px
  }
  return 0;
}

// Drag-and-drop state för sortering av barn
const draggedChildId = ref<number | null>(null);
const dragOverChildId = ref<number | null>(null);
let dragLeaveTimeout: ReturnType<typeof setTimeout> | null = null;

function handleDragStart(event: DragEvent, childId: number) {
  draggedChildId.value = childId;
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = "move";
  }
}

function handleDragOver(event: DragEvent, childId: number) {
  event.preventDefault();
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = "move";
  }
  // Avbryt tidigare leave-timeout för att undvika blinking
  if (dragLeaveTimeout) {
    clearTimeout(dragLeaveTimeout);
    dragLeaveTimeout = null;
  }
  dragOverChildId.value = childId;
}

function handleDragLeave() {
  // Fördröj nollställningen för att undvika blinking vid snabb hover
  dragLeaveTimeout = setTimeout(() => {
    dragOverChildId.value = null;
    dragLeaveTimeout = null;
  }, 100);
}

async function handleDrop(event: DragEvent, dropChildId: number) {
  event.preventDefault();
  // Rensa timeout och nollställ state
  if (dragLeaveTimeout) {
    clearTimeout(dragLeaveTimeout);
    dragLeaveTimeout = null;
  }
  dragOverChildId.value = null;

  if (draggedChildId.value === null || draggedChildId.value === dropChildId) {
    draggedChildId.value = null;
    return;
  }

  const draggedId = draggedChildId.value;

  // Reparent: ta bort från nuvarande förälder och lägg till som barn till dropChildId
  try {
    // Ta bort relation med nuvarande förälder
    await removeParentRelation(draggedId, goalId.value);
    goalsStore.removeRelation(draggedId, goalId.value);
    
    // Räkna befintliga barn till dropTargetChild för att bestämma ordning
    const targetChildrenCount = children.value.filter(c => {
      const hasRelation = goalData.value?.goal?.childRelations?.some(
        r => r.child_id === c.id
      );
      return hasRelation;
    }).length;
    
    await addChildRelation(draggedId, dropChildId, targetChildrenCount);
    goalsStore.addRelation(draggedId, dropChildId);
    
    await refresh();
  } catch (error) {
    console.error("Failed to reparent goal:", error);
    await refresh();
  }

  draggedChildId.value = null;
}

// Flytta markerat barn upp i listan
async function moveChildUp() {
  if (
    selectedChildIndex.value <= 0 ||
    isParentMode.value ||
    isGoalSelected.value
  ) {
    return;
  }

  const currentIndex = selectedChildIndex.value;
  const displayedChildren = filteredChildren.value;

  if (currentIndex >= displayedChildren.length) {
    return;
  }

  // Hämta de två målen som ska bytas
  const movingChild = displayedChildren[currentIndex];
  const aboveChild = displayedChildren[currentIndex - 1];

  console.log('moveChildUp: moving child', movingChild.id, 'with order', movingChild.order, 'to position of', aboveChild.id, 'with order', aboveChild.order);

  try {
    // Byt ordningen mellan de två målen
    const movingOrder = movingChild.order;
    const aboveOrder = aboveChild.order;

    await Promise.all([
      updateGoalOrder(goalId.value, movingChild.id, aboveOrder),
      updateGoalOrder(goalId.value, aboveChild.id, movingOrder),
    ]);

    console.log('moveChildUp: order updated, refreshing with force');

    // Uppdatera markerad index
    selectedChildIndex.value = currentIndex - 1;

    // Uppdatera data från server med force refresh
    await refresh(true);
    
    console.log('moveChildUp: refresh complete, children count:', children.value.length);
  } catch (error) {
    console.error("Failed to move child up:", error);
    await refresh(true);
  }
}

// Flytta markerat barn ner i listan
async function moveChildDown() {
  if (
    selectedChildIndex.value < 0 ||
    isParentMode.value ||
    isGoalSelected.value
  ) {
    return;
  }

  const currentIndex = selectedChildIndex.value;
  const displayedChildren = filteredChildren.value;

  if (currentIndex >= displayedChildren.length - 1) {
    return;
  }

  // Hämta de två målen som ska bytas
  const movingChild = displayedChildren[currentIndex];
  const belowChild = displayedChildren[currentIndex + 1];

  console.log('moveChildDown: moving child', movingChild.id, 'with order', movingChild.order, 'to position of', belowChild.id, 'with order', belowChild.order);

  try {
    // Byt ordningen mellan de två målen
    const movingOrder = movingChild.order;
    const belowOrder = belowChild.order;

    await Promise.all([
      updateGoalOrder(goalId.value, movingChild.id, belowOrder),
      updateGoalOrder(goalId.value, belowChild.id, movingOrder),
    ]);

    console.log('moveChildDown: order updated, refreshing with force');

    // Uppdatera markerad index
    selectedChildIndex.value = currentIndex + 1;

    // Uppdatera data från server med force refresh
    await refresh(true);
    
    console.log('moveChildDown: refresh complete, children count:', children.value.length);
  } catch (error) {
    console.error("Failed to move child down:", error);
    await refresh(true);
  }
}

// Vim-style navigering för undermål
const selectedChildIndex = ref(0);
const selectedParentIndex = ref(0);
const isParentMode = ref(false);

// Vim modes
const mode = ref<"normal" | "insert">("normal");
const editingGoalId = ref<number | null>(null);
const editTitle = ref("");
const editInputRef = ref<HTMLInputElement | null>(null);

// Goal selection (själva målet är markerat, inte barn eller föräldrar)
const isGoalSelected = ref(false);

// Delete confirmation
const showDeleteConfirmation = ref(false);
const goalToDelete = ref<{ id: number; title: string } | null>(null);
const deleteDialogSelection = ref<"cancel" | "confirm">("cancel");

// Leader key state
const isLeaderModalOpen = ref(false);
let leaderModalJustOpened = false;
let leaderFirstKey = "";

// Gå in i insert mode för att redigera ett mål
async function enterInsertMode(
  goalId: number,
  title: string,
  atBeginning: boolean = false
) {
  mode.value = "insert";
  editingGoalId.value = goalId;
  editTitle.value = title;

  // Vänta tills DOM uppdateras
  await nextTick();
  await nextTick();

  // Hitta input-elementet som nu är synligt
  const input = document.querySelector(
    'input[type="text"]:focus, input[type="text"]:not([style*="display: none"])'
  ) as HTMLInputElement;
  if (input) {
    input.focus();
    if (atBeginning) {
      // 'a' sätter cursor i början
      input.setSelectionRange(0, 0);
    } else {
      // 'i' sätter cursor i slutet
      input.setSelectionRange(title.length, title.length);
    }
  } else {
    // Fallback: försök hitta vilket input-element som helst
    setTimeout(() => {
      const fallbackInput = document.querySelector(
        "input.border-blue-500"
      ) as HTMLInputElement;
      if (fallbackInput) {
        fallbackInput.focus();
        if (atBeginning) {
          fallbackInput.setSelectionRange(0, 0);
        } else {
          fallbackInput.setSelectionRange(title.length, title.length);
        }
      }
    }, 50);
  }
}

// Spara ändringar och gå tillbaka till normal mode
async function saveEdit() {
  if (!editingGoalId.value) {
    mode.value = "normal";
    return;
  }

  // Om titeln är tom, ta bort målet
  if (!editTitle.value.trim()) {
    await handleDeleteGoal(editingGoalId.value);
    mode.value = "normal";
    editingGoalId.value = null;
    return;
  }

  try {
    await updateGoalTitle(editingGoalId.value, editTitle.value.trim());
    goalsStore.updateGoal(editingGoalId.value, {
      title: editTitle.value.trim(),
    });
    await refresh();
  } catch (error) {
    console.error("Failed to update goal title:", error);
  }

  mode.value = "normal";
  editingGoalId.value = null;
}

// Avbryt redigering
async function cancelEdit() {
  // Om vi avbryter ett nytt mål med tom titel, ta bort det
  if (editingGoalId.value && editTitle.value === "") {
    await handleDeleteGoal(editingGoalId.value);
  }
  mode.value = "normal";
  editingGoalId.value = null;
  editTitle.value = "";
}

// Visa bekräftelse för borttagning
function confirmDeleteGoal(goalId: number, title: string) {
  goalToDelete.value = { id: goalId, title };
  deleteDialogSelection.value = "confirm"; // Starta med "Ta bort mål" valt
  showDeleteConfirmation.value = true;
}

// Avbryt borttagning
function cancelDelete() {
  showDeleteConfirmation.value = false;
  goalToDelete.value = null;
  deleteDialogSelection.value = "cancel";
}

// Bekräfta och ta bort målet
async function executeDelete() {
  console.log("executeDelete called, goalToDelete:", goalToDelete.value);
  if (!goalToDelete.value) {
    console.log("No goal to delete");
    return;
  }

  console.log("Calling handleDeleteGoal with id:", goalToDelete.value.id);
  await handleDeleteGoal(goalToDelete.value.id);
  console.log("Delete completed");
  showDeleteConfirmation.value = false;
  goalToDelete.value = null;
  deleteDialogSelection.value = "cancel";
}

// Hantera navigering i delete-dialog
function handleDeleteDialogKey(event: KeyboardEvent) {
  if (!showDeleteConfirmation.value) return;

  console.log(
    "Delete dialog key:",
    event.key,
    "Selection:",
    deleteDialogSelection.value
  );

  if (event.key === "h") {
    event.preventDefault();
    event.stopPropagation();
    deleteDialogSelection.value = "cancel";
  } else if (event.key === "l") {
    event.preventDefault();
    event.stopPropagation();
    deleteDialogSelection.value = "confirm";
  } else if (event.key === "Enter") {
    event.preventDefault();
    event.stopPropagation();
    console.log("Enter pressed, selection:", deleteDialogSelection.value);
    if (deleteDialogSelection.value === "cancel") {
      console.log("Canceling delete");
      cancelDelete();
    } else {
      console.log("Executing delete");
      executeDelete();
    }
  } else if (event.key === "Escape") {
    event.preventDefault();
    event.stopPropagation();
    cancelDelete();
  }
}

// Ta bort ett mål (använder API från composable)
async function handleDeleteGoal(goalId: number) {
  try {
    await deleteGoal(goalId);
    goalsStore.removeGoal(goalId);
    await refresh();
  } catch (error) {
    console.error("Failed to delete goal:", error);
  }
}

// Skapa nytt mål på samma nivå som markerat mål
async function createSiblingGoal() {
  const userId = user.value?.id;
  if (!userId) {
    console.error("No user logged in");
    return;
  }

  try {
    const newGoal = await createGoal("", userId);
    await addParentRelation(newGoal.id, goalId.value);
    goalsStore.addGoal(newGoal);
    goalsStore.addRelation(newGoal.id, goalId.value);
    await refresh();
    await enterInsertMode(newGoal.id, "", false);
  } catch (error) {
    console.error("Failed to create sibling goal:", error);
  }
}

// Toggla finished status på ett mål
async function toggleFinished(goalToToggle: Goal) {
  try {
    const newFinishedValue = goalToToggle.finished ? null : new Date().toISOString();

    await toggleGoalFinished(goalToToggle.id, newFinishedValue);
    goalsStore.updateGoal(goalToToggle.id, { finished: newFinishedValue });

    // Uppdatera lokal state i component
    if (goalData.value?.allGoals) {
      const goalIndex = goalData.value.allGoals.findIndex(
        (g) => g.id === goalToToggle.id
      );
      if (goalIndex !== -1) {
        goalData.value.allGoals[goalIndex].finished = newFinishedValue;
      }
    }

    // Uppdatera data från server
    await refresh();
  } catch (error) {
    console.error("Failed to toggle goal finished status:", error);
  }
}

// Toggla started status på ett mål
async function toggleStarted(goalToToggle: Goal) {
  try {
    const newStartedValue = goalToToggle.started ? null : new Date().toISOString();
    await toggleGoalStarted(goalToToggle.id, newStartedValue);
    goalsStore.updateGoal(goalToToggle.id, { started: newStartedValue });
    if (goalData.value?.allGoals) {
      const goalIndex = goalData.value.allGoals.findIndex((g) => g.id === goalToToggle.id);
      if (goalIndex !== -1) {
        goalData.value.allGoals[goalIndex].started = newStartedValue;
      }
    }
    await refresh();
  } catch (error) {
    console.error("Failed to toggle goal started status:", error);
  }
}

// Uppdatera goal icon
async function updateGoalIcon(newIcon: string) {
  try {
    // Bestäm vilket mål som ska uppdateras
    const targetGoalId = editingIconGoalId.value || goalId.value;

    await updateGoalIconApi(targetGoalId, newIcon);

    // Uppdatera lokal state i component
    if (editingIconGoalId.value) {
      // Uppdatera barn-mål
      const childIndex = goalData.value?.allGoals?.findIndex(g => g.id === editingIconGoalId.value);
      if (childIndex !== -1 && goalData.value?.allGoals) {
        goalData.value.allGoals[childIndex].icon = newIcon;
      }
    } else if (goalData.value?.goal) {
      // Uppdatera huvudmål
      goalData.value.goal.icon = newIcon;
    }

    // Uppdatera data från server
    await refresh();

    // Stäng ikonväljaren
    editingIconGoalId.value = null;
    editingIconGoalIcon.value = '';
  } catch (error) {
    console.error("Failed to update goal icon:", error);
  }
}

// Weight editing functions
function getWeightStyle(weight: number): { color: string; opacity: number; fontWeight?: string } {
  return { color: '#FFFFFF', opacity: 1 }
}

function startWeightEdit(child: GoalWithWeight) {
  weightEditingChildId.value = child.id
  tempWeight.value = child.weight
}

async function saveWeight() {
  if (weightEditingChildId.value !== null) {
    try {
      await updateGoalWeight(goalId.value, weightEditingChildId.value, tempWeight.value)
      await refresh()
    } catch (error) {
      console.error("Failed to update weight:", error)
    }
    weightEditingChildId.value = null
  }
}

function cancelWeightEdit() {
  weightEditingChildId.value = null
}

// Exekvera leader key kommando
function executeLeaderCommand(key: string): boolean {
  // Hantera kombinerade tangentsekvenser från modal (t.ex. "fd", "gi", "gg")
  if (key === "fd") {
    showCompleted.value = !showCompleted.value;
    isLeaderModalOpen.value = false;
    return true;
  } else if (key === "gi") {
    isLeaderModalOpen.value = false;
    router.push('/');
    return true;
  } else if (key === "gg") {
    isLeaderModalOpen.value = false;
    router.push('/goals');
    return true;
  }

  if (key === "f") {
    leaderFirstKey = "f";
    return false;
  } else if (key === "d" && leaderFirstKey === "f") {
    showCompleted.value = !showCompleted.value;
    leaderFirstKey = "";
    isLeaderModalOpen.value = false;
    return true;
  } else if (key === "g") {
    leaderFirstKey = "g";
    return false;
  } else if (key === "i" && leaderFirstKey === "g") {
    isLeaderModalOpen.value = false;
    router.push('/');
    leaderFirstKey = "";
    return true;
  } else if (key === "g" && leaderFirstKey === "g") {
    isLeaderModalOpen.value = false;
    router.push('/goals');
    leaderFirstKey = "";
    return true;
  } else if (key === "i") {
    isLeaderModalOpen.value = false;
    if (isGoalSelected.value && goal.value) {
      editingIconGoalId.value = null;
      showIconPicker.value = true;
    } else if (isParentMode.value && parents.value.length > 0) {
      editingIconGoalId.value = parents.value[selectedParentIndex.value].id;
      showIconPicker.value = true;
    } else if (filteredChildren.value.length > 0) {
      editingIconGoalId.value = filteredChildren.value[selectedChildIndex.value].id;
      showIconPicker.value = true;
    }
    leaderFirstKey = "";
    return true;
  }

  leaderFirstKey = "";
  return true;
}

// Hantera Vim-kommandon
function handleKeydown(event: KeyboardEvent) {
  // Don't handle keys when search is open
  const { isSearchOpen } = useSearchState();
  if (isSearchOpen.value) return;

  // Hantera delete-dialog navigering först
  if (showDeleteConfirmation.value) {
    handleDeleteDialogKey(event);
    return;
  }

  // Ignorera om sökfält eller ikonväljare är aktiva
  if (showParentSearch.value || showChildSearch.value || showIconPicker.value) return;

  // Insert mode - ignorera alla tangenter (hanteras direkt på input-elementet)
  if (mode.value === "insert") {
    return;
  }

  // ===============================================
  // LEADER MODE CHECKS - PRIORITERA FÖRE ALLT ANNAT
  // ===============================================

  // Hantera ESC för att stänga leader mode
  if (event.key === "Escape" && isLeaderModalOpen.value) {
    event.preventDefault();
    isLeaderModalOpen.value = false;
    return;
  }

  // Hantera leader commands när i leader mode
  if (isLeaderModalOpen.value) {
    event.preventDefault();
    const key = event.key.toLowerCase();
    executeLeaderCommand(key);
    return;
  }

  // Hantera leader key (space) - enter leader mode
  if (event.key === " ") {
    event.preventDefault();
    isLeaderModalOpen.value = true;
    return;
  }

  // ===============================================
  // NORMAL MODE KOMMANDON
  // ===============================================

  // Hantera 'd' för att toggla finished status
  if (event.key === "d") {
    event.preventDefault();
    if (isParentMode.value && parents.value.length > 0) {
      const parent = parents.value[selectedParentIndex.value];
      toggleFinished(parent);
    } else if (
      filteredChildren.value.length > 0 &&
      selectedChildIndex.value < filteredChildren.value.length
    ) {
      const child = filteredChildren.value[selectedChildIndex.value];
      toggleFinished(child);
    }
    return;
  }

  // Hantera 's' för att toggla started status
  if (event.key === "s") {
    event.preventDefault();
    if (isParentMode.value && parents.value.length > 0) {
      const parent = parents.value[selectedParentIndex.value];
      toggleStarted(parent);
    } else if (isGoalSelected.value && goal.value) {
      toggleStarted(goal.value);
    } else if (
      filteredChildren.value.length > 0 &&
      selectedChildIndex.value >= 0 &&
      selectedChildIndex.value < filteredChildren.value.length
    ) {
      const child = filteredChildren.value[selectedChildIndex.value];
      toggleStarted(child);
    }
    return;
  }

  // Hantera 'x' för att ta bort mål
  if (event.key === "x") {
    event.preventDefault();
    if (isParentMode.value && parents.value.length > 0) {
      const parent = parents.value[selectedParentIndex.value];
      parentToRemove.value = parent.id;
      showRemoveConfirmation.value = true;
    } else if (filteredChildren.value.length > 0) {
      const child = filteredChildren.value[selectedChildIndex.value];
      confirmDeleteGoal(child.id, child.title);
    }
    return;
  }

  // Hantera 'i' och 'a' för att gå in i insert mode
  if (event.key === "i" || event.key === "a") {
    event.preventDefault();
    const atBeginning = event.key === "i";

    if (isGoalSelected.value && goal.value) {
      enterInsertMode(goal.value.id, goal.value.title, atBeginning);
    } else if (isParentMode.value && parents.value.length > 0) {
      const parent = parents.value[selectedParentIndex.value];
      enterInsertMode(parent.id, parent.title, atBeginning);
    } else if (filteredChildren.value.length > 0) {
      const child = filteredChildren.value[selectedChildIndex.value];
      enterInsertMode(child.id, child.title, atBeginning);
    }
    return;
  }

  if (isParentMode.value) {
    if (event.key === "h") {
      event.preventDefault();
      selectedParentIndex.value = Math.max(selectedParentIndex.value - 1, 0);
    } else if (event.key === "l") {
      event.preventDefault();
      const parentsCount = parents.value.length;
      selectedParentIndex.value = Math.min(selectedParentIndex.value + 1, parentsCount - 1);
    } else if (event.key === "k") {
      event.preventDefault();
      const selectedParent = parents.value[selectedParentIndex.value];
      if (selectedParent) {
        if (selectedParent.id === 1) {
          router.push('/goals');
        } else {
          router.push(`/goal/${selectedParent.id}`);
        }
      }
    } else if (event.key === "j") {
      event.preventDefault();
      isParentMode.value = false;
      isGoalSelected.value = true;
      selectedChildIndex.value = -1;
    } else if (event.key === "Enter") {
      event.preventDefault();
      toggleParentSearch();
    }
  } else {
    if (event.key === "j") {
      event.preventDefault();
      if (isGoalSelected.value) {
        isGoalSelected.value = false;
        selectedChildIndex.value = 0;
      } else {
        const childrenCount = filteredChildren.value.length;
        if (childrenCount > 0) {
          selectedChildIndex.value = Math.min(selectedChildIndex.value + 1, childrenCount - 1);
        }
      }
    } else if (event.key === "k") {
      event.preventDefault();
      if (selectedChildIndex.value === 0 && !isGoalSelected.value) {
        // Go to goal title from first child
        isGoalSelected.value = true;
        selectedChildIndex.value = -1;
      } else if (isGoalSelected.value) {
        // From goal title, go to parent mode or /goals
        if (parents.value.length === 0 || parents.value[0].id === 1) {
          // No parents or root is parent -> go to /goals
          router.push('/goals');
        } else {
          // Has real parents -> go to parent mode
          isGoalSelected.value = false;
          isParentMode.value = true;
          selectedParentIndex.value = 0;
        }
      } else {
        // Move up in children list
        const childrenCount = filteredChildren.value.length;
        if (childrenCount > 0) {
          selectedChildIndex.value = Math.max(selectedChildIndex.value - 1, 0);
        }
      }
    } else if (event.key === "Enter") {
      event.preventDefault();
      toggleChildSearch();
    } else if (event.key === "l") {
      event.preventDefault();
      console.log('Navigation: l pressed, isGoalSelected:', isGoalSelected.value, 'selectedChildIndex:', selectedChildIndex.value);
      if (isGoalSelected.value) {
        isGoalSelected.value = false;
        selectedChildIndex.value = 0;
      } else {
        const childrenCount = filteredChildren.value.length;
        if (childrenCount > 0) {
          const selectedChild = filteredChildren.value[selectedChildIndex.value];
          if (selectedChild) {
            router.push(`/goal/${selectedChild.id}`);
          }
        }
      }
    } else if (event.key === "h") {
      event.preventDefault();
      console.log('Navigation: h pressed, parents.length:', parents.value.length);
      if (parents.value.length > 0) {
        router.push(`/goal/${parents.value[0].id}`);
      }
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      moveChildUp();
    } else if (event.key === "ArrowDown") {
      event.preventDefault();
      moveChildDown();
    }
  }
}

// Lägg till och ta bort event listener
onMounted(() => {
  window.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleKeydown);
});

// Återställ selectedChildIndex när children ändras
watch(filteredChildren, () => {
  if (selectedChildIndex.value >= filteredChildren.value.length) {
    selectedChildIndex.value = Math.max(0, filteredChildren.value.length - 1);
  }
});

// Debug: Log när isLeaderModalOpen ändras
watch(
  isLeaderModalOpen,
  (newVal, oldVal) => {
    if (newVal !== oldVal) {
      console.log('🔴 isLeaderModalOpen changed from', oldVal, 'to', newVal);
      console.trace('Stack trace of who changed it:');
    }
  }
);

// Återställ parent mode när målet ändras
watch(
  () => route.params.id,
  () => {
    isParentMode.value = false;
    isGoalSelected.value = false;
    selectedChildIndex.value = 0;
    selectedParentIndex.value = 0;
  }
);

// Scrolla till markerat child när selectedChildIndex ändras
watch(selectedChildIndex, async () => {
  await nextTick();
  const selectedElement = document.querySelector(`[data-child-index="${selectedChildIndex.value}"]`) as HTMLElement;
  if (selectedElement) {
    selectedElement.scrollIntoView({
      behavior: 'smooth',
      block: 'nearest'
    });
  }
});

// Scrolla till markerad parent när selectedParentIndex ändras
watch(selectedParentIndex, async () => {
  await nextTick();
  const selectedElement = document.querySelector(`[data-parent-index="${selectedParentIndex.value}"]`) as HTMLElement;
  if (selectedElement) {
    selectedElement.scrollIntoView({
      behavior: 'smooth',
      block: 'nearest'
    });
  }
});
</script>

<template>
  <div v-if="user" class="max-w-4xl mx-auto p-6 h-screen flex flex-col overflow-hidden">
    <div v-if="!goal" class="text-center py-12">
      <p class="text-gray-500">Laddar mål...</p>
    </div>

    <div v-else class="flex flex-col h-full gap-6">
      <!-- Breadcrumb / Föräldrar -->
      <div class="flex items-center justify-between flex-shrink-0">
        <div class="flex items-center gap-2 text-sm flex-wrap">
          <NuxtLink v-if="parents.length === 0" to="/goals"
            class="text-blue-400 hover:text-blue-300 transition-colors font-medium">
            Root
          </NuxtLink>

          <!-- Insert mode för förälder -->
          <div v-for="(parent, index) in parents" :key="parent.id">
            <div v-if="mode === 'insert' && editingGoalId === parent.id" class="inline-block">
              <input ref="editInputRef" v-model="editTitle" type="text"
                class="px-3 py-1 bg-gray-800 border border-blue-500 rounded text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-400"
                @keydown.enter.prevent="saveEdit" @keydown.esc.prevent="cancelEdit" />
            </div>

            <!-- Normal mode för förälder -->
            <NuxtLink v-else :to="`/goal/${parent.id}`" :data-parent-index="index"
              class="px-2 py-1 rounded transition-all select-none inline-block" :class="isParentMode && selectedParentIndex === index
                ? 'text-gray-100 bg-blue-500 font-medium'
                : 'text-gray-500 hover:text-gray-300'
                " @mousedown="handleParentMouseDown(parent.id)" @mouseup="handleParentMouseUp"
              @mouseleave="handleParentMouseUp" @touchstart="handleParentMouseDown(parent.id)"
              @touchend="handleParentMouseUp" @touchcancel="handleParentMouseUp">
              {{ parent.title }}
            </NuxtLink>
          </div>
        </div>

        <!-- + knapp för att lägga till förälder -->
        <button @click="toggleParentSearch"
          class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800"
          title="Lägg till förälder">
          <Icon name="lucide:plus" class="h-5 w-5" />
        </button>
      </div>

      <!-- Sökfält för föräldrar -->
      <div v-if="showParentSearch" class="border border-gray-700 rounded-lg p-4 bg-gray-800 flex-shrink-0">
        <div class="mb-2">
          <input ref="parentSearchInput" v-model="parentSearchQuery" @keydown="handleSearchKeydown" type="text"
            placeholder="Sök efter mål eller skapa nytt (tryck Enter)"
            class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-gray-200 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
            autofocus />
        </div>

        <!-- Sökresultat -->
        <div v-if="searchResults.length > 0" class="space-y-1 mt-2">
          <button v-for="result in searchResults" :key="result.id" @click="addExistingParent(result.id)"
            class="w-full text-left px-4 py-2 hover:bg-gray-700 rounded-lg transition-colors">
            <div class="font-medium text-gray-200">{{ result.title }}</div>
            <div class="text-xs text-gray-500">
              {{ new Date(result.created).toLocaleDateString("sv-SE") }}
            </div>
          </button>
        </div>

        <!-- Meddelande när inga resultat -->
        <div v-else-if="parentSearchQuery.trim()" class="text-sm text-gray-500 px-4 py-2">
          Tryck Enter för att skapa "{{ parentSearchQuery }}" som nytt mål
        </div>
      </div>

      <!-- Huvudmål -->
      <div v-if="mode === 'insert' && editingGoalId === goal?.id" class="mb-4">
        <input ref="editInputRef" v-model="editTitle" type="text"
          class="w-full px-3 py-2 bg-gray-800 border border-blue-500 rounded text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-400"
          @keydown.enter.prevent="saveEdit" @keydown.esc.prevent="cancelEdit" />
        <div class="text-xs text-gray-500 mt-2">
          Enter för att spara, Escape för att avbryta
        </div>
      </div>
      <div v-else class="flex items-center gap-4 flex-shrink-0">
        <h1 :class="['text-4xl font-bold transition-colors px-3 py-2 rounded flex-1',
          isGoalSelected ? 'border border-blue-500 text-gray-100' : 'text-gray-100'
        ]">{{ goal?.title }}</h1>
        <!-- Started-toggle (s) -->
        <button v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)"
          @click.stop="toggleStarted(goal)"
          class="p-2 rounded transition-colors"
          :class="goal.started
            ? 'text-yellow-400 hover:text-yellow-300'
            : 'text-gray-600 hover:text-gray-400'"
          title="Påbörja (s)">
          <Icon name="lucide:circle-play" class="w-6 h-6"
            :style="{ opacity: goal.started ? 1 : 0.25 }" />
        </button>
        <button v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)" @click.stop="showIconPicker = true"
          class="text-gray-400 hover:text-gray-200 transition-colors p-2 rounded hover:bg-gray-800" title="Ändra ikon">
          <Icon :name="goal.icon || 'heroicons:star'" class="w-8 h-8 text-white" />
        </button>
        <!-- Toggle dependencies visibility -->
        <button @click.stop="showDependencies = !showDependencies"
          class="p-2 rounded transition-colors"
          :class="showDependencies ? 'text-orange-400 hover:text-orange-300' : 'text-gray-600 hover:text-gray-400'"
          title="Beroenden (blockerade mål)">
          <Icon name="lucide:git-branch" class="w-6 h-6" />
        </button>
      </div>

      <!-- Dependencies-sektion -->
      <div v-if="showDependencies" class="border border-gray-700 rounded-lg p-4 bg-gray-800/50 mb-4">
        <!-- Väntar på -->
        <div class="mb-4">
          <div class="flex items-center justify-between mb-2">
            <h3 class="text-sm font-medium text-gray-400 flex items-center gap-2">
              <Icon name="lucide:clock" class="w-4 h-4" />
              Väntar på ({{ dependsOn.length }})
            </h3>
            <button @click="toggleDependencySearch"
              class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-700"
              title="Lägg till beroende">
              <Icon name="lucide:plus" class="w-4 h-4" />
            </button>
          </div>
          
          <!-- Sök för att lägga till dependency -->
          <div v-if="showDependencySearch" class="mb-3">
            <input ref="dependencySearchInput" v-model="dependencySearchQuery" @keydown="handleDependencySearchKeydown"
              type="text"
              placeholder="Sök efter mål att vänta på..."
              class="w-full px-3 py-2 bg-gray-900 border border-gray-700 rounded text-gray-200 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-orange-500 text-sm"
              autofocus />
            <div v-if="filteredDependencySearchResults.length > 0" class="mt-2 space-y-1">
              <button v-for="result in filteredDependencySearchResults.slice(0, 5)" :key="result.id"
                @click="addExistingDependency(result.id)"
                class="w-full text-left px-3 py-2 hover:bg-gray-700 rounded transition-colors flex items-center gap-2">
                <Icon :name="result.icon || 'heroicons:star'" class="w-4 h-4 text-gray-400" />
                <span class="text-gray-200 text-sm">{{ result.title }}</span>
                <span v-if="result.finished" class="text-xs text-green-400 ml-auto">Klar</span>
              </button>
            </div>
            <div v-else-if="dependencySearchQuery.trim() && filteredDependencySearchResults.length === 0" class="text-xs text-gray-500 mt-2 px-2">
              Inga matchande mål hittades
            </div>
          </div>
          
          <!-- Lista över dependencies -->
          <div v-if="dependsOn.length > 0" class="space-y-2">
            <div v-for="dep in dependsOn" :key="dep.depends_on_id"
              class="flex items-center gap-2 px-3 py-2 bg-gray-900 rounded hover:bg-gray-700 transition-colors">
              <Icon :name="dep.icon || 'heroicons:star'" class="w-4 h-4 text-gray-400 flex-shrink-0" />
              <NuxtLink :to="`/goal/${dep.depends_on_id}`" class="flex-1 text-sm text-gray-200 hover:text-white">
                {{ dep.title }}
              </NuxtLink>
              <span v-if="dep.finished" class="text-xs text-green-400">Klar</span>
              <span v-else class="text-xs text-orange-400">Väntar...</span>
              <button @click.stop="removeExistingDependency(dep.depends_on_id)"
                class="text-gray-500 hover:text-red-400 transition-colors p-1">
                <Icon name="lucide:x" class="w-3 h-3" />
              </button>
            </div>
          </div>
          <div v-else class="text-sm text-gray-500 px-2">
            Inga beroenden. Lägg till mål som måste göras först.
          </div>
        </div>
        
        <!-- Blockerar -->
        <div v-if="blocking.length > 0">
          <h3 class="text-sm font-medium text-gray-400 flex items-center gap-2 mb-2">
            <Icon name="lucide:lock" class="w-4 h-4" />
            Blockerar ({{ blocking.length }})
          </h3>
          <div class="space-y-2">
            <div v-for="block in blocking" :key="block.goal_id"
              class="flex items-center gap-2 px-3 py-2 bg-gray-900 rounded">
              <Icon :name="block.icon || 'heroicons:star'" class="w-4 h-4 text-gray-400 flex-shrink-0" />
              <NuxtLink :to="`/goal/${block.goal_id}`" class="flex-1 text-sm text-gray-200 hover:text-white">
                {{ block.title }}
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>

      <!-- Undermål - scrollbar container -->
      <div class="flex-1 overflow-y-auto min-h-0">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-2xl font-semibold text-gray-300"></h2>
          <div class="flex items-center gap-2">
            <button @click="showCompleted = !showCompleted"
              class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800" :title="showCompleted ? 'Dölj avklarade mål' : 'Visa avklarade mål'
                ">
              <Icon :name="showCompleted ? 'lucide:eye' : 'lucide:eye-off'" class="h-5 w-5" />
            </button>
            <!-- Visa endast påbörjade -->
            <button @click="showStarted = !showStarted"
              class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800"
              :class="showStarted ? 'text-yellow-400' : ''"
              :title="showStarted ? 'Visa alla mål' : 'Visa endast påbörjade mål'">
              <Icon name="lucide:circle-play" class="h-5 w-5"
                :style="{ opacity: showStarted ? 1 : 0.3 }" />
            </button>
            <button @click="toggleChildSearch"
              class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800"
              title="Lägg till undermål">
              <Icon name="lucide:plus" class="h-5 w-5" />
            </button>
          </div>
        </div>

        <!-- Sökfält för undermål -->
        <div v-if="showChildSearch" class="border border-gray-700 rounded-lg p-4 bg-gray-800 mb-4">
          <div class="mb-2">
             <input ref="childSearchInput" v-model="childSearchQuery" @keydown.enter.prevent="handleChildSearchKeydown" @keydown.escape.prevent="showChildSearch = false; childSearchQuery = ''" type="text"
              placeholder="Sök efter mål eller skapa nytt (tryck Enter)"
              class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-gray-200 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
              autofocus />
          </div>

          <!-- Sökresultat -->
          <div v-if="childSearchResults.length > 0" class="space-y-1 mt-2">
            <button v-for="result in childSearchResults" :key="result.id" @click="addExistingChild(result.id)"
              class="w-full text-left px-4 py-2 hover:bg-gray-700 rounded-lg transition-colors">
              <div class="font-medium text-gray-200">{{ result.title }}</div>
              <div class="text-xs text-gray-500">
                {{ new Date(result.created).toLocaleDateString("sv-SE") }}
              </div>
            </button>
          </div>

          <!-- Meddelande när inga resultat -->
          <div v-else-if="childSearchQuery.trim()" class="text-sm text-gray-500 px-4 py-2">
            Tryck Enter för att skapa "{{ childSearchQuery }}" som nytt undermål
          </div>
        </div>

        <ul v-if="filteredChildren.length > 0" class="space-y-3" :class="{ 'select-none': isDepDragging }">
          <li v-for="(child, index) in filteredChildren" :key="child.id" :data-child-index="index" draggable="true"
            @dragstart="handleDragStart($event, child.id)" @dragover="handleDragOver($event, child.id)"
            @dragleave="handleDragLeave" @drop="handleDrop($event, child.id)"
            class="relative overflow-hidden rounded-lg transition-opacity"
            :class="[
              isChildBlocked(child.id) ? 'ml-4 pl-3' : '',
              dragOverChildId === child.id ? 'ring-2 ring-purple-500 ring-offset-2 ring-offset-gray-900' : 'opacity-100',
              draggedChildId === child.id ? 'opacity-30' : '',
              depDragOverChildId === child.id && depDragChildId !== child.id ? 'ring-2 ring-orange-500 ring-offset-2 ring-offset-gray-900' : ''
            ]">
            <!-- Reparent indicator -->
            <div v-if="dragOverChildId === child.id && draggedChildId !== child.id" 
              class="absolute inset-0 bg-purple-500/20 flex items-center justify-center z-10 rounded-lg">
              <div class="bg-purple-600 text-white px-3 py-1 rounded-full text-sm font-medium">
                Flytta in i detta mål
              </div>
            </div>
            <!-- Dependency indicator -->
            <div v-if="depDragOverChildId === child.id && depDragChildId !== child.id" 
              class="absolute inset-0 bg-orange-500/20 flex items-center justify-center z-10 rounded-lg">
              <div class="bg-orange-600 text-white px-3 py-1 rounded-full text-sm font-medium">
                🔗 Skapa beroende
              </div>
            </div>
            <!-- Swipe bakgrund -->
            <div class="absolute inset-0 flex items-center justify-start px-6"
              :class="child.finished ? 'bg-red-900/50' : 'bg-transparent'">
              <span class="text-2xl">{{ child.finished ? "↩️" : "" }}</span>
            </div>

            <!-- Huvudinnehåll -->
            <div class="relative rounded-lg transition-all bg-gray-900" :class="{
              'border border-blue-500': selectedChildIndex === index
            }" :style="{
              transform: 'translateX(' + getSwipeOffset(child.id) + 'px)',
              transition: swipeState.isSwiping ? 'none' : 'transform 0.3s ease',
            }"
              draggable="true"
              @dragstart.stop="handleDragStart($event, child.id)" 
              @dragover.stop="handleDragOver($event, child.id)"
              @dragleave.stop="handleDragLeave" 
              @drop.stop.prevent="handleDrop($event, child.id)"
              @touchstart="handleTouchStart($event, child.id)" @touchmove="handleTouchMove($event)"
              @touchend="handleTouchEnd(child)">
              <!-- Insert mode - visa input -->
              <div v-if="mode === 'insert' && editingGoalId === child.id" class="p-4">
                <input ref="editInputRef" v-model="editTitle" type="text"
                  class="w-full px-3 py-2 bg-gray-800 border border-blue-500 rounded text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-400"
                  @keydown.enter.prevent="saveEdit" @keydown.esc.prevent="cancelEdit" />
                <div class="text-xs text-gray-500 mt-2">
                  Enter för att spara, Escape för att avbryta
                </div>
              </div>

              <!-- Normal mode - visa länk -->
              <div v-else class="flex items-center gap-2">
                 <button v-if="!(mode === 'insert' && editingGoalId === child.id)"
                   @click.stop="editingIconGoalId = child.id; showIconPicker = true"
                   :class="['flex-shrink-0 transition-colors rounded p-1', isChildBlocked(child.id) || child.finished ? 'text-gray-600 hover:text-gray-600' : 'text-white hover:text-white hover:bg-gray-600']"
                   title="Ändra ikon">
                    <Icon :name="child.icon || 'heroicons:star'" class="w-5 h-5" :style="isChildBlocked(child.id) || child.finished ? { color: '#4B5563' } : { color: getWeightStyle(child.weight).color, opacity: getWeightStyle(child.weight).opacity }" />
                 </button>
                 <!-- Started-toggle (s) -->
                 <button
                   @click.stop="toggleStarted(child)"
                   :class="['flex-shrink-0 p-1 rounded transition-colors', isChildBlocked(child.id) || child.finished ? 'text-gray-600' : child.started ? 'text-yellow-400 hover:text-yellow-300' : 'text-white hover:text-gray-300']"
                   title="Påbörja (s)">
                   <Icon name="lucide:circle-play" class="w-5 h-5"
                     :style="{ opacity: isChildBlocked(child.id) || child.finished ? 0.6 : child.started ? 1 : 0.25 }" />
                 </button>
                 <!-- Dependency-drag handle -->
                 <button
                   @mousedown.stop="handleDepMouseDown($event, child.id)"
                   :class="[
                     'flex-shrink-0 p-1 rounded transition-colors hover:bg-gray-700',
                     isDepDragging && depDragChildId === child.id ? 'cursor-grabbing' : 'cursor-grab',
                     isChildBlocked(child.id) ? 'text-gray-600' : 'text-white hover:text-orange-400'
                   ]"
                   title="Dra för att skapa beroende">
                   <Icon name="lucide:link" class="w-5 h-5" />
                 </button>
                 <NuxtLink :to="`/goal/${child.id}`" class="flex-1 p-4 block cursor-pointer">
                  <h3 class="text-lg font-normal select-none" :class="child.finished || isChildBlocked(child.id) ? 'text-gray-600' : 'text-white'" :style="child.finished || isChildBlocked(child.id) ? {} : getWeightStyle(child.weight)">
                    {{ child.title }}
                  </h3>
                  <p v-if="getChildWaitingForTitle(child.id)" class="text-xs text-gray-600 mt-1">
                    🔗 Väntar på: {{ getChildWaitingForTitle(child.id) }}
                  </p>
                </NuxtLink>
                <button
                  v-if="weightEditingChildId !== child.id"
                  @click.stop="startWeightEdit(child)"
                  :class="[
                    'flex-shrink-0 transition-colors p-2 rounded',
                    isChildBlocked(child.id) ? 'text-gray-600' : 'text-white hover:text-gray-300 hover:bg-gray-800'
                  ]"
                  title="Ändra vikt">
                  <span :class="['text-xs font-mono px-2 py-1 rounded bg-gray-800', isChildBlocked(child.id) ? 'text-gray-600' : 'text-white']">
                    {{ child.weight }}
                  </span>
                </button>
                <div v-if="weightEditingChildId === child.id" class="px-4 pb-4">
                  <div class="flex items-center gap-2">
                    <input v-model.number="tempWeight" type="range" min="1" max="200" step="1" class="flex-1" />
                    <span class="text-sm text-gray-400 w-8">{{ tempWeight }}</span>
                    <button @click="saveWeight" class="text-green-400 hover:text-green-300">✓</button>
                    <button @click="cancelWeightEdit" class="text-red-400 hover:text-red-300">✗</button>
                  </div>
                </div>
              </div>
            </div>
          </li>
        </ul>

        <div v-else class="text-gray-500 p-6 border border-gray-700 rounded-lg text-center">
          Inga undermål ännu. Skapa ett för att dela upp detta mål i mindre
          delar.
        </div>
      </div>
    </div>

    <!-- Vim mode indikator -->
    <div v-if="mode === 'insert'"
      class="fixed bottom-4 left-1/2 transform -translate-x-1/2 px-4 py-2 bg-blue-500 text-white font-mono text-sm rounded shadow-lg">
      -- INSERT --
    </div>

    <!-- Modal för att bekräfta borttagning av föräldrarelation -->
    <UModal v-model:open="showRemoveConfirmation" title="Ta bort föräldrarelation?">
      <p v-if="parentToRemove" class="text-gray-400">
        Vill du ta bort relationen till
        <strong>{{
          parents.find((p) => p.id === parentToRemove)?.title
        }}</strong>? Själva målet kommer inte tas bort.
      </p>

      <template #footer="{ close }">
        <div class="flex justify-end gap-2">
          <UButton color="gray" variant="ghost" @click="cancelRemoveParent">
            Avbryt
          </UButton>
          <UButton color="red" @click="confirmRemoveParent">
            Ta bort relation
          </UButton>
        </div>
      </template>
    </UModal>

    <!-- Modal för att bekräfta borttagning av mål -->
    <UModal v-model:open="showDeleteConfirmation" title="Ta bort mål?">
      <template #footer="{ close }">
        <div class="flex justify-end gap-2">
          <UButton color="gray" variant="ghost" @click="cancelDelete" :class="deleteDialogSelection === 'cancel' ? 'ring-2 ring-blue-500' : ''
            ">
            Avbryt
          </UButton>
          <UButton color="red" @click="executeDelete" :class="deleteDialogSelection === 'confirm' ? 'ring-2 ring-blue-400' : ''
            ">
            Ta bort
          </UButton>
        </div>
        <!-- <div class="text-xs text-gray-500 mt-3 text-center"> -->
        <!--   h/l = navigera, Enter = Bekräfta, Escape = Avbryt -->
        <!-- </div> -->
      </template>
    </UModal>

    <!-- Icon Picker Component -->
    <IconPicker v-if="goal"
      :modelValue="editingIconGoalId ? goalData?.allGoals?.find(g => g.id === editingIconGoalId)?.icon || '' : goal.icon"
      :open="showIconPicker" @update:open="(value) => {
        showIconPicker = value;
        if (!value) {
          editingIconGoalId = null;
        }
      }" @update:modelValue="updateGoalIcon" />

    <!-- Leader Key Modal -->
    <LeaderKeyModal
      :open="isLeaderModalOpen"
      @execute-command="executeLeaderCommand($event)"
    />
  </div>
  <div v-else class="max-w-4xl mx-auto p-6">
    <div class="text-center">
      <h1 class="text-3xl font-bold text-gray-300 mb-4">Du måste logga in</h1>
      <p class="text-gray-600 mb-6">
        För att se dina mål måste du vara inloggad.
      </p>
      <NuxtLink to="/login">
        <UButton>Logga in</UButton>
      </NuxtLink>
    </div>
  </div>
</template>
