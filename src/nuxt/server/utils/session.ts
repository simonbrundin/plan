import { useSession } from "h3";

// Define the session schema for nuxt-auth-utils
export const sessionConfig = {
  password: process.env.NUXT_SESSION_PASSWORD || "",

  // Cookie settings
  cookie: {
    name: "nuxt-session",
    maxAge: 60 * 60 * 24 * 7, // 7 days
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax" as const,
  },
};

// Type definitions for the session
export interface User {
  id: string;
  sub: string;
  email?: string;
  name?: string;
  accessToken?: string;
  refreshToken?: string;
  idToken?: string;
}

export interface SessionData {
  user: User;
  loggedInAt?: number;
}

export interface Session {
  id: string;
  user: User;
  loggedInAt?: number;
}
