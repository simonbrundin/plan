import { createError } from "h3";

// Hämtar kostnadsdata för agentkörningar.
//
// Plan äger kostnaderna. Nuxt är bara en framför gång - sessionen skickas
// med, så att det är människans session som läser kostnaderna.
export default defineEventHandler(async (event) => {
  await requireSession(event);

  try {
    return await forwardToGo(event, "agent-usage", { method: "GET" });
  } catch (error: any) {
    throw createError({
      statusCode: 502,
      statusMessage: "Plan API:t svarar inte",
      data: error?.message,
    });
  }
});
