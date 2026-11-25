import { createError, getRouterParam } from "h3";
import { forwardToAgent } from "../../../../utils/agentApi";

export default defineEventHandler(async (event) => {
  await requireSession(event);
  const id = getRouterParam(event, "id");
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: "Missing workflow id" });
  }
  return await forwardToAgent(event, `workflow/${id}/reject`, { method: "POST" });
});
