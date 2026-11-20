<script setup lang="ts">
import LeaderKeyModal from '~/components/LeaderKeyModal.vue'
import GoalHeader from '~/components/GoalHeader.vue'
import GoalParentSection from '~/components/GoalParentSection.vue'
import GoalChildrenList from '~/components/GoalChildrenList.vue'
import GoalWeightVerticalEditor from '~/components/GoalWeightVerticalEditor.vue'
import GoalChildSearch from '~/components/GoalChildSearch.vue'
import ConfirmDeleteGoalModal from '~/components/ConfirmDeleteGoalModal.vue'
import ConfirmRemoveParentModal from '~/components/ConfirmRemoveParentModal.vue'
import type { Goal, GoalWithWeight, GetGoalResponse } from '~/types/goal'
import type { Place } from '~/types/place'
import { insertChildId, newChildInsertIndex, nextChildOrder } from '~/utils/childOrder'
import { buildDependencyTree, moveChildInTree } from '~/utils/childDependencyOrder'
import { buildGoalData } from '~/utils/goalGraph'
import { canBeFinished } from '~/utils/goalRules'
import { descriptionPreview } from '~/utils/goalDescription'
import { dueTone, formatDueDate, isFutureDate, todayDateString } from '~/utils/dueDate'
import { describeGoalRecurrence, isRecurringGoal, isWaitingForNextOccurrence, recurrenceCount } from '~/utils/recurrence'
import type { RecurrenceMode } from '~/utils/recurrence'
import { useGoalDrag } from '~/composables/useGoalDrag'
import { ETERNAL_ACCENT, ETERNAL_ACCENT_HOVER, ETERNAL_COLOR, ETERNAL_TEXT, ETERNAL_TEXT_HOVER } from '~/utils/eternalStyle'
import { normalizeGoalWeights } from '~/utils/goalWeights'

// Make this page client-only to ensure GraphQL requests work properly and prevent Pinia hydration issues
definePageMeta({
  ssr: false,
});

const route = useRoute();
const router = useRouter();
const goalId = computed(() => parseInt(route.params.id as string));
const { user } = useUserSession();
const { updateGoalTitle, updateGoalIcon: updateGoalIconApi, toggleGoalFinished, toggleGoalStarted, setGoalEternal, setGoalAgent, setGoalDue, setGoalStartDate, setGoalEstimate, deleteGoal, addParentRelation, removeParentRelation, addChildRelation, reorderChildren, updateGoalWeights, createGoal, addDependency, removeDependency, updateGoalDescription, setGoalRecurrence, completeGoalOccurrence } = useGoalApi();
const { startDrag: startGoalDrag } = useGoalDrag();

// Ångra-stapeln - tangenten "z" kan ångra handlingar från alla sidor.
const { revision: undoRevision } = useUndoStack();
// Snabbtillägget kan skapa ett nytt barn från vilken sida som helst.
const { revision: quickAddRevision } = useQuickAdd();
const { undoFinished, undoOccurrence, undoStarted, undoEternal, undoAgent, undoDue, undoStartDate, undoEstimate, undoRecurrence } = useGoalUndo();
const goalDates = useGoalDatesModal();

// Använd goals store
const goalsStore = useGoalsStore();

// Ordningen som ska sparas till servern.
//
// Deklareras här (inte vid persistChildOrder) eftersom hydrate() kör under
// setup och kan behöva spara en ny ordning direkt.
let orderSaveChain: Promise<void> = Promise.resolve();
let pendingOrder: number[] | null = null;
let isSavingOrder = false;

// Hämta målet med dess relationer
const goalData = ref<GetGoalResponse | null>(null);
const isLoading = ref(true);
const error = ref<string | null>(null);

// Sätter sidans data ur grafen.
function hydrate() {
  const data = buildGoalData(
    {
      goals: goalsStore.goals,
      relations: goalsStore.relations,
      dependencies: goalsStore.dependencies,
    },
    goalId.value,
  );
  if (!data) {
    goalData.value = null;
    error.value = "Målet hittades inte";
    return;
  }

  goalData.value = data;
  error.value = null;

  // Ordningen i listan ska följa beroendena - ett undermål som väntar på
  // ett annat undermål i samma lista ligger direkt under det.
  enforceDependencyOrder();
}

// Skriver om ordningen så att ett undermål som väntar på ett annat
// undermål i samma lista ligger direkt under det.
//
// Listan visar order-fältet, så ordningen måste ändras i fältet - annars
// skulle listan visa en ordning som inte gäller när den sparas.
function enforceDependencyOrder() {
  const list = goalData.value?.children;
  if (!list || list.length < 2) {
    return;
  }

  const currentIds = list.map((c) => c.id);
  const nextIds = buildDependencyTree(list).map((item) => item.child.id);

  if (currentIds.every((id, index) => id === nextIds[index])) {
    return;
  }

  applyLocalOrder(nextIds);
  persistChildOrder();
}

// Hämtar grafen och bygger om sidans data.
//
// force = true hämtar alltid om från servern - det används efter en
// ändring. Annars används den graf som redan finns, vilket går på
// millisekunder.
const refresh = async (force = true) => {
  isLoading.value = true;
  error.value = null;

  try {
    await goalsStore.loadGraph(force);
    hydrate();
  } catch (err) {
    console.error("Fetch goal error:", err);
    error.value = err instanceof Error ? err.message : "Unknown error";
  } finally {
    isLoading.value = false;
  }
};

// Ett mål som skapats i snabbfältet och fått denna sida som förälder finns
// redan i den lokala grafen. Visa det direkt och synka sedan mot servern.
watch(quickAddRevision, () => {
  hydrate();
  void refresh(true);
});

// Sidans data byggs ur grafen. Layouten laddar grafen redan vid start,
// så här är det oftast ingenting att hämta.
await refresh(false);

// Hämtar om grafen i bakgrunden om den börjar bli gammal.
//
// Det håller data färska om något ändrats i en annan webbläsare, utan att
// navigeringen behöver vänta på servern.
function scheduleRevalidate() {
  if (Date.now() - goalsStore.loadedAt < 15000) {
    return;
  }

  void goalsStore
    .loadGraph(true)
    .then(() => hydrate())
    .catch(() => {
      // Tyst - navigeringen fungerar ändå med den data vi redan har.
    });
}

// Hämta användarens platser - de behövs för att kunna koppla platser
// till mål.
const { places: allPlaces, fetchPlaces, fetchGoalPlaces, setGoalPlaces } = usePlaces();
await fetchPlaces().catch(() => {
  console.error('Failed to fetch places');
});

// Målet vars platser just nu redigeras. Null = ingen öppen platsväljare.
const placeTargetId = ref<number | null>(null);
const placeDraft = ref<Place[]>([]);

// Öppnar eller stänger platsväljaren för ett mål.
async function togglePlacePicker(targetId: number) {
  if (placeTargetId.value === targetId) {
    closePlacePicker();
    return;
  }
  try {
    placeDraft.value = await fetchGoalPlaces(targetId);
    placeTargetId.value = targetId;
  } catch (error) {
    console.error('Failed to load places for goal:', error);
  }
}

function closePlacePicker() {
  placeTargetId.value = null;
  placeDraft.value = [];
}

// Sparar platserna på målet och uppdaterar den lokala datan.
async function savePlaces(targetId: number) {
  try {
    const saved = await setGoalPlaces(targetId, placeDraft.value.map((p) => p.id));
    applyPlaces(targetId, saved);
    closePlacePicker();
  } catch (error) {
    console.error('Failed to save places for goal:', error);
  }
}

// Skriver tillbaka platserna i den hämtade måldatan.
function applyPlaces(targetId: number, saved: Place[]) {
  const data = goalData.value;
  if (!data) return;
  if (data.goal?.id === targetId) data.goal.places = saved;
  const child = data.children?.find((c) => c.id === targetId);
  if (child) child.places = saved;
  const parent = data.parents?.find((p) => p.id === targetId);
  if (parent) parent.places = saved;
}

// Byt målsida när ett nytt mål väljs - utan serverfråga, grafen räcker.
watch(
  () => route.params.id,
  async () => {
    hydrate();
    isLoading.value = false;
    scheduleRevalidate();
  }
);

const goal = computed(() => goalData.value?.goal);

// Extrahera alla föräldrar (REST API format)
const parents = computed(
  () => goalData.value?.parents || []
);

// Barn med vikt och ordning (REST API returnerar detta direkt).
//
// Listan visar order-fältet - den sorteras inte om i vyn. Beroenden
// påverkar i stället order-fältet (se enforceDependencyOrder), så att
// piltangenternas flyttar och det som sparas alltid är samma sak.
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

// Indenteringen kommer från trädet: ett undermål som väntar på ett annat
// undermål ligger direkt under det, och indenteras en nivå djupare.
function childIndentStyle(childId: number): Record<string, string> {
  const level = childLevels.value[childId] ?? 0;
  if (level === 0) return {};
  return { marginLeft: `${level}rem`, paddingLeft: '0.75rem' };
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

// Beroendeväljaren (kortkommandot 'w').
//
// Den gör det markerade målet (förälder, huvudmål eller undermål)
// beroende av ett annat mål i systemet. Sökningen sker bland alla
// mål i grafen - inte bara bland målets egna barn.
const dependencyPickerOpen = ref(false);
const dependencyPickerTargetId = ref<number | null>(null);
const dependencyPickerTargetTitle = ref("");

function openDependencyPicker(target: Goal | null) {
  if (!target) return;
  dependencyPickerTargetId.value = target.id;
  dependencyPickerTargetTitle.value = target.title;
  dependencyPickerOpen.value = true;
}

async function addDependencyFromPicker(dependsOnId: number) {
  const targetId = dependencyPickerTargetId.value;
  dependencyPickerOpen.value = false;
  if (targetId === null) return;

  try {
    await addDependency(targetId, dependsOnId);
    await refresh();
    // Visa beroendesektionen så att resultatet syns direkt.
    showDependencies.value = true;
  } catch (error: any) {
    console.error("Failed to add dependency:", error);
    alert(error?.data?.message || error?.message || "Kunde inte lägga till beroende");
  }
}


// Visa/dölj avklarade mål
const showCompleted = ref(false);
const showStarted = ref(false);
const weightMode = ref(false);
const weightSelectedIndex = ref(0);
let weightSaveChain: Promise<void> = Promise.resolve();

// Icon picker state
const showIconPicker = ref(false);
const editingIconGoalId = ref<number | null>(null);
const editingIconGoalIcon = ref<string>('');



// Undermål som syns, filtrerade på showCompleted och showStarted.
const visibleChildren = computed(() => {
  let result = children.value;
  if (!showCompleted.value) {
    result = result.filter((c) => !c.finished);
  }
  if (showStarted.value) {
    result = result.filter((c) => c.started !== null);
  }
  return result;
});

// Trädet av de synliga undermålen.
//
// Ett undermål som väntar på ett annat undermål i samma lista ligger
// direkt under det - som ett underelement. Då hänger indenteringen ihop.
const visibleTree = computed(() => buildDependencyTree(visibleChildren.value));

// Undermålen i den ordning de visas: trädets ordning.
const filteredChildren = computed(() => visibleTree.value.map((item) => item.child));

// Viktläget arbetar bara med pågående undermål. Slutförda mål ska inte
// ta plats i fördelningsvyn.
const weightChildren = computed(() => children.value.filter((child) => !child.finished));

// Nivå i trädet per undermål - styr indenteringen i listan.
//
// Ett undermål som väntar på ett mål utanför listan får en nivå, så att
// det fortfarande syns att det väntar på något.
const childLevels = computed(() => {
  const levels: Record<number, number> = {};

  for (const item of visibleTree.value) {
    const blocked = isChildBlocked(item.child.id);
    levels[item.child.id] = item.depth > 0 ? item.depth : blocked ? 1 : 0;
  }

  return levels;
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

// Nästa lediga ordningsnummer för ett nytt undermål
function nextChildOrderNumber(): number {
  return nextChildOrder(children.value.map((c) => c.order ?? 0));
}

// Var ett nytt undermål ska placeras i listan.
//
// Det läggs under det markerade undermålet. Om huvudmålet är markerat
// läggs det först i listan - alltså som barn nr 1.
function newChildInsertIndexNumber(): number {
  return newChildInsertIndex(
    children.value.map((c) => c.id),
    filteredChildren.value.map((c) => c.id),
    selectedChildIndex.value,
    isGoalSelected.value,
  );
}

// Lägg till ett undermål på rätt plats i listan.
//
// Relationen får först ett unikt ordningsnummer (så att den inte krockar
// med ett befintligt barn), och sedan sätts hela ordningen om i ett svep -
// då hamnar det nya målet under det markerade målet.
async function addChildAtPosition(childId: number): Promise<void> {
  const insertIndex = newChildInsertIndexNumber();
  const order = insertChildId(
    children.value.map((c) => c.id),
    insertIndex,
    childId,
  );

  await addChildRelation(childId, goalId.value, nextChildOrderNumber());
  await reorderChildren(goalId.value, order);
}

// Lägg till befintligt mål som barn
async function addExistingChild(childId: number) {
  try {
    await addChildAtPosition(childId);
    goalsStore.addRelation(childId, goalId.value);
    // Bygg om sidan från den lokala grafen direkt. Servern verifieras sedan
    // i bakgrunden, men användaren ska inte behöva vänta på eller uppdatera sidan.
    hydrate();
    showChildSearch.value = false;
    childSearchQuery.value = "";
    void refresh(true);
  } catch (error) {
    console.error("Failed to add child:", error);
  }
}

  // Skapa nytt mål som barn
  async function createNewChild() {
    if (!childSearchQuery.value.trim()) return;

    try {
      const newGoal = await createGoal(childSearchQuery.value.trim());
      await addChildAtPosition(newGoal.id);
      goalsStore.addGoal({ ...newGoal, has_parent: true });
      goalsStore.addRelation(newGoal.id, goalId.value);
      // Visa det nya barnet från den lokala grafen direkt. En bakgrundsrefresh
      // synkroniserar sedan ordning och övriga fält mot servern.
      hydrate();
      showChildSearch.value = false;
      childSearchQuery.value = "";
      void refresh(true);
      // Markera det nya målet direkt
      await nextTick();
      const newChildIndex = filteredChildren.value.findIndex(c => c.id === newGoal.id);
      if (newChildIndex !== -1) {
        selectedChildIndex.value = newChildIndex;
        isGoalSelected.value = false;
        isParentMode.value = false;
      }
    } catch (error) {
      console.error("Failed to create new child:", error);
    }
  }

  // Hantera Enter-tangent för undermål
  function handleChildSearchKeydown(event: KeyboardEvent) {
    if (event.key !== "Enter") return;
    event.preventDefault();
    event.stopPropagation();
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

  // Samma dragning kan både användas för att sortera/reparenta barnet på
  // målsidan och för att släppa barnet i kalendern.
  const child = children.value.find((candidate) => candidate.id === childId);
  if (child) {
    startGoalDrag(event, child);
  }

  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = "copyMove";
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

  // Släppet öppnar ett val: byta förälder eller skapa ett beroende.
  const draggedId = draggedChildId.value;
  draggedChildId.value = null;
  dropChoice.value = { draggedId, targetId: dropChildId };
}

// Reparent: ta bort från nuvarande förälder och lägg till som barn till targetId
async function reparentChild(draggedId: number, targetId: number) {
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

    await addChildRelation(draggedId, targetId, targetChildrenCount);
    goalsStore.addRelation(draggedId, targetId);

    await refresh();
  } catch (error) {
    console.error("Failed to reparent goal:", error);
    await refresh();
  }
}

// Skapa eller ta bort beroendet mellan två barnmål.
async function toggleDependency(draggedId: number, targetId: number) {
  const childInfo = childDependencies.value[draggedId];
  const existingDep = childInfo?.dependsOn?.find((d: any) => d.id === targetId);

  if (existingDep) {
    await removeDependency(draggedId, targetId);
  } else {
    await addDependency(draggedId, targetId);
  }
  await refresh();
}

// Valet efter ett släpp: byta förälder eller skapa beroende.
const dropChoice = ref<{ draggedId: number; targetId: number } | null>(null);

const dropChoiceTarget = computed(() => {
  if (!dropChoice.value) return null;
  return children.value.find((c) => c.id === dropChoice.value!.targetId) ?? null;
});

const dropChoiceHasDependency = computed(() => {
  if (!dropChoice.value) return false;
  const info = childDependencies.value[dropChoice.value.draggedId];
  return !!info?.dependsOn?.some((d: any) => d.id === dropChoice.value!.targetId);
});

function cancelDropChoice() {
  dropChoice.value = null;
}

async function applyDropAsChild() {
  const choice = dropChoice.value;
  if (!choice) return;
  dropChoice.value = null;
  await reparentChild(choice.draggedId, choice.targetId);
}

async function applyDropAsDependency() {
  const choice = dropChoice.value;
  if (!choice) return;
  dropChoice.value = null;
  await toggleDependency(choice.draggedId, choice.targetId);
}

// Flytta det markerade barnet upp eller ner i listan.
//
// Ett mål på översta nivån flyttas tillsammans med sina undermål - de
// hör ihop. Ett undermål flyttas bara inom sin grupp, så att det aldrig
// hamnar utanför det mål det väntar på.
//
// Ordningen uppdateras lokalt först (optimistiskt) så att tangenttryck
// direkt ger respons, och sparas sedan till servern. Sparandet serialiseras
// och slås ihop, så att snabba tangenttryck inte skriver över varandra.
function moveChild(direction: "up" | "down") {
  if (isParentMode.value || isGoalSelected.value) {
    return;
  }

  const visible = filteredChildren.value;
  const index = selectedChildIndex.value;

  if (index < 0 || index >= visible.length) {
    return;
  }

  const movingChild = visible[index];
  const targetIndex = direction === "up" ? index - 1 : index + 1;

  const allIds = children.value.map((c) => c.id);
  const reordered = moveChildInTree(allIds, visibleTree.value, movingChild.id, direction);

  if (!reordered) {
    return;
  }

  applyLocalOrder(reordered);
  selectedChildIndex.value = targetIndex;

  persistChildOrder();
}

// Sätter den nya ordningen lokalt utan att vänta på servern
function applyLocalOrder(ids: number[]) {
  const data = goalData.value;
  if (!data?.children) {
    return;
  }

  const byId = new Map(data.children.map((c) => [c.id, c]));
  const reordered = ids
    .map((id) => byId.get(id))
    .filter((c): c is GoalWithWeight => Boolean(c));

  reordered.forEach((child, index) => {
    child.order = index;
  });

  data.children = reordered;

  // Grafen håller den nya ordningen, så att sidan inte hoppar tillbaka
  // till den gamla ordningen innan servern svarat.
  goalsStore.setChildOrder(goalId.value, ids);
}

// Sparar ordningen till servern. Anropen körs i kedjan och senaste ordningen
// vinner - flera snabba tangenttryck slås ihop till ett anrop.
function persistChildOrder() {
  pendingOrder = goalData.value?.children?.map((c) => c.id) ?? [];

  if (isSavingOrder) {
    return;
  }

  isSavingOrder = true;
  orderSaveChain = orderSaveChain.then(flushChildOrder);
}

async function flushChildOrder() {
  try {
    while (pendingOrder) {
      const ids = pendingOrder;
      pendingOrder = null;

      try {
        await reorderChildren(goalId.value, ids);
      } catch (error) {
        console.error("Failed to save child order:", error);
        await refresh(true);
      }
    }
  } finally {
    isSavingOrder = false;
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

// Redigera beskrivning (t)
const editingDescriptionId = ref<number | null>(null);
const descriptionDraft = ref("");

function startEditDescription(target: { id: number; description?: string }) {
  editingDescriptionId.value = target.id;
  descriptionDraft.value = target.description ?? "";

  nextTick(() => {
    const el = document.querySelector(
      `[data-description-input="${target.id}"]`
    ) as HTMLTextAreaElement | null;
    el?.focus();
  });
}

function cancelEditDescription() {
  editingDescriptionId.value = null;
  descriptionDraft.value = "";
}

function handleDescriptionKeydown(event: KeyboardEvent, id: number) {
  if (event.key !== "Enter" || (!event.ctrlKey && !event.metaKey)) return;

  event.preventDefault();
  event.stopPropagation();
  void saveDescription(id);
}

async function saveDescription(id: number) {
  const description = descriptionDraft.value;

  try {
    await updateGoalDescription(id, description);

    if (goalData.value) {
      if (goalData.value.goal.id === id) {
        goalData.value.goal.description = description;
      }
      const child = goalData.value.children.find((c) => c.id === id);
      if (child) {
        child.description = description;
      }
    }
    goalsStore.updateGoal(id, { description });

    cancelEditDescription();
  } catch (error) {
    console.error("Failed to update goal description:", error);
  }
}

// Delete confirmation
const showDeleteConfirmation = ref(false);
const goalToDelete = ref<{ id: number; title: string } | null>(null);
const deleteDialogSelection = ref<"cancel" | "confirm">("cancel");

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
        "input.border-accent"
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
  // Eviga mål pågår hela livet och kan inte slutföras.
  if (!canBeFinished(goalToToggle)) {
    return;
  }

  // Ett återkommande mål slutförs en förekomst i taget. Målet ligger kvar
  // i listan och väntar på nästa upprepning.
  if (isRecurringGoal(goalToToggle)) {
    // Förekomsten har inte kommit än - då finns inget att checka av.
    if (isWaitingForNextOccurrence(goalToToggle)) {
      return;
    }
    try {
      await completeGoalOccurrence(goalToToggle.id, goalToToggle.recurrence_next ?? undefined, todayDateString());
      undoOccurrence(goalToToggle);
      await refresh();
    } catch (error) {
      console.error("Failed to complete goal occurrence:", error);
    }
    return;
  }

  try {
    const previousFinished = goalToToggle.finished;
    const newFinishedValue = previousFinished ? null : new Date().toISOString();

    await toggleGoalFinished(goalToToggle.id, newFinishedValue);
    undoFinished(goalToToggle, previousFinished);
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

// Upprepning för ett mål. Målets förekomster loggas var för sig, så
// målet kan checkas av och återkomma därefter.
const recurrenceTargetId = ref<number | null>(null);

function recurrenceTarget(): Goal | null {
  if (recurrenceTargetId.value === null) return null;

  if (goalData.value?.goal && goalData.value.goal.id === recurrenceTargetId.value) {
    return goalData.value.goal;
  }

  const child = filteredChildren.value.find((c) => c.id === recurrenceTargetId.value);
  if (child) return child;

  const parent = parents.value.find((p) => p.id === recurrenceTargetId.value);
  return parent ?? null;
}

function openRecurrenceEditor(target: Goal) {
  recurrenceTargetId.value = target.id;
}

function closeRecurrenceEditor() {
  recurrenceTargetId.value = null;
}

async function saveRecurrence(rule: string, mode: RecurrenceMode) {
  const target = recurrenceTarget();
  if (!target) return;

  try {
    const previousRule = target.recurrence_rule;
    const previousMode = target.recurrence_mode;
    await setGoalRecurrence(target.id, rule, mode);
    undoRecurrence(target, previousRule, previousMode);
    closeRecurrenceEditor();
    await refresh();
  } catch (error) {
    console.error("Failed to save goal recurrence:", error);
  }
}

async function removeRecurrence() {
  const target = recurrenceTarget();
  if (!target) return;

  try {
    const previousRule = target.recurrence_rule;
    const previousMode = target.recurrence_mode;
    await setGoalRecurrence(target.id, null);
    undoRecurrence(target, previousRule, previousMode);
    closeRecurrenceEditor();
    await refresh();
  } catch (error) {
    console.error("Failed to remove goal recurrence:", error);
  }
}

// Markera ett mål som evigt (eller upphäva evigt-flaggan).
// Ett evigt mål kan inte slutföras - det markeras bara som pågående.
async function toggleEternal(goalToToggle: Goal) {
  const newEternal = !goalToToggle.eternal;

  try {
    const previousEternal = goalToToggle.eternal;
    await setGoalEternal(goalToToggle.id, newEternal);
    undoEternal(goalToToggle, previousEternal);
    goalsStore.updateGoal(goalToToggle.id, { eternal: newEternal });

    // Uppdatera lokal state i component
    if (goalData.value?.allGoals) {
      const goalIndex = goalData.value.allGoals.findIndex(
        (g) => g.id === goalToToggle.id
      );
      if (goalIndex !== -1) {
        goalData.value.allGoals[goalIndex].eternal = newEternal;
      }
    }
    if (goalData.value?.goal && goalData.value.goal.id === goalToToggle.id) {
      goalData.value.goal.eternal = newEternal;
    }

    // Uppdatera data från server
    await refresh();
  } catch (error) {
    console.error("Failed to toggle eternal status:", error);
  }
}

// Markera ett mål som ett mål för en agent (eller upphäva agent-flaggan).
async function toggleAgent(goalToToggle: Goal) {
  const newAgent = !goalToToggle.agent;

  try {
    const previousAgent = goalToToggle.agent ?? false;
    await setGoalAgent(goalToToggle.id, newAgent);
    undoAgent(goalToToggle, previousAgent);
    goalsStore.updateGoal(goalToToggle.id, { agent: newAgent });

    if (goalData.value?.allGoals) {
      const goalIndex = goalData.value.allGoals.findIndex(
        (g) => g.id === goalToToggle.id
      );
      if (goalIndex !== -1) {
        goalData.value.allGoals[goalIndex].agent = newAgent;
      }
    }
    if (goalData.value?.goal && goalData.value.goal.id === goalToToggle.id) {
      goalData.value.goal.agent = newAgent;
    }

    await refresh();
  } catch (error) {
    console.error("Failed to toggle agent status:", error);
  }
}

// Det mål som tangenterna och leader-kommandona gäller.
//
// Samma prioritet som övriga tangenter: föräldraläge först, sedan det
// markerade huvudmålet, annars det markerade undermålet.
function agentTarget(): Goal | null {
  if (isParentMode.value && parents.value.length > 0) {
    return parents.value[selectedParentIndex.value] ?? null;
  }
  if (isGoalSelected.value && goal.value) {
    return goal.value;
  }
  if (
    filteredChildren.value.length > 0 &&
    selectedChildIndex.value >= 0 &&
    selectedChildIndex.value < filteredChildren.value.length
  ) {
    return filteredChildren.value[selectedChildIndex.value] ?? null;
  }
  return null;
}

// Due-datum (deadline) för ett mål. Datumet är ett kalenderdatum
// ("YYYY-MM-DD") - tom sträng rensar det.
//
// dueTargetId anger VILKET mål väljaren är öppen för, så att tangenten
// 'c' kan följa det markerade målet (förälder, huvudmål eller undermål)
// och väljaren visas bredvid just det målet.
const dueTargetId = ref<number | null>(null);
const dueDraft = ref('');
const dueInputRef = ref<HTMLInputElement | null>(null);

function dueToneClass(due: string | null): string {
	const tone = dueTone(due);
	if (tone === 'overdue') return 'text-red-400';
	if (tone === 'today') return 'text-amber-400';
	return 'text-ink-muted';
}

function openDuePicker(target: Goal) {
	dueDraft.value = target.due ?? '';
	dueTargetId.value = target.id;

	nextTick(() => {
		const input = document.querySelector(
			`[data-due-input="${target.id}"]`,
		) as (HTMLInputElement & { showPicker?: () => void }) | null;

		input?.focus();
		// showPicker öppnar den inbyggda kalender-popupen direkt. Äldre
		// webbläsare saknar API:t, så fokus är fallback där.
		try {
			input?.showPicker?.();
		} catch {
			// Webbläsaren kan neka showPicker utan user activation.
		}
	});
}

function closeDuePicker() {
	dueTargetId.value = null;
	dueDraft.value = '';
}

async function saveDue(target: Goal, value: string) {
	const due = value ? value : null;

	try {
		const previousDue = target.due;
		await setGoalDue(target.id, due);
		undoDue(target, previousDue);
		applyDue(target.id, due);
		closeDuePicker();
		await refresh();
	} catch (error) {
		console.error('Failed to save goal due date:', error);
	}
}

function applyDue(id: number, due: string | null) {
	goalsStore.updateGoal(id, { due });

	if (goalData.value?.allGoals) {
		const goalIndex = goalData.value.allGoals.findIndex((g: Goal) => g.id === id);
		if (goalIndex !== -1) {
			goalData.value.allGoals[goalIndex].due = due;
		}
	}
	if (goalData.value?.goal && goalData.value.goal.id === id) {
		goalData.value.goal.due = due;
	}
	if (goalData.value?.children) {
		const childIndex = goalData.value.children.findIndex((c) => c.id === id);
		if (childIndex !== -1) {
			goalData.value.children[childIndex].due = due;
		}
	}
	if (goalData.value?.parents) {
		const parentIndex = goalData.value.parents.findIndex((p) => p.id === id);
		if (parentIndex !== -1) {
			goalData.value.parents[parentIndex].due = due;
		}
	}
}

// Startdatum för ett mål - det datum då målet tidigast får dyka upp i
// prioriteringslistan. Det är ett kalenderdatum ("YYYY-MM-DD") - tom
// sträng rensar det.
//
// startTargetId anger VILKET mål väljaren är öppen för, så att tangenten
// 'b' kan följa det markerade målet (förälder, huvudmål eller undermål).
const startTargetId = ref<number | null>(null);
const startDraft = ref('');

function openStartPicker(target: Goal) {
  startDraft.value = target.start_date ?? '';
  startTargetId.value = target.id;

  nextTick(() => {
    const input = document.querySelector(
      `[data-start-input="${target.id}"]`,
    ) as (HTMLInputElement & { showPicker?: () => void }) | null;

    input?.focus();
    // showPicker öppnar den inbyggda kalender-popupen direkt. Äldre
    // webbläsare saknar API:t, så fokus är fallback där.
    try {
      input?.showPicker?.();
    } catch {
      // Webbläsaren kan neka showPicker utan user activation.
    }
  });
}

function closeStartPicker() {
  startTargetId.value = null;
  startDraft.value = '';
}

async function saveStartDate(target: Goal, value: string) {
  const startDate = value ? value : null;

  try {
    const previousStartDate = target.start_date ?? null;
    await setGoalStartDate(target.id, startDate);
    undoStartDate(target, previousStartDate);
    applyStartDate(target.id, startDate);
    closeStartPicker();
    await refresh();
  } catch (error) {
    console.error('Failed to save goal start date:', error);
  }
}

function applyStartDate(id: number, startDate: string | null) {
  goalsStore.updateGoal(id, { start_date: startDate });

  if (goalData.value?.allGoals) {
    const goalIndex = goalData.value.allGoals.findIndex((g: Goal) => g.id === id);
    if (goalIndex !== -1) {
      goalData.value.allGoals[goalIndex].start_date = startDate;
    }
  }
  if (goalData.value?.goal && goalData.value.goal.id === id) {
    goalData.value.goal.start_date = startDate;
  }
  if (goalData.value?.children) {
    const childIndex = goalData.value.children.findIndex((c) => c.id === id);
    if (childIndex !== -1) {
      goalData.value.children[childIndex].start_date = startDate;
    }
  }
  if (goalData.value?.parents) {
    const parentIndex = goalData.value.parents.findIndex((p) => p.id === id);
    if (parentIndex !== -1) {
      goalData.value.parents[parentIndex].start_date = startDate;
    }
  }
}

// Uppskattad tidsåtgång för ett mål.
//
// estimateTargetId anger VILKET mål fältet är öppet för, så att tangenten
// 'm' kan följa det markerade målet (förälder, huvudmål eller undermål).
const estimateTargetId = ref<number | null>(null);

function openEstimateEditor(target: Goal) {
	estimateTargetId.value = target.id;
}

function closeEstimateEditor() {
	estimateTargetId.value = null;
}

async function saveEstimate(target: Goal, minutes: number | null) {
	try {
		const previous = target.estimated_minutes ?? null;
		await setGoalEstimate(target.id, minutes);
		undoEstimate(target, previous);
		goalsStore.updateGoal(target.id, { estimated_minutes: minutes });
		closeEstimateEditor();
		await refresh();
	} catch (error) {
		console.error('Failed to save goal estimate:', error);
	}
}

// Toggla started status på ett mål
async function toggleStarted(goalToToggle: Goal) {
  try {
    const previousStarted = goalToToggle.started;
    const newStartedValue = previousStarted ? null : new Date().toISOString();
    await toggleGoalStarted(goalToToggle.id, newStartedValue);
    undoStarted(goalToToggle, previousStarted);
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

function getWeightStyle(weight: number): { color: string; opacity: number; fontWeight?: string } {
  return { color: '#FFFFFF', opacity: 1 }
}

function weightsForPersistence(activeWeights: Array<{ childId: number; weight: number }>) {
  const allWeights = normalizeGoalWeights(children.value)
  const activeIds = new Set(activeWeights.map((item) => item.childId))
  const hidden = allWeights.filter((item) => !activeIds.has(item.childId))
  const available = 100 - hidden.reduce((sum, item) => sum + item.weight, 0)
  const raw = activeWeights.map((item) => (item.weight / 100) * available)
  const scaled = raw.map(Math.floor)
  let remainder = available - scaled.reduce((sum, weight) => sum + weight, 0)

  raw
    .map((value, index) => ({ index, fraction: value - scaled[index] }))
    .sort((a, b) => b.fraction - a.fraction)
    .forEach(({ index }) => {
      if (remainder <= 0) return
      scaled[index] += 1
      remainder -= 1
    })

  return [
    ...hidden,
    ...activeWeights.map((item, index) => ({ childId: item.childId, weight: scaled[index] })),
  ]
}

function applyWeights(weights: Array<{ childId: number; weight: number }>) {
  const data = goalData.value
  if (!data) return

  for (const weight of weights) {
    const child = data.children.find((item) => item.id === weight.childId)
    if (child) child.weight = weight.weight

    const relation = data.goal.childRelations.find((item) => item.child_id === weight.childId)
    if (relation) relation.weight = weight.weight
  }
}

function previewWeights(weights: Array<{ childId: number; weight: number }>) {
  applyWeights(weightMode.value ? weightsForPersistence(weights) : weights)
}

function persistWeights(weights: Array<{ childId: number; weight: number }>) {
  weightSaveChain = weightSaveChain
    .then(() => updateGoalWeights(goalId.value, weights))
    .catch(async (error) => {
      console.error("Failed to update goal weights:", error)
      await refresh()
    })
  return weightSaveChain
}

async function saveWeights(weights: Array<{ childId: number; weight: number }>) {
  const persisted = weightMode.value ? weightsForPersistence(weights) : weights
  applyWeights(persisted)
  await persistWeights(persisted)
}

function toggleWeightMode() {
  weightMode.value = !weightMode.value
  if (weightMode.value) {
    const selectedId = filteredChildren.value[selectedChildIndex.value]?.id
    const index = weightChildren.value.findIndex((child) => child.id === selectedId)
    weightSelectedIndex.value = index >= 0 ? index : 0
  }
}

// Öppna ikonväljaren för det som är markerat (mål, förälder eller barn)
function openIconPickerForSelection() {
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
}

// Kommandon som bara finns på målsidan registreras i det globala
// leader-läget, så att samma tangentsekvenser fungerar överallt.
const { registerPageCommands } = useLeaderKey();
let unregisterLeaderCommands: (() => void) | null = null;

onMounted(() => {
  unregisterLeaderCommands = registerPageCommands([
    {
      keys: "fd",
      label: "F+D",
      title: "Toggle completed goals",
      run: () => {
        showCompleted.value = !showCompleted.value;
      },
    },
    {
      keys: "i",
      label: "I",
      title: "Change icon",
      run: openIconPickerForSelection,
    },
    {
      keys: "a",
      label: "A",
      title: "Växla agent-märke",
      run: () => {
        const target = agentTarget();
        if (target) {
          toggleAgent(target);
        }
      },
    },
    {
      keys: "d",
      label: "D",
      title: "Sätt startdatum och deadline",
      run: () => {
        const target = agentTarget();
        if (target) {
          goalDates.openFor(target);
        }
      },
    },
  ]);
});

onUnmounted(() => {
  unregisterLeaderCommands?.();
  unregisterLeaderCommands = null;
});

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

  // Beroendeväljaren hanterar sina egna tangenter.
  if (dependencyPickerOpen.value) return;

  // Insert mode - ignorera alla tangenter (hanteras direkt på input-elementet)
  if (mode.value === "insert") {
    return;
  }

  // Tangenter i formulärfält (t.ex. viktslidern) ska inte styra listan
  const eventTarget = event.target as HTMLElement | null;
  const isFormField =
    !!eventTarget &&
    (["INPUT", "TEXTAREA", "SELECT"].includes(eventTarget.tagName) ||
      eventTarget.isContentEditable);
  if (isFormField) {
    return;
  }

  // ===============================================
  // LEADER MODE - hanteras globalt, här bara ignorerar tangenten
  // ===============================================

  if (isLeaderConsumed(event)) {
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

  // Hantera 'e' för att växla evigt-flaggan
  if (event.key === "e") {
    event.preventDefault();
    if (isParentMode.value && parents.value.length > 0) {
      toggleEternal(parents.value[selectedParentIndex.value]);
    } else if (isGoalSelected.value && goal.value) {
      toggleEternal(goal.value);
    } else if (
      filteredChildren.value.length > 0 &&
      selectedChildIndex.value >= 0 &&
      selectedChildIndex.value < filteredChildren.value.length
    ) {
      toggleEternal(filteredChildren.value[selectedChildIndex.value]);
    }
    return;
  }

  // Hantera 't' för att redigera beskrivningen
  if (event.key === "t") {
    event.preventDefault();
    if (isGoalSelected.value && goal.value) {
      startEditDescription(goal.value);
    } else if (
      filteredChildren.value.length > 0 &&
      selectedChildIndex.value >= 0 &&
      selectedChildIndex.value < filteredChildren.value.length
    ) {
      startEditDescription(filteredChildren.value[selectedChildIndex.value]);
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

  // Hantera 'p' för att sätta plats (var målet kan slutföras)
  if (event.key === "p") {
    event.preventDefault();
    // Tangenten följer det markerade målet: förälder (parent-läge),
    // huvudmålet eller det markerade undermålet.
    if (isParentMode.value && parents.value.length > 0) {
      togglePlacePicker(parents.value[selectedParentIndex.value].id);
    } else if (isGoalSelected.value && goal.value) {
      togglePlacePicker(goal.value.id);
    } else if (
      filteredChildren.value.length > 0 &&
      selectedChildIndex.value >= 0 &&
      selectedChildIndex.value < filteredChildren.value.length
    ) {
      togglePlacePicker(filteredChildren.value[selectedChildIndex.value].id);
    } else if (goal.value) {
      // Inga undermål att markera - då gäller platsen huvudmålet.
      togglePlacePicker(goal.value.id);
    }
    return;
  }

  // Hantera 'c' för att sätta deadline (due-datum)
  if (event.key === "c") {
    event.preventDefault();
    // Tangenten följer det markerade målet: förälder (parent-läge),
    // huvudmålet eller det markerade undermålet.
    if (isParentMode.value && parents.value.length > 0) {
      openDuePicker(parents.value[selectedParentIndex.value]);
    } else if (isGoalSelected.value && goal.value) {
      openDuePicker(goal.value);
    } else if (
      filteredChildren.value.length > 0 &&
      selectedChildIndex.value >= 0 &&
      selectedChildIndex.value < filteredChildren.value.length
    ) {
      openDuePicker(filteredChildren.value[selectedChildIndex.value]);
    }
    return;
  }

  // Hantera 'b' för att sätta startdatum (tidigast synligt i /priority)
  if (event.key === "b") {
    event.preventDefault();
    // Tangenten följer det markerade målet: förälder (parent-läge),
    // huvudmålet eller det markerade undermålet.
    if (isParentMode.value && parents.value.length > 0) {
      openStartPicker(parents.value[selectedParentIndex.value]);
    } else if (isGoalSelected.value && goal.value) {
      openStartPicker(goal.value);
    } else if (
      filteredChildren.value.length > 0 &&
      selectedChildIndex.value >= 0 &&
      selectedChildIndex.value < filteredChildren.value.length
    ) {
      openStartPicker(filteredChildren.value[selectedChildIndex.value]);
    }
    return;
  }

  // Hantera 'm' för att sätta uppskattad tidsåtgång
  if (event.key === "m") {
    event.preventDefault();
    // Tangenten följer det markerade målet: förälder (parent-läge),
    // huvudmålet eller det markerade undermålet.
    if (isParentMode.value && parents.value.length > 0) {
      openEstimateEditor(parents.value[selectedParentIndex.value]);
    } else if (isGoalSelected.value && goal.value) {
      openEstimateEditor(goal.value);
    } else if (
      filteredChildren.value.length > 0 &&
      selectedChildIndex.value >= 0 &&
      selectedChildIndex.value < filteredChildren.value.length
    ) {
      openEstimateEditor(filteredChildren.value[selectedChildIndex.value]);
    }
    return;
  }

  // Hantera 'r' för att sätta upprepning
  if (event.key === "r") {
    event.preventDefault();
    // Tangenten följer det markerade målet: förälder (parent-läge),
    // huvudmålet eller det markerade undermålet.
    if (isParentMode.value && parents.value.length > 0) {
      openRecurrenceEditor(parents.value[selectedParentIndex.value]);
    } else if (isGoalSelected.value && goal.value) {
      openRecurrenceEditor(goal.value);
    } else if (
      filteredChildren.value.length > 0 &&
      selectedChildIndex.value >= 0 &&
      selectedChildIndex.value < filteredChildren.value.length
    ) {
      openRecurrenceEditor(filteredChildren.value[selectedChildIndex.value]);
    } else if (goal.value) {
      // Inget undermål är markerat - då gäller upprepningen huvudmålet.
      openRecurrenceEditor(goal.value);
    }
    return;
  }

  // Hantera 'w' för att göra det markerade målet beroende av ett annat mål
  if (event.key === "w") {
    event.preventDefault();
    // Inget markerat -> sidans eget mål.
    openDependencyPicker(agentTarget() ?? goal.value ?? null);
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
      moveChild("up");
    } else if (event.key === "ArrowDown") {
      event.preventDefault();
      moveChild("down");
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

// När något har ångras kan målen ha ändrats - bygg om sidans data ur
// grafen (store är redan uppdaterad av ångrandet).
watch(undoRevision, () => {
  hydrate();
});
watch(goalDates.savedRevision, () => {
  hydrate();
});

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
  <div v-if="user" class="h-screen flex flex-col overflow-hidden" :class="weightMode ? '' : 'max-w-4xl mx-auto px-6 py-6'">
    <div v-if="!goal" class="text-center py-12">
      <p class="text-ink-dim">Laddar mål...</p>
    </div>

    <div v-else-if="!weightMode" class="flex flex-col h-full gap-5">
      <!-- Breadcrumb / Föräldrar -->
      <div class="flex items-center justify-between flex-shrink-0">
        <div class="flex items-center gap-1.5 text-sm flex-wrap">
          <NuxtLink v-if="parents.length === 0" to="/goals"
            class="px-2.5 py-1 rounded-lg bg-surface-2 border border-line text-accent hover:text-accent-light hover:border-line-strong transition-colors font-medium">
            Root
          </NuxtLink>

          <!-- Insert mode för förälder -->
          <div v-for="(parent, index) in parents" :key="parent.id" class="flex items-center gap-1.5">
            <div v-if="mode === 'insert' && editingGoalId === parent.id" class="inline-block">
              <input ref="editInputRef" v-model="editTitle" type="text"
                class="rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
                @keydown.enter.prevent="saveEdit" @keydown.esc.prevent="cancelEdit" />
            </div>

            <!-- Normal mode för förälder -->
            <NuxtLink v-else :to="`/goal/${parent.id}`" :data-parent-index="index"
              class="px-2.5 py-1 rounded-lg border transition-all select-none inline-block" :class="isParentMode && selectedParentIndex === index
                ? 'text-ink bg-accent-strong border-accent-strong font-medium'
                : parent.eternal
                ? ETERNAL_TEXT + ' ' + ETERNAL_TEXT_HOVER + ' bg-surface-2 border-line'
                : 'text-ink-dim hover:text-ink hover:bg-surface-2 border-transparent'
                " @mousedown="handleParentMouseDown(parent.id)" @mouseup="handleParentMouseUp"
              @mouseleave="handleParentMouseUp" @touchstart="handleParentMouseDown(parent.id)"
              @touchend="handleParentMouseUp" @touchcancel="handleParentMouseUp">
              {{ parent.title }}
            </NuxtLink>

            <!-- Deadline för förälder (bara den markerade föräldern) -->
            <template v-if="isParentMode && selectedParentIndex === index">
              <template v-if="dueTargetId === parent.id">
                <input :data-due-input="parent.id" :value="dueDraft" type="date" autofocus draggable="false"
                  class="rounded-lg border border-line bg-surface-2 px-2 py-1 text-ink text-sm focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
                  @click.stop @input="dueDraft = ($event.target as HTMLInputElement).value"
                  @change="saveDue(parent, ($event.target as HTMLInputElement).value)"
                  @keydown.enter.prevent="saveDue(parent, ($event.target as HTMLInputElement).value)"
                  @keydown.esc.prevent="closeDuePicker" />
                <button @click.stop="saveDue(parent, '')"
                  class="rounded-lg p-1 text-ink-dim transition-colors hover:bg-elevated hover:text-ink"
                  title="Rensa deadline">
                  <Icon name="lucide:x" class="w-4 h-4" />
                </button>
              </template>
              <button v-else @click.stop="openDuePicker(parent)"
                class="rounded-lg p-1 transition-colors"
                :class="parent.due ? dueToneClass(parent.due) : 'text-ink-dim hover:text-ink-muted'"
                :title="parent.due ? `Deadline: ${formatDueDate(parent.due)}` : 'Sätt deadline (c)'">
                <Icon name="lucide:calendar-clock" class="w-4 h-4"
                  :style="{ opacity: parent.due ? 1 : 0.25 }" />
              </button>
              <EstimateBadge
                :goal="parent"
                size="sm"
                :editor-open="estimateTargetId === parent.id"
                @open="openEstimateEditor(parent)"
                @save="saveEstimate(parent, $event)"
                @cancel="closeEstimateEditor" />
              <!-- Platser för förälder (p) -->
              <div class="relative flex items-center">
                <button @click.stop="togglePlacePicker(parent.id)"
                  class="rounded-lg p-1 transition-colors"
                  :class="parent.places?.length ? 'text-accent' : 'text-ink-dim hover:text-ink-muted'"
                  :title="parent.places?.length ? `Platser: ${parent.places.map((p) => p.name).join(', ')}` : 'Sätt plats (p)'">
                  <Icon name="lucide:map-pin" class="w-4 h-4"
                    :style="{ opacity: parent.places?.length ? 1 : 0.25 }" />
                </button>
                <PlacePicker v-if="placeTargetId === parent.id"
                  class="absolute right-0 top-full z-30 mt-1"
                  :selected="placeDraft"
                  :available="allPlaces"
                  @update:selected="placeDraft = $event"
                  @save="savePlaces(parent.id)"
                  @cancel="closePlacePicker" />
              </div>
            </template>
          </div>
        </div>

        <!-- + knapp för att lägga till förälder -->
        <button @click="toggleParentSearch" class="rounded-lg p-1.5 text-ink-dim transition-colors hover:bg-elevated hover:text-ink" title="Lägg till förälder">
          <Icon name="lucide:plus" class="h-5 w-5" />
        </button>
      </div>

      <!-- Sökfält för föräldrar -->
      <div v-if="showParentSearch" class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 p-4 flex-shrink-0">
        <div class="mb-2">
          <input ref="parentSearchInput" v-model="parentSearchQuery" @keydown="handleSearchKeydown" type="text"
            placeholder="Sök efter mål eller skapa nytt (tryck Enter)"
            class="rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30 w-full"
            autofocus />
        </div>

        <!-- Sökresultat -->
        <div v-if="searchResults.length > 0" class="space-y-1 mt-2">
          <button v-for="result in searchResults" :key="result.id" @click="addExistingParent(result.id)"
            class="w-full text-left px-4 py-2 hover:bg-elevated rounded-lg transition-colors">
            <div class="font-medium text-ink">{{ result.title }}</div>
            <div class="text-xs text-ink-dim">
              {{ new Date(result.created).toLocaleDateString("sv-SE") }}
            </div>
          </button>
        </div>

        <!-- Meddelande när inga resultat -->
        <div v-else-if="parentSearchQuery.trim()" class="text-sm text-ink-dim px-4 py-2">
          Tryck Enter för att skapa "{{ parentSearchQuery }}" som nytt mål
        </div>
      </div>

      <!-- Huvudmål -->
      <div v-if="mode === 'insert' && editingGoalId === goal?.id" class="mb-1">
        <input ref="editInputRef" v-model="editTitle" type="text" class="rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30 w-full text-lg"
          @keydown.enter.prevent="saveEdit" @keydown.esc.prevent="cancelEdit" />
        <div class="text-xs text-ink-dim mt-2">
          Enter för att spara, Escape för att avbryta
        </div>
      </div>
      <div
        v-else
        data-goal-row
        :draggable="!!goal"
        @dragstart.stop="goal && startGoalDrag($event, goal)"
        class="goal-drag-row flex items-center gap-3 flex-shrink-0 cursor-grab active:cursor-grabbing opt-row"
      >
        <button v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)" @click.stop="showIconPicker = true"
          class="rounded-lg p-1.5 text-ink-dim transition-colors hover:bg-elevated hover:text-ink" title="Ändra ikon">
          <Icon :name="goal.icon || 'heroicons:star'" class="w-8 h-8 text-ink" />
        </button>
        <h1
          :class="['text-3xl font-bold tracking-tight transition-colors px-3 py-2 rounded-xl flex-1 border',
            isGoalSelected ? 'border-accent bg-accent/10' : 'border-transparent',
            goal?.eternal ? ETERNAL_TEXT : 'text-ink'
          ]"
        >{{ goal?.title }}</h1>
        <!-- Ikonerna grupperas tätt ihop så att raden inte ser utsträckt ut. -->
        <div class="flex items-center gap-1 flex-shrink-0">
        <!-- Started-toggle (s) -->
        <span v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)"
          class="opt-icon" :class="{ 'is-on': !!goal.started }">
          <button @click.stop="toggleStarted(goal)"
            class="rounded-lg p-1.5 transition-colors hover:bg-elevated"
            :class="goal.started ? 'text-yellow-400' : 'text-ink-dim hover:text-ink-muted'"
            :title="goal.started ? 'Avbörja (s)' : 'Påbörja (s)'">
            <Icon name="lucide:circle-play" class="w-6 h-6"
              :style="{ opacity: goal.started ? 1 : 0.25 }" />
          </button>
        </span>
        <!-- Evigt mål-toggle (e) -->
        <span v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)"
          class="opt-icon" :class="{ 'is-on': !!goal.eternal }">
          <button @click.stop="toggleEternal(goal)"
            class="rounded-lg p-1.5 transition-colors hover:bg-elevated"
            :class="goal.eternal
              ? ETERNAL_ACCENT + ' ' + ETERNAL_ACCENT_HOVER
              : ''"
            :title="goal.eternal ? 'Ta bort evigt-märket (e)' : 'Markera som evigt mål (e)'">
            <Icon name="lucide:infinity" class="w-6 h-6"
              :style="{ opacity: goal.eternal ? 1 : 0.25 }" />
          </button>
        </span>
        <!-- Agent-märke -->
        <span v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)"
          class="opt-icon" :class="{ 'is-on': !!goal.agent }">
          <AgentBadge :active="!!goal.agent" @toggle="toggleAgent(goal)" />
        </span>
        <!-- Upprepning: antalet slutförda förekomster (r) -->
        <div v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)"
          class="opt-icon relative flex items-center gap-1.5 flex-shrink-0"
          :class="{ 'is-on': isRecurringGoal(goal) || recurrenceTargetId === goal.id }" @click.stop>
          <button @click.stop="openRecurrenceEditor(goal)"
            class="rounded-lg p-1.5 transition-colors hover:bg-elevated"
            :class="isRecurringGoal(goal) ? 'text-accent' : 'text-ink-dim hover:text-ink-muted'"
            :title="isRecurringGoal(goal) ? `Upprepas ${describeGoalRecurrence(goal.recurrence_rule)} - sätt upprepning (r)` : 'Sätt upprepning (r)'">
            <Icon name="lucide:repeat" class="w-6 h-6"
              :style="{ opacity: isRecurringGoal(goal) ? 1 : 0.25 }" />
          </button>
          <span v-if="isRecurringGoal(goal)" class="text-sm tabular-nums text-ink-dim">
            {{ recurrenceCount(goal) }}
          </span>
          <RecurrenceEditor v-if="recurrenceTargetId === goal.id"
            class="absolute right-0 top-full z-30 mt-1"
            :rule="goal.recurrence_rule ?? null"
            :mode="goal.recurrence_mode ?? 'schedule'"
            @save="saveRecurrence"
            @remove="removeRecurrence"
            @cancel="closeRecurrenceEditor" />
        </div>
        <!-- Due-datum / deadline (c) -->
        <div v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)"
          class="opt-icon flex items-center gap-1.5 flex-shrink-0"
          :class="{ 'is-on': !!goal.due || dueTargetId === goal.id }">
          <template v-if="dueTargetId === goal.id">
            <input ref="dueInputRef" :data-due-input="goal.id" v-model="dueDraft" type="date"
              class="rounded-lg border border-line bg-surface-2 px-2 py-1 text-ink text-sm focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
              @change="saveDue(goal, dueDraft)"
              @keydown.enter.prevent="saveDue(goal, dueDraft)"
              @keydown.esc.prevent="closeDuePicker" />
            <button @click.stop="saveDue(goal, '')"
              class="rounded-lg p-1 text-ink-dim transition-colors hover:bg-elevated hover:text-ink"
              title="Rensa deadline">
              <Icon name="lucide:x" class="w-4 h-4" />
            </button>
          </template>
          <template v-else>
            <button @click.stop="openDuePicker(goal)"
              class="rounded-lg p-1.5 transition-colors hover:bg-elevated"
              :class="goal.due ? dueToneClass(goal.due) : 'text-ink-dim hover:text-ink-muted'"
              :title="goal.due ? `Deadline: ${formatDueDate(goal.due)}` : 'Sätt deadline (c)'">
              <Icon name="lucide:calendar-clock" class="w-6 h-6"
                :style="{ opacity: goal.due ? 1 : 0.25 }" />
            </button>
            <span v-if="goal.due" class="text-sm tabular-nums" :class="dueToneClass(goal.due)">
              {{ formatDueDate(goal.due) }}
            </span>
          </template>
        </div>
        <!-- Startdatum: tidigast målet får dyka upp i /priority (b) -->
        <div v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)"
          class="opt-icon flex items-center gap-1.5 flex-shrink-0"
          :class="{ 'is-on': !!goal.start_date || startTargetId === goal.id }">
          <template v-if="startTargetId === goal.id">
            <input :data-start-input="goal.id" v-model="startDraft" type="date"
              class="rounded-lg border border-line bg-surface-2 px-2 py-1 text-ink text-sm focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
              @change="saveStartDate(goal, startDraft)"
              @keydown.enter.prevent="saveStartDate(goal, startDraft)"
              @keydown.esc.prevent="closeStartPicker" />
            <button @click.stop="saveStartDate(goal, '')"
              class="rounded-lg p-1 text-ink-dim transition-colors hover:bg-elevated hover:text-ink"
              title="Rensa startdatum">
              <Icon name="lucide:x" class="w-4 h-4" />
            </button>
          </template>
          <template v-else>
            <button @click.stop="openStartPicker(goal)"
              class="rounded-lg p-1.5 transition-colors hover:bg-elevated"
              :class="goal.start_date ? (isFutureDate(goal.start_date) ? 'text-sky-300' : 'text-ink-muted') : 'text-ink-dim hover:text-ink-muted'"
              :title="goal.start_date ? `Startdatum: ${formatDueDate(goal.start_date)} - målet syns i prioriteringslistan från detta datum` : 'Sätt startdatum (b)'">
              <Icon name="lucide:calendar-days" class="w-6 h-6"
                :style="{ opacity: goal.start_date ? 1 : 0.25 }" />
            </button>
            <span v-if="goal.start_date" class="text-sm tabular-nums"
              :class="isFutureDate(goal.start_date) ? 'text-sky-300' : 'text-ink-dim'">
              {{ formatDueDate(goal.start_date) }}
            </span>
          </template>
        </div>
        <span v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)"
          class="opt-icon" :class="{ 'is-on': !!goal.estimated_minutes || estimateTargetId === goal.id }">
          <EstimateBadge
            :goal="goal"
            :editor-open="estimateTargetId === goal.id"
            @open="openEstimateEditor(goal)"
            @save="saveEstimate(goal, $event)"
            @cancel="closeEstimateEditor" />
        </span>
        <!-- Platser: var målet kan slutföras (p) -->
        <div v-if="goal && !(mode === 'insert' && editingGoalId === goal?.id)"
          class="opt-icon relative flex items-center gap-1.5 flex-shrink-0"
          :class="{ 'is-on': !!goal.places?.length || placeTargetId === goal.id }">
          <button @click.stop="togglePlacePicker(goal.id)"
            class="rounded-lg p-1.5 transition-colors hover:bg-elevated"
            :class="goal.places?.length ? 'text-accent' : 'text-ink-dim hover:text-ink-muted'"
            :title="goal.places?.length ? `Platser: ${goal.places.map((p) => p.name).join(', ')}` : 'Sätt plats (p)'">
            <Icon name="lucide:map-pin" class="w-6 h-6"
              :style="{ opacity: goal.places?.length ? 1 : 0.25 }" />
          </button>
          <span v-if="goal.places?.length" class="text-sm text-ink-dim">
            {{ goal.places.map((p) => p.name).join(', ') }}
          </span>
          <PlacePicker v-if="placeTargetId === goal.id"
            class="absolute right-0 top-full z-30 mt-1"
            :selected="placeDraft"
            :available="allPlaces"
            @update:selected="placeDraft = $event"
            @save="savePlaces(goal.id)"
            @cancel="closePlacePicker" />
        </div>
        <!-- Toggle dependencies visibility -->
        <button @click.stop="showDependencies = !showDependencies"
          class="rounded-lg p-1.5 transition-colors hover:bg-elevated"
          :class="showDependencies ? 'text-orange-400' : 'text-ink-dim hover:text-ink-muted'"
          title="Beroenden (blockerade mål)">
          <Icon name="lucide:git-branch" class="w-6 h-6" />
        </button>
        </div>
      </div>

      <!-- Beskrivning -->
      <div v-if="goal" class="flex-shrink-0">
        <div v-if="editingDescriptionId === goal.id">
          <textarea :data-description-input="goal.id" v-model="descriptionDraft" rows="4"
            class="rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30 w-full resize-y"
            placeholder="Beskrivning - vad vill du uppnå?"
            @keydown="handleDescriptionKeydown($event, goal.id)"
            @keydown.esc.prevent="cancelEditDescription" />
          <div class="flex items-center justify-between mt-2 gap-4">
            <span class="text-xs text-ink-dim">Ctrl+Enter spara, Escape avbryta</span>
            <div class="flex gap-2">
              <button @click="cancelEditDescription" class="rounded-lg border border-line bg-surface-2 px-4 py-2 text-ink-muted transition-colors hover:border-line-strong hover:text-ink">Avbryt</button>
              <button @click="saveDescription(goal.id)"
                class="px-3 py-1.5 bg-accent-strong hover:bg-accent rounded-lg text-white transition-colors">
                Spara
              </button>
            </div>
          </div>
        </div>
        <p v-else-if="goal.description" class="text-sm text-ink-muted whitespace-pre-wrap">
          {{ goal.description }}
        </p>
      </div>

      <!-- Dependencies-sektion -->
      <div v-if="showDependencies" class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 p-4 mb-1 flex-shrink-0">
        <!-- Väntar på -->
        <div class="mb-4">
          <div class="flex items-center justify-between mb-2">
            <h3 class="text-sm font-medium text-ink-muted flex items-center gap-2">
              <Icon name="lucide:clock" class="w-4 h-4" />
              Väntar på ({{ dependsOn.length }})
            </h3>
            <button @click="toggleDependencySearch" class="rounded-lg p-1.5 text-ink-dim transition-colors hover:bg-elevated hover:text-ink" title="Lägg till beroende">
              <Icon name="lucide:plus" class="w-4 h-4" />
            </button>
          </div>

          <!-- Sök för att lägga till dependency -->
          <div v-if="showDependencySearch" class="mb-3">
            <input ref="dependencySearchInput" v-model="dependencySearchQuery" @keydown="handleDependencySearchKeydown"
              type="text"
              placeholder="Sök efter mål att vänta på..."
              class="rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30 w-full text-sm"
              autofocus />
            <div v-if="filteredDependencySearchResults.length > 0" class="mt-2 space-y-1">
              <button v-for="result in filteredDependencySearchResults.slice(0, 5)" :key="result.id"
                @click="addExistingDependency(result.id)"
                class="w-full text-left px-3 py-2 hover:bg-elevated rounded-lg transition-colors flex items-center gap-2">
                <Icon :name="result.icon || 'heroicons:star'" class="w-4 h-4 text-ink-muted" />
                <span class="text-ink text-sm">{{ result.title }}</span>
                <span v-if="result.finished" class="text-xs text-green-400 ml-auto">Klar</span>
              </button>
            </div>
            <div v-else-if="dependencySearchQuery.trim() && filteredDependencySearchResults.length === 0" class="text-xs text-ink-dim mt-2 px-2">
              Inga matchande mål hittades
            </div>
          </div>

          <!-- Lista över dependencies -->
          <div v-if="dependsOn.length > 0" class="space-y-2">
            <div v-for="dep in dependsOn" :key="dep.depends_on_id"
              class="flex items-center gap-2 px-3 py-2 bg-surface-2 rounded-lg border border-line hover:border-line-strong transition-colors">
              <Icon :name="dep.icon || 'heroicons:star'" class="w-4 h-4 text-ink-muted flex-shrink-0" />
              <NuxtLink :to="`/goal/${dep.depends_on_id}`" class="flex-1 text-sm text-ink hover:text-white truncate">
                {{ dep.title }}
              </NuxtLink>
              <span v-if="dep.finished" class="text-xs text-green-400">Klar</span>
              <span v-else class="text-xs text-orange-400">Väntar...</span>
              <button @click.stop="removeExistingDependency(dep.depends_on_id)"
                class="text-ink-dim hover:text-red-400 transition-colors p-1">
                <Icon name="lucide:x" class="w-3 h-3" />
              </button>
            </div>
          </div>
          <div v-else class="text-sm text-ink-dim px-2">
            Inga beroenden. Lägg till mål som måste göras först.
          </div>
        </div>

        <!-- Blockerar -->
        <div v-if="blocking.length > 0">
          <h3 class="text-sm font-medium text-ink-muted flex items-center gap-2 mb-2">
            <Icon name="lucide:lock" class="w-4 h-4" />
            Blockerar ({{ blocking.length }})
          </h3>
          <div class="space-y-2">
            <div v-for="block in blocking" :key="block.goal_id"
              class="flex items-center gap-2 px-3 py-2 bg-surface-2 rounded-lg border border-line">
              <Icon :name="block.icon || 'heroicons:star'" class="w-4 h-4 text-ink-muted flex-shrink-0" />
              <NuxtLink :to="`/goal/${block.goal_id}`" class="flex-1 text-sm text-ink hover:text-white truncate">
                {{ block.title }}
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>

      <!-- Undermål - scrollbar container -->
      <div class="flex-1 overflow-y-auto min-h-0 pb-6">
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-sm font-semibold uppercase tracking-wider text-ink-dim">
            Undermål
          </h2>
          <div class="flex items-center gap-1">
            <button @click="showCompleted = !showCompleted"
              class="rounded-lg p-1.5 text-ink-dim transition-colors hover:bg-elevated hover:text-ink" :title="showCompleted ? 'Dölj avklarade mål' : 'Visa avklarade mål'">
              <Icon :name="showCompleted ? 'lucide:eye' : 'lucide:eye-off'" class="h-5 w-5" />
            </button>
            <!-- Viktläge -->
            <button @click="toggleWeightMode"
              class="rounded-lg p-1.5 transition-colors hover:bg-elevated"
              :class="weightMode ? 'text-accent' : 'text-ink-dim hover:text-ink-muted'"
              :title="weightMode ? 'Stäng viktning' : 'Ändra viktning'">
              <Icon name="lucide:scale" class="h-5 w-5" />
            </button>
            <!-- Visa endast påbörjade -->
            <button @click="showStarted = !showStarted"
              class="rounded-lg p-1.5 transition-colors hover:bg-elevated"
              :class="showStarted ? 'text-yellow-400' : 'text-ink-dim hover:text-ink-muted'"
              :title="showStarted ? 'Visa alla mål' : 'Visa endast påbörjade mål'">
              <Icon name="lucide:circle-play" class="h-5 w-5"
                :style="{ opacity: showStarted ? 1 : 0.3 }" />
            </button>
            <button @click="toggleChildSearch" class="rounded-lg p-1.5 text-ink-dim transition-colors hover:bg-elevated hover:text-ink" title="Lägg till undermål">
              <Icon name="lucide:plus" class="h-5 w-5" />
            </button>
          </div>
        </div>

        <!-- Sökfält för undermål -->
        <div v-if="showChildSearch" class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 p-4 mb-4">
          <div class="mb-2">
            <input ref="childSearchInput" v-model="childSearchQuery" @keydown.enter.prevent="handleChildSearchKeydown" @keydown.escape.prevent="showChildSearch = false; childSearchQuery = ''" type="text"
              placeholder="Sök efter mål eller skapa nytt (tryck Enter)"
              class="rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30 w-full"
              autofocus />
          </div>

          <!-- Sökresultat -->
          <div v-if="childSearchResults.length > 0" class="space-y-1 mt-2">
            <button v-for="result in childSearchResults" :key="result.id" @click="addExistingChild(result.id)"
              class="w-full text-left px-4 py-2 hover:bg-elevated rounded-lg transition-colors">
              <div class="font-medium text-ink">{{ result.title }}</div>
              <div class="text-xs text-ink-dim">
                {{ new Date(result.created).toLocaleDateString("sv-SE") }}
              </div>
            </button>
          </div>

          <!-- Meddelande när inga resultat -->
          <div v-else-if="childSearchQuery.trim()" class="text-sm text-ink-dim px-4 py-2">
            Tryck Enter för att skapa "{{ childSearchQuery }}" som nytt undermål
          </div>
        </div>

        <ul v-else-if="filteredChildren.length > 0" class="space-y-2">
          <li v-for="(child, index) in filteredChildren" :key="child.id" :data-child-index="index" draggable="true"
            @dragstart="handleDragStart($event, child.id)" @dragover="handleDragOver($event, child.id)"
            @dragleave="handleDragLeave" @drop="handleDrop($event, child.id)"
            class="relative overflow-hidden rounded-2xl transition-opacity"
            :style="childIndentStyle(child.id)"
            :class="[
              dragOverChildId === child.id ? 'ring-2 ring-fuchsia-500 ring-offset-2 ring-offset-canvas' : 'opacity-100',
              draggedChildId === child.id ? 'opacity-30' : ''
            ]">
            <!-- Indicator när något släpps här - valet (flytta eller beroende) görs efteråt -->
            <div v-if="dragOverChildId === child.id && draggedChildId !== child.id"
              class="absolute inset-0 bg-fuchsia-500/20 flex items-center justify-center z-10 rounded-2xl">
              <div class="bg-fuchsia-600 text-white px-3 py-1 rounded-full text-sm font-medium">
                Släpp för att välja
              </div>
            </div>
            <!-- Swipe bakgrund -->
            <div class="absolute inset-0 flex items-center justify-start px-6"
              :class="child.finished ? 'bg-red-900/50' : 'bg-transparent'">
              <span class="text-2xl">{{ child.finished ? "↩️" : "" }}</span>
            </div>

            <!-- Huvudinnehåll -->
            <div class="relative rounded-2xl border transition-all bg-surface hover:bg-surface-2 group opt-row" :class="{
              'border-accent shadow-[0_0_0_1px_rgb(129_140_248/0.35)]': selectedChildIndex === index,
              'border-line': selectedChildIndex !== index
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
                <input ref="editInputRef" v-model="editTitle" type="text" class="rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30 w-full"
                  @keydown.enter.prevent="saveEdit" @keydown.esc.prevent="cancelEdit" />
                <div class="text-xs text-ink-dim mt-2">
                  Enter för att spara, Escape för att avbryta
                </div>
              </div>

              <!-- Normal mode - visa länk -->
              <div v-else class="flex items-center gap-1.5 px-2">
                 <button v-if="!(mode === 'insert' && editingGoalId === child.id)"
                   @click.stop="editingIconGoalId = child.id; showIconPicker = true"
                   :class="['flex-shrink-0 w-7 h-7 flex items-center justify-center rounded-lg transition-colors', child.eternal ? ETERNAL_TEXT + ' ' + ETERNAL_TEXT_HOVER : isChildBlocked(child.id) || child.finished ? 'text-ink-dim hover:text-ink-dim' : 'text-ink hover:text-white hover:bg-elevated']"
                   title="Ändra ikon">
                    <Icon :name="child.icon || 'heroicons:star'" class="w-5 h-5" :style="child.eternal ? { color: ETERNAL_COLOR } : isChildBlocked(child.id) || child.finished ? { color: '#63708a' } : { color: getWeightStyle(child.weight).color, opacity: getWeightStyle(child.weight).opacity }" />
                 </button>
                 <div class="flex-1 min-w-0">
                  <NuxtLink :to="`/goal/${child.id}`" class="py-1.5 block cursor-pointer">
                    <h3 class="text-base font-normal select-none" :class="child.eternal ? ETERNAL_TEXT : child.finished || isChildBlocked(child.id) ? 'text-ink-dim' : 'text-ink'" :style="child.eternal ? { color: ETERNAL_COLOR } : child.finished || isChildBlocked(child.id) ? {} : getWeightStyle(child.weight)">
                      {{ child.title }}
                    </h3>
                    <p v-if="getChildWaitingForTitle(child.id)" class="text-xs text-ink-dim mt-0.5">
                      🔗 Väntar på: {{ getChildWaitingForTitle(child.id) }}
                    </p>
                    <p v-if="child.description && editingDescriptionId !== child.id"
                      class="text-xs text-ink-dim mt-0.5 truncate">
                      {{ descriptionPreview(child.description, 80) }}
                    </p>
                  </NuxtLink>
                </div>
                <!-- Ikonerna grupperas tätt ihop så att raden inte ser utsträckt ut. -->
                <div class="flex items-center gap-0.5 flex-shrink-0">
                 <!-- Started-toggle (s) -->
                 <span class="opt-icon" :class="{ 'is-on': !!child.started }">
                   <button
                     @click.stop="toggleStarted(child)"
                     :class="['flex-shrink-0 w-7 h-7 flex items-center justify-center rounded-lg transition-colors', isChildBlocked(child.id) || child.finished ? 'text-ink-dim' : child.started ? 'text-yellow-400 hover:text-yellow-300' : 'text-ink hover:text-ink-muted']"
                     :title="child.started ? 'Avbörja (s)' : 'Påbörja (s)'">
                     <Icon name="lucide:circle-play" class="w-5 h-5"
                       :style="{ opacity: isChildBlocked(child.id) || child.finished ? 0.6 : child.started ? 1 : 0.25 }" />
                   </button>
                 </span>
                 <!-- Agent-märke -->
                 <span class="opt-icon" :class="{ 'is-on': !!child.agent }">
                   <AgentBadge :active="!!child.agent" @toggle="toggleAgent(child)" />
                 </span>
                <!-- Deadline för undermål (c) -->
                <div class="opt-icon flex-shrink-0 flex items-center gap-1"
                  :class="{ 'is-on': !!child.due || dueTargetId === child.id }">
                  <template v-if="dueTargetId === child.id">
                    <input :data-due-input="child.id" :value="dueDraft" type="date" autofocus draggable="false"
                      class="rounded-lg border border-line bg-surface-2 px-2 py-1 text-ink text-sm focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30"
                      @click.stop @input="dueDraft = ($event.target as HTMLInputElement).value"
                      @change="saveDue(child, ($event.target as HTMLInputElement).value)"
                      @keydown.enter.prevent="saveDue(child, ($event.target as HTMLInputElement).value)"
                      @keydown.esc.prevent="closeDuePicker" />
                    <button @click.stop="saveDue(child, '')"
                      class="rounded-lg p-1 text-ink-dim transition-colors hover:bg-elevated hover:text-ink"
                      title="Rensa deadline">
                      <Icon name="lucide:x" class="w-4 h-4" />
                    </button>
                  </template>
                  <template v-else>
                    <button @click.stop="openDuePicker(child)"
                      class="w-7 h-7 flex items-center justify-center rounded-lg transition-colors"
                      :class="child.due ? dueToneClass(child.due) : 'text-ink-dim hover:text-ink-muted'"
                      :title="child.due ? `Deadline: ${formatDueDate(child.due)}` : 'Sätt deadline (c)'">
                      <Icon name="lucide:calendar-clock" class="w-5 h-5"
                        :style="{ opacity: child.due ? 1 : 0.25 }" />
                    </button>
                  </template>
                </div>
                <span class="opt-icon" :class="{ 'is-on': !!child.estimated_minutes || estimateTargetId === child.id }">
                  <EstimateBadge
                    :goal="child"
                    size="sm"
                    :editor-open="estimateTargetId === child.id"
                    @open="openEstimateEditor(child)"
                    @save="saveEstimate(child, $event)"
                    @cancel="closeEstimateEditor" />
                </span>
                <!-- Upprepning för undermål -->
                <div class="opt-icon relative flex-shrink-0 flex items-center gap-1"
                  :class="{ 'is-on': isRecurringGoal(child) || recurrenceTargetId === child.id }">
                  <button @click.stop="openRecurrenceEditor(child)"
                    class="w-7 h-7 flex items-center justify-center rounded-lg transition-colors"
                    :class="isRecurringGoal(child) ? 'text-accent' : 'text-ink-dim hover:text-ink-muted'"
                    :title="isRecurringGoal(child) ? `Upprepas ${describeGoalRecurrence(child.recurrence_rule)} - sätt upprepning (r)` : 'Sätt upprepning (r)'">
                    <Icon name="lucide:repeat" class="w-5 h-5"
                      :style="{ opacity: isRecurringGoal(child) ? 1 : 0.25 }" />
                  </button>
                  <span v-if="isRecurringGoal(child)" class="text-xs tabular-nums text-ink-dim">
                    {{ recurrenceCount(child) }}
                  </span>
                  <div v-if="recurrenceTargetId === child.id"
                    class="absolute right-0 top-full z-30 mt-1" @click.stop @keydown.stop>
                    <RecurrenceEditor :rule="child.recurrence_rule ?? null"
                      :mode="child.recurrence_mode ?? 'schedule'"
                      @save="saveRecurrence"
                      @remove="removeRecurrence"
                      @cancel="closeRecurrenceEditor" />
                  </div>
                </div>
                <!-- Platser för undermål (p) -->
                <div class="opt-icon relative flex-shrink-0 flex items-center gap-1"
                  :class="{ 'is-on': !!child.places?.length || placeTargetId === child.id }">
                  <button @click.stop="togglePlacePicker(child.id)"
                    class="w-7 h-7 flex items-center justify-center rounded-lg transition-colors"
                    :class="child.places?.length ? 'text-accent' : 'text-ink-dim hover:text-ink-muted'"
                    :title="child.places?.length ? `Platser: ${child.places.map((p) => p.name).join(', ')}` : 'Sätt plats (p)'">
                    <Icon name="lucide:map-pin" class="w-5 h-5"
                      :style="{ opacity: child.places?.length ? 1 : 0.25 }" />
                  </button>
                  <PlacePicker v-if="placeTargetId === child.id"
                    class="absolute right-0 top-full z-30 mt-1"
                    :selected="placeDraft"
                    :available="allPlaces"
                    @update:selected="placeDraft = $event"
                    @save="savePlaces(child.id)"
                    @cancel="closePlacePicker" />
                </div>
                </div>
              </div>
            </div>

            <!-- Redigera beskrivning för undermål (t) -->
            <div v-if="editingDescriptionId === child.id" class="relative rounded-2xl border border-accent bg-surface p-4 mt-2">
              <textarea :data-description-input="child.id" v-model="descriptionDraft" rows="3"
                class="rounded-lg border border-line bg-surface-2 px-3.5 py-2 text-ink placeholder-ink-dim transition-colors focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30 w-full resize-y"
                placeholder="Beskrivning - vad vill du uppnå?"
                @keydown="handleDescriptionKeydown($event, child.id)"
                @keydown.esc.prevent="cancelEditDescription" />
              <div class="flex items-center justify-between mt-2 gap-4">
                <span class="text-xs text-ink-dim">Ctrl+Enter spara, Escape avbryta</span>
                <div class="flex gap-2">
                  <button @click="cancelEditDescription" class="rounded-lg border border-line bg-surface-2 px-4 py-2 text-ink-muted transition-colors hover:border-line-strong hover:text-ink">Avbryt</button>
                  <button @click="saveDescription(child.id)"
                    class="px-3 py-1.5 bg-accent-strong hover:bg-accent rounded-lg text-white transition-colors">
                    Spara
                  </button>
                </div>
              </div>
            </div>
          </li>
        </ul>

        <div v-else class="text-ink-dim p-10 border border-dashed border-line rounded-2xl text-center">
          Inga undermål ännu. Skapa ett för att dela upp detta mål i mindre
          delar.
        </div>
      </div>
    </div>

    <!-- Viktläge: använd hela viewporten och dölj övrig målnavigering. -->
    <div v-else class="fixed inset-0 z-[60] flex flex-col bg-canvas">
      <div class="flex min-h-0 flex-1 items-center justify-center">
        <GoalWeightVerticalEditor
          v-if="weightChildren.length > 1"
          :children="weightChildren"
          :selected-index="weightSelectedIndex"
          @select="weightSelectedIndex = $event"
          @close="weightMode = false"
          @update="previewWeights"
          @commit="saveWeights"
        />
        <p v-else class="text-sm text-ink-dim">Det behövs minst två pågående undermål för att ändra viktning.</p>
      </div>
    </div>

    <!-- Vim mode indicator -->
    <div v-if="mode === 'insert'"
      class="fixed bottom-4 left-1/2 transform -translate-x-1/2 px-4 py-2 bg-accent-strong text-white font-mono text-sm rounded-xl shadow-lg">
      -- INSERT --
    </div>

    <!-- Modal för att bekräfta borttagning av föräldrarelation -->
    <UModal v-model:open="showRemoveConfirmation" title="Ta bort föräldrarelation?">
      <p v-if="parentToRemove" class="text-ink-muted">
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

    <!-- Val efter att ett barnmål släppts över ett annat -->
    <UModal :open="dropChoice !== null" title="Vad vill du göra?"
      @update:open="(value) => { if (!value) cancelDropChoice(); }">
      <p v-if="dropChoiceTarget" class="text-ink-muted">
        <strong>{{ dropChoiceTarget.title }}</strong>
      </p>

      <template #footer="{ close }">
        <div class="flex justify-end gap-2">
          <UButton color="gray" variant="ghost" @click="cancelDropChoice">
            Avbryt
          </UButton>
          <UButton color="gray" @click="applyDropAsChild">
            Flytta in som undermål
          </UButton>
          <UButton color="primary" @click="applyDropAsDependency">
            {{ dropChoiceHasDependency ? 'Ta bort beroendet' : 'Skapa beroende' }}
          </UButton>
        </div>
      </template>
    </UModal>

    <!-- Modal för att bekräfta borttagning av mål -->
    <UModal v-model:open="showDeleteConfirmation" title="Ta bort mål?">
      <template #footer="{ close }">
        <div class="flex justify-end gap-2">
          <UButton color="gray" variant="ghost" @click="cancelDelete" :class="deleteDialogSelection === 'cancel' ? 'ring-2 ring-accent' : ''
            ">
            Avbryt
          </UButton>
          <UButton color="red" @click="executeDelete" :class="deleteDialogSelection === 'confirm' ? 'ring-2 ring-accent' : ''
            ">
            Ta bort
          </UButton>
        </div>
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

    <!-- Beroendeväljare (w) -->
    <GoalDependencyPicker
      :open="dependencyPickerOpen"
      :target-id="dependencyPickerTargetId"
      :target-title="dependencyPickerTargetTitle"
      @select="addDependencyFromPicker"
      @close="dependencyPickerOpen = false" />
  </div>
  <div v-else class="max-w-4xl mx-auto px-6 py-8">
    <div class="text-center">
      <h1 class="text-3xl font-bold tracking-tight text-ink mb-4">Du måste logga in</h1>
      <p class="text-ink-dim mb-6">
        För att se dina mål måste du vara inloggad.
      </p>
      <NuxtLink to="/login">
        <UButton>Logga in</UButton>
      </NuxtLink>
    </div>
  </div>
</template>
