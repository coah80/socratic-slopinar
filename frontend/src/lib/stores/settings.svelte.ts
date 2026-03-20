import type { Config } from '$lib/types';

let config = $state<Config>({ api_key: '', models: [], tavily_api_key: '' });
let loading = $state(false);
let error = $state<string | null>(null);

export function getConfig(): Config {
	return config;
}

export function isLoading(): boolean {
	return loading;
}

export function getError(): string | null {
	return error;
}

export async function loadConfig(): Promise<void> {
	loading = true;
	error = null;
	try {
		const res = await fetch('/api/config');
		if (!res.ok) throw new Error(`Failed to load config: ${res.statusText}`);
		const data = await res.json();
		config = { api_key: data.api_key ?? '', models: data.models ?? [], tavily_api_key: data.tavily_api_key ?? '' };
	} catch (e) {
		error = e instanceof Error ? e.message : 'Failed to load config';
	} finally {
		loading = false;
	}
}

export async function saveConfig(newConfig: Config): Promise<void> {
	loading = true;
	error = null;
	try {
		const res = await fetch('/api/config', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(newConfig)
		});
		if (!res.ok) throw new Error(`Failed to save config: ${res.statusText}`);
		config = { ...newConfig };
	} catch (e) {
		error = e instanceof Error ? e.message : 'Failed to save config';
	} finally {
		loading = false;
	}
}

export async function addModel(modelId: string): Promise<void> {
	if (!modelId.trim()) return;
	const trimmed = modelId.trim();
	if (config.models.includes(trimmed)) return;

	try {
		const res = await fetch('/api/config/models', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ model: trimmed })
		});
		if (!res.ok) throw new Error(`Failed to add model: ${res.statusText}`);
		config = { ...config, models: [...config.models, trimmed] };
	} catch (e) {
		error = e instanceof Error ? e.message : 'Failed to add model';
	}
}

export async function removeModel(modelId: string): Promise<void> {
	try {
		const res = await fetch(`/api/config/models/${encodeURIComponent(modelId)}`, {
			method: 'DELETE'
		});
		if (!res.ok) throw new Error(`Failed to remove model: ${res.statusText}`);
		config = { ...config, models: config.models.filter((m) => m !== modelId) };
	} catch (e) {
		error = e instanceof Error ? e.message : 'Failed to remove model';
	}
}
