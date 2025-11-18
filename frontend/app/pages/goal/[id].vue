<script setup lang="ts">
// Make this page client-only to ensure GraphQL requests work properly and prevent Pinia hydration issues
definePageMeta({
  ssr: false
 });



 const route = useRoute();
const router = useRouter();
const goalId = parseInt(route.params.id as string);
const { user } = useUserSession();
const config = useRuntimeConfig();

interface Goal {
  id: number;
  title: string;
  created: string;
  finished: string | null;
}

interface ChildRelation {
  child_id: number;
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
    const response = await fetch('http://localhost:8080/v1/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-hasura-admin-secret': config.public.hasuraAdminSecret
      },
      body: JSON.stringify({
        query: `
          query GetGoal($id: Int!) {
            goal: goals_by_pk(id: $id) {
              id
              title
              created
              finished

              # Barn - relationer där detta mål är parent
              childRelations: goalRelationsByParentId {
                child_id
              }

              # Förälder - relationer där detta mål är child
              parentRelations: goal_relations {
                goalByParentId {
                  id
                  title
                }
              }
            }

            # Hämta alla mål för att kunna matcha children
            allGoals: goals {
              id
              title
              finished
              created
            }
          }
        `,
        variables: { id: goalId }
      })
    });

    const result = await response.json();
    console.log('GetGoal result:', result);

    if (result.errors) {
      throw new Error(result.errors[0].message);
    }

    goalData.value = result.data;
  } catch (err) {
    console.error('Fetch goal error:', err);
    error.value = err instanceof Error ? err.message : 'Unknown error';
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
  () => goal.value?.parentRelations?.map((r) => r.goalByParentId) || [],
);

// Matcha barn-ID:n med faktiska goal-objekt
const children = computed(() => {
  if (!goal.value?.childRelations || !goalData.value?.allGoals) return [];

  const childIds = goal.value.childRelations.map((r) => r.child_id);
  return goalData.value.allGoals.filter((g) => childIds.includes(g.id));
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
// Temporarily disabled to debug
// onMounted(async () => {
//   if (!goalsStore.isLoaded) {
//     await goalsStore.loadGoals();
//   }
// });

// Visa/dölj avklarade mål
const showCompleted = ref(true);

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
    (g) => g.id !== goalId && !currentParentIds.includes(g.id),
  );
});

// Lägg till befintligt mål som förälder
async function addExistingParent(parentId: number) {
  try {
    const response = await fetch('http://localhost:8080/v1/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-hasura-admin-secret': config.public.hasuraAdminSecret
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
        variables: { childId: goalId, parentId }
      })
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
        insert_goals_one(object: { title: $title }) {
          id
          title
          created
          finished
        }
      }
    `;

    const goalResponse = await $fetch('http://localhost:8080/v1/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-hasura-admin-secret': config.public.hasuraAdminSecret
      },
      body: JSON.stringify({
        query: createGoalQuery,
        variables: { title: parentSearchQuery.value.trim() }
      })
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

      const userGoalResponse = await $fetch('http://localhost:8080/v1/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'x-hasura-admin-secret': config.public.hasuraAdminSecret
        },
        body: JSON.stringify({
          query: createUserGoalQuery,
          variables: { userId, goalId: newGoal.id }
        })
      });

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

      const relationResponse = await $fetch('http://localhost:8080/v1/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'x-hasura-admin-secret': config.public.hasuraAdminSecret
        },
        body: JSON.stringify({
          query: addParentRelationQuery,
          variables: { childId: goalId, parentId: newGoal.id }
        })
      });

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
    (g) => g.id !== goalId && !currentChildIds.includes(g.id),
  );
});

// Lägg till befintligt mål som barn
async function addExistingChild(childId: number) {
  try {
    const response = await fetch('http://localhost:8080/v1/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-hasura-admin-secret': config.public.hasuraAdminSecret
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
        variables: { childId, parentId: goalId }
      })
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
      childSearchQuery.value.trim(),
    );

    // Skapa nytt mål
    const createGoalQuery = `
      mutation CreateGoal($title: String!) {
        insert_goals_one(object: { title: $title }) {
          id
          title
          created
          finished
        }
      }
    `;

    const goalResponse = await $fetch('http://localhost:8080/v1/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-hasura-admin-secret': config.public.hasuraAdminSecret
      },
      body: JSON.stringify({
        query: createGoalQuery,
        variables: { title: childSearchQuery.value.trim() }
      })
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
        goalId,
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

      const userGoalResponse = await $fetch('http://localhost:8080/v1/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'x-hasura-admin-secret': config.public.hasuraAdminSecret
        },
        body: JSON.stringify({
          query: createUserGoalQuery,
          variables: { userId, goalId: newGoal.id }
        })
      });

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

      const relationResponse = await $fetch('http://localhost:8080/v1/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'x-hasura-admin-secret': config.public.hasuraAdminSecret
        },
        body: JSON.stringify({
          query: addParentRelationQuery,
          variables: { childId: newGoal.id, parentId: goalId }
        })
      });

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
    const response = await fetch('http://localhost:8080/v1/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-hasura-admin-secret': config.public.hasuraAdminSecret
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
        variables: { childId: goalId, parentId: parentToRemove.value }
      })
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
    try {
      const newFinishedValue = child.finished ? null : new Date().toISOString();

      const response = await fetch('http://localhost:8080/v1/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'x-hasura-admin-secret': config.public.hasuraAdminSecret
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
          variables: { id: child.id, finished: newFinishedValue }
        })
      });

      const result = await response.json();
      if (result.errors) {
        throw new Error(result.errors[0].message);
      }

      // Uppdatera lokal state i store
      goalsStore.updateGoal(child.id, { finished: newFinishedValue });

      // Uppdatera lokal state i component
      if (goalData.value?.allGoals) {
        const goalIndex = goalData.value.allGoals.findIndex(
          (g) => g.id === child.id,
        );
        if (goalIndex !== -1) {
          goalData.value.allGoals[goalIndex].finished = newFinishedValue;
        }
      }

      // Uppdatera data från server
      await refresh();
    } catch (error) {
      console.error("Failed to update goal finished status:", error);
    }
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

// Vim-style navigering för undermål
const selectedChildIndex = ref(0);

// Hantera Vim-kommandon
function handleKeydown(event: KeyboardEvent) {
  // Ignorera om sökfält är aktiva
  if (showParentSearch.value || showChildSearch.value) return;

  if (event.key === "j") {
    event.preventDefault();
    const childrenCount = filteredChildren.value.length;
    if (childrenCount > 0) {
      selectedChildIndex.value = Math.min(selectedChildIndex.value + 1, childrenCount - 1);
    }
  } else if (event.key === "k") {
    event.preventDefault();
    const childrenCount = filteredChildren.value.length;
    if (childrenCount > 0) {
      selectedChildIndex.value = Math.max(selectedChildIndex.value - 1, 0);
    }
  } else if (event.key === "Enter" || event.key === "l") {
    event.preventDefault();
    const childrenCount = filteredChildren.value.length;
    if (childrenCount > 0) {
      const selectedChild = filteredChildren.value[selectedChildIndex.value];
      if (selectedChild) {
        router.push(`/goal/${selectedChild.id}`);
      }
    }
  } else if (event.key === "h") {
    event.preventDefault();
    // Kolla om föregående sida var en förälder
    const referrer = document.referrer;
    const parentIds = parents.value.map(p => p.id);

    // Parse goal ID från referrer (om det finns)
    const referrerMatch = referrer.match(/\/goal\/(\d+)/);
    const referrerId = referrerMatch ? parseInt(referrerMatch[1]) : null;

    if (referrerId && parentIds.includes(referrerId)) {
      // Föregående sida var en förälder - gå tillbaka
      router.back();
    } else if (parents.value.length > 0) {
      // Navigera till första föräldern
      router.push(`/goal/${parents.value[0].id}`);
    } else {
      // Inga föräldrar - gå till root-goals
      router.push('/root-goals');
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
</script>

<template>
  <div v-if="user" class="max-w-4xl mx-auto p-6">
    <div v-if="!goal" class="text-center py-12">
      <p class="text-gray-500">Laddar mål...</p>
    </div>

    <div v-else class="space-y-8">
      <!-- Breadcrumb / Föräldrar -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2 text-sm flex-wrap">
          <NuxtLink
            v-if="parents.length === 0"
            to="/root-goals"
            class="text-blue-400 hover:text-blue-300 transition-colors font-medium"
          >
            Grundmål
          </NuxtLink>
          <NuxtLink
            v-for="parent in parents"
            :key="parent.id"
            :to="`/goal/${parent.id}`"
            class="text-gray-500 hover:text-gray-300 transition-colors select-none"
            @mousedown="handleParentMouseDown(parent.id)"
            @mouseup="handleParentMouseUp"
            @mouseleave="handleParentMouseUp"
            @touchstart="handleParentMouseDown(parent.id)"
            @touchend="handleParentMouseUp"
            @touchcancel="handleParentMouseUp"
          >
            {{ parent.title }}
          </NuxtLink>
        </div>

        <!-- + knapp för att lägga till förälder -->
        <button
          @click="toggleParentSearch"
          class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800"
          title="Lägg till förälder"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-5 w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M12 4v16m8-8H4"
            />
          </svg>
        </button>
      </div>

      <!-- Sökfält för föräldrar -->
      <div
        v-if="showParentSearch"
        class="border border-gray-700 rounded-lg p-4 bg-gray-800"
      >
        <div class="mb-2">
          <input
            ref="parentSearchInput"
            v-model="parentSearchQuery"
            @keydown="handleSearchKeydown"
            type="text"
            placeholder="Sök efter mål eller skapa nytt (tryck Enter)"
            class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-gray-200 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
            autofocus
          />
        </div>

        <!-- Sökresultat -->
        <div v-if="searchResults.length > 0" class="space-y-1 mt-2">
          <button
            v-for="result in searchResults"
            :key="result.id"
            @click="addExistingParent(result.id)"
            class="w-full text-left px-4 py-2 hover:bg-gray-700 rounded-lg transition-colors"
          >
            <div class="font-medium text-gray-200">{{ result.title }}</div>
            <div class="text-xs text-gray-500">
              {{ new Date(result.created).toLocaleDateString("sv-SE") }}
            </div>
          </button>
        </div>

        <!-- Meddelande när inga resultat -->
        <div
          v-else-if="parentSearchQuery.trim()"
          class="text-sm text-gray-500 px-4 py-2"
        >
          Tryck Enter för att skapa "{{ parentSearchQuery }}" som nytt mål
        </div>
      </div>

      <!-- Huvudmål -->
      <h1 class="text-4xl font-bold text-gray-100">{{ goal.title }}</h1>

      <!-- Undermål -->
      <div>
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-2xl font-semibold text-gray-300"></h2>
          <div class="flex items-center gap-2">
            <button
              @click="showCompleted = !showCompleted"
              class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800"
              :title="
                showCompleted ? 'Dölj avklarade mål' : 'Visa avklarade mål'
              "
            >
              <svg
                v-if="showCompleted"
                xmlns="http://www.w3.org/2000/svg"
                class="h-5 w-5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                />
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                />
              </svg>
              <svg
                v-else
                xmlns="http://www.w3.org/2000/svg"
                class="h-5 w-5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"
                />
              </svg>
            </button>
            <button
              @click="toggleChildSearch"
              class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800"
              title="Lägg till undermål"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-5 w-5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 4v16m8-8H4"
                />
              </svg>
            </button>
          </div>
        </div>

        <!-- Sökfält för undermål -->
        <div
          v-if="showChildSearch"
          class="border border-gray-700 rounded-lg p-4 bg-gray-800 mb-4"
        >
          <div class="mb-2">
            <input
              ref="childSearchInput"
              v-model="childSearchQuery"
              @keydown="handleChildSearchKeydown"
              type="text"
              placeholder="Sök efter mål eller skapa nytt (tryck Enter)"
              class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-gray-200 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
              autofocus
            />
          </div>

          <!-- Sökresultat -->
          <div v-if="childSearchResults.length > 0" class="space-y-1 mt-2">
            <button
              v-for="result in childSearchResults"
              :key="result.id"
              @click="addExistingChild(result.id)"
              class="w-full text-left px-4 py-2 hover:bg-gray-700 rounded-lg transition-colors"
            >
              <div class="font-medium text-gray-200">{{ result.title }}</div>
              <div class="text-xs text-gray-500">
                {{ new Date(result.created).toLocaleDateString("sv-SE") }}
              </div>
            </button>
          </div>

          <!-- Meddelande när inga resultat -->
          <div
            v-else-if="childSearchQuery.trim()"
            class="text-sm text-gray-500 px-4 py-2"
          >
            Tryck Enter för att skapa "{{ childSearchQuery }}" som nytt undermål
          </div>
        </div>

        <ul v-if="filteredChildren.length > 0" class="space-y-3">
          <li
            v-for="(child, index) in filteredChildren"
            :key="child.id"
            class="relative overflow-hidden rounded-lg"
          >
            <!-- Swipe bakgrund -->
            <div
              class="absolute inset-0 flex items-center justify-start px-6"
              :class="child.finished ? 'bg-red-900/50' : 'bg-green-900/50'"
            >
              <span class="text-2xl">{{ child.finished ? "↩️" : "✓" }}</span>
            </div>

            <!-- Huvudinnehåll -->
            <div
              class="relative border rounded-lg hover:border-gray-600 transition-all bg-gray-900"
              :class="selectedChildIndex === index ? 'border-blue-500 bg-blue-900/20' : 'border-gray-700'"
              :style="{
                transform: `translateX(${getSwipeOffset(child.id)}px)`,
                transition: swipeState.isSwiping
                  ? 'none'
                  : 'transform 0.3s ease',
              }"
              @touchstart="handleTouchStart($event, child.id)"
              @touchmove="handleTouchMove($event)"
              @touchend="handleTouchEnd(child)"
            >
              <NuxtLink
                :to="`/goal/${child.id}`"
                class="p-4 flex items-start justify-between block"
              >
                <div class="flex-1">
                  <h3
                    class="text-lg font-medium"
                    :class="child.finished ? 'text-gray-500' : 'text-gray-200'"
                  >
                    {{ child.title }}
                  </h3>
                </div>
              </NuxtLink>
            </div>
          </li>
        </ul>

        <div
          v-else
          class="text-gray-500 p-6 border border-gray-700 rounded-lg text-center"
        >
          Inga undermål ännu. Skapa ett för att dela upp detta mål i mindre
          delar.
        </div>
      </div>
    </div>

    <!-- Modal för att bekräfta borttagning av föräldrarelation -->
    <UModal
      v-model:open="showRemoveConfirmation"
      title="Ta bort föräldrarelation?"
    >
      <p v-if="parentToRemove" class="text-gray-400">
        Vill du ta bort relationen till
        <strong>{{
          parents.find((p) => p.id === parentToRemove)?.title
        }}</strong
        >? Själva målet kommer inte tas bort.
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
  </div>
  <div v-else class="max-w-4xl mx-auto p-6">
    <div class="text-center">
      <h1 class="text-3xl font-bold text-gray-300 mb-4">Du måste logga in</h1>
      <p class="text-gray-600 mb-6">För att se dina mål måste du vara inloggad.</p>
      <NuxtLink to="/login">
        <UButton>Logga in</UButton>
      </NuxtLink>
    </div>
  </div>
</template>
