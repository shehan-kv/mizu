<script lang="ts">
	import { getProjectTasks, type ProjectTask } from '$lib/api/projects';
	import CellSignalFull from 'phosphor-svelte/lib/CellSignalFull';
	import CellSignalLow from 'phosphor-svelte/lib/CellSignalLow';
	import CellSignalMedium from 'phosphor-svelte/lib/CellSignalMedium';
	import { onMount } from 'svelte';
	import Spinner from './Spinner.svelte';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import { formatMinutes } from '$lib/utils/formatMinutes';
	import ErrorMessage from './ErrorMessage.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { formatDate } from '$lib/utils/formatDate';

	interface Props {
		projectId: number;
		status: string;
	}
	let { projectId, status }: Props = $props();

	let isLoading = $state(false);

	let page = $state(1);
	let limit = $state(30);

	let tasks: ProjectTask[] = $state([]);

	let abortController: AbortController | null = null;
	let loadError: unknown | null = $state(null);

	async function loadItems() {
		isLoading = true;
		if (abortController) {
			abortController.abort();
		}

		abortController = new AbortController();

		try {
			const res = await getProjectTasks(projectId, { status, page, limit }, abortController.signal);
			tasks.push(...res.data);
			return res;
		} catch (error) {
			loadError = error;
		} finally {
			isLoading = false;
		}
	}

	let container: HTMLDivElement | null = null;

	async function handleScroll() {
		if (!container || isLoading) return;
		const { scrollTop, clientHeight, scrollHeight } = container;

		if (scrollTop + clientHeight >= scrollHeight) {
			page++;
			const res = await loadItems();

			// Reset the page counter so each new scroll request
			// fetches the next sequential page relative to the
			// pages already loaded; for example, if the current
			// page is 1, repeated scrolls will request page 2
			// when the next fetch is performed.
			if (!res || res.data.length == 0) page--;
		}
	}

	onMount(() => {
		loadItems();
	});
</script>

<div class="h-full space-y-2 overflow-y-auto" bind:this={container} onscroll={handleScroll}>
	{#if loadError}
		<div>
			{#if loadError instanceof APIBadRequestError}
				<ErrorMessage variant="warn" text="Invalid Request" retry={loadItems} />
			{:else if loadError instanceof APIForbiddenError}
				<ErrorMessage
					variant="warn"
					text="You Don't Have Permission To View These Project Tasks"
					retry={loadItems}
				/>
			{:else if loadError instanceof APINotFoundError}
				<ErrorMessage variant="info" text="Not Found" retry={loadItems} />
			{:else if loadError instanceof APIServerError}
				<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadItems} />
			{/if}
		</div>
	{/if}

	{#if !loadError}
		{#if tasks.length > 0}
			{#each tasks as task}
				<div class="rounded bg-neutral-50 p-6 dark:bg-neutral-900">
					<div class="flex items-start justify-between text-xs">
						<div class="space-x-2">
							{#if task.priority == 'high'}
								<p class="inline-flex items-center gap-1 text-red-800 dark:text-red-200">
									<CellSignalFull
										weight="duotone"
										size={14}
										class="text-rose-600 dark:text-rose-400"
									/>
									High Priority
								</p>
							{:else if task.priority == 'medium'}
								<p class="inline-flex items-center gap-2 text-amber-800 dark:text-amber-200">
									<CellSignalMedium
										weight="duotone"
										size={14}
										class="text-yellow-600 dark:text-amber-500"
									/>
									Medium Priority
								</p>
							{:else if task.priority == 'low'}
								<p class="inline-flex items-center gap-1 text-emerald-800 dark:text-emerald-200">
									<CellSignalLow
										weight="duotone"
										size={14}
										class="text-emerald-600 dark:text-emerald-500"
									/>
									Low Priority
								</p>
							{:else}
								<p>{toTitleCase(task.priority)}</p>
							{/if}
						</div>
						<div>
							<p>Added - {formatDate(task.createdAt)}</p>
							<p class="text-neutral-4400 mt-0.5 text-right text-xs">
								{formatMinutes(task.estTimeMinutes)} Estimated
							</p>
						</div>
					</div>
					<p class="mt-4 text-sm font-bold">{task.name}</p>
					<p class="mt-0.5 text-xs text-neutral-700 dark:text-neutral-400">
						{task.description}
					</p>
					<div class="mt-6">
						<div>
							<p class="text-neutral-00 text-xs">Assigned To</p>

							<p class="text-sm">
								{task.assignees.map((a) => `${a.firstName} ${a.lastName}`).join(', ')}
							</p>
						</div>
					</div>
				</div>
			{/each}
		{:else}
			<div>
				<ErrorMessage variant="info" text="Tasks Not Found" retry={loadItems} />
			</div>
		{/if}
	{/if}

	<div>
		{#if isLoading}
			<Spinner />
		{/if}
	</div>
</div>
