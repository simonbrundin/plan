/**
 * Fokus ska inte fastna på knappar efter ett musklick.
 *
 * Appen ritar fokusringen med :focus-visible, så en knapp som klickats med
 * musen ser ut att sakna fokus - men är fortfarande aktiv. Då aktiverar
 * mellanslag knappen i stället för att öppna leader-läget, vilket gör att
 * leader-kommandona tydligen "inte fungerar" på de sidor där man klickar
 * på knappar (t.ex. filtren på /now).
 *
 * Därför tas fokus bort efter ett musklick. Klick från tangentbordet har
 * detail = 0 och lämnas orörda, så att knappar fortfarande kan aktiveras
 * med mellanslag när man navigerat dit med Tab.
 */
export default defineNuxtPlugin(() => {
	function handleClick(event: MouseEvent) {
		// detail > 0 = riktigt klick med mus. Tangentbordsklick har detail 0.
		if (event.detail === 0) return;

		const target = event.target as HTMLElement | null;
		const button = target?.closest("button");
		if (!button) return;

		// Blur sker först efter att knappens egen klick-hanterare kört.
		// Många knappar stoppar händelsens propagering (@click.stop), så
		// listenern måste köras i capture-fasen för att över huvud taget
		// nå knappen.
		setTimeout(() => {
			// Knappen kan ha slutat vara aktiv under tiden - då gör
			// blur() ingenting, vilket är precis vad vi vill.
			button.blur();
		}, 0);
	}

	document.addEventListener("click", handleClick, true);
});
