import { browser } from '$app/environment';
import { aiState } from './state.svelte';
import * as chrome from './providers/chrome';
import * as local from './providers/local';

export { aiState } from './state.svelte';

type Provider = 'chrome' | 'local';

let provider: Provider | null = null;
let initializationPromise: Promise<void> | null = null;
let enablePromise: Promise<void> | null = null;

async function selectProvider(): Promise<Provider | null> {
	if (chrome.isSupported()) {
		try {
			const availability = await chrome.availability();

			if (availability !== 'unavailable') {
				return 'chrome';
			}
		} catch {
			// Fall through to the local provider.
		}
	}

	if (local.isSupported()) {
		return 'local';
	}

	return null;
}

export async function initialize(): Promise<void> {
	if (!browser || aiState.status !== 'disabled') {
		return;
	}

	if (initializationPromise) {
		return initializationPromise;
	}

	initializationPromise = (async () => {
		aiState.status = 'checking';

		try {
			provider = await selectProvider();

			if (!provider) {
				aiState.status = 'unavailable';
				return;
			}

			aiState.status = 'available';
		} catch {
			provider = null;
			aiState.status = 'error';
		} finally {
			initializationPromise = null;
		}
	})();

	return initializationPromise;
}

async function loadProvider(): Promise<void> {
	if (!provider) {
		throw new Error('No AI provider selected.');
	}

	aiState.status = 'loading';
	aiState.progress = 0;

	if (provider === 'chrome') {
		try {
			const availability = await chrome.availability();

			if (availability === 'unavailable') {
				throw new Error('Chrome AI is unavailable.');
			}

			await chrome.create((progress) => {
				aiState.status = 'downloading';
				aiState.progress = progress;
			});

			return;
		} catch {
			// Chrome AI was selected but could not be initialized.
			// Fall back to the local browser model.
			provider = 'local';
		}
	}

	await local.load((progress) => {
		aiState.status = 'downloading';
		aiState.progress = progress;
	});
}

export async function enable(): Promise<void> {
	if (!browser || aiState.enabled) {
		return;
	}

	if (enablePromise) {
		return enablePromise;
	}

	enablePromise = (async () => {
		try {
			if (!provider) {
				await initialize();
			}

			if (!provider) {
				aiState.status = 'unavailable';
				return;
			}

			await loadProvider();

			aiState.status = 'ready';
			aiState.enabled = true;
			aiState.progress = 1;
		} catch {
			aiState.enabled = false;
			aiState.status = 'error';
			aiState.progress = 0;
		} finally {
			enablePromise = null;
		}
	})();

	return enablePromise;
}

export function disable(): void {
	aiState.enabled = false;
}

export async function suggest(text: string, signal?: AbortSignal): Promise<string | null> {
	if (!aiState.enabled || aiState.status !== 'ready' || !provider) {
		return null;
	}

	if (provider === 'chrome') {
		return chrome.suggest(text, signal);
	}

	return local.suggest(text, signal);
}
