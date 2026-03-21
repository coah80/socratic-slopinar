<script lang="ts">
	import { getHistoryItems, isHistoryLoading, loadHistory, loadDiscussion, deleteDiscussion } from '$lib/stores/history.svelte';
	import { loadViewState } from '$lib/stores/websocket.svelte';

	let { onclose }: { onclose: () => void } = $props();

	const items = $derived(getHistoryItems());
	const loading = $derived(isHistoryLoading());

	$effect(() => {
		loadHistory();
	});

	function handleOverlayClick(e: MouseEvent) {
		if (e.target === e.currentTarget) onclose();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onclose();
	}

	async function handleSelect(id: string) {
		const discussion = await loadDiscussion(id);
		if (discussion) {
			loadViewState({
				id: discussion.id,
				messages: discussion.messages,
				sharedNotes: discussion.sharedNotes,
				executionPrompt: discussion.executionPrompt
			});
			onclose();
		}
	}

	async function handleDelete(e: MouseEvent, id: string) {
		e.stopPropagation();
		await deleteDiscussion(id);
	}

	function truncate(text: string, max: number): string {
		if (text.length <= max) return text;
		return text.slice(0, max) + '...';
	}
</script>

<!-- svelte-ignore a11y_interactive_supports_focus -->
<div class="overlay" role="dialog" aria-modal="true" onclick={handleOverlayClick} onkeydown={handleKeydown}>
	<div class="panel">
		<div class="panel-header">
			<div class="header-left">
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<circle cx="12" cy="12" r="10"/>
					<polyline points="12 6 12 12 16 14"/>
				</svg>
				<span>History<span class="accent">.</span></span>
			</div>
			<button class="close-btn" onclick={onclose} aria-label="Close">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<line x1="18" y1="6" x2="6" y2="18"/>
					<line x1="6" y1="6" x2="18" y2="18"/>
				</svg>
			</button>
		</div>

		<div class="panel-body">
			{#if loading}
				<div class="empty-state">Loading...</div>
			{:else if items.length === 0}
				<div class="empty-state">No past discussions</div>
			{:else}
				{#each items as item (item.id)}
					<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div class="history-item" onclick={() => handleSelect(item.id)} onkeydown={(e) => { if (e.key === 'Enter') handleSelect(item.id); }} role="button" tabindex="0">
						<div class="item-prompt">{truncate(item.prompt, 60)}</div>
						<div class="item-meta">
							<span class="meta-tag">
								<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
									<rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
									<line x1="8" y1="21" x2="16" y2="21"/>
									<line x1="12" y1="17" x2="12" y2="21"/>
								</svg>
								{item.model_count}
							</span>
							<span class="meta-tag">
								<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
									<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
								</svg>
								{item.message_count}
							</span>
							<span class="meta-date">{item.date}</span>
						</div>
						<button
							class="delete-btn"
							onclick={(e) => handleDelete(e, item.id)}
							aria-label="Delete discussion"
						>
							<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<line x1="18" y1="6" x2="6" y2="18"/>
								<line x1="6" y1="6" x2="18" y2="18"/>
							</svg>
						</button>
					</div>
				{/each}
			{/if}
		</div>
	</div>
</div>

<style>
	.overlay {
		position: fixed;
		inset: 0;
		z-index: 60;
		background: rgba(17, 17, 27, 0.5);
		backdrop-filter: blur(4px);
		-webkit-backdrop-filter: blur(4px);
	}

	.panel {
		position: absolute;
		left: 0;
		top: 0;
		bottom: 0;
		width: 380px;
		max-width: 90vw;
		display: flex;
		flex-direction: column;
		background: var(--ctp-base);
		border-right: 1px solid var(--ctp-surface1);
		animation: slide-in 0.25s ease both;
	}

	@keyframes slide-in {
		from { transform: translateX(-100%); }
		to { transform: translateX(0); }
	}

	.panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 20px 24px;
		border-bottom: 1px solid var(--ctp-surface0);
		flex-shrink: 0;
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: 8px;
		font-family: var(--font-heading);
		font-size: 16px;
		font-weight: 700;
		color: var(--ctp-text);
	}

	.accent {
		color: var(--ctp-mauve);
	}

	.close-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 10px;
		color: var(--ctp-subtext0);
		transition: all 0.15s ease;
	}

	.close-btn:hover {
		background: rgba(49, 50, 68, 0.4);
		color: var(--ctp-text);
	}

	.panel-body {
		flex: 1;
		overflow-y: auto;
		padding: 12px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.empty-state {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--ctp-overlay1);
		font-size: 13px;
		font-style: italic;
	}

	.history-item {
		position: relative;
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding: 14px 16px;
		border-radius: 16px;
		border: 1px solid rgba(69, 71, 90, 0.5);
		background: rgba(49, 50, 68, 0.4);
		text-align: left;
		transition: all 0.15s ease;
		cursor: pointer;
	}

	.history-item:hover {
		border-color: rgba(203, 166, 247, 0.3);
		background: rgba(49, 50, 68, 0.6);
	}

	.item-prompt {
		font-size: 14px;
		font-weight: 500;
		color: var(--ctp-text);
		line-height: 1.4;
		padding-right: 28px;
	}

	.item-meta {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.meta-tag {
		display: flex;
		align-items: center;
		gap: 4px;
		font-family: var(--font-data);
		font-size: 12px;
		color: var(--ctp-overlay1);
	}

	.meta-date {
		font-family: var(--font-data);
		font-size: 12px;
		color: var(--ctp-overlay0);
		margin-left: auto;
	}

	.delete-btn {
		position: absolute;
		top: 12px;
		right: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 26px;
		height: 26px;
		border-radius: 8px;
		color: var(--ctp-overlay0);
		transition: all 0.15s ease;
	}

	.delete-btn:hover {
		background: rgba(243, 139, 168, 0.15);
		color: var(--ctp-red);
	}
</style>
