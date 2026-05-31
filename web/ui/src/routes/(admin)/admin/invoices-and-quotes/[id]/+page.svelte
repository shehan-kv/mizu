<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import type { Invoice } from '$lib/api/invoices';
	import ViewInvoice from '$lib/components/ViewInvoice.svelte';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';

	let id = page.params.id || '';
	let invoice: Invoice | null = $state(null);

	function setState(inv: Invoice) {
		invoice = inv;
		document.title = `View ${inv.isInvoice ? 'Invoice' : 'Quote'} #${inv.id}`;
	}

	const role = 'administrator';
</script>

<svelte:head>
	<title>View Invoice/Quote</title>
</svelte:head>

<div class="mx-auto grid h-full grid-rows-[min-content_1fr] gap-4 lg:container">
	<div class="flex w-fit items-center gap-3 text-sm text-neutral-700 dark:text-neutral-400">
		<a href={resolve('/admin/invoices-and-quotes')} class="underline">Invoices & Quotes</a>
		<ChevronRight size={18} />
		{#if invoice}
			<p>{invoice.isInvoice ? 'Invoice' : 'Quote'} #{invoice.id}</p>
		{:else}
			<p class="">...</p>
		{/if}
	</div>
	<ViewInvoice invoiceId={id} {role} onLoad={setState} />
</div>
