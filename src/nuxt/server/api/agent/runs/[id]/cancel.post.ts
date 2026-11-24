import { createError, getRouterParam } from "h3";

// Avbryter en agentkörning.
//
// Beslutet tas i Plan av en människa - sessionen skickas med, så att
// agentens egen nyckel aldrig kan avbryta någons körning. När Plan
// markerar körningen som avbruten städar jobbcontrollern jobbet.
export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, "id");
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: "Missing run id" });
  }

  await requireSession(event);

  try {
    return await forwardToGo(event, `agent-runs/${id}/cancel`, {
      method: "POST",
    });
  } catch (error: any) {
    throw createError({
      statusCode: 502,
      statusMessage: "Plan API:t svarar inte",
      data: error?.message,
    });
  }
});
