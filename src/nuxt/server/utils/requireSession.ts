import { createError } from "h3";

/**
 * Kräver en giltig session - agent-API:t får bara användas av inloggade
 * användare. Sessionen valideras av Plan:s Go API (den äger sessionen).
 */
export async function requireSession(event: any): Promise<void> {
  try {
    const session = await forwardToGo(event, "auth/session", { method: "GET" });
    const parsed = typeof session === "string" ? JSON.parse(session) : session;
    // Plan:s session-svar innehåller user_id (och ibland sub/email).
    if (!parsed || (!parsed.user_id && !parsed.user && !parsed.sub)) {
      throw new Error("no session");
    }
  } catch (error: any) {
    throw createError({
      statusCode: 401,
      statusMessage: "Inte inloggad",
    });
  }
}
