<script lang="ts">
	import * as Dialog from '$lib/components/dialogs';
	import * as Table from '$lib/components/ui/table';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
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
	import { getChangeRequestsByProject, type ChangeRequest } from '$lib/api/changeRequest';
	import { onMount } from 'svelte';
	import type { UserRole } from '$lib/api/users';
	import DotsThree from 'phosphor-svelte/lib/DotsThree';
	import { createDialogState } from './dialogs/createDialogState.svelte';

	interface Props {
		projectId: number;
		role?: UserRole;
	}
	let { projectId, role = 'client' }: Props = $props();

	let reqPromise: Promise<PaginatedResponse<ChangeRequest>> | null = $state(null);
	let abort: AbortController | null = null;
	function loadReqs() {
		if (abort) {
			abort.abort();
		}
		abort = new AbortController();

		reqPromise = getChangeRequestsByProject(projectId, { page: 1, limit: 20 }, abort.signal);
	}

	export function refresh() {
		loadReqs();
	}

	onMount(() => {
		loadReqs();
	});

	type ActionsAllowed = 'closed';
	type SelectedReq = ChangeRequest & { action: ActionsAllowed };
	let selectedReq: SelectedReq | null = $state(null);
	let setStatusDialog = createDialogState();
	function openStatusDialog(req: ChangeRequest, action: ActionsAllowed) {
		selectedReq = { ...req, action };
		setStatusDialog.open();
	}

	// svelte-ignore non_reactive_update
	let linksPrefix = '';
	if (role == 'admin') linksPrefix = '/admin';
	if (role == 'staff') linksPrefix = '/staff';
</script>

<div class="grid h-full w-full grid-rows-[min-content_1fr] overflow-hidden rounded border">
	<div class="flex items-center justify-between bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
		<p class="text-sm">Change Requests</p>
		<a
			href={`${linksPrefix}/projects/${projectId}/change-requests`}
			class="flex items-center gap-1 text-sm"
		>
			<span>View All</span>
			<ArrowRight />
		</a>
	</div>
	<div class="overflow-scroll px-6 py-2">
		{#await reqPromise}
			<Spinner />
		{:then res}
			{#if res && res.data.length > 0}
				<Table.Root>
					<Table.Body>
						{#each res.data as req (req)}
							<Table.Row
								class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
								dark:text-neutral-400 dark:hover:text-neutral-50"
							>
								<Table.Cell class="pl-0">
									{req.title}
								</Table.Cell>
								<Table.Cell class="flex items-center gap-1">
									{toTitleCase(req.status)}
									{#if req.status == 'closed'}
										<Checks size={18} class="text-emerald-500" />
									{/if}
								</Table.Cell>
								<Table.Cell>
									Started By {req.requestedBy.firstName}
									{req.requestedBy.lastName}
								</Table.Cell>
								<Table.Cell>
									Created On {formatDate(req.createdAt)}
								</Table.Cell>
								<Table.Cell class="pr-0" align="right">
									<div
										class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5
										*:hover:text-neutral-950 dark:text-neutral-400
										*:dark:hover:text-neutral-50"
									>
										<a
											href={`${linksPrefix}/change-requests/${req.id}`}
											class="inline-block"
											title="View"
										>
											<ArrowRight size={18} />
										</a>

										{#if role == 'admin' || role == 'staff'}
											<DropdownMenu.Root>
												<DropdownMenu.Trigger
													class="cursor-pointer p-1 hover:text-neutral-950 dark:hover:text-neutral-50"
												>
													<DotsThree size={18} />
												</DropdownMenu.Trigger>
												<DropdownMenu.Content class="mr-4 *:text-xs">
													{#if req.status != 'closed'}
														<DropdownMenu.Group class="text-xs">
															<DropdownMenu.Label class="text-xs">Mark As</DropdownMenu.Label>
															<DropdownMenu.Item
																class="pl-4 text-xs"
																onclick={() => openStatusDialog(req, 'closed')}
															>
																Closed
															</DropdownMenu.Item>
														</DropdownMenu.Group>
													{:else}
														<div class="flex items-center gap-2 px-2 py-1.5 text-xs">
															<Checks /> Already {toTitleCase(req.status)}
														</div>
													{/if}
												</DropdownMenu.Content>
											</DropdownMenu.Root>
										{/if}
									</div>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{:else}
				<ErrorMessage variant="warn" text="Change Requests Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof APIBadRequestError}
				<ErrorMessage variant="warn" text="Invalid Request" retry={loadReqs} />
			{:else if err instanceof APIForbiddenError}
				<ErrorMessage
					variant="warn"
					text="You Don't Have Permission To View Change Requests"
					retry={loadReqs}
				/>
			{:else if err instanceof APINotFoundError}
				<ErrorMessage variant="info" text="Not Found" retry={loadReqs} />
			{:else if err instanceof APIServerError}
				<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadReqs} />
			{:else}
				<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadReqs} />
			{/if}
		{/await}
	</div>
</div>

{#if selectedReq}
	<Dialog.ChangeReqStatusConfirm
		bind:open={setStatusDialog.isOpen}
		requestId={selectedReq.id}
		status={selectedReq.action}
		onSuccess={loadReqs}
	/>
{/if}
