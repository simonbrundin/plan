import type { Tag } from "~/types/tag";
import { apiRequest } from "~/utils/api";

/**
 * useTags ger åtkomst till användarens taggar och till målens tagglistor.
 *
 * Taggar skapas på demand - skriver man in ett nytt namn i taggväljaren
 * skapas taggen i API:t.
 */
export function useTags() {
	/** Alla användarens taggar, med antalet mål som har dem. */
	const tags = useState<Tag[]>("tags", () => []);

	/** Laddar användarens taggar. */
	const fetchTags = async (): Promise<void> => {
		const result = await apiRequest<Tag[]>("tags");
		tags.value = result ?? [];
	};

	/** Hämtar taggarna kopplade till ett mål. */
	const fetchGoalTags = async (goalId: number): Promise<string[]> => {
		const result = await apiRequest<string[]>(`goals/${goalId}/tags`);
		return result ?? [];
	};

	/**
	 * Ersätter taggarna på ett mål.
	 *
	 * Tom lista tar bort alla taggar. Namn som inte finns ännu skapas
	 * av API:t. Returnerar den nya tagglistan.
	 */
	const setGoalTags = async (
		goalId: number,
		names: string[],
	): Promise<string[]> => {
		const result = await apiRequest<string[]>(`goals/${goalId}/tags`, {
			method: "PUT",
			body: { tags: names },
		});
		return result ?? [];
	};

	return {
		tags,
		fetchTags,
		fetchGoalTags,
		setGoalTags,
	};
}
