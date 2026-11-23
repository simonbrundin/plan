// En tagg (etikett) på ett mål - t.ex. "ute" eller "hem".
export interface Tag {
	name: string;
	// Antalet mål som har taggen. Används för att se vilka taggar
	// som faktiskt används.
	goal_count: number;
}
