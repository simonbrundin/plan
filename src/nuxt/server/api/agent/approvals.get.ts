import { createError } from "h3";

// Lista över förslag som väntar på ett beslut i Plan.
export default defineEventHandler(async (event) => {
  await requireSession(event);

  try {
    return await forwardToAgent(event, "approvals");
  } catch (error: any) {
    throw createError({
      statusCode: 502,
      statusMessage: "Agent-API:t svarar inte",
      data: error?.message,
    });
  }
});
