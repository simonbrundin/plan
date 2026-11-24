import { createError } from "h3";

// Lista över förslag som väntar på ett beslut.
//
// Godkännandena ligger i Plan, inte i agentens minne - därför går listan
// till Plan API:t med användarens session.
export default defineEventHandler(async (event) => {
  await requireSession(event);

  try {
    const response = await forwardToGo(event, "agent-approvals");

    // Plan svarar med JSON som text - den tolkas här så att klienten får
    // en lista och inte en sträng.
    let approvals: unknown[] = [];
    if (typeof response === "string") {
      try {
        const parsed = JSON.parse(response);
        approvals = Array.isArray(parsed) ? parsed : [];
      } catch {
        approvals = [];
      }
    } else if (Array.isArray(response)) {
      approvals = response;
    }

    return { approvals };
  } catch (error: any) {
    throw createError({
      statusCode: 502,
      statusMessage: "Plan API:t svarar inte",
      data: error?.message,
    });
  }
});
