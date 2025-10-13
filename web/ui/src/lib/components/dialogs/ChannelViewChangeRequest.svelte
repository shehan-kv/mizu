<script lang="ts">
	import ViewChangeRequest from '../ViewChangeRequest.svelte';
	import FullScreenDialog from './FullScreenDialog.svelte';

	interface Props {
		open: Boolean;
		requestId: number;
		requestTitle: string;
	}
	let { open = $bindable(), requestId, requestTitle }: Props = $props();
</script>

<FullScreenDialog bind:open>
	<div class="grid auto-rows-[min-content_1fr] gap-6 overflow-y-auto">
		<div class="container mx-auto space-y-0.5">
			<p class="text-xs text-neutral-500">Change Request</p>
			<p class="font-bold">{requestTitle}</p>
		</div>

		{#if open}
			<ViewChangeRequest {requestId} />
		{/if}
	</div>
</FullScreenDialog>
