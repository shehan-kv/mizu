<script lang="ts">
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import Envelope from 'phosphor-svelte/lib/Envelope';

	import * as Table from '$lib/components/ui/table';
	import Pagination from '../Pagination.svelte';
	import SearchBar from '../SearchBar.svelte';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import Spinner from '../Spinner.svelte';
	import { createDialogState } from './createDialogState.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import ChannelViewContract from './ChannelViewContract.svelte';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import Checks from 'phosphor-svelte/lib/Checks';
	import { formatDate } from '$lib/utils/formatDate';
	import type { Channel } from '$lib/api/messages';
	import { getContractOverviewsByProject, type ContractOverview } from '$lib/api/contracts';
	import type { PaginatedResponse } from '$lib/api/page';

	interface Props {
		open: boolean;
		channel: Channel;
	}
	let { open = $bindable(), channel }: Props = $props();

	let _q = $state('');
	let q = $state('');
	let status = $state('');

	let page = $state(1);
	let limit = $state(30);

	let selectedContract: ContractOverview | null = $state(null);
	let contractViewDialog = createDialogState();

	let contractsPromise: Promise<PaginatedResponse<ContractOverview>> | null = $state(null);

	let abortController: AbortController | null = null;
	function loadContracts() {
		if (!channel.projectId) {
			return;
		}

		if (abortController) {
			abortController.abort();
		}

		abortController = new AbortController();

		contractsPromise = getContractOverviewsByProject(
			channel.projectId,
			{ q, status, page, limit },
			abortController.signal
		);
	}

	function handleSearch() {
		// $effect automatically runs the loadContracts function when
		// q changes. This function is used as a workaround to
		// set page to 1 when a user searches for a file.
		page = 1;
		q = _q;
	}

	$effect(() => {
		if (!open) return;
		loadContracts();
	});
</script>

<FullScreenDialog bind:open>
	<div class="grid auto-rows-[min-content_1fr_min-content] gap-6 overflow-y-auto px-5">
		<div class="flex-none">
			<div class="container mx-auto flex items-end justify-between gap-4">
				<p class="font-bold">Contracts of {channel.name}</p>
				<div class="w-full max-w-xs">
					<SearchBar bind:value={_q} onchange={handleSearch} />
				</div>
			</div>
		</div>

		{#if !channel.projectId}
			<ErrorMessage variant="warn" text="Project ID Not Found" />
		{:else}
			{#await contractsPromise}
				<Spinner />
			{:then res}
				{#if res && res.items}
					<div class="overflow-y-auto">
						{#if res.items.length == 0}
							<ErrorMessage variant="info" text="Contracts Not Found" />
						{/if}
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
													<button
														onclick={() => {
															selectedContract = contract;
															contractViewDialog.open();
														}}
														title="View Contract"
													>
														<ArrowRight size={18} />
													</button>
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
					{#if res.items.length > 0}
						<div class="container mx-auto flex justify-end">
							<Pagination bind:page count={res.totalCount} perPage={res.limit} />
						</div>
					{/if}
				{/if}
			{:catch err}
				<ErrorMessage variant="warn" text={err} retry={loadContracts} />
			{/await}
		{/if}
	</div>
</FullScreenDialog>

{#if selectedContract}
	<ChannelViewContract bind:open={contractViewDialog.isOpen} contract={selectedContract} />
{/if}
