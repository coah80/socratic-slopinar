<script lang="ts">
	let { content }: { content: string } = $props();

	function renderMarkdown(text: string): string {
		return text
			.replace(/^### (.+)$/gm, '<h3>$1</h3>')
			.replace(/^## (.+)$/gm, '<h2>$1</h2>')
			.replace(/^# (.+)$/gm, '<h1>$1</h1>')
			.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
			.replace(/\*(.*?)\*/g, '<em>$1</em>')
			.replace(/`([^`]+)`/g, '<code>$1</code>')
			.replace(/^- (.+)$/gm, '<li>$1</li>')
			.replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>')
			.replace(/\n\n/g, '</p><p>')
			.replace(/\n/g, '<br>');
	}
</script>

<div class="shared-doc">
	<div class="doc-header">
		<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
			<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
			<polyline points="14 2 14 8 20 8"/>
			<line x1="16" y1="13" x2="8" y2="13"/>
			<line x1="16" y1="17" x2="8" y2="17"/>
			<polyline points="10 9 9 9 8 9"/>
		</svg>
		<span>Shared Notes<span class="accent">.</span></span>
	</div>

	<div class="doc-body">
		{#if content}
			<div class="markdown">
				{@html renderMarkdown(content)}
			</div>
		{:else}
			<p class="placeholder">No notes yet...</p>
		{/if}
	</div>
</div>

<style>
	.shared-doc {
		display: flex;
		flex-direction: column;
		height: 100%;
	}

	.doc-header {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 16px;
		border-bottom: 1px solid var(--ctp-surface0);
		font-family: var(--font-heading);
		font-size: 13px;
		font-weight: 600;
		color: var(--ctp-subtext0);
		flex-shrink: 0;
	}

	.accent {
		color: var(--ctp-mauve);
	}

	.doc-body {
		flex: 1;
		overflow-y: auto;
		padding: 16px;
	}

	.placeholder {
		color: var(--ctp-overlay1);
		font-style: italic;
		font-size: 13px;
	}

	.markdown {
		font-size: 13px;
		line-height: 1.65;
		color: var(--ctp-subtext1);
	}

	.markdown :global(h1) {
		font-family: var(--font-heading);
		font-size: 18px;
		font-weight: 700;
		color: var(--ctp-text);
		margin: 0 0 8px;
	}

	.markdown :global(h2) {
		font-family: var(--font-heading);
		font-size: 15px;
		font-weight: 600;
		color: var(--ctp-text);
		margin: 12px 0 6px;
	}

	.markdown :global(h3) {
		font-family: var(--font-heading);
		font-size: 14px;
		font-weight: 600;
		color: var(--ctp-subtext1);
		margin: 10px 0 4px;
	}

	.markdown :global(strong) {
		color: var(--ctp-text);
	}

	.markdown :global(code) {
		background: var(--ctp-surface0);
		padding: 2px 6px;
		border-radius: 4px;
		font-family: var(--font-data);
		font-size: 0.9em;
	}

	.markdown :global(ul) {
		padding-left: 20px;
		margin: 6px 0;
	}

	.markdown :global(li) {
		margin: 2px 0;
	}
</style>
