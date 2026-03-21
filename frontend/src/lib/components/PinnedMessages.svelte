<script lang="ts">
	import type { PinnedMessage } from '$lib/types';
	import { getModelColor } from '$lib/types';

	let { pins }: { pins: PinnedMessage[] } = $props();
</script>

<div class="pinned">
	<div class="pinned-header">
		<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
			<line x1="12" y1="17" x2="12" y2="22"/>
			<path d="M5 17h14v-1.76a2 2 0 0 0-1.11-1.79l-1.78-.9A2 2 0 0 1 15 10.76V6h1a2 2 0 0 0 0-4H8a2 2 0 0 0 0 4h1v4.76a2 2 0 0 1-1.11 1.79l-1.78.9A2 2 0 0 0 5 15.24z"/>
		</svg>
		<span>Pinned<span class="accent">.</span></span>
	</div>

	<div class="pinned-body">
		{#if pins.length === 0}
			<p class="placeholder">No pinned messages</p>
		{:else}
			{#each pins as pin, i (i)}
				<div class="pin-card" style="border-left-color: {getModelColor(pin.model_id)};">
					<span class="pin-model" style="color: {getModelColor(pin.model_id)};">{pin.display_name}</span>
					<p class="pin-content">{pin.content}</p>
				</div>
			{/each}
		{/if}
	</div>
</div>

<style>
	.pinned {
		display: flex;
		flex-direction: column;
		height: 100%;
	}

	.pinned-header {
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

	.pinned-body {
		flex: 1;
		overflow-y: auto;
		padding: 12px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.placeholder {
		color: var(--ctp-overlay1);
		font-style: italic;
		font-size: 13px;
		padding: 4px;
	}

	.pin-card {
		padding: 10px 14px;
		border-radius: 12px;
		border: 1px solid rgba(69, 71, 90, 0.5);
		background: rgba(49, 50, 68, 0.3);
		border-left: 3px solid;
	}

	.pin-model {
		font-family: var(--font-data);
		font-size: 11px;
		font-weight: 600;
		display: block;
		margin-bottom: 4px;
	}

	.pin-content {
		font-size: 13px;
		line-height: 1.5;
		color: var(--ctp-subtext1);
	}
</style>
