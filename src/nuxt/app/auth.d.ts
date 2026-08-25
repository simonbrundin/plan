declare module '#auth-utils' {
  interface User {
    id: string
    sub: string
    email: string
    accessToken?: string
  }

  interface UserSession {
    loggedInAt: number
  }
}

export {}
