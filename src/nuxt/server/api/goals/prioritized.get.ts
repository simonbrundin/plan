import { sql } from "../../utils/db";
import { getCurrentUser } from "../../utils/auth";

export default defineEventHandler(async (event) => {
	const user = await getCurrentUser(event);

	if (!user) {
		throw createError({
			statusCode: 401,
			message: "Unauthorized",
		});
	}

	// Hämta alla mål användaren har tillgång till, med vikt från
	// goal_relations. Om ett mål har flera parent-relationer tar vi MAX(weight)
	// (den mest prioriterade relationen vinner). parent_title är titeln på den
	// parent som har den vikten, så UI:t kan visa "under Root" osv.
	// Vi filtrerar bort:
	//   - avklarade mål
	//   - inbox-mål (ännu inte prioriterade/inkapslade)
	//   - root goal (ID 1) som delas av alla användare
	//
	// Om kolumnen "started" saknas (före migrering) så fallback:ar vi
	// till en query utan den kolumnen.
	let goals: any[];
	try {
		goals = await sql<any[]>`
      SELECT
        g.id,
        g.title,
        g.icon,
        g.created,
        g.started,
        g.finished,
        g.inbox,
        COALESCE(
          (SELECT MAX(weight) FROM goal_relations WHERE child_id = g.id),
          10
        ) AS weight,
        (
          SELECT g2.title
          FROM goal_relations gr
          INNER JOIN goals g2 ON g2.id = gr.parent_id
          WHERE gr.child_id = g.id
          ORDER BY gr.weight DESC
          LIMIT 1
        ) AS "parentTitle",
        (
          SELECT gr.parent_id
          FROM goal_relations gr
          WHERE gr.child_id = g.id
          ORDER BY gr.weight DESC
          LIMIT 1
        ) AS "parentId"
      FROM goals g
      INNER JOIN user_goals ug ON g.id = ug.goal_id
      WHERE ug.user_id = ${user.id}
        AND g.finished IS NULL
        AND g.inbox = 0
        AND g.id <> 1
      ORDER BY weight DESC, g.created DESC
	  `;
	} catch (_) {
		// Kolumnen started finns inte ännu — fallback utan den
		goals = await sql<any[]>`
		  SELECT
        g.id,
        g.title,
        g.icon,
        g.created,
        NULL AS started,
        g.finished,
        g.inbox,
        COALESCE(
          (SELECT MAX(weight) FROM goal_relations WHERE child_id = g.id),
          10
        ) AS weight,
        (
          SELECT g2.title
          FROM goal_relations gr
          INNER JOIN goals g2 ON g2.id = gr.parent_id
          WHERE gr.child_id = g.id
          ORDER BY gr.weight DESC
          LIMIT 1
        ) AS "parentTitle",
        (
          SELECT gr.parent_id
          FROM goal_relations gr
          WHERE gr.child_id = g.id
          ORDER BY gr.weight DESC
          LIMIT 1
        ) AS "parentId"
      FROM goals g
      INNER JOIN user_goals ug ON g.id = ug.goal_id
      WHERE ug.user_id = ${user.id}
        AND g.finished IS NULL
        AND g.inbox = 0
        AND g.id <> 1
      ORDER BY weight DESC, g.created DESC
		`;
	}

	return goals;
});
