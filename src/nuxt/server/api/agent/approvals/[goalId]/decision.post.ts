import { createError, getRouterParam, readBody } from "h3";

// Godkänner eller avvisar ett förslag. Beslutet tas i Plan - Slack
// används inte.
export default defineEventHandler(async (event) => {
  const goalId = getRouterParam(event, "goalId");
  if (!goalId) {
    throw createError({ statusCode: 400, statusMessage: "Missing goal id" });
  }

  const body = await readBody(event).catch(() => ({}));

  await requireSession(event);

  try {
    return await forwardToAgent(event, `approvals/${goalId}/decision`, {
      method: "POST",
      body,
    });
  } catch (error: any) {
    throw createError({
      statusCode: 502,
      statusMessage: "Agent-API:t svarar inte",
      data: error?.message,
    });
  }
});
