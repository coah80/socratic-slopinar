<script lang="ts">
	import type { Message } from '$lib/types';

	let { message, color, muted = false, onmute, onunmute }: {
		message: Message;
		color: string;
		muted?: boolean;
		onmute?: (modelId: string) => void;
		onunmute?: (modelId: string) => void;
	} = $props();

	const isGod = $derived(message.role === 'god');

	function formatContent(text: string): string {
		return text
			.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
			.replace(/`([^`]+)`/g, '<code>$1</code>')
			.replace(/\n/g, '<br>');
	}
</script>

<div
	class="message"
	class:muted
	class:god-message={isGod}
	style="border-left-color: {isGod ? 'var(--ctp-yellow)' : color};"
>
	<div class="message-header">
		<span class="model-dot" style="background: {isGod ? 'var(--ctp-yellow)' : color};"></span>
		<span class="model-name" style="color: {isGod ? 'var(--ctp-yellow)' : color};">{message.display_name}</span>
		{#if !isGod && (onmute || onunmute)}
			{#if muted}
				<button class="mute-btn" onclick={() => onunmute?.(message.model_id)} aria-label="Unmute {message.display_name}">
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/>
						<line x1="23" y1="9" x2="17" y2="15"/>
						<line x1="17" y1="9" x2="23" y2="15"/>
					</svg>
				</button>
			{:else}
				<button class="mute-btn" onclick={() => onmute?.(message.model_id)} aria-label="Mute {message.display_name}">
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/>
						<path d="M19.07 4.93a10 10 0 0 1 0 14.14M15.54 8.46a5 5 0 0 1 0 7.07"/>
					</svg>
				</button>
			{/if}
		{/if}
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

	.message.muted {
		opacity: 0.4;
	}

	.message.god-message {
		background: rgba(249, 226, 175, 0.06);
		border-color: rgba(249, 226, 175, 0.25);
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

	.mute-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		border-radius: 6px;
		color: var(--ctp-overlay0);
		margin-left: auto;
		transition: all 0.15s ease;
	}

	.mute-btn:hover {
		background: rgba(49, 50, 68, 0.6);
		color: var(--ctp-text);
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
