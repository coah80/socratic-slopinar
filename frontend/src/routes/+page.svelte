<script lang="ts">
	import Chat from '$lib/components/Chat.svelte';
	import SharedDoc from '$lib/components/SharedDoc.svelte';
	import ExecutionPrompt from '$lib/components/ExecutionPrompt.svelte';
	import PinnedMessages from '$lib/components/PinnedMessages.svelte';
	import { getDiscussion, startDiscussion, stopDiscussion, muteModel, unmuteModel, injectMessage } from '$lib/stores/websocket.svelte';
	import { exportDiscussion } from '$lib/stores/history.svelte';

	let prompt = $state('');
	let codebasePath = $state('');
	let rounds = $state(5);
	let godInput = $state('');

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

	function handleGodKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleGodSend();
		}
	}

	function handleGodSend() {
		if (!godInput.trim()) return;
		injectMessage(godInput.trim());
		godInput = '';
	}

	function handleExport() {
		if (discussion.discussionId) {
			exportDiscussion(discussion.discussionId);
		}
	}
</script>

<div class="page">
	<div class="left-panel">
		<Chat
			messages={discussion.messages}
			activeModel={discussion.activeModel}
			activeDisplayName={discussion.activeDisplayName}
			mutedModels={discussion.mutedModels}
			onmute={muteModel}
			onunmute={unmuteModel}
		/>

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

		{#if discussion.status === 'running'}
			<div class="god-bar">
				<div class="god-input-group">
					<svg class="god-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--ctp-yellow)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
					</svg>
					<input
						type="text"
						bind:value={godInput}
						placeholder="Intervene as God..."
						onkeydown={handleGodKeydown}
					/>
				</div>
				<button class="god-send-btn" onclick={handleGodSend} disabled={!godInput.trim()}>
					Send
				</button>
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
		{#if discussion.pinnedMessages.length > 0}
			<div class="right-card pinned-card">
				<PinnedMessages pins={discussion.pinnedMessages} />
			</div>
		{/if}
		<div class="right-card">
			<div class="shared-doc-wrapper">
				<SharedDoc content={discussion.sharedNotes} />
			</div>
		</div>
		<div class="right-card">
			<div class="exec-prompt-wrapper">
				<ExecutionPrompt content={discussion.executionPrompt} />
				{#if discussion.status === 'complete' && discussion.discussionId}
					<div class="export-area">
						<button class="export-btn" onclick={handleExport}>
							<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
								<polyline points="7 10 12 15 17 10"/>
								<line x1="12" y1="15" x2="12" y2="3"/>
							</svg>
							Export
						</button>
					</div>
				{/if}
			</div>
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

	.pinned-card {
		flex: 0 0 auto;
		max-height: 200px;
	}

	.shared-doc-wrapper {
		height: 100%;
	}

	.exec-prompt-wrapper {
		display: flex;
		flex-direction: column;
		height: 100%;
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

	.god-bar {
		display: flex;
		gap: 8px;
		align-items: center;
		padding: 10px 16px;
		margin: 0 0 8px;
		border-radius: 16px;
		border: 1px solid rgba(249, 226, 175, 0.25);
		background: rgba(249, 226, 175, 0.06);
	}

	.god-input-group {
		flex: 1;
		display: flex;
		align-items: center;
		gap: 8px;
		border-radius: 12px;
		border: 1px solid rgba(249, 226, 175, 0.2);
		background: rgba(49, 50, 68, 0.4);
		padding: 0 14px;
	}

	.god-input-group:focus-within {
		border-color: rgba(249, 226, 175, 0.5);
	}

	.god-input-group input {
		flex: 1;
		padding: 10px 0;
		background: transparent;
		font-size: 14px;
		font-family: var(--font-body);
		color: var(--ctp-text);
		min-width: 0;
	}

	.god-input-group input::placeholder {
		color: var(--ctp-overlay1);
	}

	.god-icon {
		flex-shrink: 0;
	}

	.god-send-btn {
		padding: 10px 20px;
		background: var(--ctp-yellow);
		border-radius: 12px;
		font-size: 14px;
		font-weight: 700;
		color: var(--ctp-crust);
		white-space: nowrap;
		transition: filter 0.15s ease;
		flex-shrink: 0;
		font-family: var(--font-body);
	}

	.god-send-btn:hover:not(:disabled) {
		filter: brightness(1.1);
	}

	.god-send-btn:active:not(:disabled) {
		filter: brightness(0.9);
	}

	.god-send-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
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

	.export-area {
		padding: 12px 16px;
		border-top: 1px solid var(--ctp-surface0);
		flex-shrink: 0;
	}

	.export-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 16px;
		border-radius: 12px;
		font-size: 13px;
		font-weight: 600;
		color: var(--ctp-subtext0);
		border: 1px solid var(--ctp-surface1);
		background: rgba(49, 50, 68, 0.4);
		transition: all 0.15s ease;
		width: 100%;
		justify-content: center;
		font-family: var(--font-body);
	}

	.export-btn:hover {
		background: var(--ctp-surface1);
		color: var(--ctp-text);
	}
</style>
