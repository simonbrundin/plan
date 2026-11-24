import {
  getRequestHeader,
  getRouterParam,
  readBody,
  setResponseHeader,
  setResponseStatus,
} from "h3";

/**
 * Agent-API:t är en separat tjänst från samma repository (src/agent) som
 * håller agentens konversationer och körningar. Plan proxar till den - Slack
 * används inte, all kommunikation med agenten sker i Plan.
 */

// agentApiBase returns the server-side URL to the agent API.
export function agentApiBase(): string {
  const config = useRuntimeConfig();
  return config.agentApiInternalUrl || "http://localhost:8081/api/v1";
}

/**
 * agentApiKey returns the server-side key used to talk to the agent API.
 * Empty means no key is sent - the agent API may run without auth in dev.
 */
export function agentApiKey(): string {
  const config = useRuntimeConfig();
  return config.agentApiKey || "";
}

interface AgentRequestOptions {
  method?: string;
  body?: unknown;
  query?: Record<string, string | number | undefined>;
}

/**
 * forwardToAgent forwards a request to the agent API and returns the
 * response body as text. Cookies are not forwarded - the agent API is
 * called server to server.
 */
export async function forwardToAgent(
  event: any,
  path: string,
  options: AgentRequestOptions = {},
): Promise<string> {
  const method = (options.method || event.method || "GET").toUpperCase();

  const headers: Record<string, string> = {};
  const key = agentApiKey();
  if (key) {
    headers.authorization = `Bearer ${key}`;
  }

  let body: string | undefined;
  if (!["GET", "HEAD"].includes(method)) {
    const payload = options.body ?? (await readBody(event).catch(() => undefined));
    if (payload !== undefined && payload !== null) {
      body = JSON.stringify(payload);
      headers["content-type"] = "application/json";
    }
  }

  const query = options.query ?? {};
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === null) continue;
    search.append(key, String(value));
  }
  const queryString = search.toString();

  const response = await fetch(
    `${agentApiBase()}/${path}${queryString ? `?${queryString}` : ""}`,
    { method, headers, body },
  );

  const contentType = response.headers.get("content-type");
  if (contentType) {
    setResponseHeader(event, "content-type", contentType);
  }
  setResponseStatus(event, response.status);

  return await response.text();
}

/**
 * pipeFromAgent pipes the response body from the agent API straight through
 * to the browser. Used for the chat: the agent's answer arrives while it
 * is being written (NDJSON) and new messages arrive over SSE.
 */
export async function pipeFromAgent(
  event: any,
  path: string,
  options: { method?: string; body?: unknown; query?: Record<string, string | number | undefined> } = {},
): Promise<ReadableStream | null> {
  const method = (options.method || "GET").toUpperCase();
  const headers: Record<string, string> = {};
  const key = agentApiKey();
  if (key) {
    headers.authorization = `Bearer ${key}`;
  }

  let body: string | undefined;
  if (!["GET", "HEAD"].includes(method)) {
    if (options.body !== undefined && options.body !== null) {
      body = JSON.stringify(options.body);
      headers["content-type"] = "application/json";
    }
  }

  const search = new URLSearchParams();
  for (const [k, v] of Object.entries(options.query ?? {})) {
    if (v === undefined || v === null) continue;
    search.append(k, String(v));
  }
  const queryString = search.toString();

  const response = await fetch(
    `${agentApiBase()}/${path}${queryString ? `?${queryString}` : ""}`,
    { method, headers, body },
  );

  const contentType = response.headers.get("content-type");
  if (contentType) {
    setResponseHeader(event, "content-type", contentType);
  }
  setResponseHeader(event, "cache-control", "no-cache");
  setResponseHeader(event, "x-accel-buffering", "no");
  setResponseStatus(event, response.status);

  // Strömmen returneras som den är - webbläsaren får agentens svar medan
  // de skrivs.
  return response.body;
}

/**
 * agentPath builds a path under the agent API from the event parameters.
 */
export function agentPath(event: any, prefix: string): string {
  const id = getRouterParam(event, "id");
  const goalId = getRouterParam(event, "goalId");
  const parts = [prefix, id, goalId].filter(Boolean);
  return parts.join("/");
}

/**
 * forwards the session cookie if the agent API ever needs it. Kept as a
 * separate helper so the intent is explicit.
 */
export function agentCookie(event: any): string | undefined {
  return getRequestHeader(event, "cookie");
}
