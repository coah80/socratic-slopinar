<script lang="ts">
	import { getProviderInfo } from '$lib/types';

	let { provider, size = 16 }: { provider: string; size?: number } = $props();

	const info = $derived(getProviderInfo(provider));

	const LOBEHUB_CDN = 'https://unpkg.com/@lobehub/icons-static-svg@1.82.0/icons';

	const iconMap: Record<string, string> = {
		openrouter: 'openrouter',
		openai: 'openai',
		anthropic: 'anthropic',
		google: 'google',
		xai: 'xai',
		deepseek: 'deepseek',
		mistral: 'mistral',
		groq: 'groq',
		together: 'together',
		minimax: 'minimax',
		cohere: 'cohere',
		perplexity: 'perplexity',
		fireworks: 'fireworks',
		cerebras: 'cerebras',
		nvidia: 'nvidia',
		ai21: 'ai21',
		sambanova: 'sambanova',
		moonshot: 'moonshot',
		lambda: 'lambda',
		novita: 'novita',
	};

	const slug = $derived(iconMap[provider] ?? provider);
	const src = $derived(`${LOBEHUB_CDN}/${slug}.svg`);
	let failed = $state(false);
</script>

{#if failed}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<circle cx="12" cy="12" r="8" stroke={info.color} stroke-width="2" />
	</svg>
{:else}
	<img
		{src}
		alt={info.name}
		width={size}
		height={size}
		onerror={() => failed = true}
		style="display:inline-block;vertical-align:middle;filter:drop-shadow(0 0 0 transparent);"
	/>
{/if}
