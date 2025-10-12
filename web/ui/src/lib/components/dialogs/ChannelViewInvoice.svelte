<script lang="ts">
	import FullScreenDialog from './FullScreenDialog.svelte';
	import ViewInvoice from '../ViewInvoice.svelte';

	interface Props {
		open: boolean;
		invoiceId: number;
		isInvoice: boolean;
	}

	let { open = $bindable(), invoiceId, isInvoice }: Props = $props();
</script>

<FullScreenDialog bind:open>
	<div class="grid auto-rows-[min-content_1fr_min-content] gap-6 overflow-y-auto px-5">
		<div class="flex-none">
			<div class="container mx-auto">
				<p class="font-bold">
					View {isInvoice ? 'Invoice' : 'Quote'} - #{invoiceId}
				</p>
			</div>
		</div>

		<div class="overflow-y-auto">
			<div class="container mx-auto">
				{#if open}
					<ViewInvoice {invoiceId} />
				{/if}
			</div>
		</div>
	</div>
</FullScreenDialog>
