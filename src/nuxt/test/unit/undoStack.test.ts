import { describe, expect, it } from "bun:test";
import {
	popUndoEntry,
	pushUndoEntry,
	UNDO_STACK_LIMIT,
	type UndoEntry,
} from "../../app/utils/undoStack";

function entry(label: string): UndoEntry {
	return { label, undo: () => {} };
}

describe("pushUndoEntry", () => {
	it("lägger handlingen överst i stapeln", () => {
		const first = entry("första");
		const second = entry("andra");

		const stack = pushUndoEntry(pushUndoEntry([], first), second);

		expect(stack.map((e) => e.label)).toEqual(["första", "andra"]);
	});

	it("tappar den äldsta posten när gränsen nås", () => {
		let stack: UndoEntry[] = [];
		for (let i = 0; i < UNDO_STACK_LIMIT + 3; i++) {
			stack = pushUndoEntry(stack, entry(`handling ${i}`));
		}

		expect(stack).toHaveLength(UNDO_STACK_LIMIT);
		expect(stack[0]?.label).toBe("handling 3");
		expect(stack[stack.length - 1]?.label).toBe(
			`handling ${UNDO_STACK_LIMIT + 2}`,
		);
	});
});

describe("popUndoEntry", () => {
	it("tar bort den senaste handlingen", () => {
		const stack = pushUndoEntry(pushUndoEntry([], entry("första")), entry("andra"));

		const { entries, entry: popped } = popUndoEntry(stack);

		expect(popped?.label).toBe("andra");
		expect(entries.map((e) => e.label)).toEqual(["första"]);
	});

	it("returnerar null när stapeln är tom", () => {
		const { entries, entry: popped } = popUndoEntry([]);

		expect(popped).toBeNull();
		expect(entries).toEqual([]);
	});
});
