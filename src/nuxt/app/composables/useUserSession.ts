interface UserSession {
  sub?: string
  email?: string
}

interface UserSessionData {
  user: UserSession | null
  loggedIn: boolean
  loggedInAt?: number
}

/**
 * The session lives in the Go API. The `plan_user` cookie is a readable
 * hint (set by the Go API) so the UI can render synchronously - the server
 * is always the authority and validates the session itself.
 */
export function useUserSession() {
  const userCookie = useCookie('plan_user')

  const user = computed<UserSession | null>(() => {
    if (!userCookie.value) return null
    try {
      return JSON.parse(atob(userCookie.value))
    } catch {
      return null
    }
  })

  const loggedIn = computed(() => !!user.value)

  const session = computed<UserSessionData>(() => ({
    user: user.value,
    loggedIn: loggedIn.value,
  }))

  /** Confirms the session with the Go API and refreshes the hint cookie. */
  const fetch = async () => {
    try {
      const data = await $fetch<{ sub?: string; email?: string }>('/api/auth/session')
      const payload = JSON.stringify({ sub: data.sub, email: data.email })
      userCookie.value = btoa(payload)
      return data
    } catch {
      userCookie.value = null
      return null
    }
  }

  /** Ends the session server side and clears the local hint cookie. */
  const clear = async () => {
    await $fetch('/api/auth/logout', { method: 'POST' }).catch(() => {})
    userCookie.value = null
  }

  return {
    user,
    loggedIn,
    session,
    fetch,
    clear,
  }
}
