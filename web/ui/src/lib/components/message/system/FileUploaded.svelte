<script lang="ts">
	import { formatBytes } from '$lib/utils/formatBytes';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import ArrowSquareOut from 'phosphor-svelte/lib/ArrowSquareOut';
	import Files from 'phosphor-svelte/lib/Files';
	import { downloadChannelFile, openFileInNewTab } from '$lib/api/messages';

	export interface FileUploadedPayload {
		type: 'message.file.created';
		fileId: string;
		name: string;
		mimeType: string;
		size: number;
		userId: string;
		firstName: string;
		lastName: string;
		uploadedAt: string;
	}

	interface Props {
		message: FileUploadedPayload;
	}

	// eslint-disable-next-line svelte/no-unused-props
	let { message }: Props = $props();

	const imgMimeTypes = [
		'image/apng',
		'image/avif',
		'image/bmp',
		'image/gif',
		'image/x-icon',
		'image/vnd.microsoft.icon',
		'image/jpeg',
		'image/png',
		'image/svg+xml',
		'image/webp'
	];

	let isImage = imgMimeTypes.includes(message.mimeType);
</script>

<div
	class="my-1 ml-auto w-fit min-w-xs gap-4 rounded
		bg-neutral-100 p-4 text-sm text-neutral-700 dark:bg-neutral-900
		dark:text-neutral-300"
>
	<p class="text-xs text-neutral-500">
		#{message.fileId.replaceAll('-', '').slice(-8).toUpperCase()}
	</p>
	<div class="flex items-center gap-1">
		<Files size={16} weight="fill" />
		<p class="grow">File Uploaded</p>

		{#if isImage}
			<button
				type="button"
				title="Open Image In New Tab"
				class="p-1 text-neutral-500 hover:text-neutral-950 dark:hover:text-neutral-50"
				onclick={() => openFileInNewTab(message.fileId)}
			>
				<ArrowSquareOut size={16} />
			</button>
		{/if}
		<button
			onclick={() => downloadChannelFile(message.fileId)}
			title="Download"
			class="cursor-pointer p-1 text-neutral-500 hover:text-neutral-950 dark:hover:text-neutral-50"
		>
			<DownloadSimple size={16} />
		</button>
	</div>

	<div class="mt-3">
		<p>{message.name}</p>
		<p class="text-neutral-500">{formatBytes(message.size)}</p>
	</div>

	{#if isImage}
		<div class="mt-3 max-w-lg overflow-hidden rounded">
			<img alt={message.name} src={`/api/v1/messages/files/${message.fileId}`} />
		</div>
	{/if}

	<div class="mt-3">
		<p
			class="w-fit rounded bg-neutral-200
				px-3 py-1.5 text-xs dark:bg-neutral-800"
		>
			{message.firstName + ' ' + message.lastName}
		</p>
	</div>
</div>
