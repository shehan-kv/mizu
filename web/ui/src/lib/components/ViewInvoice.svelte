<script lang="ts">
	import { getInvoiceDetails, type InvoiceDetails } from '$lib/api/invoices';
	import * as Table from '$lib/components/ui/table';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import { formatDate } from '$lib/utils/formatDate';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Note from 'phosphor-svelte/lib/Note';
	import ClockCounterClockwise from 'phosphor-svelte/lib/ClockCounterClockwise';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import Spinner from './Spinner.svelte';
	import User from './User.svelte';
	import ErrorMessage from './ErrorMessage.svelte';
	import { onMount } from 'svelte';

	interface Props {
		invoiceId: number;
	}

	let { invoiceId }: Props = $props();

	let invoicePromise: Promise<InvoiceDetails> | null = $state(null);
	let abortController: AbortController | null = null;

	function loadInvoice() {
		if (abortController) {
			abortController.abort();
		}
		abortController = new AbortController();

		invoicePromise = getInvoiceDetails(invoiceId, abortController.signal);
	}

	onMount(() => {
		loadInvoice();
	});
</script>

{#await invoicePromise}
	<Spinner />
{:then invoice}
	{#if invoice}
		<div class="flex gap-20">
			<div class="space-y-2">
				<div>
					<p class="text-xs text-neutral-500">Project</p>
					<p>{invoice.projectName}</p>
				</div>
				<div>
					<p class="text-xs text-neutral-500">Status</p>
					<p>{toTitleCase(invoice.status)}</p>
				</div>
				<div>
					<p class="text-xs text-neutral-500">Issued On</p>
					<p>{formatDate(invoice.issuedAt)}</p>
				</div>
				<div>
					<p class="text-xs text-neutral-500">Due On</p>
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
					<p>{currencyFormatter(invoice.currencyCode, invoice.discount)}</p>
				</div>
				<div>
					<p class="text-xs text-neutral-500">Tax</p>
					<p>{currencyFormatter(invoice.currencyCode, invoice.tax)}</p>
				</div>
				<div>
					<p class="text-xs text-neutral-500">Total</p>
					<p>{currencyFormatter(invoice.currencyCode, invoice.total)}</p>
				</div>
			</div>
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
					{#each invoice.items as item, idx}
						<Table.Row>
							<Table.Cell>{item.description}</Table.Cell>
							<Table.Cell>{item.qty}</Table.Cell>
							<Table.Cell>
								{currencyFormatter(invoice.currencyCode, item.unitPrice)}
							</Table.Cell>
							<Table.Cell>
								{#if item.discountType == 'percentage'}
									{item.unitDiscount}%
								{:else}
									{currencyFormatter(invoice.currencyCode, item.unitDiscount)}
								{/if}
							</Table.Cell>
							<Table.Cell>
								{#if item.taxType == 'percentage'}
									{item.unitTax}%
								{:else}
									{currencyFormatter(invoice.currencyCode, item.unitTax)}
								{/if}
							</Table.Cell>
							<Table.Cell>
								{currencyFormatter(invoice.currencyCode, item.totalDiscount)}
							</Table.Cell>
							<Table.Cell>
								{currencyFormatter(invoice.currencyCode, item.totalTax)}
							</Table.Cell>
							<Table.Cell align="right">
								{currencyFormatter(invoice.currencyCode, item.total)}
							</Table.Cell>
						</Table.Row>
					{/each}
					<Table.Row>
						<Table.Cell colspan={7} align="right" class="border-r">Total Discount</Table.Cell>
						<Table.Cell align="right">
							{currencyFormatter(invoice.currencyCode, invoice.discount)}
						</Table.Cell>
					</Table.Row>
					<Table.Row>
						<Table.Cell colspan={7} align="right" class="border-r">Total Tax</Table.Cell>
						<Table.Cell align="right">
							{currencyFormatter(invoice.currencyCode, invoice.tax)}
						</Table.Cell>
					</Table.Row>
					<Table.Row class="bg-neutral-100 dark:bg-neutral-900">
						<Table.Cell colspan={7} align="right" class="border-r">Invoice Total</Table.Cell>
						<Table.Cell align="right">
							{currencyFormatter(invoice.currencyCode, invoice.total)}
						</Table.Cell>
					</Table.Row>
				</Table.Body>
			</Table.Root>
		</div>

		<div class="mt-8">
			<p class="inline-flex items-center gap-1"><Note size={20} /> Note</p>
			<p>{invoice.note || 'N/A'}</p>
		</div>

		<div class="mt-8">
			<p class="inline-flex items-center gap-1">
				<ClockCounterClockwise size={20} /> History
			</p>
			<div
				class="before:content-[' '] dark:before:-z-1 relative mt-2
								space-y-10 before:absolute before:left-[6px] before:top-1
								before:z-auto before:min-h-full before:w-1 before:border-l-[1px]
								before:border-dashed before:border-neutral-600"
			>
				{#each invoice.history as entry}
					<div class="flex items-start gap-3">
						<div
							class="z-1 mt-1.5 size-3 shrink-0 rounded-full border border-2
											border-emerald-500 bg-neutral-50 dark:z-auto dark:border-emerald-700
											dark:bg-neutral-950"
						></div>
						<div>
							<p class="text-sm">
								{entry.isInvoice ? 'Invoice' : 'Quote'}
								{toTitleCase(entry.event)}
							</p>
							<p class=" text-xs">
								On {formatDate(entry.recoredAt)}
							</p>
							{#if entry.lastStatus && entry.newStatus}
								<p
									class="inline-flex items-center gap-2 text-xs text-neutral-600 dark:text-neutral-500"
								>
									{toTitleCase(entry.lastStatus)}
									<ArrowRight class="text-emerald-500" />
									{toTitleCase(entry.newStatus)}
								</p>
							{/if}
							<div class="mt-2">
								<User
									image={entry.user.image}
									name={`${entry.user.firstName} ${entry.user.lastName}`}
									role={entry.user.role}
									title={entry.user.title}
								/>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{:else}
		<ErrorMessage variant="info" text="Invoice/Quote Not Found" />
	{/if}
{:catch err}
	{#if err instanceof APIBadRequestError}
		<ErrorMessage variant="warn" text="Invalid Request" retry={loadInvoice} />
	{:else if err instanceof APIForbiddenError}
		<ErrorMessage
			variant="warn"
			text="You Don't Have Permission To View This Invoice/Quote"
			retry={loadInvoice}
		/>
	{:else if err instanceof APINotFoundError}
		<ErrorMessage variant="info" text="Not Found" retry={loadInvoice} />
	{:else if err instanceof APIServerError}
		<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadInvoice} />
	{:else}
		<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadInvoice} />
	{/if}
{/await}
