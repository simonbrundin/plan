import type { AgentApproval, AgentConversation, AgentMessage } from '~/types/agent';
import { apiRequest } from '~/utils/api';

/**
 * Chatt med agenten - Plan är gränssnittet, Slack används inte.
 *
 * Samtalen går via Nuxt:s server routes, som proxar till agent-API:t (nyckeln
 * ska inte nå webbläsaren). Godkännanden och beslut ligger i Plan API:t och
 * anropas direkt.
 */
export function useAgentChat(goalId: Ref<number | undefined>) {
  const messages = ref<AgentMessage[]>([]);
  const conversationId = ref<string | null>(null);
  const approval = ref<AgentApproval | null>(null);
  const streaming = ref(false);
  const error = ref<string | null>(null);
  const available = ref(true);

  // Laddar meddelandena och eventuellt väntande förslag.
  const load = async (id?: string | null) => {
    const convId = id ?? conversationId.value;
    if (!convId) return;

    try {
      const response = await $fetch<{ messages: AgentMessage[] }>(
        `/api/agent/conversations/${convId}/messages`,
      );
      messages.value = (response?.messages ?? []).map((message) => ({
        ...message,
        id: message.id ?? `${message.role}-${message.created_at ?? ''}`,
      }));
    } catch (err: any) {
      available.value = false;
      error.value = 'Kunde inte hämta chatten från agenten.';
    }

    await loadApproval();
  };

  // Hämtar ett förslag som väntar på beslut.
  //
  // Godkännandena ligger i Plan - därför läses de därifrån.
  const loadApproval = async () => {
    if (!goalId.value) return;
    try {
      const response = await apiRequest<{ approvals: AgentApproval[] }>('agent-approvals');
      const match = (response?.approvals ?? []).find(
        (item) => item.goal_id === goalId.value,
      );
      approval.value = match ?? null;
    } catch {
      // Plan svarar inte - inget att godkänna.
      approval.value = null;
    }
  };

  // Startar en konversation för målet.
  const startConversation = async (title: string): Promise<string | null> => {
    try {
      const response = await $fetch<AgentConversation>('/api/agent/conversations', {
        method: 'POST',
        body: { title, model: '', provider: '' },
      });
      conversationId.value = response?.id ?? null;
      return conversationId.value;
    } catch (err: any) {
      available.value = false;
      error.value = 'Kunde inte starta en konversation med agenten.';
      return null;
    }
  };

  // Skickar ett meddelande och strömmar svaret in i chatten.
  const send = async (
    text: string,
    goalContext?: { id: number; title: string; description?: string },
  ) => {
    const content = text.trim();
    if (!content || streaming.value) return;

    if (!conversationId.value) return;

    streaming.value = true;
    error.value = null;

    const userMessage: AgentMessage = {
      id: `local-${Date.now()}`,
      role: 'user',
      content,
      created_at: new Date().toISOString(),
    };
    messages.value = [...messages.value, userMessage];

    try {
      const response = await fetch(
        `/api/agent/conversations/${conversationId.value}/messages`,
        {
          method: 'POST',
          headers: { 'content-type': 'application/json' },
          body: JSON.stringify({ content, goal_context: goalContext }),
        },
      );

      if (!response.ok || !response.body) {
        throw new Error(`Agent-API svarade ${response.status}`);
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      for (;;) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() ?? '';

        for (const line of lines) {
          const trimmed = line.trim();
          if (!trimmed) continue;
          handleChunk(trimmed);
        }
      }
    } catch (err: any) {
      error.value = 'Kunde inte skicka meddelandet till agenten.';
    } finally {
      streaming.value = false;
      await loadApproval();
    }
  };

  // Tolkar en NDJSON-chunk från agent-API:t.
  const handleChunk = (line: string) => {
    let chunk: any;
    try {
      chunk = JSON.parse(line);
    } catch {
      return;
    }

    const type = chunk?.type;

    // Agenten svarar - texten ligger i message.content.
    if (type === 'assistant_message' && chunk.message?.content) {
      messages.value = [
        ...messages.value,
        {
          id: chunk.message.id ?? `assistant-${Date.now()}`,
          role: 'assistant',
          content: chunk.message.content,
          created_at: chunk.message.created_at,
        },
      ];
      return;
    }

    // Agentens tankar - visas inte i chatten men sparas.
    if (type === 'thinking') {
      return;
    }

    // Verktygsanrop och resultat.
    if (type === 'tool_request' || type === 'tool_result') {
      const message = chunk.message ?? {};
      messages.value = [
        ...messages.value,
        {
          id: message.id ?? `tool-${Date.now()}-${messages.value.length}`,
          role: 'tool',
          content: message.content ?? message.tool_result ?? '',
          tool_name: message.tool_name,
          tool_result: message.tool_result,
          is_error: message.is_error,
        },
      ];
      return;
    }

    if (type === 'error' || chunk.error) {
      error.value = chunk.error ?? 'Agenten stötte på ett problem.';
    }
  };

  // Godkänner eller avvisar ett förslag.
  //
  // Beslutet tas i Plan av en människa - agenten kan inte godkänna sig
  // själv, och det tekniska hindret ligger i Plan, inte i prompten.
  const decide = async (decision: 'approve' | 'reject', feedback = '') => {
    const pending = approval.value;
    if (!pending) return;
    try {
      await apiRequest(`agent-approvals/${pending.id}/decision`, {
        method: 'POST',
        body: { decision, feedback },
      });
      approval.value = null;
    } catch {
      error.value = 'Kunde inte spara beslutet.';
    }
  };

  return {
    messages,
    conversationId,
    approval,
    streaming,
    error,
    available,
    load,
    loadApproval,
    startConversation,
    send,
    decide,
  };
}
