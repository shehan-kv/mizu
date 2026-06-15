<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { getInvoices, type InvoiceOverview } from '$lib/api/invoices';
	import * as Table from '$lib/components/ui/table';
	import type { PaginatedResponse } from '$lib/api/page';
	import DashboardCard from './DashboardCard.svelte';
	import Spinner from './Spinner.svelte';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import Checks from 'phosphor-svelte/lib/Checks';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import ErrorMessage from './ErrorMessage.svelte';
	import { ApiError } from '$lib/api/client';
	import type { UserRole } from '$lib/api/users';

	interface Props {
		class?: string;
		role: UserRole;
	}
	let { class: className = '', role = 'client' }: Props = $props();

	let promise: Promise<PaginatedResponse<InvoiceOverview>> | null = $state(null);
	let aborter: AbortController | null = null;

	function reload() {
		aborter?.abort();

		aborter = new AbortController();

		promise = getInvoices({ page: 1, limit: 20 }, aborter.signal);
	}

	onMount(reload);

	onDestroy(() => {
		aborter?.abort();
	});

	// svelte-ignore non_reactive_update
	let linksPrefix = '';
	if (role == 'administrator') linksPrefix = '/admin';
	if (role == 'staff') linksPrefix = '/staff';
</script>

<DashboardCard title="Recent Invoices" class={className}>
	<div class="overflow-scroll px-6 py-2">
		{#await promise}
			<Spinner />
		{:then res}
			{#if res && res.items.length > 0}
				<Table.Root>
					<Table.Body>
						{#each res.items as invoice (invoice.id)}
							<Table.Row
								class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
								dark:text-neutral-400 dark:hover:text-neutral-50"
							>
								<Table.Cell class="pl-0">
									{invoice.isInvoice ? 'Invoice' : 'Quote'} #{invoice.id}
								</Table.Cell>
								<Table.Cell align="right">
									{currencyFormatter(invoice.currencyCode, invoice.subTotal)} Total
								</Table.Cell>
								<Table.Cell class="flex items-center gap-1">
									{toTitleCaseDashed(invoice.status)}
									{#if invoice.status == 'paid' || invoice.status == 'accepted'}
										<Checks size={18} class="text-emerald-500" />
									{/if}
								</Table.Cell>
								<Table.Cell class="pr-0" align="right">
									<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
									<a href={`${linksPrefix}/invoices-and-quotes/${invoice.id}`} title="View">
										<ArrowRight size={18} />
									</a>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{:else}
				<ErrorMessage variant="warn" text="Invoices/Quotes Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof ApiError}
				<ErrorMessage variant="warn" text={err.message} retry={reload} />
			{:else}
				<ErrorMessage variant="warn" text="An Error Occurred" retry={reload} />
			{/if}
		{/await}
	</div>
</DashboardCard>
