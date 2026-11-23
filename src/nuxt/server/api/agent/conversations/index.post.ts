import { createError } from "h3";

// Startar en konversation hos agent-API:t.
export default defineEventHandler(async (event) => {
  const body = await readBody(event).catch(() => ({}));

  try {
    const response = await forwardToAgent(event, "conversations", {
      method: "POST",
      body,
    });
    return response;
  } catch (error: any) {
    throw createError({
      statusCode: 502,
      statusMessage: "Agent-API:t svarar inte",
      data: error?.message,
    });
  }
});
