<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import { page } from '$app/state';
	import { getFilesByProject, type File } from '$lib/api/files';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import FilterInput from '$lib/components/FilterInput.svelte';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import { onMount } from 'svelte';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import { formatDate } from '$lib/utils/formatDate';
	import { formatBytes } from '$lib/utils/formatBytes';
	import Pagination from '$lib/components/Pagination.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { getProjectDetails, type ProjectDetails } from '$lib/api/projects';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import { goto } from '$app/navigation';

	const MAX_LIMIT = 100;
	const MIN_LIMIT = 1;
	const DEFAULT_LIMIT = 30;
	const DEFAULT_PAGE_NUMBER = 1;

	const params = new URLSearchParams(page.url.searchParams.toString());

	let id = Number(page.params.id);

	let q = $state(params.get('q') || '');
	let pageNum = $state(Number(params.get('page')) || DEFAULT_PAGE_NUMBER);
	let limit = $state(Math.min(Number(params.get('limit')) || DEFAULT_LIMIT, MAX_LIMIT));

	let filesPromise: Promise<PaginatedResponse<File>> | null = $state(null);
	let fileAbort: AbortController | null = null;
	function loadFiles() {
		if (fileAbort) {
			fileAbort.abort();
		}

		fileAbort = new AbortController();

		filesPromise = getFilesByProject(id, q, pageNum, limit, fileAbort.signal);
	}

	let projectPromise: Promise<ProjectDetails> | null = $state(null);
	let projectAbort: AbortController | null = null;
	function loadProject() {
		if (projectAbort) {
			projectAbort.abort();
		}

		projectAbort = new AbortController();

		projectPromise = getProjectDetails(id, projectAbort.signal).then((d) => {
			document.title = 'Files - ' + d.name;
			return d;
		});
	}

	function updateUrlParam() {
		if (q) {
			params.set('q', q);
		} else {
			params.delete('q');
		}

		params.set('page', pageNum.toString());
		params.set('limit', limit.toString());

		goto(`?${params.toString()}`, { replaceState: true, keepFocus: true });
	}

	function handleFilter() {
		pageNum = 1;
		if (limit > MAX_LIMIT) limit = MAX_LIMIT;
		if (limit < MIN_LIMIT) limit = MIN_LIMIT;
		updateUrlParam();
		loadFiles();
	}

	onMount(() => {
		loadProject();
		loadFiles();
	});
</script>

<svelte:head>
	<title>Files</title>
</svelte:head>

<div class="grid h-full auto-rows-[min-content_1fr] gap-6">
	<div class="mx-auto space-y-4 lg:container">
		<div class="flex w-fit items-center gap-3 text-sm text-neutral-700 dark:text-neutral-400">
			{#await projectPromise}
				<p class="">...</p>
			{:then res}
				<a href={`/projects/${res?.id}`} class="underline">{res?.name}</a>
			{/await}

			<ChevronRight size={18} />
			<p>Files</p>
		</div>
		<div class="flex gap-4">
			<div class="w-full max-w-80">
				<SearchBar bind:value={q} onchange={handleFilter} />
			</div>
			<div>
				<FilterInput
					id="limit"
					max={MAX_LIMIT}
					min={MIN_LIMIT}
					label="Limit"
					type="number"
					bind:value={limit}
					onchange={handleFilter}
				/>
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
							{#each res.data as file (file)}
								<Table.Row>
									<Table.Cell>{file.originalName}</Table.Cell>
									<Table.Cell>{formatBytes(file.size)}</Table.Cell>
									<Table.Cell>{formatDate(file.uploadedAt)}</Table.Cell>
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
					<Pagination bind:page={pageNum} count={res.count} perPage={res.limit} />
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
