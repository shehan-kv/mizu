<script lang="ts">
	import FullScreenDialog from './FullScreenDialog.svelte';

	import { getContract, type Contract } from '$lib/api/contracts';
	import * as Dialog from '$lib/components/dialogs';

	import ViewContract from '../ViewContract.svelte';
	import { createDialogState } from './createDialogState.svelte';
	import { onDestroy, onMount } from 'svelte';
	import Spinner from '../Spinner.svelte';
	import { auth } from '$lib/auth/auth.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import { ApiError } from '$lib/api/client';

	interface Props {
		open: boolean;
		contractId: string;
	}
	let { open = $bindable(), contractId }: Props = $props();

	let contractPromise: Promise<Contract> | null = $state(null);
	let abort: AbortController | null = null;

	function loadContract() {
		if (abort) {
			abort.abort();
		}

		abort = new AbortController();

		contractPromise = getContract(contractId, abort.signal);
	}

	let manageSignatoriesDialog = createDialogState();

	onMount(() => {
		loadContract();
	});

	onDestroy(() => {
		abort?.abort();
	});
</script>

<FullScreenDialog bind:open>
	{#await contractPromise}
		<Spinner />
	{:then contract}
		{#if contract}
			<div class="container mx-auto grid auto-rows-[min-content_1fr] gap-6 overflow-y-auto">
				<div class="flex">
					<div class="grow space-y-0.5">
						<p class="text-xs text-neutral-500">Contract</p>
						<p class="font-bold">{contract.name}</p>
					</div>

					{#if auth.role == 'administrator'}
						<button
							onclick={manageSignatoriesDialog.open}
							class="hover:bg-netural-200 cursor-pointer rounded bg-neutral-200 px-4
							py-2 text-xs text-neutral-700 transition hover:bg-neutral-300 disabled:cursor-not-allowed
							dark:bg-neutral-800 dark:text-neutral-300 hover:dark:bg-neutral-700"
						>
							Manage Signatories
						</button>
					{/if}
				</div>

				<ViewContract {contract} />
			</div>
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
</FullScreenDialog>

<Dialog.ManageContractSignatories
	bind:open={manageSignatoriesDialog.isOpen}
	{contractId}
	onSuccess={loadContract}
/>
