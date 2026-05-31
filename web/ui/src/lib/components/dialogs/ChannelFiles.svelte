<script lang="ts">
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import * as Table from '$lib/components/ui/table';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import Spinner from '../Spinner.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import { formatBytes } from '$lib/utils/formatBytes';
	import { formatDate } from '$lib/utils/formatDate';
	import {
		downloadChannelFile,
		getChannelFiles,
		type Channel,
		type ChannelFile
	} from '$lib/api/messages';
	import type { PaginatedResponse } from '$lib/api/page';

	interface Props {
		open: boolean;
		channel: Channel;
	}
	let { open = $bindable(), channel }: Props = $props();

	let _q = $state('');
	let q = $state('');

	let page = $state(1);
	let limit = $state(30);

	let filesPromise: Promise<PaginatedResponse<ChannelFile>> | null = $state(null);

	let abortController: AbortController | null = null;
	function loadFiles() {
		if (abortController) {
			abortController.abort();
		}

		abortController = new AbortController();

		filesPromise = getChannelFiles(channel.id, { q, page, limit }, abortController.signal);
	}

	function handleSearch() {
		// $effect automatically runs the loadFiles function when
		// q changes. This function is used as a workaround to
		// set page to 1 when a user searches for a file.
		page = 1;
		q = _q;
	}

	$effect(() => {
		if (!open) return;
		loadFiles();
	});
</script>

<FullScreenDialog bind:open>
	<div class="grid auto-rows-[min-content_1fr_min-content] gap-6 overflow-y-auto px-5">
		<div class="flex-none">
			<div class="container mx-auto flex items-end justify-between gap-4">
				<p class="font-bold">Uploaded Files in {channel.name}</p>
				<div class="w-full max-w-xs">
					<SearchBar bind:value={_q} onchange={handleSearch} />
				</div>
			</div>
		</div>

		{#await filesPromise}
			<Spinner />
		{:then res}
			{#if res && res.items}
				<div class="overflow-y-auto">
					{#if res.items.length == 0}
						<ErrorMessage variant="info" text="Files Not Found" />
					{/if}
					{#if res.items.length > 0}
						<Table.Root class="container mx-auto">
							<Table.Header>
								<Table.Row>
									<Table.Head class="font-bold">File Name</Table.Head>
									<Table.Head class="font-bold">Size</Table.Head>
									<Table.Head class="font-bold">Uploaded Date</Table.Head>
									<Table.Head class="font-bold">Uploaded By</Table.Head>
									<Table.Head class="font-bold">Actions</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each res.items as file (file.id)}
									<Table.Row>
										<Table.Cell>{file.name}</Table.Cell>
										<Table.Cell>{formatBytes(file.size)}</Table.Cell>
										<Table.Cell>{formatDate(file.uploadedAt)}</Table.Cell>
										<Table.Cell>{file.user.firstName} {file.user.lastName}</Table.Cell>
										<Table.Cell>
											<button
												onclick={() => downloadChannelFile(file.id)}
												class="block w-fit cursor-pointer px-2 text-neutral-600
										transition hover:text-neutral-950 dark:text-neutral-400 dark:hover:text-neutral-50"
											>
												<DownloadSimple size={18} />
											</button>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</div>
				{#if res.items.length > 0}
					<div class="container mx-auto flex justify-end">
						<Pagination bind:page count={res.totalCount} perPage={res.limit} />
					</div>
				{/if}
			{/if}
		{:catch err}
			<ErrorMessage variant="warn" text={err} retry={loadFiles} />
		{/await}
	</div>
</FullScreenDialog>
