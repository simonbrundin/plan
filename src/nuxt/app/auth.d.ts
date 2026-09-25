declare module '#auth-utils' {
  interface User {
    id: string
    sub: string
    email?: string
    name?: string
    accessToken?: string
    refreshToken?: string
    expiresAt?: number
  }

  interface UserSession {
    loggedInAt: number
  }
}

export {}
