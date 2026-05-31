<script lang="ts">
	import Checks from 'phosphor-svelte/lib/Checks';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Envelope from 'phosphor-svelte/lib/Envelope';
	import * as Table from '$lib/components/ui/table';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Dialog from '$lib/components/dialogs';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import Spinner from '$lib/components/Spinner.svelte';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import FilterSelect from '$lib/components/FilterSelect.svelte';
	import FilterInput from '$lib/components/FilterInput.svelte';
	import { formatDate } from '$lib/utils/formatDate';
	import { getContractOverviews, type Contract, type ContractOverview } from '$lib/api/contracts';
	import DotsThree from 'phosphor-svelte/lib/DotsThree';
	import { createDialogState } from '$lib/components/dialogs/createDialogState.svelte';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import type { PaginatedResponse } from '$lib/api/page';
	import { ApiError } from '$lib/api/client';
	import { CONTRACT_STATUS } from '$lib/constants/contract';

	const MAX_LIMIT = 100;
	const MIN_LIMIT = 1;
	const DEFAULT_LIMIT = 30;
	const DEFAULT_PAGE_NUMBER = 1;

	const params = new SvelteURLSearchParams(page.url.searchParams.toString());

	let q = $state(params.get('q') || '');
	let status = $state(CONTRACT_STATUS.find((s) => s === params.get('status')) ?? '');
	let pageNum = $state(Number(params.get('page')) || DEFAULT_PAGE_NUMBER);
	let limit = $state(Math.min(Number(params.get('limit')) || DEFAULT_LIMIT, MAX_LIMIT));

	let contractsPromise: Promise<PaginatedResponse<ContractOverview>> | null = $state(null);

	let abort: AbortController | null = null;
	function loadContracts() {
		if (abort) {
			abort.abort();
		}

		abort = new AbortController();

		contractsPromise = getContractOverviews({ q, status, page: pageNum, limit }, abort.signal);
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

	let signDialog = createDialogState();
	let rejectDialog = createDialogState();

	type ActionsAllowed = 'sign' | 'reject';
	type SelectedContract = Contract & { action?: ActionsAllowed };
	let selectedContract: SelectedContract | null = $state(null);

	function openStatusDialog(contract: Contract, action: ActionsAllowed) {
		selectedContract = { ...contract, action };

		if (action == 'sign') {
			signDialog.open();
		} else if (action == 'reject') {
			rejectDialog.open();
		}
	}
</script>

<svelte:head>
	<title>Contracts</title>
</svelte:head>

<div class="grid h-full auto-rows-[min-content_1fr_min-content] gap-6">
	<div class="mx-auto flex gap-2 lg:container">
		<div class="max-w-96">
			<SearchBar bind:value={q} onchange={handleFilter} />
		</div>
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

	{#await contractsPromise}
		<Spinner />
	{:then res}
		{#if res && res.items}
			<div class="mx-auto gap-4 overflow-y-auto lg:container">
				{#if res.items.length == 0}
					<ErrorMessage variant="info" text="Contracts Not Found" />
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
												class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5 *:hover:text-neutral-950 dark:text-neutral-400 *:dark:hover:text-neutral-50"
											>
												<a
													href={resolve(`/admin/contracts/${contract.id}`)}
													title="View Contract"
													class="inline-block"
												>
													<ArrowRight size={18} />
												</a>
												<button title="Download the Latest Version as PDF">
													<DownloadSimple size={18} />
												</button>
												<button title="Email Me"><Envelope size={18} /></button>

												<DropdownMenu.Root>
													<DropdownMenu.Trigger
														class="cursor-pointer p-1 hover:text-neutral-950 dark:hover:text-neutral-50"
													>
														<DotsThree size={18} />
													</DropdownMenu.Trigger>
													<DropdownMenu.Content class="mr-4 *:text-xs">
														{#if contract.status == 'pending'}
															{#if contract.userSignature != 'signed' && contract.userSignature != 'rejected'}
																<DropdownMenu.Item
																	class="pl-4 text-xs"
																	onclick={() => openStatusDialog(contract, 'sign')}
																>
																	Sign Latest Version
																</DropdownMenu.Item>

																<DropdownMenu.Item
																	class="pl-4 text-xs"
																	onclick={() => openStatusDialog(contract, 'reject')}
																>
																	Reject Latest Version
																</DropdownMenu.Item>
															{:else}
																<div class="flex items-center gap-2 px-2 py-1.5 text-xs">
																	<Checks /> You've Already {toTitleCase(contract.userSignature)}
																</div>
															{/if}
														{:else}
															<div class="flex items-center gap-2 px-2 py-1.5 text-xs">
																<Checks /> Already {toTitleCase(contract.status)}
															</div>
														{/if}
													</DropdownMenu.Content>
												</DropdownMenu.Root>
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

{#if selectedContract}
	<Dialog.ConfirmSignContract
		bind:open={signDialog.isOpen}
		contractId={selectedContract.id}
		onSuccess={loadContracts}
	/>

	<Dialog.ConfirmRejectContract
		bind:open={rejectDialog.isOpen}
		contractId={selectedContract.id}
		onSuccess={loadContracts}
	/>
{/if}
