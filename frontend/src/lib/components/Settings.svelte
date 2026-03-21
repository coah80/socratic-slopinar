<script lang="ts">
	import { getConfig, loadConfig, saveConfig, addModel, removeModel } from '$lib/stores/settings.svelte';
	import { PROVIDERS } from '$lib/types';
	import ProviderLogo from './ProviderLogo.svelte';

	let { onclose }: { onclose: () => void } = $props();

	let providerKeys = $state<Record<string, string>>({});
	let tavilyKey = $state('');
	let newModelId = $state('');
	let visibleKeys = $state<Set<string>>(new Set());
	let showTavilyKey = $state(false);
	let saving = $state(false);

	const config = $derived(getConfig());

	$effect(() => {
		loadConfig().then(() => {
			const c = getConfig();
			providerKeys = { ...c.provider_keys };
			tavilyKey = c.tavily_api_key;
		});
	});

	function toggleKeyVisibility(providerId: string) {
		const next = new Set(visibleKeys);
		if (next.has(providerId)) {
			next.delete(providerId);
		} else {
			next.add(providerId);
		}
		visibleKeys = next;
	}

	function updateProviderKey(providerId: string, value: string) {
		providerKeys = { ...providerKeys, [providerId]: value };
	}

	async function handleSave() {
		saving = true;
		await saveConfig({
			api_key: providerKeys.openrouter ?? '',
			models: config.models,
			tavily_api_key: tavilyKey,
			provider_keys: providerKeys,
		});
		saving = false;
		onclose();
	}

	async function handleAddModel() {
		if (!newModelId.trim()) return;
		await addModel(newModelId);
		newModelId = '';
	}

	function handleModelKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') handleAddModel();
	}

	function handleOverlayClick(e: MouseEvent) {
		if (e.target === e.currentTarget) onclose();
	}
</script>

<!-- svelte-ignore a11y_interactive_supports_focus -->
<div class="overlay" role="dialog" aria-modal="true" onclick={handleOverlayClick} onkeydown={(e) => { if (e.key === 'Escape') onclose(); }}>
	<div class="modal">
		<div class="modal-header">
			<h2>Settings<span class="accent">.</span></h2>
			<button class="close-btn" onclick={onclose} aria-label="Close">
				<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<line x1="18" y1="6" x2="6" y2="18"/>
					<line x1="6" y1="6" x2="18" y2="18"/>
				</svg>
			</button>
		</div>

		<div class="modal-body">
			<div class="field">
				<!-- svelte-ignore a11y_label_has_associated_control -->
				<label>API Providers</label>
				<div class="providers-list">
					{#each PROVIDERS as provider (provider.id)}
						{@const keyValue = providerKeys[provider.id] ?? ''}
						{@const isVisible = visibleKeys.has(provider.id)}
						<div class="provider-row">
							<div class="provider-info">
								<ProviderLogo provider={provider.id} size={18} />
								<span class="provider-name">{provider.name}</span>
								<span class="key-indicator" class:active={keyValue.length > 0}></span>
							</div>
							<div class="key-input-wrapper">
								<input
									type={isVisible ? 'text' : 'password'}
									value={keyValue}
									oninput={(e) => updateProviderKey(provider.id, e.currentTarget.value)}
									placeholder={provider.id === 'openrouter' ? 'sk-or-...' : 'sk-...'}
									class="input"
								/>
								<button class="toggle-vis" onclick={() => toggleKeyVisibility(provider.id)} aria-label={isVisible ? 'Hide' : 'Show'}>
									{#if isVisible}
										<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
											<path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
											<line x1="1" y1="1" x2="23" y2="23"/>
										</svg>
									{:else}
										<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
											<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
											<circle cx="12" cy="12" r="3"/>
										</svg>
									{/if}
								</button>
							</div>
						</div>
					{/each}
				</div>
			</div>

			<div class="field">
				<label for="tavily-key">API Key (Tavily)</label>
				<div class="key-input-wrapper">
					<input
						id="tavily-key"
						type={showTavilyKey ? 'text' : 'password'}
						bind:value={tavilyKey}
						placeholder="tvly-..."
						class="input"
					/>
					<button class="toggle-vis" onclick={() => showTavilyKey = !showTavilyKey} aria-label={showTavilyKey ? 'Hide' : 'Show'}>
						{#if showTavilyKey}
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
								<line x1="1" y1="1" x2="23" y2="23"/>
							</svg>
						{:else}
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
								<circle cx="12" cy="12" r="3"/>
							</svg>
						{/if}
					</button>
				</div>
			</div>

			<div class="field">
				<!-- svelte-ignore a11y_label_has_associated_control -->
				<label>Models</label>
				<div class="models-list">
					{#each config.models as model}
						<div class="model-pill">
							<span class="model-pill-text">{model}</span>
							<button class="model-remove" onclick={() => removeModel(model)} aria-label="Remove {model}">
								<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
									<line x1="18" y1="6" x2="6" y2="18"/>
									<line x1="6" y1="6" x2="18" y2="18"/>
								</svg>
							</button>
						</div>
					{/each}
					{#if config.models.length === 0}
						<p class="no-models">No models added yet</p>
					{/if}
				</div>
				<div class="add-model">
					<input
						type="text"
						bind:value={newModelId}
						placeholder="anthropic/claude-sonnet-4-6"
						class="input"
						onkeydown={handleModelKeydown}
					/>
					<button class="add-btn" onclick={handleAddModel}>Add</button>
				</div>
			</div>
		</div>

		<div class="modal-footer">
			<button class="cancel-btn" onclick={onclose}>Cancel</button>
			<button class="save-btn" onclick={handleSave} disabled={saving}>
				{saving ? 'Saving...' : 'Save'}
			</button>
		</div>
	</div>
</div>

<style>
	.overlay {
		position: fixed;
		inset: 0;
		z-index: 50;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(17, 17, 27, 0.6);
		backdrop-filter: blur(4px);
		-webkit-backdrop-filter: blur(4px);
	}

	.modal {
		width: 520px;
		max-width: 90vw;
		max-height: 85vh;
		display: flex;
		flex-direction: column;
		background: var(--ctp-base);
		border: 1px solid var(--ctp-surface1);
		border-radius: 16px;
		overflow: hidden;
	}

	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 20px 24px;
		border-bottom: 1px solid var(--ctp-surface0);
	}

	.modal-header h2 {
		font-family: var(--font-heading);
		font-size: 18px;
		font-weight: 700;
		color: var(--ctp-text);
	}

	.accent {
		color: var(--ctp-mauve);
	}

	.close-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 10px;
		color: var(--ctp-subtext0);
		transition: all 0.15s ease;
	}

	.close-btn:hover {
		background: rgba(49, 50, 68, 0.4);
		color: var(--ctp-text);
	}

	.modal-body {
		flex: 1;
		overflow-y: auto;
		padding: 24px;
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.field label {
		display: block;
		font-family: var(--font-heading);
		font-size: 13px;
		font-weight: 600;
		color: var(--ctp-subtext0);
		margin-bottom: 8px;
	}

	.providers-list {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.provider-row {
		padding: 10px 14px;
		border-radius: 12px;
		border: 1px solid rgba(69, 71, 90, 0.4);
		background: rgba(49, 50, 68, 0.25);
		transition: border-color 0.15s ease;
	}

	.provider-row:focus-within {
		border-color: rgba(69, 71, 90, 0.8);
	}

	.provider-info {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 8px;
	}

	.provider-name {
		font-family: var(--font-data);
		font-size: 13px;
		font-weight: 600;
		color: var(--ctp-subtext1);
	}

	.key-indicator {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: var(--ctp-surface2);
		margin-left: auto;
		transition: background 0.15s ease;
	}

	.key-indicator.active {
		background: var(--ctp-green);
	}

	.input {
		width: 100%;
		padding: 10px 14px;
		background: rgba(49, 50, 68, 0.4);
		border: 1px solid var(--ctp-surface1);
		border-radius: 10px;
		color: var(--ctp-text);
		font-size: 13px;
		font-family: var(--font-body);
		transition: border-color 0.15s ease;
		outline: none;
	}

	.input:focus {
		border-color: rgba(203, 166, 247, 0.5);
	}

	.input::placeholder {
		color: var(--ctp-overlay1);
	}

	.key-input-wrapper {
		position: relative;
	}

	.key-input-wrapper .input {
		padding-right: 40px;
	}

	.toggle-vis {
		position: absolute;
		right: 6px;
		top: 50%;
		transform: translateY(-50%);
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border-radius: 8px;
		color: var(--ctp-overlay1);
		transition: color 0.15s ease;
	}

	.toggle-vis:hover {
		color: var(--ctp-text);
	}

	.models-list {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		margin-bottom: 10px;
		min-height: 32px;
	}

	.model-pill {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 12px;
		background: rgba(203, 166, 247, 0.1);
		border-radius: 9999px;
		font-family: var(--font-data);
		font-size: 12px;
	}

	.model-pill-text {
		color: var(--ctp-subtext1);
	}

	.model-remove {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 18px;
		height: 18px;
		border-radius: 50%;
		color: var(--ctp-overlay1);
		transition: all 0.15s ease;
	}

	.model-remove:hover {
		background: var(--ctp-surface1);
		color: var(--ctp-red);
	}

	.no-models {
		color: var(--ctp-overlay1);
		font-size: 13px;
		font-style: italic;
	}

	.add-model {
		display: flex;
		gap: 8px;
	}

	.add-model .input {
		flex: 1;
	}

	.add-btn {
		padding: 10px 16px;
		border: 1px solid var(--ctp-surface1);
		background: rgba(49, 50, 68, 0.4);
		border-radius: 10px;
		font-size: 14px;
		font-weight: 500;
		color: var(--ctp-subtext1);
		transition: all 0.15s ease;
	}

	.add-btn:hover {
		background: var(--ctp-surface1);
		color: var(--ctp-text);
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: 10px;
		padding: 16px 24px;
		border-top: 1px solid var(--ctp-surface0);
	}

	.cancel-btn {
		padding: 10px 16px;
		border: 1px solid var(--ctp-surface1);
		background: rgba(49, 50, 68, 0.4);
		border-radius: 12px;
		font-size: 14px;
		color: var(--ctp-subtext1);
		transition: all 0.15s ease;
	}

	.cancel-btn:hover {
		background: var(--ctp-surface1);
		color: var(--ctp-text);
	}

	.save-btn {
		padding: 10px 24px;
		background: var(--ctp-mauve);
		border-radius: 12px;
		font-size: 14px;
		font-weight: 700;
		color: var(--ctp-crust);
		transition: filter 0.15s ease;
	}

	.save-btn:hover {
		filter: brightness(1.1);
	}

	.save-btn:active {
		filter: brightness(0.9);
	}

	.save-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style>
