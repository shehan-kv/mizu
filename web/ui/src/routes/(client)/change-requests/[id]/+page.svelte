<script lang="ts">
	import { page } from '$app/state';
	import ViewChangeRequest from '$lib/components/ViewChangeRequest.svelte';

	let id = Number(page.params.id);
</script>

<svelte:head>
	<title>View Change Request</title>
</svelte:head>

<div class="relative h-full overflow-y-auto">
	<ViewChangeRequest requestId={id} showTitle />
</div>
