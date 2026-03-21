import type { HistoryItem, HistoryDiscussion } from '$lib/types';

let items = $state<HistoryItem[]>([]);
let loading = $state(false);
let error = $state<string | null>(null);

export function getHistoryItems(): HistoryItem[] {
	return items;
}

export function isHistoryLoading(): boolean {
	return loading;
}

export function getHistoryError(): string | null {
	return error;
}

export async function loadHistory(): Promise<void> {
	loading = true;
	error = null;
	try {
		const res = await fetch('/api/history');
		if (!res.ok) throw new Error(`Failed to load history: ${res.statusText}`);
		const data = await res.json();
		items = Array.isArray(data) ? data : [];
	} catch (e) {
		error = e instanceof Error ? e.message : 'Failed to load history';
	} finally {
		loading = false;
	}
}

export async function loadDiscussion(id: string): Promise<HistoryDiscussion | null> {
	try {
		const res = await fetch(`/api/history/${encodeURIComponent(id)}`);
		if (!res.ok) throw new Error(`Failed to load discussion: ${res.statusText}`);
		return await res.json();
	} catch (e) {
		error = e instanceof Error ? e.message : 'Failed to load discussion';
		return null;
	}
}

export async function deleteDiscussion(id: string): Promise<void> {
	try {
		const res = await fetch(`/api/history/${encodeURIComponent(id)}`, { method: 'DELETE' });
		if (!res.ok) throw new Error(`Failed to delete discussion: ${res.statusText}`);
		items = items.filter((item) => item.id !== id);
	} catch (e) {
		error = e instanceof Error ? e.message : 'Failed to delete discussion';
	}
}

export function exportDiscussion(id: string): void {
	const link = document.createElement('a');
	link.href = `/api/export/${encodeURIComponent(id)}`;
	link.download = `discussion-${id}.zip`;
	document.body.appendChild(link);
	link.click();
	document.body.removeChild(link);
}
