export type MessageRole = 'assistant' | 'system' | 'tool' | 'god';

export type ToolCall = {
	name: string;
	arguments: string;
	result?: string;
};

export type Message = {
	id: string;
	model_id: string;
	display_name: string;
	role: MessageRole;
	content: string;
	tool_calls: ToolCall[];
	timestamp: number;
};

export type PinnedMessage = {
	model_id: string;
	display_name: string;
	content: string;
};

export type WSMessage = {
	type: 'message' | 'tool_call' | 'tool_result' | 'notes_update' | 'execution_prompt' | 'status' | 'error' | 'pin';
	model_id: string;
	display_name?: string;
	content: unknown;
};

export type DiscussionRequest = {
	prompt: string;
	codebase_path: string;
	rounds: number;
};

export type Config = {
	api_key: string;
	models: string[];
	tavily_api_key: string;
};

export type DiscussionState = {
	messages: Message[];
	sharedNotes: string;
	executionPrompt: string;
	activeModel: string | null;
	activeDisplayName: string | null;
	status: 'idle' | 'running' | 'complete' | 'error';
	error: string | null;
	discussionId: string | null;
	pinnedMessages: PinnedMessage[];
	mutedModels: Set<string>;
};

export type HistoryItem = {
	id: string;
	prompt: string;
	model_count: number;
	message_count: number;
	date: string;
};

export type HistoryDiscussion = {
	id: string;
	prompt: string;
	messages: Message[];
	sharedNotes: string;
	executionPrompt: string;
};

const PALETTE = [
	'#cba6f7',
	'#89b4fa',
	'#a6e3a1',
	'#f9e2af',
	'#f38ba8',
	'#89dceb',
	'#fab387',
	'#f5c2e7',
	'#94e2d5',
	'#b4befe',
] as const;

const colorMap = new Map<string, string>();

export function getModelColor(modelId: string): string {
	const existing = colorMap.get(modelId);
	if (existing) return existing;

	const idx = colorMap.size % PALETTE.length;
	const color = PALETTE[idx];
	colorMap.set(modelId, color);
	return color;
}
