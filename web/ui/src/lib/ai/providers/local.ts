import { pipeline, StoppingCriteria, type TextGenerationPipeline } from '@huggingface/transformers';

const MODEL = 'onnx-community/SmolLM2-360M-ONNX';

let generator: TextGenerationPipeline | null = null;
let loadingPromise: Promise<void> | null = null;

export function isSupported(): boolean {
	return typeof window !== 'undefined';
}

export async function load(onProgress?: (progress: number) => void): Promise<void> {
	if (generator) {
		onProgress?.(1);
		return;
	}

	if (loadingPromise) {
		return loadingPromise;
	}

	loadingPromise = (async () => {
		const useWebGPU = !!navigator.gpu;

		generator = await pipeline('text-generation', MODEL, {
			device: useWebGPU ? 'webgpu' : 'wasm',
			dtype: useWebGPU ? 'q4f16' : 'q4',

			progress_callback(info) {
				if (info.status !== 'progress_total') {
					return;
				}

				const progress = Number(info.progress);

				if (!Number.isFinite(progress)) {
					return;
				}

				onProgress?.(Math.min(Math.max(progress / 100, 0), 1));
			}
		});
	})();

	try {
		await loadingPromise;
	} catch (error) {
		generator = null;
		throw error;
	} finally {
		loadingPromise = null;
	}
}

function createPrompt(text: string): string {
	return `Complete the text.
Input: Hello
Output: world
---
Complete the text.
Input: How are
Output: you doing?
---
Complete the text.
Input: Thank you so
Output: much
---
Complete the text.
Input: ${text}
Output:`;
}

// Safely isolates only what the model appended right after "Input: [your text]\nOutput:"
function extractCompletion(generatedText: string, originalInput: string): string {
	const anchor = `Input: ${originalInput}\nOutput:`;
	const anchorIndex = generatedText.lastIndexOf(anchor);

	if (anchorIndex === -1) {
		return '';
	}

	// Extract everything after our specific input anchor
	let completion = generatedText.slice(anchorIndex + anchor.length);

	// Strip out any trailing examples or markers if the stop sequence didn't catch them
	const stopMarkers = ['\n', '---', 'Input:'];
	for (const marker of stopMarkers) {
		const index = completion.indexOf(marker);
		if (index !== -1) {
			completion = completion.slice(0, index);
		}
	}

	return completion;
}

// Cleans up any trailing artifacts or spaces from the isolated completion
function cleanCompletion(completion: string): string {
	// Preserves leading spaces (critical for inline typing!) but trims trailing junk
	return completion.trimEnd();
}

class AutocompleteStoppingCriteria extends StoppingCriteria {
	constructor(private readonly promptTokenLength: number) {
		super();
	}

	_call(input_ids: number[][]): boolean[] {
		if (!generator) {
			return [true];
		}

		const generatedTokens = input_ids[0].slice(this.promptTokenLength);

		if (generatedTokens.length === 0) {
			return [false];
		}

		const generatedText = generator.tokenizer.decode(generatedTokens, {
			skip_special_tokens: true
		});

		return [
			generatedText.endsWith('\n') ||
				generatedText.endsWith('---') ||
				generatedText.endsWith('Input:')
		];
	}
}
interface TextGenItem {
	generated_text: string;
}

export async function suggest(text: string, signal?: AbortSignal): Promise<string | null> {
	// Validate presence of generator, non-empty text, and active abort signal
	if (!generator || !text || signal?.aborted) {
		return null;
	}

	const prompt = createPrompt(text);

	try {
		const encoded = generator.tokenizer(prompt);
		const promptTokenLength = encoded.input_ids.dims[1] ?? 0;

		const output = await generator(prompt, {
			max_new_tokens: 12,
			do_sample: false,
			stopping_criteria: [new AutocompleteStoppingCriteria(promptTokenLength)],
			signal
		});

		if (signal?.aborted) {
			return null;
		}

		const outputResult = output as TextGenItem | TextGenItem[];

		// Hugging Face pipelines return an array of generation objects
		const generated = Array.isArray(outputResult)
			? outputResult[0]?.generated_text
			: outputResult?.generated_text;

		if (typeof generated !== 'string') {
			return null;
		}

		const completion = extractCompletion(generated, text);

		if (!completion) {
			return null;
		}

		return cleanCompletion(completion) || null;
	} catch {
		return null;
	}
}
