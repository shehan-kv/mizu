<script lang="ts">
	import { page } from '$app/state';
	import { getContractById, type Contract } from '$lib/api/contracts';
	import ViewContract from '$lib/components/ViewContract.svelte';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';

	let id = Number(page.params.id);

	let contractStatPromise: Promise<Contract> | null = $state(null);
	let abort: AbortController | null = null;

	function loadContractStat() {
		if (abort) {
			abort.abort();
		}

		abort = new AbortController();

		contractStatPromise = getContractById(id, abort.signal);
	}

	$effect(() => {
		loadContractStat();
	});
</script>

<div class="mx-auto grid h-full grid-rows-[min-content_1fr] gap-4 lg:container">
	<div class="flex w-fit items-center gap-3 text-sm text-neutral-700 dark:text-neutral-400">
		<a href="/contracts" class="underline">Contracts</a>
		<ChevronRight size={18} />
		{#await contractStatPromise}
			<p class="">...</p>
		{:then res}
			<p>{res?.name}</p>
		{/await}
	</div>
	<ViewContract contractId={id} />
</div>
