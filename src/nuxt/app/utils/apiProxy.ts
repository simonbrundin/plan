import type { FetchOptions } from 'ofetch';

/**
 * Anrop mot Go API:t går via Nuxt:s proxy, som vidarebefordrar
 * sessioncookien. Valideringen sker i Go API:t.
 *
 * $fetch castas till en enkel funktion. Den typade varianten jämför varje
 * anrop mot varje route i appen, och med dynamiska sökvägar blir
 * jämförelsen så djup att TypeScript ger upp ("excessive stack depth").
 */
const proxyFetch = $fetch as unknown as <T>(
  url: string,
  options?: FetchOptions
) => Promise<T>;

/**
 * Ett anrop mot Go API:t via serverns proxy.
 *
 * Körs i webbläsaren - sessioncookien skickas med automatiskt.
 */
export async function apiProxy<T>(
  path: string,
  options: FetchOptions = {}
): Promise<T> {
  return await proxyFetch<T>(`/api/proxy/${path}`, options);
}

/**
 * Sessionen har gått ut - användaren måste logga in igen.
 */
export class AuthenticationError extends Error {
  constructor(message = 'Sessionen har gått ut. Vänligen logga in igen.') {
    super(message);
    this.name = 'AuthenticationError';
  }
}

/**
 * Samma anrop som apiProxy, men med ett tydligt fel när sessionen
 * har gått ut.
 */
export async function apiProxyWithAuth<T>(
  path: string,
  options: FetchOptions = {}
): Promise<T> {
  try {
    return await apiProxy<T>(path, options);
  } catch (error: any) {
    if (error?.statusCode === 401) {
      throw new AuthenticationError();
    }
    throw error;
  }
}
