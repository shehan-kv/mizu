<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Dialog from '$lib/components/dialogs';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Spinner from './Spinner.svelte';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import Checks from 'phosphor-svelte/lib/Checks';
	import { formatDate } from '$lib/utils/formatDate';
	import ErrorMessage from './ErrorMessage.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { getContractsByProject, type Contract } from '$lib/api/contracts';
	import { onMount } from 'svelte';
	import type { UserRole } from '$lib/api/users';
	import DotsThree from 'phosphor-svelte/lib/DotsThree';
	import { createDialogState } from './dialogs/createDialogState.svelte';

	interface Props {
		projectId: number;
		role?: UserRole;
	}
	let { projectId, role = 'client' }: Props = $props();

	let contractsPromise: Promise<PaginatedResponse<Contract>> | null = $state(null);
	let abort: AbortController | null = null;
	function loadContracts() {
		if (abort) {
			abort.abort();
		}
		abort = new AbortController();

		contractsPromise = getContractsByProject(projectId, '', '', 1, 20, abort.signal);
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

	function openRevisionDialog(contract: Contract) {
		selectedContract = contract;
		revisionDialog.open();
	}

	// svelte-ignore non_reactive_update
	let linksPrefix = '';
	if (role == 'admin') linksPrefix = '/admin';
	if (role == 'staff') linksPrefix = '/staff';
</script>

<div class="grid h-full w-full grid-rows-[min-content_1fr] overflow-hidden rounded border">
	<div class="flex items-center justify-between bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
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
			{#if res && res.data.length > 0}
				<Table.Root>
					<Table.Body>
						{#each res.data as contract (contract)}
							<Table.Row
								class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
								dark:text-neutral-400 dark:hover:text-neutral-50"
							>
								<Table.Cell class="pl-0">
									{contract.name}
								</Table.Cell>
								<Table.Cell>
									<div class="flex items-center gap-1">
										{toTitleCase(contract.status)}
										{#if contract.status == 'signed'}
											<Checks size={18} class="text-emerald-500" />
										{/if}
									</div>
								</Table.Cell>
								<Table.Cell>
									Created On {formatDate(contract.createdAt)}
								</Table.Cell>
								<Table.Cell>
									{contract.versions}
									{contract.versions == 1 ? 'Version' : 'Versions'}
									(Latest {contract.latestVersion.version})
								</Table.Cell>
								<Table.Cell class="pr-0" align="right">
									<div
										class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5
										*:hover:text-neutral-950 dark:text-neutral-400
										*:dark:hover:text-neutral-50"
									>
										<a href={`/admin/contracts/${contract.id}`} class="inline-block" title="View">
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
													{#if contract.userSignature != 'signed' && contract.userSignature != 'rejected'}
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
															<Checks /> You've Already {toTitleCase(contract.userSignature)}
														</div>
													{/if}

													{#if role == 'client' && contract.userSignature != 'rejected' && contract.userSignature != 'signed'}
														<DropdownMenu.Item
															class="pl-4 text-xs"
															onclick={() => openRevisionDialog(contract)}
														>
															Request Revision
														</DropdownMenu.Item>
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
			{:else}
				<ErrorMessage variant="warn" text="Contracts Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof APIBadRequestError}
				<ErrorMessage variant="warn" text="Invalid Request" retry={loadContracts} />
			{:else if err instanceof APIForbiddenError}
				<ErrorMessage
					variant="warn"
					text="You Don't Have Permission To View Contracts"
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
</div>

{#if selectedContract}
	<Dialog.ConfirmSignContract
		bind:open={signDialog.isOpen}
		versionId={selectedContract.latestVersion.id}
		onSuccess={loadContracts}
	/>

	<Dialog.ConfirmRejectContract
		bind:open={rejectDialog.isOpen}
		versionId={selectedContract.latestVersion.id}
		onSuccess={loadContracts}
	/>

	<Dialog.RequestRevision
		contractId={selectedContract.id}
		bind:open={revisionDialog.isOpen}
		onSuccess={loadContracts}
	/>
{/if}
