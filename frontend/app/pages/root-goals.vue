<script setup lang="ts">
// Make this page client-only to ensure user session is available and prevent Pinia hydration issues
definePageMeta({
  ssr: false
});

interface ChildRelation {
  child_id: number;
}

interface Goal {
  id: number;
  title: string;
  created: string;
  finished: string | null;
  goalRelationsByParentId: ChildRelation[];
}

interface UserGoal {
  goal: Goal;
}

interface GoalRelation {
  child_id: number;
}

interface GetRootGoalsResponse {
  user_goals: UserGoal[];
  goal_relations: GoalRelation[];
}

const { user } = useUserSession();
const config = useRuntimeConfig();

console.log('Initial user session:', user.value);
console.log('Initial user ID:', user.value?.id);

// Watch user changes
watch(user, (newUser, oldUser) => {
  console.log('User changed from:', oldUser, 'to:', newUser);
}, { immediate: true });

// Direct Hasura fetch approach
const goalsData = ref<GetRootGoalsResponse | null>(null);
const isLoading = ref(false);
const error = ref(null);

const fetchGoals = async () => {
  isLoading.value = true;
  error.value = null;

  try {
    console.log('Fetching root goals');
    const query = `
      query GetRootGoals {
        user_goals {
          goal {
            id
            title
            created
            finished
            childRelations: goalRelationsByParentId {
              child_id
            }
          }
        }
        goal_relations {
          child_id
        }
      }
    `;

    const response = await $fetch('http://localhost:8080/v1/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-hasura-admin-secret': config.public.hasuraAdminSecret
      },
      body: JSON.stringify({ query })
    });

    if (response.errors) {
      throw new Error(response.errors[0].message);
    }

    if (response.data) {
      goalsData.value = response.data;
      console.log('Fetched goals data:', response.data);
    }
  } catch (err) {
    console.error('Fetch error:', err);
    error.value = err;
  } finally {
    isLoading.value = false;
  }
};



const goals = computed(() => {
  const allUserGoals = goalsData.value?.user_goals || [];
  const allRelations = goalsData.value?.goal_relations || [];

  // Get set of all goal IDs that are children (have parents)
  const childGoalIds = new Set(allRelations.map(r => r.child_id));

  // Filter to only include root goals (goals that are not children)
  const result = allUserGoals
    .map(ug => ug.goal)
    .filter(goal => !childGoalIds.has(goal.id));

  console.log('All user goals:', allUserGoals);
  console.log('Child goal IDs:', Array.from(childGoalIds));
  console.log('Filtered root goals:', result);
  return result;
});

console.log('Goals data:', goalsData.value);

// Beräkna progress för varje mål
const getProgress = (goal: Goal) => {
  if (!goal.goalRelationsByParentId || goal.goalRelationsByParentId.length === 0) return null;
  return goal.goalRelationsByParentId.length;
};

// Använd goals store
const goalsStore = useGoalsStore();

// Ladda goals om de inte är laddade
const loadUserGoals = async () => {
  if (!user.value) return;

  const userId = user.value.id;
  console.log('Loading goals for user:', userId);

  try {
    const goalsQuery = `
      query getUserGoals($userId: Int!) {
        goals(where: {user_goals: {user_id: {_eq: $userId}}}) {
          id
          title
          created
          finished
        }
      }
    `;

    const goalsResponse = await $fetch('http://localhost:8080/v1/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-hasura-admin-secret': config.public.hasuraAdminSecret
      },
      body: JSON.stringify({ query: goalsQuery, variables: { userId } })
    });

    console.log('Goals query result:', goalsResponse.data, 'error:', goalsResponse.errors);

    if (goalsResponse.errors) {
      console.error('Failed to load goals:', goalsResponse.errors);
      return;
    }

    if (goalsResponse.data?.goals) {
      goalsStore.goals = goalsResponse.data.goals;
      console.log('Set goals in store:', goalsResponse.data.goals.length);
    }

    const relationsQuery = `
      query GetUserRelations {
        goal_relations {
          parent_id
          child_id
        }
      }
    `;

    const relationsResponse = await $fetch('http://localhost:8080/v1/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-hasura-admin-secret': config.public.hasuraAdminSecret
      },
      body: JSON.stringify({ query: relationsQuery })
    });

    console.log('Relations query result:', relationsResponse.data, 'error:', relationsResponse.errors);

    if (relationsResponse.errors) {
      console.error('Failed to load relations:', relationsResponse.errors);
      return;
    }

    if (relationsResponse.data?.goal_relations) {
      // Filter relations to only those involving user's goals
      const userGoalIds = new Set(goalsStore.goals.map(g => g.id));
      goalsStore.relations = relationsResponse.data.goal_relations.filter(r =>
        userGoalIds.has(r.parent_id) && userGoalIds.has(r.child_id)
      );
      console.log('Set relations in store:', goalsStore.relations.length);
    }

    goalsStore.isLoaded = true;
  } catch (error) {
    console.error('Error loading goals:', error);
  }
};

// Watch for user changes and load goals
watch(user, async (newUser) => {
  console.log('User changed:', newUser);
  if (newUser && !goalsStore.isLoaded) {
    await loadUserGoals();
  }
  if (newUser && !goalsData.value) {
    await fetchGoals();
  }
}, { immediate: true });

// Sökfunktion för att lägga till nya grundmål
const showSearch = ref(false);
const searchQuery = ref("");
const searchInput = ref<HTMLInputElement | null>(null);

// Fokusera input när plus-knappen klickas
async function toggleSearch() {
  showSearch.value = !showSearch.value;
  if (showSearch.value) {
    await nextTick();
    searchInput.value?.focus();
  }
}

// Sök i lokal state istället för GraphQL
const searchResults = computed(() => {
  if (!searchQuery.value.trim()) return [];

  const results = goalsStore.searchGoals(searchQuery.value);

  // Filtrera bort mål som redan är grundmål
  const currentRootGoalIds = goals.value.map((g) => g.id);
  return results.filter((g) => !currentRootGoalIds.includes(g.id));
});

// Skapa nytt grundmål
async function createNewGoal() {
  if (!searchQuery.value.trim()) return;

  const userId = user.value?.id;
  if (!userId) {
    console.error("No user logged in");
    return;
  }

  try {
    console.log("Creating goal with:", { title: searchQuery.value.trim(), userId });

    // Skapa nytt mål först
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
        variables: { title: searchQuery.value.trim() }
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

      // Uppdatera lokal state
      goalsStore.addGoal(newGoal);

      // Uppdatera data från server
      await goalsStore.loadGoals();

      // Stäng och rensa sökning
      showSearch.value = false;
      searchQuery.value = "";
    } else {
      console.error("No goal ID in response:", goalResponse.data);
    }
  } catch (error) {
    console.error("Failed to create new goal:", error);
  }
}

// Hantera Enter-tangent
function handleSearchKeydown(event: KeyboardEvent) {
  if (event.key === "Enter") {
    // Skapa nytt mål
    createNewGoal();
  } else if (event.key === "Escape") {
    showSearch.value = false;
    searchQuery.value = "";
  }
}

// Vim-style navigering
const selectedIndex = ref(0);
const router = useRouter();

// Vim modes
const mode = ref<'normal' | 'insert'>('normal');
const editingGoalId = ref<number | null>(null);
const editTitle = ref('');

// Skapa nytt grundmål
async function createNewRootGoal() {
  const userId = user.value?.id;
  if (!userId) {
    console.error("No user logged in");
    return;
  }

  try {
    // Skapa nytt mål med tom titel
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
        variables: { title: '' }
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

      // Uppdatera lokal state
      goalsStore.addGoal(newGoal);
      await fetchGoals();

      // Gå in i insert mode för det nya målet
      await enterInsertMode(newGoal.id, '');
    }
  } catch (error) {
    console.error("Failed to create root goal:", error);
  }
}

// Gå in i insert mode för att redigera ett mål
async function enterInsertMode(goalId: number, title: string) {
  mode.value = 'insert';
  editingGoalId.value = goalId;
  editTitle.value = title;

  await nextTick();
  await nextTick();

  const input = document.querySelector('input.border-blue-500') as HTMLInputElement;
  if (input) {
    input.focus();
    input.setSelectionRange(title.length, title.length);
  }
}

// Spara ändringar
async function saveEdit() {
  if (!editingGoalId.value) {
    mode.value = 'normal';
    return;
  }

  // Om titeln är tom, ta bort målet
  if (!editTitle.value.trim()) {
    await deleteGoal(editingGoalId.value);
    mode.value = 'normal';
    editingGoalId.value = null;
    return;
  }

  try {
    const response = await $fetch('http://localhost:8080/v1/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-hasura-admin-secret': config.public.hasuraAdminSecret
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
        variables: { id: editingGoalId.value, title: editTitle.value.trim() }
      })
    });

    if (response.errors) {
      throw new Error(response.errors[0].message);
    }

    goalsStore.updateGoal(editingGoalId.value, { title: editTitle.value.trim() });
    await fetchGoals();
  } catch (error) {
    console.error("Failed to update goal title:", error);
  }

  mode.value = 'normal';
  editingGoalId.value = null;
}

// Avbryt redigering
async function cancelEdit() {
  if (editingGoalId.value && editTitle.value === '') {
    await deleteGoal(editingGoalId.value);
  }
  mode.value = 'normal';
  editingGoalId.value = null;
  editTitle.value = '';
}

// Ta bort ett mål
async function deleteGoal(goalId: number) {
  try {
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

    const response = await $fetch('http://localhost:8080/v1/graphql', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-hasura-admin-secret': config.public.hasuraAdminSecret
      },
      body: JSON.stringify({
        query: deleteGoalQuery,
        variables: { id: goalId }
      })
    });

    if (response.errors) {
      throw new Error(response.errors[0].message);
    }

    goalsStore.removeGoal(goalId);
    await fetchGoals();
  } catch (error) {
    console.error("Failed to delete goal:", error);
  }
}

// Toggla finished status på ett mål
async function toggleFinished(goalToToggle: Goal) {
  try {
    const newFinishedValue = goalToToggle.finished ? null : new Date().toISOString();

    const response = await $fetch('http://localhost:8080/v1/graphql', {
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
        variables: { id: goalToToggle.id, finished: newFinishedValue }
      })
    });

    if (response.errors) {
      throw new Error(response.errors[0].message);
    }

    // Uppdatera lokal state
    goalsStore.updateGoal(goalToToggle.id, { finished: newFinishedValue });
    await fetchGoals();
  } catch (error) {
    console.error("Failed to toggle goal finished status:", error);
  }
}

// Hantera Vim-kommandon
function handleKeydown(event: KeyboardEvent) {
  // Ignorera om sökfältet är aktivt
  if (showSearch.value) return;

  // Insert mode - ignorera alla tangenter (hanteras direkt på input-elementet)
  if (mode.value === 'insert') {
    return;
  }

  const goalsCount = goals.value.length;
  if (goalsCount === 0) return;

  if (event.key === "j") {
    event.preventDefault();
    selectedIndex.value = Math.min(selectedIndex.value + 1, goalsCount - 1);
  } else if (event.key === "k") {
    event.preventDefault();
    selectedIndex.value = Math.max(selectedIndex.value - 1, 0);
  } else if (event.key === "d") {
    event.preventDefault();
    const selectedGoal = goals.value[selectedIndex.value];
    if (selectedGoal) {
      toggleFinished(selectedGoal);
    }
  } else if (event.key === "Enter") {
    event.preventDefault();
    // Skapa nytt grundmål
    createNewRootGoal();
  } else if (event.key === "l") {
    event.preventDefault();
    const selectedGoal = goals.value[selectedIndex.value];
    if (selectedGoal) {
      router.push(`/goal/${selectedGoal.id}`);
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

// Återställ selectedIndex när goals ändras
watch(goals, () => {
  if (selectedIndex.value >= goals.value.length) {
    selectedIndex.value = Math.max(0, goals.value.length - 1);
  }
});

// Scrolla till markerat mål när selectedIndex ändras
watch(selectedIndex, async () => {
  await nextTick();
  const selectedElement = document.querySelector(`[data-goal-index="${selectedIndex.value}"]`) as HTMLElement;
  if (selectedElement) {
    selectedElement.scrollIntoView({
      behavior: 'smooth',
      block: 'nearest'
    });
  }
});
</script>

<template>
  <div v-if="user" class="max-w-4xl mx-auto p-6">
    <div class="mb-8">
      <div class="flex items-center justify-between mb-2">
        <h1 class="text-3xl font-bold text-gray-300">Grundmål</h1>
        <button
          @click="toggleSearch"
          class="text-gray-500 hover:text-gray-300 transition-colors p-1 rounded hover:bg-gray-800"
          title="Skapa nytt grundmål"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-6 w-6"
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
      <p class="text-gray-600">Mål utan föräldrar - dina högsta nivåer</p>
    </div>

    <!-- Sökfält för att skapa nya grundmål -->
    <div
      v-if="showSearch"
      class="border border-gray-700 rounded-lg p-4 bg-gray-800 mb-6"
    >
      <div class="mb-2">
        <input
          ref="searchInput"
          v-model="searchQuery"
          @keydown="handleSearchKeydown"
          type="text"
          placeholder="Skriv titel på nytt grundmål (tryck Enter)"
          class="w-full px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-gray-200 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
          autofocus
        />
      </div>

      <!-- Meddelande om att skapa nytt mål -->
      <div
        v-if="searchQuery.trim()"
        class="text-sm text-gray-500 px-4 py-2"
      >
        Tryck Enter för att skapa "{{ searchQuery }}" som nytt grundmål
      </div>
    </div>

    <div
      v-if="isLoading"
      class="flex justify-center items-center p-12"
    >
      <UIcon name="i-heroicons-arrow-path" class="w-8 h-8 text-gray-400 animate-spin" />
    </div>

    <div
      v-else-if="goals.length === 0"
      class="text-gray-500 p-6 border border-gray-700 rounded-lg text-center"
    >
      Inga grundmål ännu. Skapa ett nytt mål för att komma igång!
    </div>

    <ul v-else class="space-y-3">
      <li
        v-for="(goal, index) in goals"
        :key="goal.id"
        :data-goal-index="index"
        class="border rounded-lg transition-colors"
        :class="selectedIndex === index ? 'border-blue-500 bg-blue-900/20' : 'border-gray-700'"
      >
        <!-- Insert mode - visa input -->
        <div v-if="mode === 'insert' && editingGoalId === goal.id" class="p-4">
          <input
            v-model="editTitle"
            type="text"
            class="w-full px-3 py-2 bg-gray-800 border border-blue-500 rounded text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-400"
            @keydown.enter.prevent="saveEdit"
            @keydown.esc.prevent="cancelEdit"
          />
        </div>

        <!-- Normal mode - visa länk -->
        <NuxtLink
          v-else
          :to="`/goal/${goal.id}`"
          class="p-4 block hover:bg-gray-800/50"
        >
          <div class="flex items-start justify-between mb-2">
            <h3
              class="text-lg font-medium"
              :class="goal.finished ? 'text-gray-500 line-through' : 'text-gray-200'"
            >
              {{ goal.title }}
            </h3>
            <span
              v-if="goal.finished"
              class="px-2 py-1 text-xs font-semibold rounded bg-green-900 text-green-200"
            >
              Klar
            </span>
          </div>

           <div class="flex items-center gap-4 text-sm text-gray-500">
             <span v-if="getProgress(goal)">
               {{ getProgress(goal) }} undermål
             </span>
           </div>
        </NuxtLink>
      </li>
    </ul>
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
