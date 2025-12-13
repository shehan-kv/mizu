<script lang="ts">
	import Checks from 'phosphor-svelte/lib/Checks';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Envelope from 'phosphor-svelte/lib/Envelope';
	import * as Table from '$lib/components/ui/table';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import { onMount } from 'svelte';
	import { getInvoices, type InvoiceSummary } from '$lib/api/invoices';
	import { page } from '$app/state';
	import Spinner from '$lib/components/Spinner.svelte';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import FilterSelect from '$lib/components/FilterSelect.svelte';
	import FilterInput from '$lib/components/FilterInput.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { formatDate } from '$lib/utils/formatDate';
	import { getContracts, type Contract } from '$lib/api/contracts';

	const MAX_LIMIT = 100;
	const MIN_LIMIT = 1;
	const DEFAULT_LIMIT = 30;
	const DEFAULT_PAGE_NUMBER = 1;

	const params = new URLSearchParams(page.url.searchParams.toString());

	let q = $state(params.get('q') || '');
	let status = $state(params.get('status') || '');
	let pageNum = $state(Number(params.get('page')) || DEFAULT_PAGE_NUMBER);
	let limit = $state(Math.min(Number(params.get('limit')) || DEFAULT_LIMIT, MAX_LIMIT));

	let contractsPromise: Promise<PaginatedResponse<Contract>> | null = $state(null);

	let abort: AbortController | null = null;
	function loadContracts() {
		if (abort) {
			abort.abort();
		}

		abort = new AbortController();

		contractsPromise = getContracts(q, status, pageNum, limit, abort.signal);
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
		loadContracts();
	}

	onMount(() => {
		loadContracts();
	});
</script>

<svelte:head>
	<title>Contracts</title>
</svelte:head>

<div class="grid h-full auto-rows-[min-content_1fr_min-content] gap-6">
	<div class="mx-auto flex gap-4 lg:container">
		<div class="max-w-96">
			<SearchBar bind:value={q} onchange={handleFilter} />
		</div>
		<div class="flex gap-2">
			<FilterSelect
				bind:value={status}
				onchange={handleFilter}
				name="Status"
				options={[
					{ value: '', label: 'All' },
					{ value: 'signed', label: 'Signed' },
					{ value: 'rejected', label: 'Rejected' },
					{ value: 'pending', label: 'Pending' }
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

	{#await contractsPromise}
		<Spinner />
	{:then res}
		{#if res && res.data}
			<div class="mx-auto gap-4 overflow-y-auto lg:container">
				{#if res.data.length == 0}
					<ErrorMessage variant="info" text="Contracts Not Found" />
				{/if}
				<div class="overflow-y-auto">
					{#if res.data.length > 0}
						<Table.Root class="container mx-auto">
							<Table.Header>
								<Table.Row>
									<Table.Head class="font-bold">Name</Table.Head>
									<Table.Head class="font-bold">Project</Table.Head>
									<Table.Head class="font-bold">Status</Table.Head>
									<Table.Head class="font-bold">Revisions</Table.Head>
									<Table.Head class="font-bold">Created Date</Table.Head>
									<Table.Head class="font-bold">Actions</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each res.data as contract (contract)}
									<Table.Row>
										<Table.Cell>{contract.name}</Table.Cell>
										<Table.Cell>{contract.projectName}</Table.Cell>
										<Table.Cell class="flex items-center gap-1">
											{toTitleCase(contract.status)}
											{#if contract.status == 'signed'}
												<Checks size={18} class="text-emerald-500" />
											{/if}
										</Table.Cell>
										<Table.Cell>
											{contract.acceptedRevisions} / {contract.numOfRevisions} Accepted
										</Table.Cell>
										<Table.Cell>{formatDate(contract.createdAt)}</Table.Cell>
										<Table.Cell>
											<div
												class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5 *:hover:text-neutral-950 dark:text-neutral-400 *:dark:hover:text-neutral-50"
											>
												<a
													href={`/contracts/${contract.id}`}
													title="View Contract"
													class="inline-block"
												>
													<ArrowRight size={18} />
												</a>
												<button title="Download the Latest Version as PDF">
													<DownloadSimple size={18} />
												</button>
												<button title="Email Me"><Envelope size={18} /></button>
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
			<ErrorMessage variant="warn" text="Invalid Request" retry={loadContracts} />
		{:else if err instanceof APIForbiddenError}
			<ErrorMessage
				variant="warn"
				text="You Don't Have Permission To View These Invoices/Quotes"
				retry={loadContracts}
			/>
		{:else if err instanceof APINotFoundError}
			<ErrorMessage variant="info" text="Not Found" retry={loadContracts} />
		{:else if err instanceof APIServerError}
			<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadContracts} />
		{:else}
			<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadContracts} />
		{/if}
	{/await}
</div>
