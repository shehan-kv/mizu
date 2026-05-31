<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import { page } from '$app/state';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import FilterInput from '$lib/components/FilterInput.svelte';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import { onMount } from 'svelte';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import { formatDate } from '$lib/utils/formatDate';
	import Pagination from '$lib/components/Pagination.svelte';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import FilterSelect from '$lib/components/FilterSelect.svelte';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import Checks from 'phosphor-svelte/lib/Checks';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Envelope from 'phosphor-svelte/lib/Envelope';
	import { goto } from '$app/navigation';
	import { getContractOverviewsByProject, type Contract } from '$lib/api/contracts';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import type { PaginatedResponse } from '$lib/api/page';
	import { getProject, type Project } from '$lib/api/projects';
	import { ApiError } from '$lib/api/client';
	import { CONTRACT_STATUS } from '$lib/constants/contract';

	const MAX_LIMIT = 100;
	const MIN_LIMIT = 1;
	const DEFAULT_LIMIT = 30;
	const DEFAULT_PAGE_NUMBER = 1;

	const params = new SvelteURLSearchParams(page.url.searchParams.toString());

	let id = page.params.id || '';

	let q = $state(params.get('q') || '');
	let status = $state(CONTRACT_STATUS.find((s) => s === params.get('status')) ?? '');
	let pageNum = $state(Number(params.get('page')) || DEFAULT_PAGE_NUMBER);
	let limit = $state(Math.min(Number(params.get('limit')) || DEFAULT_LIMIT, MAX_LIMIT));

	let projectPromise: Promise<Project> | null = $state(null);
	let projectAbort: AbortController | null = null;
	function loadProject() {
		if (projectAbort) {
			projectAbort.abort();
		}

		projectAbort = new AbortController();

		projectPromise = getProject(id, projectAbort.signal).then((d) => {
			document.title = 'Contracts - ' + d.name;
			return d;
		});
	}

	let contractsPromise: Promise<PaginatedResponse<Contract>> | null = $state(null);
	let contractsAbort: AbortController | null = null;
	function loadContracts() {
		if (contractsAbort) {
			contractsAbort.abort();
		}

		contractsAbort = new AbortController();

		contractsPromise = getContractOverviewsByProject(
			id,
			{ q, status, page: pageNum, limit },
			contractsAbort.signal
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

		// eslint-disable-next-line svelte/no-navigation-without-resolve
		goto(`?${params.toString()}`, { replaceState: true, keepFocus: true });
	}

	function handleFilter() {
		pageNum = 1;
		if (limit > MAX_LIMIT) limit = MAX_LIMIT;
		if (limit < MIN_LIMIT) limit = MIN_LIMIT;
		updateUrlParam();
		loadContracts();
	}

	onMount(() => {
		loadProject();
		loadContracts();
	});
</script>

<svelte:head>
	<title>Contracts</title>
</svelte:head>

<div class="grid h-full auto-rows-[min-content_1fr] gap-6">
	<div class="mx-auto space-y-4 lg:container">
		<div class="flex w-fit items-center gap-3 text-sm text-neutral-700 dark:text-neutral-400">
			{#await projectPromise}
				<p class="">...</p>
			{:then res}
				<a href={resolve(`/projects/${res?.id}`)} class="underline">{res?.name}</a>
			{/await}

			<ChevronRight size={18} />
			<p>Contracts</p>
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
	</div>

	{#await contractsPromise}
		<Spinner />
	{:then res}
		{#if res && res.items}
			<div class="mx-auto gap-4 overflow-y-auto lg:container">
				{#if res.items.length == 0}
					<ErrorMessage variant="info" text="Invoices/Quotes Not Found" />
				{/if}
				<div class="overflow-y-auto">
					{#if res.items.length > 0}
						<Table.Root class="container mx-auto">
							<Table.Header>
								<Table.Row>
									<Table.Head class="font-bold">Name</Table.Head>
									<Table.Head class="font-bold">Status</Table.Head>
									<Table.Head class="font-bold">Created Date</Table.Head>
									<Table.Head class="font-bold">Actions</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each res.items as contract (contract.id)}
									<Table.Row>
										<Table.Cell>{contract.name}</Table.Cell>
										<Table.Cell class="flex items-center gap-1">
											{toTitleCase(contract.status)}
											{#if contract.status == 'signed'}
												<Checks size={18} class="text-emerald-500" />
											{/if}
										</Table.Cell>

										<Table.Cell>{formatDate(contract.createdAt)}</Table.Cell>
										<Table.Cell>
											<div
												class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5
												*:hover:text-neutral-950 dark:text-neutral-400
												*:dark:hover:text-neutral-50"
											>
												<a
													href={resolve(`/contracts/${contract.id}`)}
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
			{#if res.items.length > 0}
				<div class="container mx-auto flex justify-end">
					<Pagination bind:page={pageNum} count={res.totalCount} perPage={limit} />
				</div>
			{/if}
		{/if}
	{:catch err}
		{#if err instanceof ApiError}
			<ErrorMessage variant="warn" text={err.message} retry={loadContracts} />
		{:else}
			<ErrorMessage variant="warn" text="An Error Occurred" retry={loadContracts} />
		{/if}
	{/await}
</div>
