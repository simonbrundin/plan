declare module '#auth-utils' {
  interface User {
    id: string
    email?: string
    name?: string
  }

  interface UserSession {
    loggedInAt: number
  }

  interface SecureSessionData {
    accessToken: string
    refreshToken?: string
    expiresAt?: number
  }
}

export {}
