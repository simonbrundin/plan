import { createError, getRouterParam } from "h3";

// Hämtar meddelandena i en konversation.
export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, "id");
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: "Missing conversation id" });
  }

  await requireSession(event);

  try {
    return await forwardToAgent(event, `conversations/${id}/messages`);
  } catch (error: any) {
    throw createError({
      statusCode: 502,
      statusMessage: "Agent-API:t svarar inte",
      data: error?.message,
    });
  }
});
