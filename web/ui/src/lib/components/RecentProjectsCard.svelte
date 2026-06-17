<script lang="ts">
	import type { PaginatedResponse } from '$lib/api/page';
	import { getProjectStats, type ProjectStat } from '$lib/api/projects';
	import * as Table from '$lib/components/ui/table';
	import { onDestroy, onMount } from 'svelte';
	import DashboardCard from './DashboardCard.svelte';
	import Spinner from './Spinner.svelte';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import ErrorMessage from './ErrorMessage.svelte';
	import { ApiError } from '$lib/api/client';
	import { resolve } from '$app/paths';

	interface Props {
		class?: string;
	}
	let { class: className = '' }: Props = $props();

	let promise: Promise<PaginatedResponse<ProjectStat>> | null = $state(null);
	let aborter: AbortController | null = null;

	function reload() {
		aborter?.abort();

		aborter = new AbortController();
		promise = getProjectStats({ page: 1, limit: 20 }, aborter.signal);
	}

	onMount(reload);

	onDestroy(() => {
		aborter?.abort();
	});
</script>

<DashboardCard title="Recent Projects" class={className}>
	<div class="overflow-scroll px-6 py-2">
		{#await promise}
			<Spinner />
		{:then res}
			{#if res && res.items.length > 0}
				<Table.Root>
					<Table.Body>
						{#each res.items as project (project.id)}
							<Table.Row
								class="text-neutral-600 hover:bg-transparent hover:text-neutral-950 
							dark:text-neutral-400 dark:hover:text-neutral-50"
							>
								<Table.Cell class="pl-0">{project.name}</Table.Cell>
								<Table.Cell class="flex items-center gap-1.5">
									{#if project.status == 'started'}
										<span class="relative flex size-2">
											<span
												class="absolute inline-flex h-full w-full animate-ping rounded-full
									bg-green-500 opacity-75 dark:bg-green-600"
											>
											</span>
											<span
												class="relative inline-flex size-2 rounded-full bg-green-500 dark:bg-green-600"
											></span>
										</span>
									{/if}
									{toTitleCaseDashed(project.status)}
								</Table.Cell>
								<Table.Cell class="pr-0" align="right">
									<a href={resolve(`/projects/${project.id}`)} title="View">
										<ArrowRight size={18} />
									</a>
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{:else}
				<ErrorMessage variant="warn" text="Projects Not Found" />
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
