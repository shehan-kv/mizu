<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { ApiError } from '$lib/api/client';
	import { getInvoice, type Invoice } from '$lib/api/invoices';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import ViewInvoice from '$lib/components/ViewInvoice.svelte';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import { onDestroy, onMount } from 'svelte';

	let id = page.params.id || '';

	let promise: Promise<Invoice> | null = $state(null);
	let abort: AbortController | null = null;

	function loadInvoice() {
		abort?.abort();
		abort = new AbortController();

		promise = getInvoice(id, abort.signal);
	}

	onMount(() => {
		loadInvoice();
	});

	onDestroy(() => {
		abort?.abort();
	});
</script>

<svelte:head>
	<title>View Invoice/Quote</title>
</svelte:head>

<div
	class="mx-auto grid h-full grid-rows-[min-content_1fr] gap-6 rounded bg-neutral-50 p-4 lg:container dark:bg-neutral-950"
>
	<div class="flex w-fit items-center gap-3 text-xs text-neutral-700 dark:text-neutral-400">
		<a href={resolve('/invoices-and-quotes')} class="underline">Invoices & Quotes</a>
		<ChevronRight size={18} />
		{#await promise}
			<p class="">...</p>
		{:then invoice}
			<p>
				{invoice?.isInvoice ? 'Invoice' : 'Quote'} #{invoice?.id
					.replaceAll('-', '')
					.slice(-8)
					.toUpperCase()}
			</p>
		{/await}
	</div>

	{#await promise}
		<Spinner />
	{:then res}
		{#if res}
			<ViewInvoice invoice={res} refresh={loadInvoice} />
		{:else}
			<ErrorMessage variant="info" text="Invoice Not Found" />
		{/if}
	{:catch err}
		{#if err instanceof ApiError}
			<ErrorMessage variant="warn" text={err.message} retry={loadInvoice} />
		{:else}
			<ErrorMessage variant="warn" text="An Error Occurred" retry={loadInvoice} />
		{/if}
	{/await}
</div>
