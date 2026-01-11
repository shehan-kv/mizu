<script lang="ts">
	import { page } from '$app/state';
	import * as Table from '$lib/components/ui/table';
	import * as Dialog from '$lib/components/dialogs';
	import { getContractRevisions, type ContractRevision } from '$lib/api/contracts';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import FilterInput from '$lib/components/FilterInput.svelte';
	import FilterSelect from '$lib/components/FilterSelect.svelte';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import { onMount } from 'svelte';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import Checks from 'phosphor-svelte/lib/Checks';
	import { formatDate } from '$lib/utils/formatDate';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Pagination from '$lib/components/Pagination.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { createDialogState } from '$lib/components/dialogs/createDialogState.svelte';

	const MAX_LIMIT = 100;
	const MIN_LIMIT = 1;
	const DEFAULT_LIMIT = 30;
	const DEFAULT_PAGE_NUMBER = 1;

	const params = new URLSearchParams(page.url.searchParams.toString());

	let q = $state(params.get('q') || '');
	let status = $state(params.get('status') || '');
	let pageNum = $state(Number(params.get('page')) || DEFAULT_PAGE_NUMBER);
	let limit = $state(Math.min(Number(params.get('limit')) || DEFAULT_LIMIT, MAX_LIMIT));

	let revisions: Promise<PaginatedResponse<ContractRevision>> | null = $state(null);
	let abort: AbortController | null = null;

	function loadRevisions() {
		if (abort) {
			abort.abort();
		}

		abort = new AbortController();

		revisions = getContractRevisions({ q, status, page: pageNum, limit }, abort.signal);
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

		history.replaceState(null, '', `?${params.toString()}`);
	}

	function handleFilter() {
		pageNum = 1;
		if (limit > MAX_LIMIT) limit = MAX_LIMIT;
		if (limit < MIN_LIMIT) limit = MIN_LIMIT;
		updateUrlParam();
		loadRevisions();
	}

	onMount(() => {
		loadRevisions();
	});

	let viewRevisionDialog = createDialogState();
	let selectedRevision: ContractRevision | null = $state(null);
	function openRevisionDialog(revision: ContractRevision) {
		selectedRevision = revision;
		viewRevisionDialog.open();
	}
</script>

<svelte:head>
	<title>Contract Revision Requests</title>
</svelte:head>

<div class="grid h-full auto-rows-[min-content_1fr_min-content] gap-6">
	<div class="flex gap-2 lg:container">
		<div class="max-w-96">
			<SearchBar bind:value={q} onchange={handleFilter} />
		</div>
		<FilterSelect
			bind:value={status}
			onchange={handleFilter}
			name="Status"
			options={[
				{ value: '', label: 'All' },
				{ value: 'pending', label: 'Pending' },
				{ value: 'accepted', label: 'Accepted' },
				{ value: 'rejected', label: 'Rejected' }
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

	{#await revisions}
		<Spinner />
	{:then res}
		{#if res && res.data}
			<div class="gap-4 overflow-y-auto lg:container">
				{#if res.data.length == 0}
					<ErrorMessage variant="info" text="Revision Requests Not Found" />
				{/if}
				<div class="overflow-y-auto">
					{#if res.data.length > 0}
						<Table.Root class="lg:container">
							<Table.Header>
								<Table.Row>
									<Table.Head class="font-bold">Title</Table.Head>
									<Table.Head class="font-bold">Contract</Table.Head>
									<Table.Head class="font-bold">Status</Table.Head>
									<Table.Head class="font-bold">Started By</Table.Head>
									<Table.Head class="font-bold">Created Date</Table.Head>
									<Table.Head class="font-bold">Updated Date</Table.Head>
									<Table.Head class="font-bold">Actions</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each res.data as revision (revision)}
									<Table.Row class="*:whitespace-normal">
										<Table.Cell class="whitespace-normal">{revision.title}</Table.Cell>
										<Table.Cell>{revision.contractName}</Table.Cell>
										<Table.Cell>
											<div class="flex items-center gap-1">
												{toTitleCase(revision.status)}
												{#if revision.status == 'signed'}
													<Checks size={18} class="text-emerald-500" />
												{/if}
											</div>
										</Table.Cell>
										<Table.Cell>
											{revision.reqUser.firstName}
											{revision.reqUser.lastName}
										</Table.Cell>
										<Table.Cell>{formatDate(revision.createdAt)}</Table.Cell>
										<Table.Cell
											>{revision.updatedAt ? formatDate(revision.updatedAt) : 'N/A'}</Table.Cell
										>
										<Table.Cell>
											<div
												class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5 *:hover:text-neutral-950 dark:text-neutral-400 *:dark:hover:text-neutral-50"
											>
												<button
													title="View Revision Request"
													onclick={() => openRevisionDialog(revision)}
												>
													<ArrowRight size={18} />
												</button>
											</div>
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
			<ErrorMessage variant="warn" text="Invalid Request" retry={loadRevisions} />
		{:else if err instanceof APIForbiddenError}
			<ErrorMessage
				variant="warn"
				text="You Don't Have Permission To View These Revision Requests"
				retry={loadRevisions}
			/>
		{:else if err instanceof APINotFoundError}
			<ErrorMessage variant="info" text="Not Found" retry={loadRevisions} />
		{:else if err instanceof APIServerError}
			<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadRevisions} />
		{:else}
			<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadRevisions} />
		{/if}
	{/await}
</div>

{#if selectedRevision}
	<Dialog.ViewContractRevision
		bind:open={viewRevisionDialog.isOpen}
		contractId={selectedRevision.contractId}
		revisionId={selectedRevision.id}
		role="admin"
		onUpdate={loadRevisions}
	/>
{/if}
