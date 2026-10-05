type LanguageModelAvailability = 'available' | 'downloadable' | 'downloading' | 'unavailable';

interface LanguageModelSession {
	prompt(
		input: string,
		options?: {
			signal?: AbortSignal;
		}
	): Promise<string>;

	destroy?(): void;
}

interface LanguageModelMonitor {
	addEventListener(type: 'downloadprogress', listener: (event: { loaded: number }) => void): void;
}

interface LanguageModelOptions {
	monitor?: (monitor: LanguageModelMonitor) => void;
}

interface LanguageModelAPI {
	availability(options?: {
		expectedInputs?: Array<{
			type: 'text';
			languages?: string[];
		}>;
		expectedOutputs?: Array<{
			type: 'text';
			languages?: string[];
		}>;
	}): Promise<LanguageModelAvailability>;

	create(options?: LanguageModelOptions): Promise<LanguageModelSession>;
}

declare global {
	interface Window {
		LanguageModel?: LanguageModelAPI;
	}
}

const languageModelOptions = {
	expectedInputs: [{ type: 'text' as const, languages: ['en'] }],
	expectedOutputs: [{ type: 'text' as const, languages: ['en'] }]
};

let session: LanguageModelSession | null = null;

export function isSupported(): boolean {
	return typeof window !== 'undefined' && !!window.LanguageModel;
}

export async function availability(): Promise<LanguageModelAvailability> {
	if (!isSupported()) {
		return 'unavailable';
	}

	return window.LanguageModel!.availability(languageModelOptions);
}

export async function create(onProgress: (progress: number) => void): Promise<void> {
	if (!isSupported()) {
		throw new Error('Chrome built-in AI is unavailable.');
	}

	const newSession = await window.LanguageModel!.create({
		monitor(monitor) {
			monitor.addEventListener('downloadprogress', (event) => {
				onProgress(Math.min(Math.max(event.loaded, 0), 1));
			});
		}
	});

	session = newSession;
}

export async function suggest(text: string, signal?: AbortSignal): Promise<string | null> {
	if (!session || !text.trim() || signal?.aborted) {
		return null;
	}

	try {
		const result = await session.prompt(
			`You are an inline autocomplete engine.

Complete the user's text with a short, natural continuation.

Rules:
- Return only the continuation.
- Never repeat words already present in the user's text.
- Do not explain your answer.
- Do not use quotes.
- Prefer 2 to 8 words.
- Preserve the user's language and writing style.
- Do not start a new sentence unless the existing text naturally calls for it.

User text:
${text}`,
			{ signal }
		);

		return result.trim() || null;
	} catch {
		return null;
	}
}

export function destroy(): void {
	session?.destroy?.();
	session = null;
}
