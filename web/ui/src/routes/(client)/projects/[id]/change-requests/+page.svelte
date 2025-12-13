<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import { page } from '$app/state';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import FilterInput from '$lib/components/FilterInput.svelte';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import { onMount } from 'svelte';
	import { formatDate } from '$lib/utils/formatDate';
	import Pagination from '$lib/components/Pagination.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { getProjectDetails, type ProjectDetails } from '$lib/api/projects';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import FilterSelect from '$lib/components/FilterSelect.svelte';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import Checks from 'phosphor-svelte/lib/Checks';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import { goto } from '$app/navigation';
	import { getChangeRequestsByProject, type ChangeRequest } from '$lib/api/changeRequest';

	const MAX_LIMIT = 100;
	const MIN_LIMIT = 1;
	const DEFAULT_LIMIT = 30;
	const DEFAULT_PAGE_NUMBER = 1;

	const params = new URLSearchParams(page.url.searchParams.toString());

	let id = Number(page.params.id);

	let q = $state(params.get('q') || '');
	let status = $state(params.get('status') || '');
	let pageNum = $state(Number(params.get('page')) || DEFAULT_PAGE_NUMBER);
	let limit = $state(Math.min(Number(params.get('limit')) || DEFAULT_LIMIT, MAX_LIMIT));

	let projectPromise: Promise<ProjectDetails> | null = $state(null);
	let projectAbort: AbortController | null = null;
	function loadProject() {
		if (projectAbort) {
			projectAbort.abort();
		}

		projectAbort = new AbortController();

		projectPromise = getProjectDetails(id, projectAbort.signal).then((d) => {
			document.title = 'Change Requests - ' + d.name;
			return d;
		});
	}

	let chReqPromise: Promise<PaginatedResponse<ChangeRequest>> | null = $state(null);
	let chReqAbort: AbortController | null = null;
	function loadChReq() {
		if (chReqAbort) {
			chReqAbort.abort();
		}

		chReqAbort = new AbortController();

		chReqPromise = getChangeRequestsByProject(
			id,
			{
				q,
				page: pageNum,
				limit,
				status
			},
			chReqAbort.signal
		);
	}

	function updateUrlParam() {
		if (q) {
			params.set('q', q);
		} else {
			params.delete('q');
		}

		params.set('page', pageNum.toString());
		params.set('limit', limit.toString());

		if (status) {
			params.set('status', status);
		} else {
			params.delete('status');
		}

		goto(`?${params.toString()}`, { replaceState: true, keepFocus: true });
	}

	function handleFilter() {
		pageNum = 1;
		if (limit > MAX_LIMIT) limit = MAX_LIMIT;
		if (limit < MIN_LIMIT) limit = MIN_LIMIT;
		updateUrlParam();
		loadChReq();
	}

	onMount(() => {
		loadProject();
		loadChReq();
	});
</script>

<svelte:head>
	<title>Change Requests</title>
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
			<p>Change Requests</p>
		</div>
		<div class="flex gap-4">
			<div class="w-full max-w-80">
				<SearchBar bind:value={q} onchange={handleFilter} />
			</div>
			<div class="flex gap-2">
				<FilterSelect
					bind:value={status}
					onchange={handleFilter}
					name="Status"
					options={[
						{ value: '', label: 'All' },
						{ value: 'in-progress', label: 'In Progress' },
						{ value: 'waiting', label: 'Waiting' },
						{ value: 'closed', label: 'Closed' }
					]}
				/>
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

	{#await chReqPromise}
		<Spinner />
	{:then res}
		{#if res && res.data}
			<div class="mx-auto gap-4 overflow-y-auto lg:container">
				{#if res.data.length == 0}
					<ErrorMessage variant="info" text="Invoices/Quotes Not Found" />
				{/if}
				<div class="overflow-y-auto">
					{#if res.data.length > 0}
						<Table.Root class="container mx-auto">
							<Table.Header>
								<Table.Row>
									<Table.Head class="font-bold">Subject</Table.Head>
									<Table.Head class="font-bold">Status</Table.Head>
									<Table.Head class="font-bold">Created At</Table.Head>
									<Table.Head class="font-bold">Created By</Table.Head>
									<Table.Head class="font-bold">Actions</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each res.data as req (req)}
									<Table.Row>
										<Table.Cell>{req.title}</Table.Cell>
										<Table.Cell class="flex items-center gap-1">
											{toTitleCase(req.status)}
											{#if req.status == 'closed'}
												<Checks size={18} class="text-emerald-500" />
											{/if}
										</Table.Cell>
										<Table.Cell>{formatDate(req.createdAt)}</Table.Cell>
										<Table.Cell>{req.requestedBy.firstName} {req.requestedBy.lastName}</Table.Cell>
										<Table.Cell>
											<a
												href={`/change-requests/${req.id}`}
												class="inline-block cursor-pointer px-1.5 text-xs
												text-neutral-500 hover:text-neutral-950 dark:text-neutral-400
												dark:hover:text-neutral-50"
												title="View"
											>
												<ArrowRight size={18} />
											</a>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</div>
			</div>
			{#if res.data.length > 0}
				<div class="container mx-auto flex justify-end">
					<Pagination bind:page={pageNum} count={res.count} perPage={limit} />
				</div>
			{/if}
		{/if}
	{:catch err}
		{#if err instanceof APIBadRequestError}
			<ErrorMessage variant="warn" text="Invalid Request" retry={loadChReq} />
		{:else if err instanceof APIForbiddenError}
			<ErrorMessage
				variant="warn"
				text="You Don't Have Permission To View These Invoices/Quotes"
				retry={loadChReq}
			/>
		{:else if err instanceof APINotFoundError}
			<ErrorMessage variant="info" text="Not Found" retry={loadChReq} />
		{:else if err instanceof APIServerError}
			<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadChReq} />
		{:else}
			<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadChReq} />
		{/if}
	{/await}
</div>
