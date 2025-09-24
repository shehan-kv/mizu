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
	import { getInvoices, type Invoice } from '$lib/api/invoices';
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

	const MAX_LIMIT = 100;
	const MIN_LIMIT = 1;
	const DEFAULT_LIMIT = 30;
	const DEFAULT_PAGE_NUMBER = 1;

	const params = new URLSearchParams(page.url.searchParams.toString());

	let q = $state(params.get('q') || '');
	let status = $state(params.get('status') || '');
	let type = $state(params.get('type') || '');
	let pageNum = $state(Number(params.get('page')) || DEFAULT_PAGE_NUMBER);
	let limit = $state(Math.min(Number(params.get('limit')) || DEFAULT_LIMIT, MAX_LIMIT));

	let invoicesPromise: Promise<PaginatedResponse<Invoice>> | null = $state(null);

	let abortController: AbortController | null = null;
	function loadInvoices() {
		if (abortController) {
			abortController.abort();
		}

		abortController = new AbortController();

		invoicesPromise = getInvoices(q, status, type, pageNum, limit, abortController.signal);
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

		if (type) {
			params.set('type', type);
		} else {
			params.delete('type');
		}

		history.replaceState(null, '', `?${params.toString()}`);
	}

	function handleFilter() {
		pageNum = 1;
		if (limit > MAX_LIMIT) limit = MAX_LIMIT;
		if (limit < MIN_LIMIT) limit = MIN_LIMIT;
		updateUrlParam();
		loadInvoices();
	}

	onMount(() => {
		loadInvoices();
	});
</script>

<svelte:head>
	<title>Invoices and Quotes</title>
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
					{ value: 'accepted', label: 'Accepted' },
					{ value: 'rejected', label: 'Rejected' },
					{ value: 'cancelled', label: 'Cancelled' },
					{ value: 'pending', label: 'Pending' }
				]}
			/>
			<FilterSelect
				bind:value={type}
				onchange={handleFilter}
				name="Type"
				options={[
					{ value: '', label: 'All' },
					{ value: 'invoice', label: 'Invoice' },
					{ value: 'quote', label: 'Quote' }
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

	{#await invoicesPromise}
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
									<Table.Head class="font-bold">Type</Table.Head>
									<Table.Head class="font-bold">#ID</Table.Head>
									<Table.Head class="font-bold">Project</Table.Head>
									<Table.Head class="font-bold">Amount</Table.Head>
									<Table.Head class="font-bold">Status</Table.Head>
									<Table.Head class="font-bold">Issued On</Table.Head>
									<Table.Head class="font-bold">Due Date</Table.Head>
									<Table.Head class="font-bold">Actions</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each res.data as invoice}
									<Table.Row>
										<Table.Cell>{invoice.isInvoice ? 'Invoice' : 'Quote'}</Table.Cell>
										<Table.Cell>#{invoice.id}</Table.Cell>
										<Table.Cell>{invoice.projectName}</Table.Cell>
										<Table.Cell>
											{currencyFormatter(invoice.currencyCode, invoice.total)}
										</Table.Cell>
										<Table.Cell class="flex items-center gap-1">
											{toTitleCase(invoice.status)}
											{#if invoice.status == 'paid' || invoice.status == 'accepted'}
												<Checks size={18} class="text-emerald-500" />
											{/if}
										</Table.Cell>
										<Table.Cell>{new Date(invoice.issuedAt).toLocaleString()}</Table.Cell>
										<Table.Cell>
											{invoice.dueAt ? new Date(invoice.dueAt).toLocaleString() : 'N/A'}
										</Table.Cell>
										<Table.Cell>
											<div
												class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5 *:hover:text-neutral-950 dark:text-neutral-400 *:dark:hover:text-neutral-50"
											>
												<button title="View">
													<ArrowRight size={18} />
												</button>
												<button title="Download as PDF">
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
			<ErrorMessage variant="warn" text="Invalid Request" retry={loadInvoices} />
		{:else if err instanceof APIForbiddenError}
			<ErrorMessage
				variant="warn"
				text="You Don't Have Permission To View These Invoices/Quotes"
				retry={loadInvoices}
			/>
		{:else if err instanceof APINotFoundError}
			<ErrorMessage variant="info" text="Not Found" retry={loadInvoices} />
		{:else if err instanceof APIServerError}
			<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadInvoices} />
		{:else}
			<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadInvoices} />
		{/if}
	{/await}
</div>
