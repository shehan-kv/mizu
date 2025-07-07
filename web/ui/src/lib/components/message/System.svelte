<script lang="ts">
	import DownloadIcon from '../icons/DownloadIcon.svelte';
	import FilesIcon from '../icons/FilesIcon.svelte';
	import InvoiceIcon from '../icons/InvoiceIcon.svelte';

	interface Props {
		variant: 'file' | 'invoice';
		message: string;
		date: string;
		download?: string;
	}
	let { variant, message, download, date }: Props = $props();
</script>

<div>
	<p class="mr-4 text-right text-xs text-neutral-500">{date}</p>
	<div
		class="my-1 ml-auto flex w-fit items-center justify-end gap-4 rounded
		bg-neutral-100 px-4 py-3 text-sm text-neutral-700 dark:bg-neutral-900
		dark:text-neutral-300"
	>
		<p class="inline-flex items-center gap-1.5">
			{#if variant == 'invoice'}
				<InvoiceIcon class="size-4.5" />
			{:else if variant == 'file'}
				<FilesIcon class="size-4.5 " />
			{/if}
			{message}
		</p>

		{#if download}
			<a
				href={download}
				title="Download"
				class="transition hover:text-neutral-950 dark:hover:text-neutral-50"
			>
				<DownloadIcon class="size-4.5 " />
			</a>
		{/if}
	</div>
</div>
