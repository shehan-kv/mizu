<script lang="ts">
	import { page } from '$app/state';
	import type { ChangeRequestDetails } from '$lib/api/changeRequest';
	import ViewChangeRequest from '$lib/components/ViewChangeRequest.svelte';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';

	let id = Number(page.params.id);
	let request: ChangeRequestDetails | null = $state(null);

	function setState(req: ChangeRequestDetails) {
		request = req;
		document.title = req.title;
	}
</script>

<svelte:head>
	<title>View Change Request</title>
</svelte:head>

<div class="mx-auto grid h-full grid-rows-[min-content_1fr] gap-4 lg:container">
	<div class="flex w-fit items-center gap-3 text-sm text-neutral-700 dark:text-neutral-400">
		<a href="/admin/change-requests" class="underline">Change Requests</a>
		<ChevronRight size={18} />
		{#if request}
			<p>{request.title}</p>
		{:else}
			<p class="">...</p>
		{/if}
	</div>
	<ViewChangeRequest requestId={id} onLoad={setState} />
</div>
