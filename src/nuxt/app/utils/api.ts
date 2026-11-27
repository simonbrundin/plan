import type { FetchOptions } from 'ofetch';

/**
 * Klienten talar direkt med API:t - det finns ingen proxy i Nuxt.
 *
 * $fetch castas till en enkel funktion. Den typade varianten jämför varje
 * anrop mot varje route i appen, och med dynamiska sökvägar blir
 * jämförelsen så djup att TypeScript ger upp ("excessive stack depth").
 *
 * Anropet sker först när funktionen körs - inte när modulen laddas. Då kan
 * modulen importeras utan en Nuxt-runtime (t.ex. i enhetstester), där
 * $fetch inte finns.
 */
function request<T>(url: string, options?: FetchOptions): Promise<T> {
  const fetch = $fetch as unknown as <T>(
    url: string,
    options?: FetchOptions
  ) => Promise<T>;
  return fetch(url, options);
}

/**
 * Bas-URL:en till API:t.
 *
 * I prod ligger appen och API:t på samma host (plan.simonbrundin.com), så
 * anropen är same-origin. I dev ligger appen på :3100 och API:t på :4100 -
 * då är det cross-origin och API:t måste svara med CORS-headrar.
 */
export function apiBaseUrl(): string {
  const config = useRuntimeConfig();
  return (config.public.goApiUrl || '/api/v1').replace(/\/+$/, '');
}

/**
 * Ett anrop mot API:t.
 *
 * Körs i webbläsaren - sessioncookien (plan_session) skickas med och API:t
 * validerar den.
 */
export async function apiRequest<T>(
  path: string,
  options: FetchOptions = {}
): Promise<T> {
  return await request<T>(`${apiBaseUrl()}/${path}`, {
    credentials: 'include',
    ...options,
  });
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
 * Samma anrop som apiRequest, men med ett tydligt fel när sessionen
 * har gått ut.
 */
export async function apiRequestWithAuth<T>(
  path: string,
  options: FetchOptions = {}
): Promise<T> {
  try {
    return await apiRequest<T>(path, options);
  } catch (error: any) {
    if (error?.statusCode === 401 || error?.response?.status === 401) {
      throw new AuthenticationError();
    }
    throw error;
  }
}
