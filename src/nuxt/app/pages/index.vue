<script setup>
import { useGoalsStore } from '~/stores/goals'
import { useGoalApi } from '~/composables/useGoalApi'

const { loggedIn: isLoggedIn } = useUserSession()
const goalsStore = useGoalsStore()
const { loadAllGoals } = useGoalApi()

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

definePageMeta({
  ssr: false, // Disable SSR for this page - icons and heavy components load better on client
})
</script>

<template>
  <div class="max-w-4xl mx-auto px-6 py-8">
    <ClientOnly>
      <div v-if="isLoggedIn" class="space-y-8">
        <div class="text-center py-6">
          <h1 class="text-3xl font-bold tracking-tight text-ink mb-2">Välkommen tillbaka!</h1>
          <p class="text-ink-muted mb-6">Fånga tankar i inboxen, bryt ner dem sedan till mål.</p>
          <div class="flex justify-center gap-3">
            <NuxtLink to="/goals">
              <UButton size="lg" icon="material-symbols:arrow-forward-rounded">
                Gå till mål
              </UButton>
            </NuxtLink>
            <NuxtLink to="/now">
              <UButton size="lg" color="neutral" variant="outline" icon="material-symbols:sort">
                Gör nu
              </UButton>
            </NuxtLink>
            <NuxtLink to="/agent">
              <UButton size="lg" color="neutral" variant="outline" icon="material-symbols:smart-toy-outline">
                Agentmål
              </UButton>
            </NuxtLink>
          </div>
        </div>

        <div class="rounded-2xl border border-line bg-surface transition-colors hover:border-line-strong hover:bg-surface-2 p-6">
          <InboxInput @add="addGoalToInbox" />
          <InboxGoalsList />
        </div>
      </div>
      <template #fallback>
        <div class="flex items-center justify-center min-h-[50vh]">
          <div class="text-center">
            <h1 class="text-3xl font-bold tracking-tight text-ink mb-6">Laddar...</h1>
          </div>
        </div>
      </template>
    </ClientOnly>

    <div v-if="!isLoggedIn" class="flex items-center justify-center min-h-[50vh]">
      <div class="text-center">
        <h1 class="text-3xl font-bold tracking-tight text-ink mb-6">Välkommen till Plan</h1>
        <NuxtLink to="/login">
          <UButton size="lg" icon="material-symbols:login-rounded">
            Logga in
          </UButton>
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
