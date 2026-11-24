import { createError, getRouterParam, readBody } from "h3";

// Godkänner eller avvisar ett förslag.
//
// Beslutet tas i Plan av en människa - sessionen skickas med, så att
// agentens egen nyckel aldrig kan besluta åt någon.
export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, "id");
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: "Missing approval id" });
  }

  const body = await readBody(event).catch(() => ({}));

  await requireSession(event);

  try {
    return await forwardToGo(event, `agent-approvals/${id}/decision`, {
      method: "POST",
      body,
    });
  } catch (error: any) {
    throw createError({
      statusCode: 502,
      statusMessage: "Plan API:t svarar inte",
      data: error?.message,
    });
  }
});
