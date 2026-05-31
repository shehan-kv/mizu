<script lang="ts">
	import { page } from '$app/state';
	import * as Table from '$lib/components/ui/table';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Pagination from '$lib/components/Pagination.svelte';
	import { getProjectStats, type ProjectStat } from '$lib/api/projects';

	import { onMount } from 'svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import FilterSelect from '$lib/components/FilterSelect.svelte';
	import FilterInput from '$lib/components/FilterInput.svelte';
	import { formatDate } from '$lib/utils/formatDate';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import type { PaginatedResponse } from '$lib/api/page';
	import { ApiError } from '$lib/api/client';
	import { PROJECT_STATUS } from '$lib/constants/project';

	const MAX_LIMIT = 100;
	const MIN_LIMIT = 1;
	const DEFAULT_LIMIT = 30;
	const DEFAULT_PAGE_NUMBER = 1;

	const params = new SvelteURLSearchParams(page.url.searchParams.toString());

	let q = $state(params.get('q') || '');
	let status = $state(PROJECT_STATUS.find((s) => s === params.get('status')) ?? '');
	let pageNum = $state(Number(params.get('page')) || DEFAULT_PAGE_NUMBER);
	let limit = $state(Math.min(Number(params.get('limit')) || DEFAULT_LIMIT, MAX_LIMIT));

	let projectsPromise: Promise<PaginatedResponse<ProjectStat>> | null = $state(null);

	let abortController: AbortController | null = null;
	function loadProjects() {
		if (abortController) {
			abortController.abort();
		}

		abortController = new AbortController();

		projectsPromise = getProjectStats(
			{ q, page: pageNum, limit, status: status },
			abortController.signal
		);
	}

	function updateUrlParam() {
		if (q) {
			params.set('q', q);
		} else {
			params.delete('q');
		}

		params.set('page', pageNum.toString());
		params.set('limit', limit.toString());

		if (status) {
			params.set('status', status);
		} else {
			params.delete('status');
		}

		history.replaceState(null, '', `?${params.toString()}`);
	}

	function handleFilter() {
		pageNum = 1;
		if (limit > MAX_LIMIT) limit = MAX_LIMIT;
		if (limit < MIN_LIMIT) limit = MIN_LIMIT;
		updateUrlParam();
		loadProjects();
	}

	onMount(() => {
		loadProjects();
	});
</script>

<svelte:head>
	<title>Projects</title>
</svelte:head>

<div class="grid h-full auto-rows-[min-content_1fr_min-content] gap-6">
	<div class="mx-auto flex gap-4 lg:container">
		<div class="max-w-96">
			<SearchBar bind:value={q} onchange={handleFilter} />
		</div>
		<div class="flex gap-2">
			<FilterSelect
				bind:value={status}
				onchange={handleFilter}
				name="Status"
				options={[
					{ value: '', label: 'All' },
					{ value: 'started', label: 'Started' },
					{ value: 'paused', label: 'Paused' },
					{ value: 'cancelled', label: 'Cancelled' },
					{ value: 'completed', label: 'Completed' }
				]}
			/>
			<FilterInput
				id="limit"
				max={MAX_LIMIT}
				min={MIN_LIMIT}
				label="Limit"
				type="number"
				bind:value={limit}
				onchange={handleFilter}
			/>
		</div>
	</div>

	{#await projectsPromise}
		<Spinner />
	{:then res}
		{#if res && res.items}
			<div class="mx-auto gap-4 overflow-y-auto lg:container">
				{#if res.items.length == 0}
					<ErrorMessage variant="info" text="Projects Not Found" />
				{/if}
				<div class="overflow-y-auto">
					{#if res.items.length > 0}
						<Table.Root class="container mx-auto">
							<Table.Header>
								<Table.Row>
									<Table.Head class="font-bold">Project Name</Table.Head>
									<Table.Head class="font-bold">Status</Table.Head>
									<Table.Head class="font-bold">Tasks</Table.Head>
									<Table.Head class="font-bold">Created Date</Table.Head>
									<Table.Head class="font-bold">Invoices</Table.Head>
									<Table.Head class="font-bold">Quotations</Table.Head>
									<Table.Head class="font-bold">Actions</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#each res.items as project (project)}
									<Table.Row>
										<Table.Cell>{project.name}</Table.Cell>
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
											{toTitleCase(project.status)}
										</Table.Cell>
										<Table.Cell>{project.tasksCompleted} Completed</Table.Cell>
										<Table.Cell>{formatDate(project.createdAt)}</Table.Cell>
										<Table.Cell>{project.invoicesPaid} / {project.totalInvoices} Paid</Table.Cell>
										<Table.Cell>{project.totalQuotes}</Table.Cell>
										<Table.Cell>
											<div
												class="text-xs text-neutral-500 *:cursor-pointer *:px-1.5 *:hover:text-neutral-950 dark:text-neutral-400 *:dark:hover:text-neutral-50"
											>
												<a
													href={resolve(`/projects/${project.id}`)}
													class="inline-block"
													title="View"
												>
													<ArrowRight size={18} />
												</a>
											</div>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</div>
			</div>
			{#if res.items.length > 0}
				<div class="container mx-auto flex justify-end">
					<Pagination bind:page={pageNum} count={res.totalCount} perPage={res.limit} />
				</div>
			{/if}
		{/if}
	{:catch err}
		{#if err instanceof ApiError}
			<ErrorMessage variant="warn" text={err.message} retry={loadProjects} />
		{:else}
			<ErrorMessage variant="warn" text="An Error Occurred" retry={loadProjects} />
		{/if}
	{/await}
</div>
