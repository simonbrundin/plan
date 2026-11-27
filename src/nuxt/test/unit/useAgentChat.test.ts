import { beforeEach, describe, expect, it } from "bun:test";
import { ref } from "vue";
import { useAgentChat } from "../../app/composables/useAgentChat";

type Anrop = { url: string; method: string; body?: unknown; query?: unknown };

const anrop: Anrop[] = [];

// Svaren varieras per test - listan sätts om i beforeEach.
let svar: (url: string, options: any) => unknown = () => ({});

beforeEach(() => {
	anrop.length = 0;
	svar = () => ({});
	(globalThis as any).$fetch = async (url: string, options: any = {}) => {
		anrop.push({
			url,
			method: options?.method ?? "GET",
			body: options?.body,
			query: options?.query,
		});
		return svar(url, options);
	};
});

describe("useAgentChat", () => {
	it("listar samtalen för ett mål", async () => {
		svar = () => ({
			conversations: [
				{ id: "nytt", title: "Senaste", model: "", provider: "", goal_id: 540 },
			],
		});

		const chatt = useAgentChat(ref(540));
		const lista = await chatt.listConversations();

		expect(lista).toHaveLength(1);
		expect(anrop[0].url).toBe("/api/agent/conversations");
		expect(anrop[0].query).toEqual({ goal_id: 540 });
		expect(chatt.conversations.value[0].id).toBe("nytt");
	});

	it("skickar målets id när ett samtal skapas", async () => {
		svar = () => ({ id: "nytt", title: "Titel", model: "", provider: "" });

		const chatt = useAgentChat(ref(540));
		const id = await chatt.startConversation("Titel");

		expect(id).toBe("nytt");
		expect(anrop[0].method).toBe("POST");
		expect((anrop[0].body as any).goal_id).toBe(540);
		expect((anrop[0].body as any).title).toBe("Titel");
	});

	it("laddar det senaste samtalet för målet", async () => {
		svar = (url) => {
			if (url === "/api/agent/conversations") {
				return {
					conversations: [
						{ id: "senaste", title: "Senaste", model: "", provider: "" },
						{ id: "äldre", title: "Äldre", model: "", provider: "" },
					],
				};
			}
			if (url === "/api/agent/conversations/senaste/messages") {
				return { messages: [{ id: "1", role: "user", content: "Hej" }] };
			}
			return { approvals: [] };
		};

		const chatt = useAgentChat(ref(540));
		const id = await chatt.loadLatestForGoal();

		expect(id).toBe("senaste");
		expect(chatt.conversationId.value).toBe("senaste");
		expect(chatt.messages.value).toHaveLength(1);
	});

	it("öppnar ett tidigare samtal", async () => {
		svar = (url) => {
			if (url === "/api/agent/conversations/äldre/messages") {
				return { messages: [{ id: "2", role: "assistant", content: "Hej" }] };
			}
			return { approvals: [] };
		};

		const chatt = useAgentChat(ref(540));
		await chatt.openConversation("äldre");

		expect(chatt.conversationId.value).toBe("äldre");
		expect(chatt.messages.value).toHaveLength(1);
	});
});
