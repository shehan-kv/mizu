<script lang="ts">
	import Checks from 'phosphor-svelte/lib/Checks';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import * as Dialog from '$lib/components/dialogs';
	import * as Table from '$lib/components/ui/table';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getChangeRequests, type ChangeRequest } from '$lib/api/changeRequest';
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
	import DotsThree from 'phosphor-svelte/lib/DotsThree';
	import { createDialogState } from '$lib/components/dialogs/createDialogState.svelte';

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

	let chReqPromise: Promise<PaginatedResponse<ChangeRequest>> | null = $state(null);
	let abortController: AbortController | null = null;
	function loadReqs() {
		if (abortController) {
			abortController.abort();
		}

		abortController = new AbortController();

		chReqPromise = getChangeRequests(q, status, pageNum, limit, abortController.signal);
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

	const role = 'admin';
</script>

<svelte:head>
	<title>Change Requests</title>
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

	{#await chReqPromise}
		<Spinner />
	{:then res}
		{#if res && res.data}
			<div class="mx-auto gap-4 overflow-y-auto lg:container">
				{#if res.data.length == 0}
					<ErrorMessage variant="info" text="Change Requests Not Found" />
				{/if}
				<div class="overflow-y-auto">
					{#if res.data.length > 0}
						<Table.Root class="container mx-auto">
							<Table.Header>
								<Table.Row>
									<Table.Head class="font-bold">Subject</Table.Head>
									<Table.Head class="font-bold">Project</Table.Head>
									<Table.Head class="font-bold">Status</Table.Head>
									<Table.Head class="font-bold">Created At</Table.Head>
									<Table.Head class="font-bold">Created By</Table.Head>
									<Table.Head class="font-bold">Actions</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each res.data as req}
									<Table.Row>
										<Table.Cell>{req.title}</Table.Cell>
										<Table.Cell>{req.project.name}</Table.Cell>
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
												href={`/admin/change-requests/${req.id}`}
												class="inline-block cursor-pointer px-1.5 text-xs
												text-neutral-500 hover:text-neutral-950 dark:text-neutral-400
												dark:hover:text-neutral-50"
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
			<ErrorMessage variant="warn" text="Invalid Request" retry={loadReqs} />
		{:else if err instanceof APIForbiddenError}
			<ErrorMessage
				variant="warn"
				text="You Don't Have Permission To View These Change Requests"
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

{#if selectedReq}
	<Dialog.ChangeReqStatusConfirm
		bind:open={setStatusDialog.isOpen}
		requestId={selectedReq.id}
		status={selectedReq.action}
		onSuccess={loadReqs}
	/>
{/if}
