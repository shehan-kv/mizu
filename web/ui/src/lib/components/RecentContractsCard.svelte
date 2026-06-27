<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { getContractOverviews, type ContractOverview } from '$lib/api/contracts';
	import type { PaginatedResponse } from '$lib/api/page';
	import * as Table from '$lib/components/ui/table';
	import DashboardCard from './DashboardCard.svelte';
	import Spinner from './Spinner.svelte';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import Checks from 'phosphor-svelte/lib/Checks';
	import { formatDate } from '$lib/utils/formatDate';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import ErrorMessage from './ErrorMessage.svelte';
	import { ApiError } from '$lib/api/client';
	import { resolve } from '$app/paths';

	interface Props {
		class?: string;
	}
	let { class: className = '' }: Props = $props();

	let promise: Promise<PaginatedResponse<ContractOverview>> | null = $state(null);
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
	<div class="overflow-auto px-6 py-2">
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
								<Table.Cell>
									<div class="flex items-center gap-1">
										{toTitleCaseDashed(contract.status)}
										{#if contract.status == 'signed'}
											<Checks size={18} class="text-emerald-500" />
										{/if}
									</div>
								</Table.Cell>
								<Table.Cell>
									{#if contract.signatories.length == 0}
										<p>N/A</p>
									{:else}
										<p>
											{contract.signatories
												.map((s, idx) => idx <= 1 && `${s.firstName} ${s.lastName}`)
												.filter(Boolean)
												.join(', ')}
										</p>
										{#if contract.signatories.length > 2}
											<span class="text-xs text-neutral-700 dark:text-neutral-300">
												+{contract.signatories.length - 2} Others
											</span>
										{/if}
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
