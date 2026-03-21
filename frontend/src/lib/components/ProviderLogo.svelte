<script lang="ts">
	import { getProviderInfo } from '$lib/types';

	let { provider, size = 16 }: { provider: string; size?: number } = $props();

	const info = $derived(getProviderInfo(provider));
</script>

{#if provider === 'openrouter'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<circle cx="12" cy="12" r="4" fill={info.color} />
		<path d="M12 2v4M12 18v4M2 12h4M18 12h4M5.64 5.64l2.83 2.83M15.54 15.54l2.83 2.83M5.64 18.36l2.83-2.83M15.54 8.46l2.83-2.83" stroke={info.color} stroke-width="2" stroke-linecap="round" />
	</svg>
{:else if provider === 'openai'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<path d="M22.28 9.37a6.04 6.04 0 0 0-.52-4.96 6.1 6.1 0 0 0-6.58-2.93A6.04 6.04 0 0 0 10.63.06a6.1 6.1 0 0 0-5.82 4.24 6.04 6.04 0 0 0-4.04 2.93 6.1 6.1 0 0 0 .75 7.14 6.04 6.04 0 0 0 .52 4.96 6.1 6.1 0 0 0 6.58 2.93A6.04 6.04 0 0 0 13.17 23.68a6.1 6.1 0 0 0 5.82-4.24 6.04 6.04 0 0 0 4.04-2.93 6.1 6.1 0 0 0-.75-7.14z" fill={info.color} opacity="0.9" />
	</svg>
{:else if provider === 'anthropic'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<path d="M13.83 3L21 21h-4.2l-1.5-3.82H9.7L8.2 21H4L11.17 3h2.66z" fill={info.color} />
	</svg>
{:else if provider === 'google'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285f4" />
		<path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84A10.997 10.997 0 0 0 12 23z" fill="#34a853" />
		<path d="M5.84 14.1a6.6 6.6 0 0 1 0-4.2V7.06H2.18A10.997 10.997 0 0 0 1 12c0 1.77.42 3.45 1.18 4.94l3.66-2.84z" fill="#fbbc05" />
		<path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z" fill="#ea4335" />
	</svg>
{:else if provider === 'xai'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<path d="M4 4l8 8m0 0l8 8m-8-8l8-8m-8 8l-8 8" stroke={info.color} stroke-width="3" stroke-linecap="round" stroke-linejoin="round" />
	</svg>
{:else if provider === 'deepseek'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<text x="2" y="18" font-family="Arial, sans-serif" font-weight="900" font-size="14" fill={info.color}>DS</text>
	</svg>
{:else if provider === 'mistral'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<path d="M12 2C8 2 6 6 6 10c0 3 1 5 3 7l3 5 3-5c2-2 3-4 3-7 0-4-2-8-6-8z" fill={info.color} />
	</svg>
{:else if provider === 'groq'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<circle cx="12" cy="12" r="9" stroke={info.color} stroke-width="3" />
		<circle cx="12" cy="12" r="4" fill={info.color} />
	</svg>
{:else if provider === 'together'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<circle cx="9" cy="12" r="6" stroke={info.color} stroke-width="2.5" fill="none" />
		<circle cx="15" cy="12" r="6" stroke={info.color} stroke-width="2.5" fill="none" />
	</svg>
{:else if provider === 'minimax'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<text x="1" y="18" font-family="Arial, sans-serif" font-weight="900" font-size="13" fill={info.color}>MM</text>
	</svg>
{:else if provider === 'cohere'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<path d="M12 4a8 8 0 0 1 8 8c0 2.5-1.2 4.7-3 6.2L12 22l-5-3.8A8 8 0 0 1 12 4z" fill={info.color} />
	</svg>
{:else if provider === 'perplexity'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<path d="M12 2L4 8v8l8 6 8-6V8l-8-6z" stroke={info.color} stroke-width="2.5" fill="none" />
		<circle cx="12" cy="12" r="3" fill={info.color} />
	</svg>
{:else if provider === 'fireworks'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<path d="M12 2l2 6h6l-5 4 2 6-5-4-5 4 2-6-5-4h6z" fill={info.color} />
	</svg>
{:else if provider === 'cerebras'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<rect x="4" y="4" width="7" height="7" rx="1" fill={info.color} />
		<rect x="13" y="4" width="7" height="7" rx="1" fill={info.color} opacity="0.7" />
		<rect x="4" y="13" width="7" height="7" rx="1" fill={info.color} opacity="0.7" />
		<rect x="13" y="13" width="7" height="7" rx="1" fill={info.color} opacity="0.4" />
	</svg>
{:else if provider === 'nvidia'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<path d="M5 9c1-3 4-5 7-5v3c-2 0-4 1-5 3l-2-1zm7-5v3c3 0 5 2 5 5s-2 5-5 5v3c5 0 8-4 8-8s-3-8-8-8z" fill={info.color} />
	</svg>
{:else if provider === 'ai21'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<text x="2" y="18" font-family="Arial, sans-serif" font-weight="900" font-size="13" fill={info.color}>21</text>
	</svg>
{:else if provider === 'sambanova'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<text x="3" y="18" font-family="Arial, sans-serif" font-weight="900" font-size="12" fill={info.color}>SN</text>
	</svg>
{:else if provider === 'moonshot'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<path d="M12 2a10 10 0 0 0 0 20 7 7 0 0 1 0-20z" fill={info.color} />
	</svg>
{:else if provider === 'lambda'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<path d="M6 20L12 4l6 16h-3.5l-2.5-8-2.5 8H6z" fill={info.color} />
	</svg>
{:else if provider === 'novita'}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<text x="4" y="18" font-family="Arial, sans-serif" font-weight="900" font-size="13" fill={info.color}>N</text>
	</svg>
{:else}
	<svg width={size} height={size} viewBox="0 0 24 24" fill="none">
		<circle cx="12" cy="12" r="8" stroke={info.color} stroke-width="2" />
	</svg>
{/if}
