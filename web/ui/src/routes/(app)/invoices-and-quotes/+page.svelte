<script lang="ts">
	import Checks from 'phosphor-svelte/lib/Checks';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Envelope from 'phosphor-svelte/lib/Envelope';
	import * as Table from '$lib/components/ui/table';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Dialog from '$lib/components/dialogs';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import Pagination from '$lib/components/Pagination.svelte';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import { onDestroy, onMount } from 'svelte';
	import {
		emailInvoice,
		getInvoices,
		type InvoiceOverview,
		type InvoiceStatus
	} from '$lib/api/invoices';
	import { page } from '$app/state';
	import Spinner from '$lib/components/Spinner.svelte';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import FilterSelect from '$lib/components/FilterSelect.svelte';

	import { formatDate } from '$lib/utils/formatDate';
	import { createDialogState } from '$lib/components/dialogs/createDialogState.svelte';
	import DotsThree from 'phosphor-svelte/lib/DotsThree';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { resolve } from '$app/paths';
	import type { PaginatedResponse } from '$lib/api/page';
	import { ApiError } from '$lib/api/client';
	import { INVOICE_STATUS, INVOICE_TYPE } from '$lib/constants/invoice';
	import { auth } from '$lib/auth/auth.svelte';
	import { toast } from 'svelte-sonner';
	import CircleNotch from 'phosphor-svelte/lib/CircleNotch';

	const MAX_LIMIT = 100;
	const DEFAULT_LIMIT = 25;
	const DEFAULT_PAGE_NUMBER = 1;

	const params = new SvelteURLSearchParams(page.url.searchParams.toString());

	let q = $state(params.get('q') || '');
	let status = $state(INVOICE_STATUS.find((s) => s === params.get('status')) ?? '');
	let type = $state(INVOICE_TYPE.find((t) => t === params.get('type')) ?? '');
	let pageNum = $state(Number(params.get('page')) || DEFAULT_PAGE_NUMBER);

	const limitParam = Number(params.get('limit'));
	let limit = $state(
		Number.isFinite(limitParam)
			? Math.min(Math.max(limitParam, DEFAULT_LIMIT), MAX_LIMIT).toString()
			: DEFAULT_LIMIT.toString()
	);

	let promise: Promise<PaginatedResponse<InvoiceOverview>> | null = $state(null);

	let abort: AbortController | null = null;
	function loadInvoices() {
		if (abort) {
			abort.abort();
		}

		abort = new AbortController();

		promise = getInvoices(
			{ q, status: status || undefined, type, page: pageNum, limit: Number(limit) },
			abort.signal
		);
	}

	function updateUrlParam() {
		// A quote cannot have a 'paid' state.
		if (status == 'paid') {
			type = 'invoice';
		}

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
		updateUrlParam();
		loadInvoices();
	}

	let isEmailSending = $state(false);
	let abortEmail: AbortController | null = null;
	async function handleEmail(inv: InvoiceOverview) {
		abortEmail?.abort();
		abortEmail = new AbortController();

		isEmailSending = true;
		selectedInvoice = inv;

		try {
			await emailInvoice(inv.id, abortEmail.signal);
			toast.success(`${inv.isInvoice ? 'Invoice' : 'Quote'} Emailed Successfully`, {
				description: `#${inv.id.replaceAll('-', '').slice(-8).toUpperCase()}`
			});
		} catch (error) {
			if (error instanceof ApiError) {
				toast.error(error.message, {
					description: `#${inv.id.replaceAll('-', '').slice(-8).toUpperCase()}`
				});
			} else {
				toast.error(`Could Not Email Contract ${inv.isInvoice ? 'Invoice' : 'Quote'}`, {
					description: `#${inv.id.replaceAll('-', '').slice(-8).toUpperCase()}`
				});
			}
		} finally {
			isEmailSending = false;
			selectedInvoice = null;
		}
	}

	onMount(() => {
		loadInvoices();
	});

	onDestroy(() => {
		abort?.abort();
	});

	type ActionsAllowed = Exclude<InvoiceStatus, 'pending'>;
	type SelectedInvoice = InvoiceOverview & { action?: ActionsAllowed };
	let selectedInvoice: SelectedInvoice | null = $state(null);
	let setStatusDialog = createDialogState();

	function openStatusDialog(invoice: InvoiceOverview, action: ActionsAllowed) {
		selectedInvoice = { ...invoice, action };
		setStatusDialog.open();
	}
</script>

<svelte:head>
	<title>Invoices and Quotes</title>
</svelte:head>

<div
	class="grid h-full auto-rows-[min-content_1fr_min-content] gap-2 rounded bg-neutral-50 p-4 dark:bg-neutral-950"
>
	<div class="mx-auto flex gap-2 lg:container">
		<div class="w-full max-w-60">
			<SearchBar bind:value={q} onchange={handleFilter} />
		</div>
		<div class="flex gap-2">
			<FilterSelect
				bind:value={status}
				onchange={handleFilter}
				name="Status"
				options={[
					{ value: '', label: 'All' },
					{ value: 'paid', label: 'Paid' },
					{ value: 'accepted', label: 'Accepted' },
					{ value: 'rejected', label: 'Rejected' },
					{ value: 'cancelled', label: 'Cancelled' },
					{ value: 'pending', label: 'Pending' }
				]}
			/>
			<FilterSelect
				bind:value={type}
				onchange={handleFilter}
				name="Type"
				options={[
					...(status !== 'paid' ? [{ value: '', label: 'All' }] : []),
					{ value: 'invoice', label: 'Invoice' },
					...(status !== 'paid' ? [{ value: 'quote', label: 'Quote' }] : [])
				]}
			/>
			<FilterSelect
				bind:value={limit}
				onchange={handleFilter}
				name="Limit"
				options={[
					{ value: '25', label: '25' },
					{ value: '50', label: '50' },
					{ value: '75', label: '75' },
					{ value: '100', label: '100' }
				]}
			/>
		</div>
	</div>

	{#await promise}
		<Spinner />
	{:then res}
		{#if res && res.items}
			<div class="mx-auto gap-4 overflow-y-auto lg:container">
				{#if res.items.length == 0}
					<ErrorMessage variant="info" text="Invoices/Quotes Not Found" />
				{/if}
				<div class="overflow-y-auto">
					{#if res.items.length > 0}
						<Table.Root class="container mx-auto">
							<Table.Header>
								<Table.Row>
									<Table.Head class="font-bold">Type</Table.Head>
									<Table.Head class="font-bold">#ID</Table.Head>
									<Table.Head class="font-bold">Project</Table.Head>
									<Table.Head class="font-bold">Amount</Table.Head>
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
											#{invoice.id.replaceAll('-', '').slice(-8).toUpperCase()}
										</Table.Cell>
										<Table.Cell>{invoice.projectName}</Table.Cell>
										<Table.Cell>
											{currencyFormatter(invoice.currencyCode, invoice.subTotal)}
										</Table.Cell>
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
												class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5 *:hover:text-neutral-950
												dark:text-neutral-400 *:dark:hover:text-neutral-50"
											>
												<a
													title="View"
													class="inline-block"
													href={resolve(`/invoices-and-quotes/${invoice.id}`)}
												>
													<ArrowRight size={18} />
												</a>
												<button
													title="Email Me"
													disabled={selectedInvoice?.id == invoice.id && isEmailSending}
													onclick={() => handleEmail(invoice)}
													class="disabled:cursor-progress"
												>
													{#if selectedInvoice?.id == invoice.id && isEmailSending}
														<CircleNotch size={18} class="animate-spin" />
													{:else}
														<Envelope size={18} />
													{/if}
												</button>

												<DropdownMenu.Root>
													<DropdownMenu.Trigger
														class="cursor-pointer p-1 hover:text-neutral-950 dark:hover:text-neutral-50"
													>
														<DotsThree size={18} />
													</DropdownMenu.Trigger>
													<DropdownMenu.Content class="mr-4 *:text-xs">
														{#if auth.role == 'administrator' || auth.role == 'staff'}
															{#if invoice.status == 'pending' || invoice.status == 'accepted'}
																<DropdownMenu.Group class="text-xs">
																	<DropdownMenu.Label class="text-xs">Mark As</DropdownMenu.Label>
																	{#if invoice.status == 'accepted'}
																		<DropdownMenu.Item
																			class="pl-4 text-xs"
																			onclick={() => openStatusDialog(invoice, 'paid')}
																		>
																			Paid
																		</DropdownMenu.Item>
																	{/if}
																	<DropdownMenu.Item
																		class="pl-4 text-xs"
																		onclick={() => openStatusDialog(invoice, 'cancelled')}
																	>
																		Cancelled
																	</DropdownMenu.Item>
																</DropdownMenu.Group>
															{:else}
																<div class="flex items-center gap-2 px-2 py-1.5 text-xs">
																	<Checks /> Already {toTitleCaseDashed(invoice.status)}
																</div>
															{/if}
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
																<Checks /> Already {toTitleCaseDashed(invoice.status)}
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
					{/if}
				</div>
			</div>
			{#if res.items.length > 0}
				<div class="container mx-auto flex justify-end">
					<Pagination bind:page={pageNum} count={res.totalCount} perPage={limit} />
				</div>
			{/if}
		{/if}
	{:catch err}
		{#if err instanceof ApiError}
			<ErrorMessage variant="warn" text={err.message} retry={loadInvoices} />
		{:else}
			<ErrorMessage variant="warn" text="An Error Occurred" retry={loadInvoices} />
		{/if}
	{/await}
</div>

{#if selectedInvoice && selectedInvoice.action}
	<Dialog.InvoiceStatusConfirm
		bind:open={setStatusDialog.isOpen}
		invoiceId={selectedInvoice.id}
		status={selectedInvoice.action}
		onSuccess={loadInvoices}
	/>
{/if}
