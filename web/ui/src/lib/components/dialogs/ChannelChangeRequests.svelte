<script lang="ts">
	import Checks from 'phosphor-svelte/lib/Checks';

	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import * as Table from '$lib/components/ui/table';
	import type { Channel } from '../message/types';
	import Pagination from '../Pagination.svelte';
	import SearchBar from '../SearchBar.svelte';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import { createDialogState } from './createDialogState.svelte';
	import ChannelViewTicket from './ChannelViewTicket.svelte';
	import { getChangeRequestsByProject, type ChangeRequest } from '$lib/api/changeRequest';
	import ErrorMessage from '../ErrorMessage.svelte';
	import Spinner from '../Spinner.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import { formatDate } from '$lib/utils/formatDate';

	interface Props {
		open: Boolean;
		channel: Channel;
	}
	let { open = $bindable(), channel }: Props = $props();

	const DEFAULT_LIMIT = 30;
	const DEFAULT_PAGE_NUMBER = 1;

	let _q = $state('');
	let q = $state('');
	let status = $state('');

	let page = $state(DEFAULT_PAGE_NUMBER);
	let limit = $state(DEFAULT_LIMIT);

	let selectedRequest: ChangeRequest | null = $state(null);
	let viewRequestDialog = createDialogState();

	let requestsPromise: Promise<PaginatedResponse<ChangeRequest>> | null = $state(null);

	let abortController: AbortController | null = null;
	function loadRequests() {
		if (!channel.projectId) {
			return;
		}

		if (abortController) {
			abortController.abort();
		}

		abortController = new AbortController();

		requestsPromise = getChangeRequestsByProject(
			channel.projectId,
			{ q, status, limit, page },
			abortController.signal
		);
	}

	function handleSearch() {
		// $effect automatically runs the loadFiles function when
		// q changes. This function is used as a workaround to
		// set page to 1 when a user searches for a file.
		page = 1;
		q = _q;
	}

	$effect(() => {
		if (!open) return;
		loadRequests();
	});
</script>

<FullScreenDialog bind:open>
	<div class="grid auto-rows-[min-content_1fr_min-content] gap-6 overflow-y-auto px-5">
		<div class="flex-none">
			<div class="container mx-auto flex items-end justify-between gap-4">
				<p class="font-bold">Change Requests - {channel?.name}</p>
				<div class="w-full max-w-xs">
					<SearchBar bind:value={_q} onchange={handleSearch} />
				</div>
			</div>
		</div>

		{#if !channel.projectId}
			<ErrorMessage variant="warn" text="Project ID Not Found" />
		{:else}
			{#await requestsPromise}
				<Spinner />
			{:then res}
				{#if res && res.data}
					<div class="overflow-y-auto">
						{#if res.data.length == 0}
							<ErrorMessage variant="info" text="Contracts Not Found" />
						{/if}
						{#if res.data.length > 0}
							<Table.Root class="container mx-auto">
								<Table.Header>
									<Table.Row>
										<Table.Head class="font-bold">Subject</Table.Head>
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
											<Table.Cell class="flex items-center gap-1">
												{toTitleCase(req.status)}
												{#if req.status == 'closed'}
													<Checks size={18} class="text-emerald-500" />
												{/if}
											</Table.Cell>
											<Table.Cell>{formatDate(req.createdAt)}</Table.Cell>
											<Table.Cell>{req.requestedBy.firstName} {req.requestedBy.lastName}</Table.Cell
											>
											<Table.Cell>
												<div
													class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5 *:hover:text-neutral-950 dark:text-neutral-400 *:dark:hover:text-neutral-50"
												>
													<button
														title="View"
														onclick={() => {
															selectedRequest = req;
															viewRequestDialog.open();
														}}
													>
														<ArrowRight size={16} />
													</button>
												</div>
											</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						{/if}
					</div>
					{#if res.data.length > 0}
						<div class="container mx-auto flex justify-end">
							<Pagination bind:page count={res.count} perPage={res.limit} />
						</div>
					{/if}
				{/if}
			{:catch err}
				{#if err instanceof APIBadRequestError}
					<ErrorMessage variant="warn" text="Invalid Request" retry={loadRequests} />
				{:else if err instanceof APIForbiddenError}
					<ErrorMessage
						variant="warn"
						text="You Don't Have Permission To View These Files"
						retry={loadRequests}
					/>
				{:else if err instanceof APINotFoundError}
					<ErrorMessage variant="info" text="Not Found" retry={loadRequests} />
				{:else if err instanceof APIServerError}
					<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadRequests} />
				{:else}
					<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadRequests} />
				{/if}
			{/await}
		{/if}
	</div>
</FullScreenDialog>

<ChannelViewTicket
	bind:open={viewRequestDialog.isOpen}
	close={viewRequestDialog.close}
	ticket={selectedRequest}
/>
