<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { getContractOverviews, type Contract } from '$lib/api/contracts';
	import type { PaginatedResponse } from '$lib/api/page';
	import * as Table from '$lib/components/ui/table';
	import DashboardCard from './DashboardCard.svelte';
	import Spinner from './Spinner.svelte';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import Checks from 'phosphor-svelte/lib/Checks';
	import { formatDate } from '$lib/utils/formatDate';
	import { resolve } from '$app/paths';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import ErrorMessage from './ErrorMessage.svelte';
	import { ApiError } from '$lib/api/client';

	let { class: className = '' }: { class?: string } = $props();

	let promise: Promise<PaginatedResponse<Contract>> | null = $state(null);
	let aborter: AbortController | null = null;

	function reload() {
		aborter?.abort();

		aborter = new AbortController();

		promise = getContractOverviews({ page: 1, limit: 20 }, aborter.signal);
	}

	onMount(reload);

	onDestroy(() => {
		aborter?.abort();
	});
</script>

<DashboardCard title="Recent Contracts" class={className}>
	<div class="overflow-scroll px-6 py-2">
		{#await promise}
			<Spinner />
		{:then res}
			{#if res && res.items.length > 0}
				<Table.Root>
					<Table.Body>
						{#each res.items as contract (contract.id)}
							<Table.Row
								class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
								dark:text-neutral-400 dark:hover:text-neutral-50"
							>
								<Table.Cell class="pl-0">
									{contract.name}
								</Table.Cell>
								<Table.Cell class="flex items-center gap-1">
									{toTitleCaseDashed(contract.status)}
									{#if contract.status == 'signed'}
										<Checks size={18} class="text-emerald-500" />
									{/if}
								</Table.Cell>
								<Table.Cell>
									Created On {formatDate(contract.createdAt)}
								</Table.Cell>
								<Table.Cell class="pr-0" align="right">
									<a href={resolve(`/contracts/${contract.id}`)} title="View">
										<ArrowRight size={18} />
									</a>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{:else}
				<ErrorMessage variant="warn" text="Contracts Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof ApiError}
				<ErrorMessage variant="warn" text={err.message} retry={reload} />
			{:else}
				<ErrorMessage variant="warn" text="An Error Occurred" retry={reload} />
			{/if}
		{/await}
	</div>
</DashboardCard>
