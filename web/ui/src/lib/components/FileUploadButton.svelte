<script lang="ts">
	import { ApiError } from '$lib/api/client';
	import { uploadMessageFile } from '$lib/api/messages';
	import UploadSimple from 'phosphor-svelte/lib/UploadSimple';
	import X from 'phosphor-svelte/lib/X';
	import { toast } from 'svelte-sonner';

	interface Props {
		channelId: string;
	}

	let { channelId }: Props = $props();

	let uploadProgress = $state(0);
	let fileName: string | undefined = $state(undefined);
	let uploading = $state(false);

	// svelte-ignore non_reactive_update
	let abort: AbortController | null = null;

	async function handleFileUpload(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		fileName = file?.name;

		if (!file) return;

		// Cancel any previous upload
		abort?.abort();
		abort = new AbortController();

		uploading = true;
		uploadProgress = 0;

		try {
			await uploadMessageFile(channelId, file, {
				signal: abort.signal,
				onProgress: (progress) => {
					uploadProgress = progress;
				}
			});
		} catch (error) {
			// Ignore deliberate cancellation
			if (error instanceof DOMException && error.name === 'AbortError') {
				return;
			}

			if (error instanceof ApiError) {
				toast.error(error.message);
			} else if (error instanceof Error) {
				toast.error(error.message);
			} else {
				toast.error('Upload failed');
			}
		} finally {
			uploading = false;
			uploadProgress = 0;

			input.value = '';
		}
	}
</script>

<div class="relative shrink-0">
	<input
		id="file-upload"
		disabled={uploading}
		type="file"
		class="hidden"
		onchange={handleFileUpload}
	/>

	<label
		for="file-upload"
		class="inline-flex cursor-pointer items-center gap-1.5 rounded p-2 text-xs
		   hover:bg-neutral-200 dark:hover:bg-neutral-800"
	>
		<UploadSimple size={16} />
		{#if uploading}
			Uploading...
		{:else}
			Upload File
		{/if}
	</label>

	<!-- Uploading status -->
	{#if uploading}
		<div
			class="absolute -top-18 right-0 min-w-max space-y-2 rounded border bg-neutral-50 p-2 text-left dark:bg-neutral-950"
		>
			<div class="flex items-center gap-4">
				<p>{fileName}</p>
				<button
					onclick={abort?.abort}
					class="cursor-pointer rounded p-1 hover:bg-neutral-200 dark:hover:bg-neutral-800"
					><X /></button
				>
			</div>

			<div class="h-1 w-full">
				<div
					class="h-full rounded-full bg-neutral-950 dark:bg-neutral-50"
					style={`width: ${uploadProgress}%`}
				></div>
			</div>
		</div>
	{/if}
</div>
