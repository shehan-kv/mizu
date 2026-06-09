<script lang="ts">
	import {
		getInvoicesSummmary,
		type InvoicesSummary,
		type InvoicesSummaryMetric
	} from '$lib/api/invoices';
	import { onDestroy, onMount } from 'svelte';
	import DashboardCard from './DashboardCard.svelte';
	import { currencyFormatter } from '$lib/utils/currencyFormatter';
	import { formatDecimalSuffix } from '$lib/utils/formatDecimalSuffix';
	import Spinner from './Spinner.svelte';
	import ErrorMessage from './ErrorMessage.svelte';
	import { ApiError } from '$lib/api/client';

	let { class: className = '' }: { class?: string } = $props();

	let promise: Promise<InvoicesSummary | null> | null = $state(null);
	let aborter: AbortController | null = null;

	function reload() {
		if (aborter) {
			aborter.abort();
		}

		aborter = new AbortController();

		promise = getInvoicesSummmary(aborter.signal);
	}

	onMount(reload);
	onDestroy(() => aborter?.abort());
</script>

<DashboardCard title="Invoices Overview" class={className}>
	<div class="space-y-4 overflow-scroll px-6 py-2">
		{#snippet overview(label: string, items: InvoicesSummaryMetric[])}
			<div>
				<p class="border-b py-2 text-sm text-neutral-600 dark:text-neutral-400">
					{label}
				</p>
				{#if items.length > 0}
					{#each items as item (item)}
						<p
							class="py-2 text-right text-3xl"
							title={currencyFormatter(item.currencyCode, item.amount)}
						>
							{formatDecimalSuffix(item.amount)}
							<span class="text-sm">
								{item.currencyCode.toUpperCase()}

								({item.count})
							</span>
						</p>
					{/each}
				{:else}
					<p class="py-2 text-right text-xl">N/A</p>
				{/if}
			</div>
		{/snippet}

		{#await promise}
			<Spinner />
		{:then res}
			{#if res}
				{@render overview('Invoices Paid', res.invoicesPaid)}
				{@render overview('Invoices Pending', res.invoicesPending)}
				{@render overview('Invoices Accepted', res.invoicesAccepted)}
				{@render overview('Invoices Rejected', res.invoicesRejected)}
				{@render overview('Invoices Cancelled', res.invoicesCancelled)}
				{@render overview('Quotes Pending', res.quotesPending)}
				{@render overview('Quotes Rejected', res.quotesRejected)}
			{:else}
				<ErrorMessage variant="warn" text="Overview Not Found" />
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
