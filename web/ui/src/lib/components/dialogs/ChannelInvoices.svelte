<script lang="ts">
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import Envelope from 'phosphor-svelte/lib/Envelope';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Checks from 'phosphor-svelte/lib/Checks';

	import * as Table from '$lib/components/ui/table';
	import type { Channel } from '../message/types';
	import Pagination from '../Pagination.svelte';
	import SearchBar from '../SearchBar.svelte';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import { createDialogState } from './createDialogState.svelte';
	import { getInvoicesByProjectId, type InvoiceWithStatus } from '$lib/api/invoices';
	import ErrorMessage from '../ErrorMessage.svelte';
	import Spinner from '../Spinner.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { toTitleCase } from '$lib/utils/toTitleCase';

	interface Props {
		open: boolean;
		channel: Channel;
	}
	let { open = $bindable(), channel }: Props = $props();

	let page = $state(1);
	let limit = $state(30);

	// For the invoice details dialog
	let selectedInvoice: InvoiceWithStatus | null = $state(null);
	let invoiceViewDialog = createDialogState();

	let invoicePromise: Promise<PaginatedResponse<InvoiceWithStatus>> | null = $state(null);

	let abortController: AbortController | null = null;
	function loadInvoices() {
		if (!channel.projectId) {
			return;
		}

		if (abortController) {
			abortController.abort();
		}
		abortController = new AbortController();

		invoicePromise = getInvoicesByProjectId(
			channel.projectId,
			{ page, limit },
			abortController.signal
		);
	}

	$effect(() => {
		if (!open) return;
		loadInvoices();
	});
</script>

<FullScreenDialog bind:open>
	<div class="grid auto-rows-[min-content_1fr_min-content] gap-6 overflow-y-auto px-5">
		<div class="flex-none">
			<div class="container mx-auto">
				<p class="font-bold">Invoices & Quotes - {channel?.name}</p>
			</div>
		</div>

		{#if !channel.projectId}
			<ErrorMessage variant="warn" text="Project ID Not Found" />
		{:else}
			{#await invoicePromise}
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
										<Table.Head class="font-bold">Type</Table.Head>
										<Table.Head class="font-bold">#ID</Table.Head>
										<Table.Head class="font-bold">Total</Table.Head>
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
											<Table.Cell>{currencyFormatter('USD', invoice.total)}</Table.Cell>
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
					{#if res.data.length > 0}
						<div class="container mx-auto flex justify-end">
							<Pagination bind:page count={res.count} perPage={res.limit} />
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
		{/if}
	</div>
</FullScreenDialog>
