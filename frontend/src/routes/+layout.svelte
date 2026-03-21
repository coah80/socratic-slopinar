<script lang="ts">
	import '../app.css';
	import Settings from '$lib/components/Settings.svelte';
	import History from '$lib/components/History.svelte';

	let { children } = $props();
	let showSettings = $state(false);
	let showHistory = $state(false);
</script>

<div class="app">
	<header class="topbar">
		<div class="brand">
			<span class="brand-text">socratic slopinar<span class="accent">.</span></span>
		</div>
		<div class="topbar-actions">
			<button class="topbar-btn" onclick={() => showHistory = true} aria-label="History">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<circle cx="12" cy="12" r="10"/>
					<polyline points="12 6 12 12 16 14"/>
				</svg>
			</button>
			<button class="topbar-btn" onclick={() => showSettings = true} aria-label="Settings">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<circle cx="12" cy="12" r="3"/>
					<path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
				</svg>
			</button>
		</div>
	</header>

	<main class="content">
		{@render children()}
	</main>
</div>

{#if showSettings}
	<Settings onclose={() => showSettings = false} />
{/if}

{#if showHistory}
	<History onclose={() => showHistory = false} />
{/if}

<style>
	.app {
		display: flex;
		flex-direction: column;
		height: 100vh;
		background: var(--ctp-crust);
	}

	.topbar {
		position: fixed;
		top: 0;
		width: 100%;
		z-index: 50;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 24px;
		height: 56px;
		background: rgba(30, 30, 46, 0.9);
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		border-bottom: 1px solid rgba(205, 214, 244, 0.05);
		flex-shrink: 0;
	}

	.brand-text {
		font-family: var(--font-heading);
		font-size: 17px;
		font-weight: 700;
		letter-spacing: -0.3px;
		color: var(--ctp-text);
	}

	.accent {
		color: var(--ctp-mauve);
	}

	.topbar-actions {
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.topbar-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border-radius: 12px;
		color: var(--ctp-subtext0);
		transition: all 0.15s ease;
	}

	.topbar-btn:hover {
		background: rgba(49, 50, 68, 0.4);
		color: var(--ctp-text);
	}

	.content {
		flex: 1;
		overflow: hidden;
		margin-top: 56px;
	}
</style>
