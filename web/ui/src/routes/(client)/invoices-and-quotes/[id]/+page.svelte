<script lang="ts">
	import { page } from '$app/state';
	import ViewInvoice from '$lib/components/ViewInvoice.svelte';

	let id = Number(page.params.id);
</script>

<svelte:head>
	<title>View Invoice/Quote</title>
</svelte:head>

<ViewInvoice invoiceId={id} showTitle />
