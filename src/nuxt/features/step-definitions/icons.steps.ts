import { Given, Then, When } from "@cucumber/cucumber";
import assert from "node:assert";

import {
	extractEmoji,
	searchEmojis,
	type EmojiEntry,
} from "../../app/utils/emojis.ts";
import { isIconifyIcon, normalizeIcon } from "../../app/utils/icons.ts";

/**
 * Steg för ikonlogiken. Tillståndet hålls i en WeakMap per scenario, precis
 * som i natural-language.steps.ts, eftersom Cucumber bara har en World per
 * körning.
 */
interface IconState {
	text: string;
	emoji: string | null;
	remaining: string;
	results: EmojiEntry[];
	icon: string;
}

const states = new WeakMap<object, IconState>();

const stateOf = (world: object): IconState => {
	let state = states.get(world);
	if (!state) {
		state = {
			text: "",
			emoji: null,
			remaining: "",
			results: [],
			icon: "",
		};
		states.set(world, state);
	}
	return state;
};

Given(/the text is "([^"]*)"/, function (this: object, value: string) {
	stateOf(this).text = value;
});

When(/the emoji is extracted/, function (this: object) {
	const state = stateOf(this);
	const result = extractEmoji(state.text);
	state.emoji = result.emoji;
	state.remaining = result.text;
});

Then(/the extracted emoji should be "([^"]*)"/, function (this: object, value: string) {
	assert.equal(stateOf(this).emoji, value);
});

Then(/no emoji should be found/, function (this: object) {
	assert.equal(stateOf(this).emoji, null);
});

Then(/the remaining text should be "([^"]*)"/, function (this: object, value: string) {
	assert.equal(stateOf(this).remaining, value);
});

Given(/the user searches for emoji "([^"]*)"/, function (this: object, value: string) {
	stateOf(this).results = searchEmojis(value);
});

Then(
	/the emoji search should find "([^"]*)"/,
	function (this: object, value: string) {
		const emojis = stateOf(this).results.map((entry) => entry.emoji);
		assert.ok(
			emojis.includes(value),
			`Expected the search to find ${value}, got ${emojis.join(", ")}`,
		);
	},
);

Then(
	/the emoji search should find at least (\d+) emojis/,
	function (this: object, count: number) {
		const found = stateOf(this).results.length;
		assert.ok(
			found >= count,
			`Expected at least ${count} emojis, got ${found}`,
		);
	},
);

Then(/the emoji search should find no emojis/, function (this: object) {
	assert.equal(stateOf(this).results.length, 0);
});

Given(/the icon is "([^"]*)"/, function (this: object, value: string) {
	stateOf(this).icon = value;
});

Then(/it should be an Iconify icon/, function (this: object) {
	assert.equal(isIconifyIcon(stateOf(this).icon), true);
});

Then(/it should not be an Iconify icon/, function (this: object) {
	assert.equal(isIconifyIcon(stateOf(this).icon), false);
});

Then(/the normalized icon should be "([^"]*)"/, function (this: object, value: string) {
	assert.equal(normalizeIcon(stateOf(this).icon), value);
});
