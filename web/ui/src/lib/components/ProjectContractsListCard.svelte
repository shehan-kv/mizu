<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Dialog from '$lib/components/dialogs';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Spinner from './Spinner.svelte';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import Checks from 'phosphor-svelte/lib/Checks';
	import { formatDate } from '$lib/utils/formatDate';
	import ErrorMessage from './ErrorMessage.svelte';

	import { getContractOverviewsByProject, type ContractOverview } from '$lib/api/contracts';
	import { onMount } from 'svelte';
	import type { UserRole } from '$lib/api/users';
	import DotsThree from 'phosphor-svelte/lib/DotsThree';
	import { createDialogState } from './dialogs/createDialogState.svelte';
	import type { PaginatedResponse } from '$lib/api/page';
	import { resolve } from '$app/paths';
	import { ApiError } from '$lib/api/client';

	interface Props {
		projectId: string;
		role?: UserRole;
	}
	let { projectId, role = 'client' }: Props = $props();

	let contractsPromise: Promise<PaginatedResponse<ContractOverview>> | null = $state(null);
	let abort: AbortController | null = null;
	function loadContracts() {
		if (abort) {
			abort.abort();
		}
		abort = new AbortController();

		contractsPromise = getContractOverviewsByProject(
			projectId,
			{ page: 1, limit: 20 },
			abort.signal
		);
	}

	export function refresh() {
		loadContracts();
	}

	onMount(() => {
		loadContracts();
	});

	let revisionDialog = createDialogState();
	let signDialog = createDialogState();
	let rejectDialog = createDialogState();

	type ActionsAllowed = 'sign' | 'reject';
	type SelectedContract = ContractOverview & { action?: ActionsAllowed };
	let selectedContract: SelectedContract | null = $state(null);

	function openStatusDialog(contract: ContractOverview, action: ActionsAllowed) {
		selectedContract = { ...contract, action };

		if (action == 'sign') {
			signDialog.open();
		} else if (action == 'reject') {
			rejectDialog.open();
		}
	}

	function openRevisionDialog(contract: ContractOverview) {
		selectedContract = contract;
		revisionDialog.open();
	}

	// svelte-ignore non_reactive_update
	let linksPrefix = '';
	if (role == 'administrator') linksPrefix = '/admin';
	if (role == 'staff') linksPrefix = '/staff';
</script>

<div class="grid h-full w-full grid-rows-[min-content_1fr] overflow-hidden">
	<div class="flex items-center justify-between border-b px-6 py-2">
		<p class="text-sm">Contracts</p>
		<a
			href={`${linksPrefix}/projects/${projectId}/contracts`}
			class="flex items-center gap-1 text-sm"
		>
			<span>View All</span>
			<ArrowRight />
		</a>
	</div>

	<div class="overflow-scroll px-6 py-2">
		{#await contractsPromise}
			<Spinner />
		{:then res}
			{#if res && res.items.length > 0}
				<Table.Root>
					<Table.Body>
						{#each res.items as contract (contract)}
							<Table.Row
								class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
								dark:text-neutral-400 dark:hover:text-neutral-50"
							>
								<Table.Cell class="pl-0">
									{contract.name}
								</Table.Cell>
								<Table.Cell>
									<div class="flex items-center gap-1">
										{toTitleCaseDashed(contract.status)}
										{#if contract.status == 'signed'}
											<Checks size={18} class="text-emerald-500" />
										{/if}
									</div>
								</Table.Cell>
								<Table.Cell>
									Created On {formatDate(contract.createdAt)}
								</Table.Cell>
								<Table.Cell class="pr-0" align="right">
									<div
										class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5
										*:hover:text-neutral-950 dark:text-neutral-400
										*:dark:hover:text-neutral-50"
									>
										<a
											href={resolve(`/admin/contracts/${contract.id}`)}
											class="inline-block"
											title="View"
										>
											<ArrowRight size={18} />
										</a>

										<DropdownMenu.Root>
											<DropdownMenu.Trigger
												class="cursor-pointer p-1 hover:text-neutral-950 dark:hover:text-neutral-50"
											>
												<DotsThree size={18} />
											</DropdownMenu.Trigger>
											<DropdownMenu.Content class="mr-4 *:text-xs">
												{#if contract.status == 'pending'}
													{#if contract.memberSignatoryStatus != 'signed' && contract.memberSignatoryStatus != 'rejected'}
														<DropdownMenu.Item
															class="pl-4 text-xs"
															onclick={() => openStatusDialog(contract, 'sign')}
														>
															Sign
														</DropdownMenu.Item>

														<DropdownMenu.Item
															class="pl-4 text-xs"
															onclick={() => openStatusDialog(contract, 'reject')}
														>
															Reject
														</DropdownMenu.Item>
													{:else}
														<div class="flex items-center gap-2 px-2 py-1.5 text-xs">
															<Checks /> You've Already {toTitleCaseDashed(
																contract.memberSignatoryStatus
															)}
														</div>
													{/if}

													{#if role == 'client' && contract.memberSignatoryStatus != 'rejected' && contract.memberSignatoryStatus != 'signed'}
														<DropdownMenu.Item
															class="pl-4 text-xs"
															onclick={() => openRevisionDialog(contract)}
														>
															Request Revision
														</DropdownMenu.Item>
													{/if}
												{:else}
													<div class="flex items-center gap-2 px-2 py-1.5 text-xs">
														<Checks /> Already {toTitleCaseDashed(contract.status)}
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
			{:else}
				<ErrorMessage variant="warn" text="Contracts Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof ApiError}
				<ErrorMessage variant="warn" text={err.message} retry={loadContracts} />
			{:else}
				<ErrorMessage variant="warn" text="An Error Occurred" retry={loadContracts} />
			{/if}
		{/await}
	</div>
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
