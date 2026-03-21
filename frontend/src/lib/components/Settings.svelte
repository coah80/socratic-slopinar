<script lang="ts">
	import { getConfig, loadConfig, addProviderKey, removeProviderKey, addModel, removeModel } from '$lib/stores/settings.svelte';
	import { PROVIDERS } from '$lib/types';
	import ProviderLogo from './ProviderLogo.svelte';

	let { onclose }: { onclose: () => void } = $props();

	let newModelId = $state('');
	let addingKey = $state(false);
	let selectedProvider = $state('');
	let newKeyValue = $state('');
	let saving = $state(false);

	const config = $derived(getConfig());

	const configuredProviders = $derived(
		PROVIDERS.filter(p => config.provider_keys[p.id])
	);

	const hasTavilyKey = $derived(config.tavily_api_key.length > 0);

	const availableProviders = $derived(
		PROVIDERS.filter(p => !config.provider_keys[p.id])
	);

	const isTavilyAvailable = $derived(!hasTavilyKey);

	$effect(() => {
		loadConfig();
	});

	function maskKey(key: string): string {
		if (key.length <= 8) return '***';
		return key.slice(0, 5) + '...' + key.slice(-3);
	}

	async function handleAddKey() {
		if (!selectedProvider || !newKeyValue.trim()) return;
		saving = true;
		if (selectedProvider === '_tavily') {
			await addProviderKey('_tavily', newKeyValue.trim());
		} else {
			await addProviderKey(selectedProvider, newKeyValue.trim());
		}
		selectedProvider = '';
		newKeyValue = '';
		addingKey = false;
		saving = false;
	}

	async function handleRemoveKey(providerId: string) {
		await removeProviderKey(providerId);
	}

	async function handleRemoveTavily() {
		await removeProviderKey('_tavily');
	}

	function handleKeyInputKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') handleAddKey();
		if (e.key === 'Escape') {
			addingKey = false;
			selectedProvider = '';
			newKeyValue = '';
		}
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
				<label>API Keys</label>
				<div class="keys-list">
					{#each configuredProviders as provider (provider.id)}
						<div class="key-row">
							<div class="key-row-left">
								<ProviderLogo provider={provider.id} size={18} />
								<span class="key-provider-name">{provider.name}</span>
								<span class="key-masked">{maskKey(config.provider_keys[provider.id])}</span>
							</div>
							<button class="key-remove" onclick={() => handleRemoveKey(provider.id)} aria-label="Remove {provider.name} key">
								<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
									<line x1="18" y1="6" x2="6" y2="18"/>
									<line x1="6" y1="6" x2="18" y2="18"/>
								</svg>
							</button>
						</div>
					{/each}

					{#if hasTavilyKey}
						<div class="key-row">
							<div class="key-row-left">
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="var(--ctp-teal)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
									<circle cx="11" cy="11" r="8"/>
									<line x1="21" y1="21" x2="16.65" y2="16.65"/>
								</svg>
								<span class="key-provider-name">Tavily</span>
								<span class="key-masked">{maskKey(config.tavily_api_key)}</span>
							</div>
							<button class="key-remove" onclick={handleRemoveTavily} aria-label="Remove Tavily key">
								<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
									<line x1="18" y1="6" x2="6" y2="18"/>
									<line x1="6" y1="6" x2="18" y2="18"/>
								</svg>
							</button>
						</div>
					{/if}

					{#if configuredProviders.length === 0 && !hasTavilyKey}
						<p class="no-keys">No API keys configured</p>
					{/if}
				</div>

				{#if addingKey}
					<div class="add-key-form">
						<select class="input select-input" bind:value={selectedProvider}>
							<option value="" disabled>Select provider...</option>
							{#each availableProviders as provider (provider.id)}
								<option value={provider.id}>{provider.name}</option>
							{/each}
							{#if isTavilyAvailable}
								<option value="_tavily">Tavily (Search)</option>
							{/if}
						</select>
						{#if selectedProvider}
							<div class="add-key-input-row">
								<input
									type="password"
									bind:value={newKeyValue}
									placeholder={selectedProvider === '_tavily' ? 'tvly-...' : 'sk-...'}
									class="input"
									onkeydown={handleKeyInputKeydown}
								/>
								<button class="save-key-btn" onclick={handleAddKey} disabled={saving || !newKeyValue.trim()}>
									{saving ? '...' : 'Save'}
								</button>
								<button class="cancel-key-btn" onclick={() => { addingKey = false; selectedProvider = ''; newKeyValue = ''; }}>
									Cancel
								</button>
							</div>
						{/if}
					</div>
				{:else}
					<button class="add-key-btn" onclick={() => { addingKey = true; }}>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
							<line x1="12" y1="5" x2="12" y2="19"/>
							<line x1="5" y1="12" x2="19" y2="12"/>
						</svg>
						Add API Key
					</button>
				{/if}
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
			<button class="close-footer-btn" onclick={onclose}>Close</button>
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

	.keys-list {
		display: flex;
		flex-direction: column;
		gap: 4px;
		margin-bottom: 10px;
	}

	.key-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 14px;
		border-radius: 10px;
		background: rgba(49, 50, 68, 0.25);
		border: 1px solid rgba(69, 71, 90, 0.3);
	}

	.key-row-left {
		display: flex;
		align-items: center;
		gap: 10px;
		min-width: 0;
	}

	.key-provider-name {
		font-family: var(--font-data);
		font-size: 13px;
		font-weight: 600;
		color: var(--ctp-subtext1);
		white-space: nowrap;
	}

	.key-masked {
		font-family: var(--font-data);
		font-size: 12px;
		color: var(--ctp-overlay1);
		white-space: nowrap;
	}

	.key-remove {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border-radius: 8px;
		color: var(--ctp-overlay1);
		flex-shrink: 0;
		transition: all 0.15s ease;
	}

	.key-remove:hover {
		background: rgba(243, 139, 168, 0.15);
		color: var(--ctp-red);
	}

	.no-keys {
		color: var(--ctp-overlay1);
		font-size: 13px;
		font-style: italic;
		padding: 8px 0;
	}

	.add-key-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		border: 1px dashed var(--ctp-surface2);
		background: transparent;
		border-radius: 10px;
		font-size: 13px;
		font-weight: 500;
		color: var(--ctp-subtext0);
		transition: all 0.15s ease;
		width: 100%;
		justify-content: center;
	}

	.add-key-btn:hover {
		border-color: var(--ctp-mauve);
		color: var(--ctp-mauve);
		background: rgba(203, 166, 247, 0.05);
	}

	.add-key-form {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.select-input {
		appearance: none;
		-webkit-appearance: none;
		cursor: pointer;
		background-image: url("data:image/svg+xml,%3Csvg width='10' height='6' viewBox='0 0 10 6' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1 1L5 5L9 1' stroke='%236c7086' stroke-width='1.5' stroke-linecap='round' stroke-linejoin='round'/%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 12px center;
		padding-right: 32px;
	}

	.select-input option {
		background: var(--ctp-surface0);
		color: var(--ctp-text);
	}

	.add-key-input-row {
		display: flex;
		gap: 8px;
	}

	.add-key-input-row .input {
		flex: 1;
	}

	.save-key-btn {
		padding: 10px 16px;
		background: var(--ctp-mauve);
		border-radius: 10px;
		font-size: 13px;
		font-weight: 700;
		color: var(--ctp-crust);
		transition: filter 0.15s ease;
		white-space: nowrap;
	}

	.save-key-btn:hover {
		filter: brightness(1.1);
	}

	.save-key-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.cancel-key-btn {
		padding: 10px 14px;
		border: 1px solid var(--ctp-surface1);
		background: transparent;
		border-radius: 10px;
		font-size: 13px;
		color: var(--ctp-subtext0);
		transition: all 0.15s ease;
		white-space: nowrap;
	}

	.cancel-key-btn:hover {
		background: var(--ctp-surface0);
		color: var(--ctp-text);
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

	.close-footer-btn {
		padding: 10px 24px;
		border: 1px solid var(--ctp-surface1);
		background: rgba(49, 50, 68, 0.4);
		border-radius: 12px;
		font-size: 14px;
		color: var(--ctp-subtext1);
		transition: all 0.15s ease;
	}

	.close-footer-btn:hover {
		background: var(--ctp-surface1);
		color: var(--ctp-text);
	}
</style>
