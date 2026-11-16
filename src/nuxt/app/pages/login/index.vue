<script setup lang="ts">
// Read user info from cookie set by auth callback
const config = useRuntimeConfig()
const userCookie = useCookie('plan_user')

const devAuthEnabled = computed(() => !!config.public.devAuthEnabled)

const user = computed(() => {
  if (!userCookie.value) return null
  try {
    return JSON.parse(atob(userCookie.value))
  } catch {
    return null
  }
})

const handleLogout = async () => {
  // Logouten sköts av Go API:t (sessionen raderas server side, cookierna
  // rensas) - här rensas bara den läsbara hint-cookien.
  await $fetch('/api/auth/logout', { method: 'POST' }).catch(() => {})
  userCookie.value = null
  window.location.reload()
}
</script>

<template>
  <div class="max-w-md mx-auto px-6 py-16">
    <div class="rounded-2xl border border-line bg-surface p-8">
      <div v-if="user">
        <h1 class="text-3xl font-bold tracking-tight text-ink mb-1">Inloggad</h1>
        <p class="text-sm text-ink-dim mb-6">Din session är aktiv.</p>

        <dl class="space-y-2 mb-6 text-sm">
          <div v-if="user.sub" class="flex items-center justify-between gap-4">
            <dt class="text-ink-dim">User ID</dt>
            <dd class="font-mono text-ink">{{ user.sub }}</dd>
          </div>
          <div v-if="user.email" class="flex items-center justify-between gap-4">
            <dt class="text-ink-dim">Email</dt>
            <dd class="font-mono text-ink truncate">{{ user.email }}</dd>
          </div>
        </dl>

        <UButton @click="handleLogout" icon="solar:logout-2-line-duotone">Logga ut</UButton>
      </div>

      <div v-else>
        <h1 class="text-3xl font-bold tracking-tight text-ink mb-1">Logga in</h1>
        <p class="text-sm text-ink-dim mb-6">Logga in för att se dina mål.</p>

        <div class="flex flex-col gap-3">
          <a href="/auth/login">
            <UButton block>Logga in med Zitadel</UButton>
          </a>
          <a v-if="devAuthEnabled" href="/auth/dev-login" class="block">
            <UButton block color="neutral" variant="outline">Dev-login (utan Zitadel)</UButton>
          </a>
        </div>
      </div>
    </div>

    <div class="mt-6 text-center">
      <NuxtLink to="/">
        <UButton color="neutral" variant="ghost" icon="material-symbols-light:home-outline-rounded"
          >Till startsidan</UButton
        >
      </NuxtLink>
    </div>
  </div>
</template>
<style scoped></style>
