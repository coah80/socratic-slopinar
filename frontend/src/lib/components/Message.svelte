<script lang="ts">
	import type { Message } from '$lib/types';

	let { message, color }: { message: Message; color: string } = $props();

	function formatContent(text: string): string {
		return text
			.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
			.replace(/`([^`]+)`/g, '<code>$1</code>')
			.replace(/\n/g, '<br>');
	}


</script>

<div class="message" style="border-left-color: {color};">
	<div class="message-header">
		<span class="model-dot" style="background: {color};"></span>
		<span class="model-name" style="color: {color};">{message.display_name}</span>
	</div>

	{#if message.content}
		<div class="message-content">
			{@html formatContent(message.content)}
		</div>
	{/if}

	{#each message.tool_calls as tool}
		<details class="tool-call">
			<summary class="tool-summary">
				<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
				</svg>
				{tool.name}
			</summary>
			<div class="tool-body">
				{#if tool.arguments}
					<div class="tool-section">
						<span class="tool-label">args</span>
						<pre>{tool.arguments}</pre>
					</div>
				{/if}
				{#if tool.result}
					<div class="tool-section">
						<span class="tool-label">result</span>
						<pre>{tool.result}</pre>
					</div>
				{/if}
			</div>
		</details>
	{/each}
</div>

<style>
	.message {
		padding: 14px 18px;
		border-radius: 16px;
		border: 1px solid rgba(69, 71, 90, 0.5);
		background: rgba(49, 50, 68, 0.4);
		border-left: 3px solid;
		margin-bottom: 8px;
	}

	.message-header {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 8px;
	}

	.model-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.model-name {
		font-family: var(--font-data);
		font-size: 13px;
		font-weight: 600;
	}

	.message-content {
		font-size: 14px;
		line-height: 1.65;
		color: var(--ctp-subtext1);
	}

	.message-content :global(strong) {
		color: var(--ctp-text);
		font-weight: 600;
	}

	.message-content :global(code) {
		background: var(--ctp-surface0);
		padding: 2px 6px;
		border-radius: 4px;
		font-family: var(--font-data);
		font-size: 0.9em;
	}

	.tool-call {
		margin-top: 10px;
		border: 1px solid var(--ctp-surface0);
		border-radius: 12px;
		overflow: hidden;
	}

	.tool-summary {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 12px;
		background: var(--ctp-mantle);
		font-family: var(--font-data);
		font-size: 12px;
		color: var(--ctp-subtext0);
		cursor: pointer;
		user-select: none;
	}

	.tool-summary:hover {
		background: var(--ctp-surface0);
	}

	.tool-body {
		padding: 10px 12px;
		background: var(--ctp-crust);
	}

	.tool-section {
		margin-bottom: 8px;
	}

	.tool-section:last-child {
		margin-bottom: 0;
	}

	.tool-label {
		font-family: var(--font-data);
		font-size: 11px;
		font-weight: 600;
		color: var(--ctp-overlay0);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		display: block;
		margin-bottom: 4px;
	}

	.tool-body pre {
		margin: 0;
		font-size: 12px;
		white-space: pre-wrap;
		word-break: break-word;
	}
</style>
