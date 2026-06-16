<script lang="ts">
	import { page } from '$app/state';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import { formatDate } from '$lib/utils/formatDate';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import { onMount } from 'svelte';
	import ProjectMembersCard from '$lib/components/ProjectMembersCard.svelte';
	import InvoiceListCard from '$lib/components/ProjectInvoiceListCard.svelte';
	import ProjectInvoicePaidChartCard from '$lib/components/ProjectInvoicePaidChartCard.svelte';
	import ProjectTasksCompletedChartCard from '$lib/components/ProjectTasksCompletedChartCard.svelte';
	import ProjectFilesListCard from '$lib/components/ProjectFilesListCard.svelte';
	import ProjectContractsListCard from '$lib/components/ProjectContractsListCard.svelte';
	import ProjectKanbanCard from '$lib/components/ProjectKanbanCard.svelte';
	import { getProject, type Project } from '$lib/api/projects';
	import { ApiError } from '$lib/api/client';

	let id = page.params.id || '';

	let projectPromise: Promise<Project> | null = $state(null);
	let projectAbort: AbortController | null = null;

	function loadProject() {
		if (projectAbort) {
			projectAbort.abort();
		}
		projectAbort = new AbortController();

		projectPromise = getProject(id, projectAbort.signal);
	}

	onMount(() => {
		loadProject();
	});
</script>

<svelte:head>
	<title>View Project</title>
</svelte:head>

<div class="mx-auto grid grid-cols-4 gap-2 lg:container">
	<div class="space-y-2 rounded bg-neutral-50 p-6 dark:bg-neutral-950">
		{#await projectPromise}
			<Spinner />
		{:then res}
			{#if res}
				<div>
					<p class="text-xs text-neutral-500">Name</p>
					<p>{res.name}</p>
				</div>
				<div>
					<p class="text-xs text-neutral-500">Status</p>
					<div class="flex items-center gap-1.5 text-sm">
						{#if res.status == 'started'}
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
						{toTitleCaseDashed(res.status)}
					</div>
				</div>
				<div>
					<p class="text-xs text-neutral-500">Created On</p>
					<p class="text-sm">{formatDate(res.createdAt)}</p>
				</div>
				<div class="mt-6 grid grid-cols-2 gap-x-20 gap-y-3">
					<div>
						<p class="text-xs text-neutral-500">Invoices</p>
						<p class="text-sm">
							{res.invoicePaidCount} / {res.invoiceCount} Paid
						</p>
					</div>
					<div>
						<p class="text-xs text-neutral-500">Quotations</p>
						<p class="text-sm">{res.quoteCount}</p>
					</div>

					<div>
						<p class="text-xs text-neutral-500">Tasks</p>
						<p class="text-sm">
							{res.taskCompletedCount} / {res.taskCount} Completed
						</p>
					</div>
					<div>
						<p class="text-xs text-neutral-500">Contracts</p>
						<p class="text-sm">
							{res.contractSignedCount} / {res.contractCount} Signed
						</p>
					</div>
					<div>
						<p class="text-xs text-neutral-500">Files</p>
						<p class="text-sm">{res.fileCount}</p>
					</div>
				</div>
			{:else}
				<ErrorMessage variant="warn" text="Project Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof ApiError}
				<ErrorMessage variant="warn" text={err.message} retry={loadProject} />
			{:else}
				<ErrorMessage variant="warn" text="An Error Occurred" retry={loadProject} />
			{/if}
		{/await}
	</div>

	<div class="max-h-100 min-h-50 rounded bg-neutral-50 dark:bg-neutral-950">
		<ProjectMembersCard projectId={id} />
	</div>

	<div class="col-span-2 min-h-80 rounded bg-neutral-50 dark:bg-neutral-950">
		<ProjectInvoicePaidChartCard projectId={id} />
	</div>

	<div class="col-span-4 grid h-84 grid-cols-4 gap-2 overflow-hidden">
		<div class="col-span-2 rounded bg-neutral-50 dark:bg-neutral-950">
			<ProjectTasksCompletedChartCard projectId={id} />
		</div>

		<div class="col-span-2 rounded bg-neutral-50 dark:bg-neutral-950">
			<ProjectFilesListCard projectId={id} />
		</div>
	</div>

	<div class="col-span-4 max-h-100 min-h-60 rounded bg-neutral-50 dark:bg-neutral-950">
		<InvoiceListCard projectId={id} role="client" />
	</div>

	<div class="col-span-4 max-h-100 min-h-60 rounded bg-neutral-50 dark:bg-neutral-950">
		<ProjectContractsListCard projectId={id} role="client" />
	</div>

	<div class="col-span-4 max-h-100 rounded bg-neutral-50 dark:bg-neutral-950">
		<ProjectKanbanCard projectId={id} role="client" />
	</div>
</div>
