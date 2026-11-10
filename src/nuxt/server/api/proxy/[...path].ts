import { createError, getRouterParam } from "h3";

// Proxy API calls to the Go API. The session cookie is forwarded as-is and
// the Go API validates it - the auth logic lives there, not here.
export default defineEventHandler(async (event) => {
  const path = getRouterParam(event, "path");

  if (!path) {
    throw createError({
      statusCode: 400,
      statusMessage: "Missing path parameter",
    });
  }

  try {
    return await forwardToGo(event, path);
  } catch (error: any) {
    throw createError({
      statusCode: 500,
      statusMessage: "Proxy error",
      data: error?.message,
    });
  }
});
