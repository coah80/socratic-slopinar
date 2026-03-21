<script lang="ts">
	import type { Message as MessageType } from '$lib/types';
	import { getModelColor } from '$lib/types';
	import MessageComponent from './Message.svelte';

	let { messages, activeModel, activeDisplayName, mutedModels = new Set(), onmute, onunmute }: {
		messages: MessageType[];
		activeModel: string | null;
		activeDisplayName: string | null;
		mutedModels?: Set<string>;
		onmute?: (modelId: string) => void;
		onunmute?: (modelId: string) => void;
	} = $props();

	let scrollContainer: HTMLDivElement | undefined = $state();

	$effect(() => {
		if (messages.length && scrollContainer) {
			requestAnimationFrame(() => {
				scrollContainer!.scrollTop = scrollContainer!.scrollHeight;
			});
		}
	});
</script>

<div class="chat" bind:this={scrollContainer}>
	{#if messages.length === 0}
		<div class="empty">
			<div class="empty-icon">
				<svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
					<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
				</svg>
			</div>
			<p class="empty-heading">No discussion yet<span class="accent">.</span></p>
			<p class="empty-sub">Enter a prompt below to start</p>
		</div>
	{:else}
		{#each messages as msg, i (msg.id)}
			<div class="msg-enter" style="animation-delay: {Math.min(i * 30, 200)}ms;">
				<MessageComponent
					message={msg}
					color={getModelColor(msg.model_id)}
					muted={mutedModels.has(msg.model_id)}
					{onmute}
					{onunmute}
				/>
			</div>
		{/each}
	{/if}

	{#if activeModel}
		<div class="typing" style="border-left-color: {getModelColor(activeModel)};">
			<span class="model-dot" style="background: {getModelColor(activeModel)};"></span>
			<span class="typing-name" style="color: {getModelColor(activeModel)};">{activeDisplayName ?? activeModel}</span>
			<span class="typing-dots">
				<span class="dot"></span>
				<span class="dot"></span>
				<span class="dot"></span>
			</span>
		</div>
	{/if}
</div>

<style>
	.chat {
		flex: 1;
		overflow-y: auto;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 0;
	}

	.empty {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 6px;
		color: var(--ctp-overlay0);
		animation: fade-up 0.5s ease both;
	}

	.empty-icon {
		opacity: 0.3;
		margin-bottom: 8px;
	}

	.empty-heading {
		font-family: var(--font-heading);
		font-size: 16px;
		font-weight: 600;
		color: var(--ctp-subtext0);
	}

	.accent {
		color: var(--ctp-mauve);
	}

	.empty-sub {
		font-size: 13px;
		color: var(--ctp-overlay1);
	}

	.msg-enter {
		animation: fade-up 0.3s ease both;
	}

	.typing {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 18px;
		border-radius: 16px;
		border: 1px solid rgba(69, 71, 90, 0.5);
		background: rgba(49, 50, 68, 0.4);
		border-left: 3px solid;
		margin-top: 8px;
	}

	.model-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.typing-name {
		font-family: var(--font-data);
		font-size: 13px;
		font-weight: 600;
	}

	.typing-dots {
		display: flex;
		gap: 4px;
		margin-left: 4px;
	}

	.dot {
		width: 5px;
		height: 5px;
		border-radius: 50%;
		background: var(--ctp-overlay0);
		animation: bounce 1.4s infinite ease-in-out both;
	}

	.dot:nth-child(1) { animation-delay: -0.32s; }
	.dot:nth-child(2) { animation-delay: -0.16s; }

	@keyframes bounce {
		0%, 80%, 100% { transform: scale(0); }
		40% { transform: scale(1); }
	}
</style>
