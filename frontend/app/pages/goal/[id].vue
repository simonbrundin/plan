<script setup lang="ts">
// Make this page client-only to ensure GraphQL requests work properly and prevent Pinia hydration issues
definePageMeta({
  ssr: false,
});

const route = useRoute();
const router = useRouter();
const goalId = parseInt(route.params.id as string);
const { user } = useUserSession();
const config = useRuntimeConfig();

interface Goal {
  id: number;
  title: string;
  icon: string;
  created: string;
  finished: string | null;
}

interface ChildRelation {
  child_id: number;
  order: number;
}

interface ParentRelation {
  goalByParentId: Goal;
}

interface GoalWithRelations extends Goal {
  childRelations: ChildRelation[];
  parentRelations: ParentRelation[];
}

interface GetGoalResponse {
  goal: GoalWithRelations;
  allGoals: Goal[];
}

interface SearchGoalsResponse {
  goals: Goal[];
}

// Hämta målet med dess relationer
const goalData = ref<GetGoalResponse | null>(null);
const isLoading = ref(true);
const error = ref<string | null>(null);

const fetchGoalData = async () => {
  isLoading.value = true;
  error.value = null;

  try {
    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          query GetGoal($id: Int!) {
            goal: goals_by_pk(id: $id) {
              id
              title
              icon
              created
              finished

              # Barn - relationer där detta mål är parent
              childRelations: goalRelationsByParentId(order_by: { order: asc }) {
                child_id
                order
              }

              # Förälder - relationer där detta mål är child
              parentRelations: goal_relations {
                goalByParentId {
                  id
                  title
                  icon
                }
              }
            }

            # Hämta alla mål för att kunna matcha children
            allGoals: goals {
              id
              title
              icon
              finished
              created
            }
          }
        `,
        variables: { id: goalId },
      }),
    });

    const result = await response.json();
    console.log("GetGoal result:", result);

    if (result.errors) {
      throw new Error(result.errors[0].message);
    }

    goalData.value = result.data;
  } catch (err) {
    console.error("Fetch goal error:", err);
    error.value = err instanceof Error ? err.message : "Unknown error";
  } finally {
    isLoading.value = false;
  }
};

const refresh = async () => {
  await fetchGoalData();
};

// Fetch initial data
await fetchGoalData();

const goal = computed(() => goalData.value?.goal);

// Extrahera alla föräldrar
const parents = computed(
  () => goal.value?.parentRelations?.map((r) => r.goalByParentId) || []
);

// Matcha barn-ID:n med faktiska goal-objekt och behåll ordningen från childRelations
const children = computed(() => {
  if (!goal.value?.childRelations || !goalData.value?.allGoals) return [];

  // Behåll ordningen från childRelations genom att mappa dem direkt
  return goal.value.childRelations
    .map((relation) =>
      goalData.value!.allGoals.find((g) => g.id === relation.child_id)
    )
    .filter((g) => g !== undefined) as Goal[];
});

// Beräkna progress baserat på färdiga undermål
const progress = computed(() => {
  if (children.value.length === 0) return 0;
  const completed = children.value.filter((c) => c.finished).length;
  return Math.round((completed / children.value.length) * 100);
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

// Icon picker state
const showIconPicker = ref(false);
const editingIconGoalId = ref<number | null>(null);
const editingIconGoalIcon = ref<string>('');

// Filtrerade undermål baserat på showCompleted
const filteredChildren = computed(() => {
  if (showCompleted.value) {
    return children.value;
  }
  return children.value.filter((c) => !c.finished);
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
    (g) => g.id !== goalId && !currentParentIds.includes(g.id)
  );
});

// Lägg till befintligt mål som förälder
async function addExistingParent(parentId: number) {
  try {
    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation AddParentRelation($childId: Int!, $parentId: Int!) {
            insert_goal_relations_one(
              object: { child_id: $childId, parent_id: $parentId }
            ) {
              child_id
              parent_id
            }
          }
        `,
        variables: { childId: goalId, parentId },
      }),
    });

    const result = await response.json();
    if (result.errors) {
      throw new Error(result.errors[0].message);
    }

    // Uppdatera lokal state
    goalsStore.addRelation(goalId, parentId);

    // Uppdatera data från server
    await refresh();

    // Stäng och rensa sökning
    showParentSearch.value = false;
    parentSearchQuery.value = "";
  } catch (error) {
    console.error("Failed to add parent:", error);
  }
}

// Skapa nytt mål som förälder
async function createNewParent() {
  if (!parentSearchQuery.value.trim()) return;

  const userId = user.value?.id;
  if (!userId) {
    console.error("No user logged in");
    return;
  }

  try {
    // Skapa nytt mål
    const createGoalQuery = `
      mutation CreateGoal($title: String!) {
        insert_goals_one(object: { title: $title, icon: "material-symbols:circle" }) {
          id
          title
          icon
          created
          finished
        }
      }
    `;

    const goalResponse = await $fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: createGoalQuery,
        variables: { title: parentSearchQuery.value.trim() },
      }),
    });

    if (goalResponse.errors) {
      throw new Error(goalResponse.errors[0].message);
    }

    if (goalResponse.data?.insert_goals_one?.id) {
      const newGoal = goalResponse.data.insert_goals_one;

      // Skapa user_goals relation
      const createUserGoalQuery = `
        mutation CreateUserGoal($userId: Int!, $goalId: Int!) {
          insert_user_goals_one(object: { user_id: $userId, goal_id: $goalId }) {
            user_id
            goal_id
          }
        }
      `;

      const userGoalResponse = await $fetch(
        "http://localhost:8080/v1/graphql",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "x-hasura-admin-secret": config.public.hasuraAdminSecret,
          },
          body: JSON.stringify({
            query: createUserGoalQuery,
            variables: { userId, goalId: newGoal.id },
          }),
        }
      );

      if (userGoalResponse.errors) {
        throw new Error(userGoalResponse.errors[0].message);
      }

      // Skapa relationen i databasen
      const addParentRelationQuery = `
        mutation AddParentRelation($childId: Int!, $parentId: Int!) {
          insert_goal_relations_one(object: { child_id: $childId, parent_id: $parentId }) {
            child_id
            parent_id
          }
        }
      `;

      const relationResponse = await $fetch(
        "http://localhost:8080/v1/graphql",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "x-hasura-admin-secret": config.public.hasuraAdminSecret,
          },
          body: JSON.stringify({
            query: addParentRelationQuery,
            variables: { childId: goalId, parentId: newGoal.id },
          }),
        }
      );

      if (relationResponse.errors) {
        throw new Error(relationResponse.errors[0].message);
      }

      // Uppdatera lokal state
      goalsStore.addGoal(newGoal);
      goalsStore.addRelation(goalId, newGoal.id);

      // Uppdatera data från server
      await goalsStore.loadGoals();

      // Stäng och rensa sökning
      showParentSearch.value = false;
      parentSearchQuery.value = "";
    }
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
    (g) => g.id !== goalId && !currentChildIds.includes(g.id)
  );
});

// Lägg till befintligt mål som barn
async function addExistingChild(childId: number) {
  try {
    // Räkna nuvarande barn för att få rätt ordning
    const nextOrder = goal.value?.childRelations?.length || 0;

    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation AddParentRelation($childId: Int!, $parentId: Int!, $order: Int!) {
            insert_goal_relations_one(
              object: { child_id: $childId, parent_id: $parentId, order: $order }
            ) {
              child_id
              parent_id
              order
            }
          }
        `,
        variables: { childId, parentId: goalId, order: nextOrder },
      }),
    });

    const result = await response.json();
    if (result.errors) {
      throw new Error(result.errors[0].message);
    }

    // Uppdatera lokal state
    goalsStore.addRelation(childId, goalId);

    // Uppdatera data från server
    await refresh();

    // Stäng och rensa sökning
    showChildSearch.value = false;
    childSearchQuery.value = "";
  } catch (error) {
    console.error("Failed to add child:", error);
  }
}

// Skapa nytt mål som barn
async function createNewChild() {
  if (!childSearchQuery.value.trim()) return;

  const userId = user.value?.id;
  if (!userId) {
    console.error("No user logged in");
    return;
  }

  try {
    console.log(
      "Creating new child with title:",
      childSearchQuery.value.trim()
    );

    // Skapa nytt mål
    const createGoalQuery = `
      mutation CreateGoal($title: String!) {
        insert_goals_one(object: { title: $title, icon: "material-symbols:circle" }) {
          id
          title
          icon
          created
          finished
        }
      }
    `;

    const goalResponse = await $fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: createGoalQuery,
        variables: { title: childSearchQuery.value.trim() },
      }),
    });

    if (goalResponse.errors) {
      throw new Error(goalResponse.errors[0].message);
    }

    if (goalResponse.data?.insert_goals_one?.id) {
      const newGoal = goalResponse.data.insert_goals_one;

      console.log(
        "Adding relation - childId:",
        newGoal.id,
        "parentId:",
        goalId
      );

      // Skapa user_goals relation
      const createUserGoalQuery = `
        mutation CreateUserGoal($userId: Int!, $goalId: Int!) {
          insert_user_goals_one(object: { user_id: $userId, goal_id: $goalId }) {
            user_id
            goal_id
          }
        }
      `;

      const userGoalResponse = await $fetch(
        "http://localhost:8080/v1/graphql",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "x-hasura-admin-secret": config.public.hasuraAdminSecret,
          },
          body: JSON.stringify({
            query: createUserGoalQuery,
            variables: { userId, goalId: newGoal.id },
          }),
        }
      );

      if (userGoalResponse.errors) {
        throw new Error(userGoalResponse.errors[0].message);
      }

      // Räkna nuvarande barn för att få rätt ordning
      const nextOrder = goal.value?.childRelations?.length || 0;

      // Skapa relationen i databasen
      const addParentRelationQuery = `
        mutation AddParentRelation($childId: Int!, $parentId: Int!, $order: Int!) {
          insert_goal_relations_one(object: { child_id: $childId, parent_id: $parentId, order: $order }) {
            child_id
            parent_id
            order
          }
        }
      `;

      const relationResponse = await $fetch(
        "http://localhost:8080/v1/graphql",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "x-hasura-admin-secret": config.public.hasuraAdminSecret,
          },
          body: JSON.stringify({
            query: addParentRelationQuery,
            variables: { childId: newGoal.id, parentId: goalId, order: nextOrder },
          }),
        }
      );

      if (relationResponse.errors) {
        throw new Error(relationResponse.errors[0].message);
      }

      // Uppdatera lokal state
      goalsStore.addGoal(newGoal);
      goalsStore.addRelation(newGoal.id, goalId);

      // Uppdatera data från server
      await goalsStore.loadGoals();

      // Stäng och rensa sökning
      showChildSearch.value = false;
      childSearchQuery.value = "";
    }
  } catch (error) {
    console.error("Failed to create new child:", error);
  }
}

// Hantera Enter-tangent för undermål
function handleChildSearchKeydown(event: KeyboardEvent) {
  if (event.key === "Enter") {
    if (childSearchResults.value.length === 0) {
      // Skapa nytt mål om inga resultat
      createNewChild();
    } else {
      // Välj första resultatet
      addExistingChild(childSearchResults.value[0].id);
    }
  } else if (event.key === "Escape") {
    showChildSearch.value = false;
    childSearchQuery.value = "";
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
    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation RemoveParentRelation($childId: Int!, $parentId: Int!) {
            delete_goal_relations(
              where: {
                child_id: {_eq: $childId}
                parent_id: {_eq: $parentId}
              }
            ) {
              affected_rows
            }
          }
        `,
        variables: { childId: goalId, parentId: parentToRemove.value },
      }),
    });

    const result = await response.json();
    if (result.errors) {
      throw new Error(result.errors[0].message);
    }

    // Uppdatera lokal state
    goalsStore.removeRelation(goalId, parentToRemove.value);

    // Uppdatera data från server
    await refresh();

    // Stäng modal
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
const draggedChildIndex = ref<number | null>(null);
const dragOverChildIndex = ref<number | null>(null);

function handleDragStart(event: DragEvent, index: number) {
  draggedChildIndex.value = index;
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = "move";
  }
}

function handleDragOver(event: DragEvent, index: number) {
  event.preventDefault();
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = "move";
  }
  dragOverChildIndex.value = index;
}

function handleDragLeave() {
  dragOverChildIndex.value = null;
}

async function handleDrop(event: DragEvent, dropIndex: number) {
  event.preventDefault();
  dragOverChildIndex.value = null;

  if (draggedChildIndex.value === null || draggedChildIndex.value === dropIndex) {
    draggedChildIndex.value = null;
    return;
  }

  const dragIndex = draggedChildIndex.value;
  const actualChildren = children.value;

  // Sortera barnen lokalt
  const sortedChildren = [...actualChildren];
  const draggedChild = sortedChildren[dragIndex];
  sortedChildren.splice(dragIndex, 1);
  sortedChildren.splice(dropIndex, 0, draggedChild);

  // Uppdatera ordning i databasen
  try {
    for (let i = 0; i < sortedChildren.length; i++) {
      const child = sortedChildren[i];

      await fetch("http://localhost:8080/v1/graphql", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "x-hasura-admin-secret": config.public.hasuraAdminSecret,
        },
        body: JSON.stringify({
          query: `
            mutation UpdateGoalOrder($parent_id: Int!, $child_id: Int!, $order: Int!) {
              update_goal_relations_by_pk(
                pk_columns: { parent_id: $parent_id, child_id: $child_id }
                _set: { order: $order }
              ) {
                parent_id
                child_id
                order
              }
            }
          `,
          variables: {
            parent_id: goalId,
            child_id: child.id,
            order: i,
          },
        }),
      });
    }

    // Uppdatera data från server
    await refresh();
  } catch (error) {
    console.error("Failed to update child order:", error);
    // Uppdatera ändå från server för att synkronisera
    await refresh();
  }

  draggedChildIndex.value = null;
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
  const allChildren = children.value;

  if (currentIndex >= displayedChildren.length) {
    return;
  }

  // Hämta de två målen som ska bytas
  const movingChild = displayedChildren[currentIndex];
  const aboveChild = displayedChildren[currentIndex - 1];

  // Hitta deras ordningspositioner i full lista
  const movingChildFullIndex = allChildren.findIndex((c) => c.id === movingChild.id);
  const aboveChildFullIndex = allChildren.findIndex((c) => c.id === aboveChild.id);

  if (movingChildFullIndex === -1 || aboveChildFullIndex === -1) {
    return;
  }

  // Uppdatera ordningen för dessa två mål i databasen
  try {
    // Byt ordningen mellan de två målen
    const movingChildOrder = goal.value?.childRelations?.[movingChildFullIndex]?.order ?? movingChildFullIndex;
    const aboveChildOrder = goal.value?.childRelations?.[aboveChildFullIndex]?.order ?? aboveChildFullIndex;

    await Promise.all([
      fetch("http://localhost:8080/v1/graphql", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "x-hasura-admin-secret": config.public.hasuraAdminSecret,
        },
        body: JSON.stringify({
          query: `
            mutation UpdateGoalOrder($parent_id: Int!, $child_id: Int!, $order: Int!) {
              update_goal_relations_by_pk(
                pk_columns: { parent_id: $parent_id, child_id: $child_id }
                _set: { order: $order }
              ) {
                parent_id
                child_id
                order
              }
            }
          `,
          variables: {
            parent_id: goalId,
            child_id: movingChild.id,
            order: aboveChildOrder,
          },
        }),
      }),
      fetch("http://localhost:8080/v1/graphql", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "x-hasura-admin-secret": config.public.hasuraAdminSecret,
        },
        body: JSON.stringify({
          query: `
            mutation UpdateGoalOrder($parent_id: Int!, $child_id: Int!, $order: Int!) {
              update_goal_relations_by_pk(
                pk_columns: { parent_id: $parent_id, child_id: $child_id }
                _set: { order: $order }
              ) {
                parent_id
                child_id
                order
              }
            }
          `,
          variables: {
            parent_id: goalId,
            child_id: aboveChild.id,
            order: movingChildOrder,
          },
        }),
      }),
    ]);

    // Uppdatera markerad index
    selectedChildIndex.value = currentIndex - 1;

    // Uppdatera data från server
    await refresh();
  } catch (error) {
    console.error("Failed to move child up:", error);
    // Uppdatera ändå från server för att synkronisera
    await refresh();
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
  const allChildren = children.value;

  if (currentIndex >= displayedChildren.length - 1) {
    return;
  }

  // Hämta de två målen som ska bytas
  const movingChild = displayedChildren[currentIndex];
  const belowChild = displayedChildren[currentIndex + 1];

  // Hitta deras ordningspositioner i full lista
  const movingChildFullIndex = allChildren.findIndex((c) => c.id === movingChild.id);
  const belowChildFullIndex = allChildren.findIndex((c) => c.id === belowChild.id);

  if (movingChildFullIndex === -1 || belowChildFullIndex === -1) {
    return;
  }

  // Uppdatera ordningen för dessa två mål i databasen
  try {
    // Byt ordningen mellan de två målen
    const movingChildOrder = goal.value?.childRelations?.[movingChildFullIndex]?.order ?? movingChildFullIndex;
    const belowChildOrder = goal.value?.childRelations?.[belowChildFullIndex]?.order ?? belowChildFullIndex;

    await Promise.all([
      fetch("http://localhost:8080/v1/graphql", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "x-hasura-admin-secret": config.public.hasuraAdminSecret,
        },
        body: JSON.stringify({
          query: `
            mutation UpdateGoalOrder($parent_id: Int!, $child_id: Int!, $order: Int!) {
              update_goal_relations_by_pk(
                pk_columns: { parent_id: $parent_id, child_id: $child_id }
                _set: { order: $order }
              ) {
                parent_id
                child_id
                order
              }
            }
          `,
          variables: {
            parent_id: goalId,
            child_id: movingChild.id,
            order: belowChildOrder,
          },
        }),
      }),
      fetch("http://localhost:8080/v1/graphql", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "x-hasura-admin-secret": config.public.hasuraAdminSecret,
        },
        body: JSON.stringify({
          query: `
            mutation UpdateGoalOrder($parent_id: Int!, $child_id: Int!, $order: Int!) {
              update_goal_relations_by_pk(
                pk_columns: { parent_id: $parent_id, child_id: $child_id }
                _set: { order: $order }
              ) {
                parent_id
                child_id
                order
              }
            }
          `,
          variables: {
            parent_id: goalId,
            child_id: belowChild.id,
            order: movingChildOrder,
          },
        }),
      }),
    ]);

    // Uppdatera markerad index
    selectedChildIndex.value = currentIndex + 1;

    // Uppdatera data från server
    await refresh();
  } catch (error) {
    console.error("Failed to move child down:", error);
    // Uppdatera ändå från server för att synkronisera
    await refresh();
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
const isLeaderMode = ref(false);
let leaderTimeout: NodeJS.Timeout | null = null;

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
    await deleteGoal(editingGoalId.value);
    mode.value = "normal";
    editingGoalId.value = null;
    return;
  }

  try {
    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation UpdateGoalTitle($id: Int!, $title: String!) {
            update_goals_by_pk(pk_columns: { id: $id }, _set: { title: $title }) {
              id
              title
            }
          }
        `,
        variables: { id: editingGoalId.value, title: editTitle.value.trim() },
      }),
    });

    const result = await response.json();
    if (result.errors) {
      throw new Error(result.errors[0].message);
    }

    // Uppdatera lokal state
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
    await deleteGoal(editingGoalId.value);
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

  console.log("Calling deleteGoal with id:", goalToDelete.value.id);
  await deleteGoal(goalToDelete.value.id);
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

// Ta bort ett mål
async function deleteGoal(goalId: number) {
  try {
    // Ta bort målet från databasen
    const deleteGoalQuery = `
      mutation DeleteGoal($id: Int!) {
        delete_goal_relations(where: { _or: [{ child_id: { _eq: $id } }, { parent_id: { _eq: $id } }] }) {
          affected_rows
        }
        delete_user_goals(where: { goal_id: { _eq: $id } }) {
          affected_rows
        }
        delete_goals_by_pk(id: $id) {
          id
        }
      }
    `;

    const response = await $fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: deleteGoalQuery,
        variables: { id: goalId },
      }),
    });

    if (response.errors) {
      throw new Error(response.errors[0].message);
    }

    // Uppdatera lokal state
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
    // Skapa nytt mål med tom titel
    const createGoalQuery = `
      mutation CreateGoal($title: String!) {
        insert_goals_one(object: { title: $title, icon: "material-symbols:circle" }) {
          id
          title
          icon
          created
          finished
        }
      }
    `;

    const goalResponse = await $fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: createGoalQuery,
        variables: { title: "" },
      }),
    });

    if (goalResponse.errors) {
      throw new Error(goalResponse.errors[0].message);
    }

    if (goalResponse.data?.insert_goals_one?.id) {
      const newGoal = goalResponse.data.insert_goals_one;

      // Skapa user_goals relation
      const createUserGoalQuery = `
        mutation CreateUserGoal($userId: Int!, $goalId: Int!) {
          insert_user_goals_one(object: { user_id: $userId, goal_id: $goalId }) {
            user_id
            goal_id
          }
        }
      `;

      const userGoalResponse = await $fetch(
        "http://localhost:8080/v1/graphql",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "x-hasura-admin-secret": config.public.hasuraAdminSecret,
          },
          body: JSON.stringify({
            query: createUserGoalQuery,
            variables: { userId, goalId: newGoal.id },
          }),
        }
      );

      if (userGoalResponse.errors) {
        throw new Error(userGoalResponse.errors[0].message);
      }

      // Skapa relation till samma förälder (nuvarande mål)
      const addParentRelationQuery = `
        mutation AddParentRelation($childId: Int!, $parentId: Int!) {
          insert_goal_relations_one(object: { child_id: $childId, parent_id: $parentId }) {
            child_id
            parent_id
          }
        }
      `;

      const relationResponse = await $fetch(
        "http://localhost:8080/v1/graphql",
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "x-hasura-admin-secret": config.public.hasuraAdminSecret,
          },
          body: JSON.stringify({
            query: addParentRelationQuery,
            variables: { childId: newGoal.id, parentId: goalId },
          }),
        }
      );

      if (relationResponse.errors) {
        throw new Error(relationResponse.errors[0].message);
      }

      // Uppdatera lokal state
      goalsStore.addGoal(newGoal);
      goalsStore.addRelation(newGoal.id, goalId);

      // Uppdatera data från server
      await refresh();

      // Gå in i insert mode för det nya målet
      await enterInsertMode(newGoal.id, "", false);
    }
  } catch (error) {
    console.error("Failed to create sibling goal:", error);
  }
}

// Toggla finished status på ett mål
async function toggleFinished(goalToToggle: Goal) {
  try {
    const newFinishedValue = goalToToggle.finished ? null : new Date().toISOString();

    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation UpdateGoalFinished($id: Int!, $finished: timestamptz) {
            update_goals_by_pk(pk_columns: { id: $id }, _set: { finished: $finished }) {
              id
              title
              created
              finished
            }
          }
        `,
        variables: { id: goalToToggle.id, finished: newFinishedValue },
      }),
    });

    const result = await response.json();
    if (result.errors) {
      throw new Error(result.errors[0].message);
    }

    // Uppdatera lokal state i store
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

// Uppdatera goal icon
async function updateGoalIcon(newIcon: string) {
  try {
    // Bestäm vilket mål som ska uppdateras
    const targetGoalId = editingIconGoalId.value || goalId;

    const response = await fetch("http://localhost:8080/v1/graphql", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-hasura-admin-secret": config.public.hasuraAdminSecret,
      },
      body: JSON.stringify({
        query: `
          mutation UpdateGoalIcon($id: Int!, $icon: String!) {
            update_goals_by_pk(pk_columns: { id: $id }, _set: { icon: $icon }) {
              id
              icon
            }
          }
        `,
        variables: { id: targetGoalId, icon: newIcon },
      }),
    });

    const result = await response.json();
    if (result.errors) {
      throw new Error(result.errors[0].message);
    }

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

// Hantera Vim-kommandon
function handleKeydown(event: KeyboardEvent) {
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

  // Normal mode - hantera alla kommandon
  // Hantera 'd' för att toggla finished status
  if (event.key === "d") {
    event.preventDefault();
    if (isParentMode.value && parents.value.length > 0) {
      const parent = parents.value[selectedParentIndex.value];
      toggleFinished(parent);
    } else if (filteredChildren.value.length > 0) {
      const child = filteredChildren.value[selectedChildIndex.value];
      toggleFinished(child);
    }
    return;
  }

  // Hantera 'x' för att ta bort mål
  if (event.key === "x") {
    event.preventDefault();
    if (isParentMode.value && parents.value.length > 0) {
      const parent = parents.value[selectedParentIndex.value];
      confirmDeleteGoal(parent.id, parent.title);
    } else if (filteredChildren.value.length > 0) {
      const child = filteredChildren.value[selectedChildIndex.value];
      confirmDeleteGoal(child.id, child.title);
    }
    return;
  }

  // Hantera 'i' och 'a' för att gå in i insert mode
  if (event.key === "i" || event.key === "a") {
    event.preventDefault();
    // 'i' = början, 'a' = slutet
    const atBeginning = event.key === "i";

    if (isGoalSelected.value) {
      // Redigera själva målets titel
      if (goal.value) {
        enterInsertMode(goal.value.id, goal.value.title, atBeginning);
      }
    } else if (isParentMode.value && parents.value.length > 0) {
      const parent = parents.value[selectedParentIndex.value];
      enterInsertMode(parent.id, parent.title, atBeginning);
    } else if (filteredChildren.value.length > 0) {
      const child = filteredChildren.value[selectedChildIndex.value];
      enterInsertMode(child.id, child.title, atBeginning);
    }
    return;
  }

  // Hantera leader key (space)
  if (event.key === " " && !isLeaderMode.value) {
    event.preventDefault();
    isLeaderMode.value = true;
    // Återställ leader mode efter 1 sekund om ingen command följer
    if (leaderTimeout) clearTimeout(leaderTimeout);
    leaderTimeout = setTimeout(() => {
      isLeaderMode.value = false;
    }, 1000);
    return;
  }

  // Hantera leader commands
  if (isLeaderMode.value) {
    event.preventDefault();
    if (leaderTimeout) clearTimeout(leaderTimeout);
    isLeaderMode.value = false;

    if (event.key === "d") {
      // Toggla visa/dölj avklarade mål
      showCompleted.value = !showCompleted.value;
    }
    return;
  }

  if (isParentMode.value) {
    // Parent mode - navigera mellan föräldrar
    if (event.key === "h") {
      event.preventDefault();
      selectedParentIndex.value = Math.max(selectedParentIndex.value - 1, 0);
    } else if (event.key === "l") {
      event.preventDefault();
      const parentsCount = parents.value.length;
      selectedParentIndex.value = Math.min(
        selectedParentIndex.value + 1,
        parentsCount - 1
      );
    } else if (event.key === "k") {
      event.preventDefault();
      // Navigera till markerad förälder
      const selectedParent = parents.value[selectedParentIndex.value];
      if (selectedParent) {
        router.push(`/goal/${selectedParent.id}`);
      }
    } else if (event.key === "j") {
      event.preventDefault();
      // Gå till goal mode
      isParentMode.value = false;
      isGoalSelected.value = true;
      selectedChildIndex.value = -1;
    } else if (event.key === "Enter") {
      event.preventDefault();
      // Öppna föräldersök (samma som + knappen)
      toggleParentSearch();
    }
  } else {
    // Child mode - navigera mellan undermål
    if (event.key === "j") {
      event.preventDefault();
      if (isGoalSelected.value) {
        // Från goal mode: markera första barnet
        isGoalSelected.value = false;
        selectedChildIndex.value = 0;
      } else {
        const childrenCount = filteredChildren.value.length;
        if (childrenCount > 0) {
          selectedChildIndex.value = Math.min(
            selectedChildIndex.value + 1,
            childrenCount - 1
          );
        }
      }
    } else if (event.key === "k") {
      event.preventDefault();
      if (selectedChildIndex.value === 0 && !isGoalSelected.value) {
        // Markera själva målet istället för att navigera vidare
        isGoalSelected.value = true;
        selectedChildIndex.value = -1; // Avmarkera barnen
      } else if (isGoalSelected.value) {
        // Från goal mode: gå in i parent mode
        if (parents.value.length > 0) {
          isGoalSelected.value = false;
          isParentMode.value = true;
          selectedParentIndex.value = 0;
        }
      } else {
        const childrenCount = filteredChildren.value.length;
        if (childrenCount > 0) {
          selectedChildIndex.value = Math.max(selectedChildIndex.value - 1, 0);
        }
      }
    } else if (event.key === "Enter") {
      event.preventDefault();
      // Skapa nytt mål på samma nivå
      createSiblingGoal();
    } else if (event.key === "l") {
      event.preventDefault();
      if (isGoalSelected.value) {
        // Från goal mode: markera första barnet
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
      // Gå till första föräldern
      if (parents.value.length > 0) {
        router.push(`/goal/${parents.value[0].id}`);
      }
    } else if (event.key === "ArrowUp") {
      event.preventDefault();
      // Flytta markerat barn upp i listan
      moveChildUp();
    } else if (event.key === "ArrowDown") {
      event.preventDefault();
      // Flytta markerat barn ner i listan
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
  if (leaderTimeout) clearTimeout(leaderTimeout);
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
          <NuxtLink v-if="parents.length === 0" to="/goal/1"
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
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
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
        <button v-if="goal" @click.stop="showIconPicker = true"
          class="text-gray-400 hover:text-gray-200 transition-colors p-2 rounded hover:bg-gray-800"
          title="Ändra ikon">
          <Icon :name="goal.icon || 'material-symbols:circle'" class="w-8 h-8" />
        </button>
      </div>

      <!-- Undermål - scrollbar container -->
      <div class="flex-1 overflow-y-auto min-h-0">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-2xl font-semibold text-gray-300"></h2>
          <div class="flex items-center gap-2">
            <button @click="showCompleted = !showCompleted"
              class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800" :title="showCompleted ? 'Dölj avklarade mål' : 'Visa avklarade mål'
                ">
              <svg v-if="showCompleted" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none"
                viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24"
                stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
              </svg>
            </button>
            <button @click="toggleChildSearch"
              class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800"
              title="Lägg till undermål">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24"
                stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Sökfält för undermål -->
        <div v-if="showChildSearch" class="border border-gray-700 rounded-lg p-4 bg-gray-800 mb-4">
          <div class="mb-2">
            <input ref="childSearchInput" v-model="childSearchQuery" @keydown="handleChildSearchKeydown" type="text"
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

        <ul v-if="filteredChildren.length > 0" class="space-y-3">
          <li v-for="(child, index) in filteredChildren" :key="child.id" :data-child-index="index"
            draggable="true" @dragstart="handleDragStart($event, index)" @dragover="handleDragOver($event, index)" @dragleave="handleDragLeave" @drop="handleDrop($event, index)"
            class="relative overflow-hidden rounded-lg transition-opacity" :class="dragOverChildIndex === index ? 'opacity-50' : 'opacity-100'">
            <!-- Swipe bakgrund -->
            <div class="absolute inset-0 flex items-center justify-start px-6"
              :class="child.finished ? 'bg-red-900/50' : 'bg-green-900/50'">
              <span class="text-2xl">{{ child.finished ? "↩️" : "✓" }}</span>
            </div>

            <!-- Huvudinnehåll -->
            <div class="relative border rounded-lg transition-all bg-gray-900" :class="selectedChildIndex === index
              ? 'border-blue-500 bg-blue-900/20'
              : 'border-gray-700'
              " :style="{
                transform: `translateX(${getSwipeOffset(child.id)}px)`,
                transition: swipeState.isSwiping
                  ? 'none'
                  : 'transform 0.3s ease',
              }" @touchstart="handleTouchStart($event, child.id)" @touchmove="handleTouchMove($event)"
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
              <NuxtLink v-else :to="`/goal/${child.id}`"
                class="p-4 flex items-start justify-between block hover:bg-gray-800/50">
                <div class="flex-1 flex items-center gap-3">
                  <button @click.stop="editingIconGoalId = child.id; showIconPicker = true"
                    class="flex-shrink-0 text-gray-400 hover:text-gray-200 transition-colors rounded p-1 hover:bg-gray-600"
                    title="Ändra ikon">
                    <Icon :name="child.icon || 'material-symbols:circle'" class="w-6 h-6" />
                  </button>
                  <h3 class="text-lg font-medium" :class="child.finished ? 'text-gray-500' : 'text-gray-200'">
                    {{ child.title }}
                  </h3>
                </div>
              </NuxtLink>
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
    <IconPicker
      v-if="goal"
      :modelValue="editingIconGoalId ? goalData?.allGoals?.find(g => g.id === editingIconGoalId)?.icon || '' : goal.icon"
      :open="showIconPicker"
      @update:open="(value) => {
        showIconPicker = value;
        if (!value) {
          editingIconGoalId = null;
        }
      }"
      @update:modelValue="updateGoalIcon"
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
