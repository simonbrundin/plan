import { beforeEach, describe, expect, it } from "bun:test";
import { useUndoStack } from "../../app/composables/useUndoStack";

describe("useUndoStack", () => {
	// Stapeln är global - varje test börjar med en tom stapel.
	beforeEach(() => {
		useUndoStack().clear();
	});

	it("ångrar den senaste handlingen", async () => {
		const stack = useUndoStack();
		let undone = 0;

		stack.push({ label: "första", undo: () => { undone += 1; } });
		stack.push({ label: "andra", undo: () => { undone += 1; } });

		expect(stack.canUndo.value).toBe(true);
		expect(stack.nextLabel.value).toBe("andra");

		const label = await stack.undoLast();

		expect(label).toBe("andra");
		expect(undone).toBe(1);
		expect(stack.nextLabel.value).toBe("första");
	});

	it("ångrar i omvänd ordning", async () => {
		const stack = useUndoStack();
		const order: string[] = [];

		stack.push({ label: "första", undo: () => { order.push("första"); } });
		stack.push({ label: "andra", undo: () => { order.push("andra"); } });

		await stack.undoLast();
		await stack.undoLast();

		expect(order).toEqual(["andra", "första"]);
		expect(stack.canUndo.value).toBe(false);
	});

	it("returnerar null när det inte finns något att ångra", async () => {
		const stack = useUndoStack();

		expect(await stack.undoLast()).toBeNull();
		expect(stack.canUndo.value).toBe(false);
	});

	it("lägger tillbaka handlingen om ångrandet misslyckas", async () => {
		const stack = useUndoStack();

		stack.push({
			label: "trasig",
			undo: () => {
				throw new Error("nope");
			},
		});

		await expect(stack.undoLast()).rejects.toThrow("nope");
		expect(stack.canUndo.value).toBe(true);
		expect(stack.nextLabel.value).toBe("trasig");
	});

	it("räknar upp revisionen när något ångras", async () => {
		const stack = useUndoStack();
		const before = stack.revision.value;

		stack.push({ label: "något", undo: () => {} });
		await stack.undoLast();

		expect(stack.revision.value).toBe(before + 1);
	});
});
