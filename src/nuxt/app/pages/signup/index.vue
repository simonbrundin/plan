<script setup lang="ts">
const form = reactive({
  email: '',
  password: '',
  passwordConfirm: '',
  username: '',
  name: '',
})

const loading = ref(false)
const error = ref('')

const validateForm = () => {
  if (!form.email || !form.password || !form.passwordConfirm || !form.username) {
    error.value = 'Alla fält måste fyllas i'
    return false
  }

  if (form.password !== form.passwordConfirm) {
    error.value = 'Lösenorden matchar inte'
    return false
  }

  if (form.password.length < 8) {
    error.value = 'Lösenordet måste vara minst 8 tecken'
    return false
  }

  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  if (!emailRegex.test(form.email)) {
    error.value = 'Ogiltig e-postadress'
    return false
  }

  return true
}

const handleSignup = async () => {
  error.value = ''

  if (!validateForm()) {
    return
  }

  loading.value = true

  try {
    const response = await $fetch('/api/auth/signup', {
      method: 'POST',
      body: {
        email: form.email,
        password: form.password,
        username: form.username,
        name: form.name,
      },
    })

    // Efter framgångsrik registrering, logga in användaren via Zitadel OAuth
    navigateTo('/auth/zitadel')
  } catch (err: any) {
    console.error('Signup error:', err)
    error.value = err.data?.message || 'Något gick fel vid registrering'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
      <div>
        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900 dark:text-white">
          Skapa ett konto
        </h2>
        <p class="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
          Eller
          <NuxtLink to="/login" class="font-medium text-primary-600 hover:text-primary-500">
            logga in på ditt befintliga konto
          </NuxtLink>
        </p>
      </div>

      <UCard>
        <form @submit.prevent="handleSignup" class="space-y-6">
          <UAlert
            v-if="error"
            color="error"
            variant="solid"
            :title="error"
            :close-button="{ icon: 'i-heroicons-x-mark-20-solid', color: 'white', variant: 'link' }"
            @close="error = ''"
          />

          <UFormGroup label="E-post" required>
            <UInput
              v-model="form.email"
              type="email"
              placeholder="din@email.com"
              icon="i-heroicons-envelope"
              :disabled="loading"
              size="lg"
            />
          </UFormGroup>

          <UFormGroup label="Användarnamn" required>
            <UInput
              v-model="form.username"
              type="text"
              placeholder="användarnamn"
              icon="i-heroicons-user"
              :disabled="loading"
              size="lg"
            />
          </UFormGroup>

          <UFormGroup label="Namn" hint="Valfritt">
            <UInput
              v-model="form.name"
              type="text"
              placeholder="Ditt namn"
              icon="i-heroicons-identification"
              :disabled="loading"
              size="lg"
            />
          </UFormGroup>

          <UFormGroup label="Lösenord" required>
            <UInput
              v-model="form.password"
              type="password"
              placeholder="Minst 8 tecken"
              icon="i-heroicons-lock-closed"
              :disabled="loading"
              size="lg"
            />
          </UFormGroup>

          <UFormGroup label="Bekräfta lösenord" required>
            <UInput
              v-model="form.passwordConfirm"
              type="password"
              placeholder="Upprepa lösenord"
              icon="i-heroicons-lock-closed"
              :disabled="loading"
              size="lg"
            />
          </UFormGroup>

          <div class="flex gap-3">
            <UButton
              type="submit"
              :loading="loading"
              :disabled="loading"
              size="lg"
              block
              icon="i-heroicons-user-plus"
            >
              Skapa konto
            </UButton>
          </div>
        </form>
      </UCard>

      <div class="text-center">
        <NuxtLink to="/" class="text-sm text-gray-600 dark:text-gray-400 hover:text-primary-600">
          ← Tillbaka till startsidan
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
