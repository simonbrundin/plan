import { describe, expect, it } from "bun:test";
import { matchLeaderKey } from "../../app/utils/leaderKeys";

const COMMANDS = ["gi", "gg", "gp", "gc", "gs", "fd", "i"];

describe("matchLeaderKey", () => {
	it("kör kommandot när sekvensen är komplett", () => {
		expect(matchLeaderKey(COMMANDS, "g", "p")).toEqual({
			type: "run",
			keys: "gp",
		});
	});

	it("väntar på fler tangenter när sekvensen är ett prefix", () => {
		expect(matchLeaderKey(COMMANDS, "", "g")).toEqual({
			type: "pending",
			keys: "g",
		});
		expect(matchLeaderKey(COMMANDS, "f", "d")).toEqual({
			type: "run",
			keys: "fd",
		});
	});

	it("återställer när tangenten inte hör till något kommando", () => {
		expect(matchLeaderKey(COMMANDS, "g", "x")).toEqual({ type: "reset" });
		expect(matchLeaderKey(COMMANDS, "", "z")).toEqual({ type: "reset" });
	});

	it("kör enkla kommandon direkt", () => {
		expect(matchLeaderKey(COMMANDS, "", "i")).toEqual({
			type: "run",
			keys: "i",
		});
	});
});
