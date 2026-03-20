<script lang="ts">
	let { content }: { content: string } = $props();

	let copied = $state(false);

	async function copyToClipboard() {
		if (!content) return;
		try {
			await navigator.clipboard.writeText(content);
			copied = true;
			setTimeout(() => { copied = false; }, 2000);
		} catch {
			// fall through
		}
	}
</script>

<div class="exec-prompt">
	<div class="prompt-header">
		<div class="header-left">
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
				<polyline points="4 17 10 11 4 5"/>
				<line x1="12" y1="19" x2="20" y2="19"/>
			</svg>
			<span>Execution Prompt<span class="accent">.</span></span>
		</div>
		{#if content}
			<button class="copy-btn" onclick={copyToClipboard}>
				{#if copied}
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--ctp-green)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<polyline points="20 6 9 17 4 12"/>
					</svg>
					<span style="color: var(--ctp-green);">Copied</span>
				{:else}
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
						<path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
					</svg>
					<span>Copy</span>
				{/if}
			</button>
		{/if}
	</div>

	<div class="prompt-body">
		{#if content}
			<pre class="prompt-text">{content}</pre>
		{:else}
			<p class="placeholder">Will be generated after discussion...</p>
		{/if}
	</div>
</div>

<style>
	.exec-prompt {
		display: flex;
		flex-direction: column;
		height: 100%;
	}

	.prompt-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 12px 16px;
		border-bottom: 1px solid var(--ctp-surface0);
		flex-shrink: 0;
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: 8px;
		font-family: var(--font-heading);
		font-size: 13px;
		font-weight: 600;
		color: var(--ctp-subtext0);
	}

	.accent {
		color: var(--ctp-mauve);
	}

	.copy-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 12px;
		border-radius: 12px;
		font-size: 12px;
		color: var(--ctp-subtext0);
		transition: all 0.15s ease;
		border: 1px solid var(--ctp-surface1);
		background: rgba(49, 50, 68, 0.4);
	}

	.copy-btn:hover {
		background: var(--ctp-surface1);
		color: var(--ctp-text);
	}

	.prompt-body {
		flex: 1;
		overflow-y: auto;
		padding: 16px;
	}

	.placeholder {
		color: var(--ctp-overlay1);
		font-style: italic;
		font-size: 13px;
	}

	.prompt-text {
		margin: 0;
		white-space: pre-wrap;
		word-break: break-word;
		font-size: 13px;
		line-height: 1.6;
		background: var(--ctp-mantle);
		border: 1px solid var(--ctp-surface0);
		border-radius: 12px;
		padding: 14px 16px;
	}
</style>
