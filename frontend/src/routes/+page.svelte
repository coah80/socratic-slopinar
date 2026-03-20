<script lang="ts">
	import Chat from '$lib/components/Chat.svelte';
	import SharedDoc from '$lib/components/SharedDoc.svelte';
	import ExecutionPrompt from '$lib/components/ExecutionPrompt.svelte';
	import { getDiscussion, startDiscussion, stopDiscussion } from '$lib/stores/websocket.svelte';

	let prompt = $state('');
	let codebasePath = $state('');
	let rounds = $state(5);

	const discussion = $derived(getDiscussion());

	function handleStart() {
		if (!prompt.trim()) return;
		startDiscussion(prompt.trim(), codebasePath.trim(), rounds);
	}

	function handleStop() {
		stopDiscussion();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey && discussion.status !== 'running') {
			e.preventDefault();
			handleStart();
		}
	}
</script>

<div class="page">
	<div class="left-panel">
		<Chat messages={discussion.messages} activeModel={discussion.activeModel} activeDisplayName={discussion.activeDisplayName} />

		{#if discussion.status === 'error' && discussion.error}
			<div class="error-bar">
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<circle cx="12" cy="12" r="10"/>
					<line x1="12" y1="8" x2="12" y2="12"/>
					<line x1="12" y1="16" x2="12.01" y2="16"/>
				</svg>
				<span>{discussion.error}</span>
			</div>
		{/if}

		<div class="input-bar">
			<div class="input-row">
				<div class="input-group prompt-input">
					<input
						type="text"
						bind:value={prompt}
						placeholder="What should they discuss?"
						onkeydown={handleKeydown}
						disabled={discussion.status === 'running'}
					/>
				</div>
			</div>
			<div class="input-row bottom-row">
				<div class="input-group path-input">
					<svg class="input-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
					</svg>
					<input
						type="text"
						bind:value={codebasePath}
						placeholder="/path/to/codebase"
						disabled={discussion.status === 'running'}
					/>
				</div>
				<div class="input-group rounds-input">
					<label for="rounds">Rounds</label>
					<input
						id="rounds"
						type="number"
						bind:value={rounds}
						min="1"
						max="20"
						disabled={discussion.status === 'running'}
					/>
				</div>
				{#if discussion.status === 'running'}
					<button class="stop-btn" onclick={handleStop}>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
							<rect x="6" y="6" width="12" height="12" rx="2"/>
						</svg>
						Stop
					</button>
				{:else}
					<button class="start-btn" onclick={handleStart} disabled={!prompt.trim()}>
						Start Discussion
					</button>
				{/if}
			</div>
		</div>
	</div>

	<div class="divider"></div>

	<div class="right-panel">
		<div class="right-card">
			<SharedDoc content={discussion.sharedNotes} />
		</div>
		<div class="right-card">
			<ExecutionPrompt content={discussion.executionPrompt} />
		</div>
	</div>
</div>

<style>
	.page {
		display: flex;
		height: 100%;
		gap: 0;
		padding: 12px;
		animation: fade-up 0.4s ease both;
	}

	.left-panel {
		flex: 6;
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.divider {
		width: 1px;
		background: var(--ctp-surface0);
		margin: 0 12px;
		flex-shrink: 0;
	}

	.right-panel {
		flex: 4;
		display: flex;
		flex-direction: column;
		gap: 12px;
		min-width: 0;
	}

	.right-card {
		flex: 1;
		overflow: hidden;
		border-radius: 16px;
		border: 1px solid rgba(69, 71, 90, 0.5);
		background: rgba(49, 50, 68, 0.4);
	}

	.error-bar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		margin: 0 0 8px;
		background: rgba(243, 139, 168, 0.1);
		border: 1px solid rgba(243, 139, 168, 0.25);
		border-radius: 12px;
		color: var(--ctp-red);
		font-size: 13px;
	}

	.input-bar {
		flex-shrink: 0;
		padding: 14px 16px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		border-radius: 16px;
		border: 1px solid rgba(69, 71, 90, 0.5);
		background: rgba(49, 50, 68, 0.4);
	}

	.input-row {
		display: flex;
		gap: 8px;
		align-items: center;
	}

	.input-group {
		display: flex;
		align-items: center;
		gap: 8px;
		border-radius: 12px;
		border: 1px solid var(--ctp-surface1);
		background: rgba(49, 50, 68, 0.4);
		padding: 0 16px;
		transition: border-color 0.15s ease;
	}

	.input-group:focus-within {
		border-color: rgba(203, 166, 247, 0.5);
	}

	.input-group input {
		flex: 1;
		padding: 12px 0;
		background: transparent;
		font-size: 14px;
		font-family: var(--font-body);
		color: var(--ctp-text);
		min-width: 0;
	}

	.input-group input::placeholder {
		color: var(--ctp-overlay1);
	}

	.input-group input:disabled {
		opacity: 0.5;
	}

	.input-icon {
		color: var(--ctp-overlay1);
		flex-shrink: 0;
	}

	.prompt-input {
		flex: 1;
	}

	.prompt-input input {
		font-size: 15px;
	}

	.bottom-row {
		display: flex;
	}

	.path-input {
		flex: 1;
	}

	.rounds-input {
		width: 120px;
		flex-shrink: 0;
	}

	.rounds-input label {
		font-size: 12px;
		color: var(--ctp-overlay1);
		white-space: nowrap;
	}

	.rounds-input input {
		width: 50px;
		text-align: center;
		font-family: var(--font-data);
		font-size: 14px;
	}

	.start-btn {
		padding: 10px 24px;
		background: var(--ctp-mauve);
		border-radius: 12px;
		font-size: 14px;
		font-weight: 700;
		color: var(--ctp-crust);
		white-space: nowrap;
		transition: filter 0.15s ease;
		flex-shrink: 0;
		font-family: var(--font-body);
	}

	.start-btn:hover:not(:disabled) {
		filter: brightness(1.1);
	}

	.start-btn:active:not(:disabled) {
		filter: brightness(0.9);
	}

	.start-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.stop-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 10px 24px;
		background: var(--ctp-red);
		border-radius: 12px;
		font-size: 14px;
		font-weight: 700;
		color: var(--ctp-crust);
		white-space: nowrap;
		transition: filter 0.15s ease;
		flex-shrink: 0;
		font-family: var(--font-body);
	}

	.stop-btn:hover {
		filter: brightness(1.1);
	}

	.stop-btn:active {
		filter: brightness(0.9);
	}
</style>
