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
  <div>
    <div v-if="user">
      <h1 class="text-3xl font-bold text-gray-300 mb-2">Inloggad</h1>

      <p v-if="user.sub" class="mb-2">
        User ID: <span class="font-mono text-sm">{{ user.sub }}</span>
      </p>

      <p v-if="user.email" class="mb-2">
        Email: <span class="font-mono text-sm">{{ user.email }}</span>
      </p>

      <UButton @click="handleLogout" icon="solar:logout-2-line-duotone">Logout</UButton>
    </div>
    <div v-else>
      <h1>Not logged in</h1>
      <a href="/auth/login">
        <UButton>Logga in med Zitadel</UButton>
      </a>
      <a v-if="devAuthEnabled" href="/auth/dev-login" class="block mt-2">
        <UButton color="neutral" variant="outline">Dev-login (utan Zitadel)</UButton>
      </a>
    </div>
    <NuxtLink to="/">
      <UButton color="neutral" icon="material-symbols-light:home-outline-rounded"
        >Hem</UButton
      >
    </NuxtLink>
  </div>
</template>
<style scoped></style>
