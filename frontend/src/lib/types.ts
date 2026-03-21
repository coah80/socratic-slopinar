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
	provider_keys: Record<string, string>;
};

export const PROVIDERS = [
	{ id: 'openrouter', name: 'OpenRouter', color: '#6366f1' },
	{ id: 'openai', name: 'OpenAI', color: '#10a37f' },
	{ id: 'anthropic', name: 'Anthropic', color: '#d4a574' },
	{ id: 'google', name: 'Google', color: '#4285f4' },
	{ id: 'xai', name: 'xAI', color: '#1da1f2' },
	{ id: 'deepseek', name: 'DeepSeek', color: '#0066ff' },
	{ id: 'mistral', name: 'Mistral', color: '#ff7000' },
	{ id: 'groq', name: 'Groq', color: '#f55036' },
	{ id: 'together', name: 'Together', color: '#5046e5' },
	{ id: 'minimax', name: 'MiniMax', color: '#7c3aed' },
] as const;

export type ProviderId = typeof PROVIDERS[number]['id'];

const MODEL_PREFIX_MAP: Record<string, ProviderId> = {
	'openai/': 'openai',
	'anthropic/': 'anthropic',
	'google/': 'google',
	'x-ai/': 'xai',
	'xai/': 'xai',
	'deepseek/': 'deepseek',
	'mistralai/': 'mistral',
	'mistral/': 'mistral',
	'groq/': 'groq',
	'together/': 'together',
	'minimax/': 'minimax',
};

export function getProviderForModel(modelId: string): ProviderId {
	const lower = modelId.toLowerCase();
	for (const [prefix, provider] of Object.entries(MODEL_PREFIX_MAP)) {
		if (lower.startsWith(prefix)) return provider;
	}
	return 'openrouter';
}

export function getProviderInfo(providerId: string) {
	return PROVIDERS.find(p => p.id === providerId) ?? PROVIDERS[0];
}

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
