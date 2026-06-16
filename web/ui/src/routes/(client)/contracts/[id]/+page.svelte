<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { ApiError } from '$lib/api/client';
	import { getContract, type Contract } from '$lib/api/contracts';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import ViewContract from '$lib/components/ViewContract.svelte';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import { onDestroy, onMount } from 'svelte';

	let id = page.params.id || '';

	let contractPromise: Promise<Contract> | null = $state(null);
	let abort: AbortController | null = null;

	function loadContract() {
		if (abort) {
			abort.abort();
		}

		abort = new AbortController();

		contractPromise = getContract(id, abort.signal).then((c) => {
			document.title = 'Contract - ' + c.name;
			return c;
		});
	}

	onMount(() => {
		loadContract();
	});

	onDestroy(() => {
		abort?.abort();
	});
</script>

<svelte:head>
	<title>Contract</title>
</svelte:head>

<div
	class="mx-auto grid h-full grid-rows-[min-content_1fr] gap-4 rounded bg-neutral-50 p-4 lg:container dark:bg-neutral-950"
>
	<div class="flex w-fit items-center gap-3 text-xs text-neutral-700 dark:text-neutral-400">
		<a href={resolve('/contracts')} class="underline">Contracts</a>
		<ChevronRight size={18} />
		{#await contractPromise}
			<p class="">...</p>
		{:then res}
			<p>{res?.name}</p>
		{/await}
	</div>

	{#await contractPromise}
		<Spinner />
	{:then res}
		{#if res}
			<ViewContract contract={res} />
		{:else}
			<ErrorMessage variant="info" text="Contract Not Found" />
		{/if}
	{:catch err}
		{#if err instanceof ApiError}
			<ErrorMessage variant="warn" text={err.message} retry={loadContract} />
		{:else}
			<ErrorMessage variant="warn" text="An Error Occurred" retry={loadContract} />
		{/if}
	{/await}
</div>
