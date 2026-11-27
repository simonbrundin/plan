import { createError } from "h3";

// Listar konversationer - med goal_id filtrets listan fram de som
// gäller ett visst mål.
export default defineEventHandler(async (event) => {
  const query = getQuery(event);
  const goalId = query.goal_id;

  try {
    const response = await forwardToAgent(event, "conversations", {
      method: "GET",
      query: {
        goal_id: goalId === undefined ? undefined : String(goalId),
        limit: query.limit === undefined ? undefined : String(query.limit),
        offset: query.offset === undefined ? undefined : String(query.offset),
      },
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
