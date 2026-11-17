/**
 * En användares kalender. Användaren kan skapa hur många kalendrar som
 * helst, och varje kalender har en egen färg som händelserna ärver.
 */
export interface Calendar {
	id: number;
	user_id: number;
	name: string;
	color: string;
	icon?: string;
	sort_order: number;
}
