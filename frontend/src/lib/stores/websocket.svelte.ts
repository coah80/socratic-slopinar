import type { Message, WSMessage, DiscussionState, ToolCall } from '$lib/types';

let state = $state<DiscussionState>({
	messages: [],
	sharedNotes: '',
	executionPrompt: '',
	activeModel: null,
	activeDisplayName: null,
	status: 'idle',
	error: null
});

let ws: WebSocket | null = null;
let messageCounter = 0;
let displayNames = new Map<string, string>();

export function getDiscussion(): DiscussionState {
	return state;
}

function createMessage(modelId: string, displayName: string, content: string, toolCalls: ToolCall[] = []): Message {
	messageCounter++;
	return {
		id: `msg-${messageCounter}-${Date.now()}`,
		model_id: modelId,
		display_name: displayName || modelId,
		role: 'assistant',
		content,
		tool_calls: toolCalls,
		timestamp: Date.now()
	};
}

function findLastMessageByModel(modelId: string): Message | undefined {
	for (let i = state.messages.length - 1; i >= 0; i--) {
		if (state.messages[i].model_id === modelId) return state.messages[i];
	}
	return undefined;
}

function handleWSMessage(event: MessageEvent): void {
	let data: WSMessage;
	try {
		data = JSON.parse(event.data);
	} catch {
		return;
	}

	const content = data.content as any;

	switch (data.type) {
		case 'message': {
			if (data.display_name) displayNames.set(data.model_id, data.display_name);
			const msg = createMessage(data.model_id, data.display_name ?? data.model_id, typeof content === 'string' ? content : '');
			state = {
				...state,
				messages: [...state.messages, msg],
				activeModel: data.model_id,
				activeDisplayName: data.display_name ?? data.model_id
			};
			break;
		}

		case 'tool_call': {
			const tc: ToolCall = { name: content?.name ?? '', arguments: content?.arguments ?? '' };
			const existing = findLastMessageByModel(data.model_id);
			if (existing) {
				const updated: Message = {
					...existing,
					tool_calls: [...existing.tool_calls, tc]
				};
				state = {
					...state,
					messages: state.messages.map((m) => (m.id === existing.id ? updated : m))
				};
			} else {
				const msg = createMessage(data.model_id, data.display_name ?? data.model_id, '', [tc]);
				state = { ...state, messages: [...state.messages, msg] };
			}
			break;
		}

		case 'tool_result': {
			const toolName = content?.name ?? '';
			const toolResult = content?.result ?? '';
			const existing2 = findLastMessageByModel(data.model_id);
			if (existing2) {
				const updatedCalls = existing2.tool_calls.map((tc) =>
					tc.name === toolName ? { ...tc, result: toolResult } : tc
				);
				const updated: Message = { ...existing2, tool_calls: updatedCalls };
				state = {
					...state,
					messages: state.messages.map((m) => (m.id === existing2.id ? updated : m))
				};
			}
			break;
		}

		case 'notes_update': {
			state = { ...state, sharedNotes: typeof content === 'string' ? content : '' };
			break;
		}

		case 'execution_prompt': {
			state = { ...state, executionPrompt: typeof content === 'string' ? content : '' };
			break;
		}

		case 'status': {
			if (data.display_name && data.model_id) displayNames.set(data.model_id, data.display_name);
			const s = typeof content === 'string' ? content : '';
			const newStatus = (s === 'done' || s === 'complete') ? 'complete' as const : s === 'running' ? 'running' as const : state.status;
			const isThinking = s.includes('thinking');
			state = {
				...state,
				status: newStatus,
				activeModel: newStatus === 'complete' ? null : (isThinking ? data.model_id : state.activeModel),
				activeDisplayName: newStatus === 'complete' ? null : (isThinking ? (data.display_name ?? displayNames.get(data.model_id) ?? data.model_id) : state.activeDisplayName)
			};
			break;
		}

		case 'error': {
			state = { ...state, status: 'error', error: typeof content === 'string' ? content : 'Unknown error', activeModel: null, activeDisplayName: null };
			break;
		}
	}
}

export function startDiscussion(prompt: string, codebasePath: string, rounds: number): void {
	stopDiscussion();

	messageCounter = 0;
	state = {
		messages: [],
		sharedNotes: '',
		executionPrompt: '',
		activeModel: null,
		status: 'running',
		error: null
	};

	const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	const wsUrl = `${protocol}//${window.location.host}/api/discuss`;
	ws = new WebSocket(wsUrl);

	ws.onopen = () => {
		ws?.send(JSON.stringify({ prompt, codebase_path: codebasePath, rounds }));
	};

	ws.onmessage = handleWSMessage;

	ws.onerror = () => {
		state = { ...state, status: 'error', error: 'WebSocket connection failed', activeModel: null };
	};

	ws.onclose = () => {
		if (state.status === 'running') {
			state = { ...state, status: 'complete', activeModel: null };
		}
	};
}

export function stopDiscussion(): void {
	if (ws) {
		if (ws.readyState === WebSocket.OPEN) {
			ws.send(JSON.stringify({ action: 'stop' }));
		}
		ws.close();
		ws = null;
	}
	if (state.status === 'running') {
		state = { ...state, status: 'idle', activeModel: null, activeDisplayName: null };
	}
}
