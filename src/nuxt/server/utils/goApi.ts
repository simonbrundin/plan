import {
  appendResponseHeader,
  getQuery,
  getRequestHeader,
  readBody,
  setResponseHeader,
  setResponseStatus,
} from "h3";

/**
 * Authen ligger i Go API:t: det skapar och validerar sessionen och sätter
 * cookierna. Nuxt är bara en tunn framför gång - den vidarebefordrar
 * requesten och kopierar tillbaka svar, cookies och redirects.
 */

// goApiBase returns the server-side URL to the Go API.
export function goApiBase(): string {
  const config = useRuntimeConfig();
  return (
    config.goApiInternalUrl ||
    config.public.goApiUrl ||
    "http://localhost:8080/api/v1"
  );
}

interface ForwardOptions {
  method?: string;
  body?: unknown;
}

/**
 * forwardToGo forwards the current request to the Go API and copies the
 * response back to the client - including Set-Cookie (session cookies) and
 * Location (OAuth redirects).
 */
export async function forwardToGo(
  event: any,
  path: string,
  options: ForwardOptions = {},
) {
  const method = (options.method || event.method || "GET").toUpperCase();

  const headers: Record<string, string> = {};
  const cookie = getRequestHeader(event, "cookie");
  if (cookie) {
    headers.cookie = cookie;
  }

  let body: string | undefined;
  if (!["GET", "HEAD"].includes(method)) {
    const payload = options.body ?? (await readBody(event).catch(() => undefined));
    if (payload !== undefined && payload !== null) {
      body = JSON.stringify(payload);
      headers["content-type"] = "application/json";
    }
  }

  const query = getQuery(event);
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === null) continue;
    if (Array.isArray(value)) {
      for (const entry of value) search.append(key, String(entry));
    } else {
      search.append(key, String(value));
    }
  }
  const queryString = search.toString();

  const response = await fetch(
    `${goApiBase()}/${path}${queryString ? `?${queryString}` : ""}`,
    {
      method,
      headers,
      body,
      // Vi hanterar redirects själva - annars skulle servern följa OAuth-
      // redirecten till Zitadel i stället för att skicka den till webbläsaren.
      redirect: "manual",
    },
  );

  // Sessioncookies sätts av Go API:t - de måste nå webbläsaren.
  const setCookies = response.headers.getSetCookie?.() ?? [];
  for (const cookie of setCookies) {
    appendResponseHeader(event, "set-cookie", cookie);
  }

  const location = response.headers.get("location");
  if (location) {
    setResponseHeader(event, "location", location);
    setResponseStatus(event, response.status);
    return "";
  }

  const contentType = response.headers.get("content-type");
  if (contentType) {
    setResponseHeader(event, "content-type", contentType);
  }
  setResponseStatus(event, response.status);

  return await response.text();
}
