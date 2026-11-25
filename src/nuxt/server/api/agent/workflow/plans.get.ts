import { forwardToAgent } from "../../../utils/agentApi";

export default defineEventHandler(async (event) => {
  await requireSession(event);
  return await forwardToAgent(event, "workflow/plans", { method: "GET" });
});
