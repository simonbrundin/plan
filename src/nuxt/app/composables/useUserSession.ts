interface UserSession {
  sub?: string
  email?: string
}

interface UserSessionData {
  user: UserSession | null
  loggedIn: boolean
  loggedInAt?: number
}

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
  
  const fetch = async () => {
    // Session is read from cookie, no fetch needed
  }
  
  const clear = async () => {
    await $fetch('/api/auth/logout', { method: 'POST' })
  }
  
  return {
    user,
    loggedIn,
    session,
    fetch,
    clear,
  }
}
