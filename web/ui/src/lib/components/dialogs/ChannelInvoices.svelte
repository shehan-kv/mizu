<script lang="ts">
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import Envelope from 'phosphor-svelte/lib/Envelope';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Checks from 'phosphor-svelte/lib/Checks';

	import * as Table from '$lib/components/ui/table';
	import Pagination from '../Pagination.svelte';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import { createDialogState } from './createDialogState.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import Spinner from '../Spinner.svelte';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import { formatDate } from '$lib/utils/formatDate';
	import type { Channel } from '$lib/api/messages';
	import ChannelViewInvoice from './ChannelViewInvoice.svelte';
	import { getInvoicesByProject, type InvoiceOverview } from '$lib/api/invoices';
	import type { PaginatedResponse } from '$lib/api/page';

	interface Props {
		open: boolean;
		channel: Channel;
	}
	let { open = $bindable(), channel }: Props = $props();

	const DEFAULT_PAGE = 1;
	const DEFAULT_LIMIT = 25;

	let page = $state(DEFAULT_PAGE);
	let limit = $state(DEFAULT_LIMIT);

	// For the invoice details dialog
	let selectedInvoice: InvoiceOverview | null = $state(null);
	let invoiceViewDialog = createDialogState();

	let invoicePromise: Promise<PaginatedResponse<InvoiceOverview>> | null = $state(null);

	let abortController: AbortController | null = null;
	function loadInvoices() {
		if (!channel.projectId) {
			return;
		}

		if (abortController) {
			abortController.abort();
		}
		abortController = new AbortController();

		invoicePromise = getInvoicesByProject(
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
				<p class="font-bold">Invoices & Quotes - {channel.name}</p>
			</div>
		</div>

		{#if !channel.projectId}
			<ErrorMessage variant="warn" text="Project ID Not Found" />
		{:else}
			{#await invoicePromise}
				<Spinner />
			{:then res}
				{#if res && res.items}
					<div class="overflow-y-auto">
						{#if res.items.length == 0}
							<ErrorMessage variant="info" text="Invoices/Quotes Not Found" />
						{/if}
						{#if res.items.length > 0}
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
									{#each res.items as invoice (invoice.id)}
										<Table.Row>
											<Table.Cell>{invoice.isInvoice ? 'Invoice' : 'Quote'}</Table.Cell>
											<Table.Cell>
												#{invoice.id.replaceAll('-', '').slice(-8).toUpperCase()}</Table.Cell
											>
											<Table.Cell>{currencyFormatter('USD', invoice.subTotal)}</Table.Cell>
											<Table.Cell class="flex items-center gap-1">
												{toTitleCaseDashed(invoice.status)}
												{#if invoice.status == 'paid' || invoice.status == 'accepted'}
													<Checks size={18} class="text-emerald-500" />
												{/if}
											</Table.Cell>
											<Table.Cell>{formatDate(invoice.createdAt)}</Table.Cell>
											<Table.Cell>
												{invoice.dueAt ? formatDate(invoice.dueAt) : 'N/A'}
											</Table.Cell>
											<Table.Cell>
												<div
													class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5 *:hover:text-neutral-950 dark:text-neutral-400 *:dark:hover:text-neutral-50"
												>
													<button
														title="View"
														onclick={() => {
															selectedInvoice = invoice;
															invoiceViewDialog.open();
														}}
													>
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
					{#if res.items.length > 0}
						<div class="container mx-auto flex justify-end">
							<Pagination bind:page count={res.totalCount} perPage={res.limit} />
						</div>
					{/if}
				{/if}
			{:catch err}
				<ErrorMessage variant="warn" text={err} retry={loadInvoices} />
			{/await}
		{/if}
	</div>
</FullScreenDialog>

{#if selectedInvoice}
	<ChannelViewInvoice
		bind:open={invoiceViewDialog.isOpen}
		invoiceId={selectedInvoice.id}
		isInvoice={selectedInvoice.isInvoice}
	/>
{/if}
