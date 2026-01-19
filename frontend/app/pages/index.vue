<script setup>
import { useGoalsStore } from '~/stores/goals'
import { useGoalApi } from '~/composables/useGoalApi'

const { loggedIn } = useUserSession()
const goalsStore = useGoalsStore()
const { loadAllGoals } = useGoalApi()

// Load goals on mount
// Temporarily disabled to debug hanging issue
// onMounted(async () => {
//   if (loggedIn) {
//     try {
//       const goals = await loadAllGoals()
//       goalsStore.goals = goals
//       goalsStore.isLoaded = true
//     } catch (error) {
//       console.error('Failed to load goals:', error)
//     }
//   }
// })

const { createGoal: createGoalApi } = useGoalApi()

const addGoalToInbox = async (goal) => {
  try {
    // For now, create goal without user association since we're in inbox
    // TODO: Add user association when user system is integrated
    const newGoal = await createGoalApi(goal.title, 1) // Using user ID 1 as placeholder
    goalsStore.addGoal({ ...newGoal, inbox: goal.inbox })
  } catch (error) {
    console.error('Failed to add goal to inbox:', error)
  }
}
</script>

<template>
  <div class="max-w-4xl mx-auto p-6">
    <div v-if="loggedIn" class="space-y-8">
      <div class="text-center">
        <h1 class="text-3xl font-bold text-gray-900 mb-6">Välkommen tillbaka!</h1>
        <NuxtLink to="/goal/1">
          <UButton size="lg" icon="material-symbols:arrow-forward-rounded" class="mr-4">
            Gå till mål
          </UButton>
        </NuxtLink>
      </div>

      <div class="bg-gray-50 rounded-lg p-6">
        <InboxInput @add="addGoalToInbox" />
        <InboxGoalsList />
      </div>
    </div>

    <div v-else class="flex items-center justify-center min-h-[50vh]">
      <div class="text-center">
        <h1 class="text-3xl font-bold text-gray-900 mb-6">Välkommen till Plan</h1>
        <NuxtLink to="/login">
          <UButton size="lg" icon="material-symbols:login-rounded">
            Logga in
          </UButton>
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
