<script lang="ts">
	import { getInvoice, type Invoice, type InvoiceStatus } from '$lib/api/invoices';
	import * as Table from '$lib/components/ui/table';
	import * as Dialog from '$lib/components/dialogs';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import { formatDate } from '$lib/utils/formatDate';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import Note from 'phosphor-svelte/lib/Note';

	import Spinner from './Spinner.svelte';
	import ErrorMessage from './ErrorMessage.svelte';
	import { onMount } from 'svelte';
	import Checks from 'phosphor-svelte/lib/Checks';
	import WarningCircle from 'phosphor-svelte/lib/WarningCircle';
	import { createDialogState } from './dialogs/createDialogState.svelte';
	import type { UserRole } from '$lib/api/users';
	import { ApiError } from '$lib/api/client';

	interface Props {
		invoiceId: string;
		role?: UserRole;
		onLoad?: (invoice: Invoice) => unknown;
	}

	let { invoiceId, role = 'client', onLoad }: Props = $props();

	let invoicePromise: Promise<Invoice> | null = $state(null);
	let abortController: AbortController | null = null;

	function loadInvoice() {
		if (abortController) {
			abortController.abort();
		}
		abortController = new AbortController();

		invoicePromise = getInvoice(invoiceId, abortController.signal).then((res) => {
			onLoad?.(res);
			return res;
		});
	}

	let setStatusDialog = createDialogState();
	type Status = Exclude<InvoiceStatus, 'pending'>;
	let selectedStatus: Status | null = $state(null);

	function openStatusDialog(status: Status) {
		selectedStatus = status;
		setStatusDialog.open();
	}

	onMount(() => {
		loadInvoice();
	});
</script>

{#await invoicePromise}
	<Spinner />
{:then invoice}
	{#if invoice}
		<div class="flex justify-between">
			<div class="flex gap-20">
				<div class="space-y-2">
					<div>
						<p class="text-xs text-neutral-500">Id</p>
						<p>
							{invoice.isInvoice ? 'Invoice' : 'Quote'} #{invoice.id
								.replaceAll('-', '')
								.slice(-8)
								.toUpperCase()}
						</p>
					</div>
					<div>
						<p class="text-xs text-neutral-500">Project</p>
						<p>{invoice.projectName}</p>
					</div>
					<div>
						<p class="text-xs text-neutral-500">Status</p>
						<p>{toTitleCaseDashed(invoice.status)}</p>
					</div>
					<div>
						<p class="text-xs text-neutral-500">Issued At</p>
						<p>{formatDate(invoice.createdAt)}</p>
					</div>
					<div>
						<p class="text-xs text-neutral-500">Due At</p>
						<p>{invoice.dueAt ? formatDate(invoice.dueAt) : 'N/A'}</p>
					</div>
				</div>
				<div class=" space-y-2">
					<div>
						<p class="text-xs text-neutral-500">Currency Code</p>
						<p>{invoice.currencyCode.toLocaleUpperCase()}</p>
					</div>
					<div>
						<p class="text-xs text-neutral-500">Discount</p>
						<p>{currencyFormatter(invoice.currencyCode, invoice.totalDiscount)}</p>
					</div>
					<div>
						<p class="text-xs text-neutral-500">Tax</p>
						<p>{currencyFormatter(invoice.currencyCode, invoice.totalTax)}</p>
					</div>
					<div>
						<p class="text-xs text-neutral-500">Total</p>
						<p>{currencyFormatter(invoice.currencyCode, invoice.subTotal)}</p>
					</div>
				</div>
			</div>
			{#if role == 'client' && invoice.status == 'pending'}
				<div class="space-x-1">
					<button
						onclick={() => openStatusDialog('accepted')}
						class="inline-flex cursor-pointer items-center gap-2
						rounded bg-neutral-950 px-4 py-3 text-xs text-neutral-50 transition
						hover:bg-neutral-800 dark:bg-neutral-100 dark:text-neutral-950
						dark:hover:bg-neutral-300"
					>
						Accept <Checks size={16} />
					</button>
					<button
						onclick={() => openStatusDialog('rejected')}
						class="inline-flex cursor-pointer items-center gap-2
						rounded bg-neutral-100 px-4 py-3 text-xs text-neutral-950 transition
						hover:bg-neutral-200 dark:bg-neutral-900 dark:text-neutral-100
						dark:hover:bg-neutral-800"
					>
						Reject <WarningCircle size={16} />
					</button>
				</div>
			{:else if invoice.status == 'pending'}
				<div class="space-x-1">
					<button
						onclick={() => openStatusDialog('paid')}
						class="inline-flex cursor-pointer items-center gap-2
						rounded bg-neutral-950 px-4 py-3 text-xs text-neutral-50 transition
						hover:bg-neutral-800 dark:bg-neutral-100 dark:text-neutral-950
						dark:hover:bg-neutral-300"
					>
						Mark As Paid <Checks size={16} />
					</button>
					<button
						onclick={() => openStatusDialog('cancelled')}
						class="inline-flex cursor-pointer items-center gap-2
						rounded bg-neutral-100 px-4 py-3 text-xs text-neutral-950 transition
						hover:bg-neutral-200 dark:bg-neutral-900 dark:text-neutral-100
						dark:hover:bg-neutral-800"
					>
						Cancel <WarningCircle size={16} />
					</button>
				</div>
			{/if}
		</div>
		<div class="mt-16">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head class="font-bold">Description</Table.Head>
						<Table.Head class="font-bold">QTY</Table.Head>
						<Table.Head class="font-bold">Unit Price</Table.Head>
						<Table.Head class="font-bold">Unit Discount</Table.Head>
						<Table.Head class="font-bold">Unit Tax</Table.Head>
						<Table.Head class="font-bold">Total Discount</Table.Head>
						<Table.Head class="font-bold">Total Tax</Table.Head>
						<Table.Head class="text-right font-bold">Item Total</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each invoice.items as item (item)}
						<Table.Row>
							<Table.Cell>{item.description}</Table.Cell>
							<Table.Cell>{item.qty}</Table.Cell>
							<Table.Cell>
								{currencyFormatter(invoice.currencyCode, item.unitPrice)}
							</Table.Cell>
							<Table.Cell>
								{#if item.discountType == 'percentage'}
									{item.discountRate}%
								{:else}
									{currencyFormatter(invoice.currencyCode, item.discountRate)}
								{/if}
							</Table.Cell>
							<Table.Cell>
								{#if item.taxType == 'percentage'}
									{item.taxRate}%
								{:else}
									{currencyFormatter(invoice.currencyCode, item.taxRate)}
								{/if}
							</Table.Cell>
							<Table.Cell>
								{currencyFormatter(invoice.currencyCode, item.lineDiscount)}
							</Table.Cell>
							<Table.Cell>
								{currencyFormatter(invoice.currencyCode, item.lineTax)}
							</Table.Cell>
							<Table.Cell align="right">
								{currencyFormatter(invoice.currencyCode, item.lineTotal)}
							</Table.Cell>
						</Table.Row>
					{/each}
					<Table.Row>
						<Table.Cell colspan={7} align="right" class="border-r">Total Discount</Table.Cell>
						<Table.Cell align="right">
							{currencyFormatter(invoice.currencyCode, invoice.totalDiscount)}
						</Table.Cell>
					</Table.Row>
					<Table.Row>
						<Table.Cell colspan={7} align="right" class="border-r">Total Tax</Table.Cell>
						<Table.Cell align="right">
							{currencyFormatter(invoice.currencyCode, invoice.totalTax)}
						</Table.Cell>
					</Table.Row>
					<Table.Row class="bg-neutral-100 dark:bg-neutral-900">
						<Table.Cell colspan={7} align="right" class="border-r">Invoice Total</Table.Cell>
						<Table.Cell align="right">
							{currencyFormatter(invoice.currencyCode, invoice.subTotal)}
						</Table.Cell>
					</Table.Row>
				</Table.Body>
			</Table.Root>
		</div>

		<div class="mt-8">
			<p class="inline-flex items-center gap-1"><Note size={20} /> Note</p>
			<p class="max-w-xl whitespace-break-spaces">{invoice.note || 'N/A'}</p>
		</div>
	{:else}
		<ErrorMessage variant="info" text="Invoice/Quote Not Found" />
	{/if}
{:catch err}
	{#if err instanceof ApiError}
		<ErrorMessage variant="warn" text="Invalid Request" retry={loadInvoice} />
	{:else}
		<ErrorMessage variant="warn" text="An Error Occurred" retry={loadInvoice} />
	{/if}
{/await}

{#if selectedStatus}
	<Dialog.InvoiceStatusConfirm
		bind:open={setStatusDialog.isOpen}
		{invoiceId}
		status={selectedStatus}
		onSuccess={() => {
			selectedStatus = null;
			loadInvoice();
		}}
	/>
{/if}
