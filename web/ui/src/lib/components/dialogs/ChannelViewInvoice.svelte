<script lang="ts">
	import FullScreenDialog from './FullScreenDialog.svelte';
	import ViewInvoice from '../ViewInvoice.svelte';
	import { getInvoice, type Invoice } from '$lib/api/invoices';
	import Spinner from '../Spinner.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import { ApiError } from '$lib/api/client';

	interface Props {
		open: boolean;
		invoiceId: string;
		isInvoice: boolean;
	}

	let { open = $bindable(), invoiceId, isInvoice }: Props = $props();

	let promise: Promise<Invoice> | null = $state(null);
	let abort: AbortController | null = null;

	function loadInvoice() {
		abort?.abort();
		abort = new AbortController();

		promise = getInvoice(invoiceId, abort.signal);
	}

	$effect(() => {
		if (open) {
			loadInvoice();
		} else {
			abort?.abort();
		}
	});
</script>

<FullScreenDialog bind:open>
	<div class="grid auto-rows-[min-content_1fr_min-content] gap-6 overflow-y-auto px-5">
		<div class="flex-none">
			<div class="container mx-auto">
				<p class="font-bold">
					View {isInvoice ? 'Invoice' : 'Quote'} - #{invoiceId
						.replaceAll('-', '')
						.slice(-8)
						.toUpperCase()}
				</p>
			</div>
		</div>

		{#await promise}
			<Spinner />
		{:then res}
			{#if res}
				<div class="overflow-y-auto">
					<div class="container mx-auto">
						<ViewInvoice invoice={res} refresh={loadInvoice} />
					</div>
				</div>
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
</FullScreenDialog>
