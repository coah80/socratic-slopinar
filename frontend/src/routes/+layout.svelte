<script lang="ts">
	import '../app.css';
	import Settings from '$lib/components/Settings.svelte';

	let { children } = $props();
	let showSettings = $state(false);
</script>

<div class="app">
	<header class="topbar">
		<div class="brand">
			<span class="brand-text">socratic slopinar<span class="accent">.</span></span>
		</div>
		<button class="settings-btn" onclick={() => showSettings = true} aria-label="Settings">
			<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
				<circle cx="12" cy="12" r="3"/>
				<path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
			</svg>
		</button>
	</header>

	<main class="content">
		{@render children()}
	</main>
</div>

{#if showSettings}
	<Settings onclose={() => showSettings = false} />
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

	.settings-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border-radius: 12px;
		color: var(--ctp-subtext0);
		transition: all 0.15s ease;
	}

	.settings-btn:hover {
		background: rgba(49, 50, 68, 0.4);
		color: var(--ctp-text);
	}

	.content {
		flex: 1;
		overflow: hidden;
		margin-top: 56px;
	}
</style>
