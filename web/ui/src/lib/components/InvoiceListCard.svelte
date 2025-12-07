<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Dialog from '$lib/components/dialogs';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Spinner from './Spinner.svelte';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import Checks from 'phosphor-svelte/lib/Checks';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import { formatDate } from '$lib/utils/formatDate';
	import ErrorMessage from './ErrorMessage.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import {
		getInvoicesByProjectId,
		type InvoiceStatus,
		type InvoiceWithStatus
	} from '$lib/api/invoices';
	import { onMount } from 'svelte';
	import DotsThree from 'phosphor-svelte/lib/DotsThree';
	import { createDialogState } from './dialogs/createDialogState.svelte';
	import type { UserRole } from '$lib/api/users';

	interface Props {
		projectId: number;
		role?: UserRole;
		page?: number;
		limit?: number;
	}

	let { projectId, role = 'client', page = 1, limit = 20 }: Props = $props();

	let invoices: Promise<PaginatedResponse<InvoiceWithStatus>> | null = $state(null);
	let abort: AbortController | null = null;
	function loadInvoices() {
		if (abort) {
			abort.abort();
		}
		abort = new AbortController();

		invoices = getInvoicesByProjectId(projectId, { page, limit }, abort.signal);
	}

	export function refresh() {
		loadInvoices();
	}

	onMount(() => {
		loadInvoices();
	});

	type ActionsAllowed = Exclude<InvoiceStatus, 'pending'>;
	type SelectedInvoice = InvoiceWithStatus & { action?: ActionsAllowed };
	let selectedInvoice: SelectedInvoice | null = $state(null);
	let setStatusDialog = createDialogState();

	function openStatusDialog(invoice: InvoiceWithStatus, action: ActionsAllowed) {
		selectedInvoice = { ...invoice, action };
		setStatusDialog.open();
	}
</script>

<div
	class="min-h-50 max-h-100 col-span-4 grid grid-rows-[min-content_1fr]
		overflow-hidden rounded border"
>
	<div class="flex items-center justify-between bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
		<p class="text-sm">Invoices / Quotes</p>
		<a href={`/admin/projects/${projectId}/invoices`} class="flex items-center gap-1 text-sm">
			<span>View All</span>
			<ArrowRight />
		</a>
	</div>
	<div class="overflow-scroll px-6 py-2">
		{#await invoices}
			<Spinner />
		{:then res}
			{#if res && res.data.length > 0}
				<Table.Root>
					<Table.Body>
						{#each res.data as invoice}
							<Table.Row
								class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
								dark:text-neutral-400 dark:hover:text-neutral-50"
							>
								<Table.Cell class="pl-0">
									{invoice.isInvoice ? 'Invoice' : 'Quote'} #{invoice.id}
								</Table.Cell>
								<Table.Cell class="flex items-center gap-1">
									{toTitleCase(invoice.status)}
									{#if invoice.status == 'accepted'}
										<Checks size={18} class="text-emerald-500" />
									{/if}
								</Table.Cell>
								<Table.Cell>
									{currencyFormatter(invoice.currencyCode, invoice.total)} Total
								</Table.Cell>
								<Table.Cell>
									Issued On {formatDate(invoice.issuedAt)}
								</Table.Cell>
								<Table.Cell class="pr-0" align="right">
									<div
										class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5
										*:hover:text-neutral-950 dark:text-neutral-400
										*:dark:hover:text-neutral-50"
									>
										<a
											href={`/admin/invoices-and-quotes/${invoice.id}`}
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
												{#if role == 'admin' || role == 'staff'}
													{#if invoice.status == 'pending' || invoice.status == 'accepted'}
														<DropdownMenu.Group class="text-xs">
															<DropdownMenu.Label class="text-xs">Mark As</DropdownMenu.Label>
															<DropdownMenu.Item
																class="pl-4 text-xs"
																onclick={() => openStatusDialog(invoice, 'paid')}
															>
																Paid
															</DropdownMenu.Item>
															<DropdownMenu.Item
																class="pl-4 text-xs"
																onclick={() => openStatusDialog(invoice, 'cancelled')}
															>
																Cancelled
															</DropdownMenu.Item>
														</DropdownMenu.Group>
													{:else}
														<div class="flex items-center gap-2 px-2 py-1.5 text-xs">
															<Checks /> Already {toTitleCase(invoice.status)}
														</div>
													{/if}

													<!-- If the role is a client -->
												{:else if invoice.status == 'pending'}
													<DropdownMenu.Item
														class="pl-4 text-xs"
														onclick={() => openStatusDialog(invoice, 'accepted')}
													>
														Accept
													</DropdownMenu.Item>
													<DropdownMenu.Item
														class="pl-4 text-xs"
														onclick={() => openStatusDialog(invoice, 'rejected')}
													>
														Reject
													</DropdownMenu.Item>
												{:else}
													<div class="flex items-center gap-2 px-2 py-1.5 text-xs">
														<Checks /> Already {toTitleCase(invoice.status)}
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
				<ErrorMessage variant="warn" text="Invoices / Quotes Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof APIBadRequestError}
				<ErrorMessage variant="warn" text="Invalid Request" retry={loadInvoices} />
			{:else if err instanceof APIForbiddenError}
				<ErrorMessage
					variant="warn"
					text="You Don't Have Permission To View Invoices/Quotes"
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
	</div>
</div>

{#if selectedInvoice && selectedInvoice.action}
	<Dialog.InvoiceStatusConfirm
		bind:open={setStatusDialog.isOpen}
		invoiceId={selectedInvoice.id}
		status={selectedInvoice.action}
		onSuccess={loadInvoices}
	/>
{/if}
