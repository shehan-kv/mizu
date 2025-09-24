<script lang="ts">
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import * as Table from '$lib/components/ui/table';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import type { Channel } from '$lib/components/message/types';
	import Spinner from '../Spinner.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import { getFilesByChannel, type File } from '$lib/api/files';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { formatBytes } from '$lib/utils/formatBytes';

	interface Props {
		open: boolean;
		close: () => void;
		channel: Channel;
	}
	let { open = $bindable(), close, channel }: Props = $props();

	let _q = $state('');
	let q = $state('');

	let page = $state(1);
	let limit = $state(30);

	let filesPromise: Promise<PaginatedResponse<File>> | null = $state(null);

	let abortController: AbortController | null = null;
	function loadFiles() {
		if (abortController) {
			abortController.abort();
		}

		abortController = new AbortController();

		filesPromise = getFilesByChannel(channel.id, q, page, limit, abortController.signal);
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
			{#if res && res.data}
				<div class="overflow-y-auto">
					{#if res.data.length == 0}
						<ErrorMessage variant="info" text="Files Not Found" />
					{/if}
					{#if res.data.length > 0}
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
								{#each res.data as file}
									<Table.Row>
										<Table.Cell>{file.originalName}</Table.Cell>
										<Table.Cell>{formatBytes(file.size)}</Table.Cell>
										<Table.Cell>{new Date(file.uploadedAt).toLocaleString()}</Table.Cell>
										<Table.Cell>{file.user.firstName} {file.user.lastName}</Table.Cell>
										<Table.Cell>
											<a
												href={file.url}
												class="block w-fit cursor-pointer px-2 text-neutral-600
										transition hover:text-neutral-950 dark:text-neutral-400 dark:hover:text-neutral-50"
											>
												<DownloadSimple size={18} />
											</a>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</div>
				{#if res.data.length > 0}
					<div class="container mx-auto flex justify-end">
						<Pagination bind:page count={res.count} perPage={res.limit} />
					</div>
				{/if}
			{/if}
		{:catch err}
			{#if err instanceof APIBadRequestError}
				<ErrorMessage variant="warn" text="Invalid Request" retry={loadFiles} />
			{:else if err instanceof APIForbiddenError}
				<ErrorMessage
					variant="warn"
					text="You Don't Have Permission To View These Files"
					retry={loadFiles}
				/>
			{:else if err instanceof APINotFoundError}
				<ErrorMessage variant="info" text="Not Found" retry={loadFiles} />
			{:else if err instanceof APIServerError}
				<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadFiles} />
			{:else}
				<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadFiles} />
			{/if}
		{/await}
	</div>
</FullScreenDialog>
