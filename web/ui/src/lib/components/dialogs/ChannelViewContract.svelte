<script lang="ts">
	import FullScreenDialog from './FullScreenDialog.svelte';

	import { type Contract } from '$lib/api/contracts';

	import ViewContract from '../ViewContract.svelte';

	interface Props {
		open: Boolean;
		contract: Contract;
	}
	let { open = $bindable(), contract }: Props = $props();
</script>

<FullScreenDialog bind:open>
	<div class="container mx-auto grid auto-rows-[min-content_1fr] gap-6 overflow-y-auto">
		<div class="space-y-0.5">
			<p class="text-xs text-neutral-500">Contract</p>
			<p class="font-bold">{contract.name}</p>
		</div>

		{#if open}
			<ViewContract contractId={contract.id} />
		{/if}
	</div>
</FullScreenDialog>
