<script lang="ts">
	import { page } from '$app/state';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Dialog from '$lib/components/dialogs';
	import { getProject, type Project, type ProjectStatus } from '$lib/api/projects';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import { formatDate } from '$lib/utils/formatDate';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import { onMount } from 'svelte';
	import CaretDown from 'phosphor-svelte/lib/CaretDown';
	import { createDialogState } from '$lib/components/dialogs/createDialogState.svelte';
	import { goto } from '$app/navigation';
	import ProjectMembersCard from '$lib/components/ProjectMembersCard.svelte';
	import ProjectInvoiceListCard from '$lib/components/ProjectInvoiceListCard.svelte';
	import ProjectInvoicePaidChartCard from '$lib/components/ProjectInvoicePaidChartCard.svelte';
	import ProjectTasksCompletedChartCard from '$lib/components/ProjectTasksCompletedChartCard.svelte';
	import ProjectFilesListCard from '$lib/components/ProjectFilesListCard.svelte';
	import ProjectContractsList from '$lib/components/ProjectContractsListCard.svelte';
	import ProjectKanbanCard from '$lib/components/ProjectKanbanCard.svelte';
	import { resolve } from '$app/paths';
	import { ApiError } from '$lib/api/client';
	import { auth } from '$lib/auth/auth.svelte';

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

	let projectMembersCard: ProjectMembersCard | null = $state(null);
	function refreshMembers() {
		projectMembersCard?.refresh();
	}

	const projectStatusDialog = createDialogState();
	const projectDeleteDialog = createDialogState();
	const manageMembersDialog = createDialogState();
	const newInvoiceDialog = createDialogState();
	const newTaskDialog = createDialogState();
	const newContractDialog = createDialogState();

	let selectedAction: ProjectStatus | null = $state(null);

	function openStatusDialog(action: ProjectStatus) {
		selectedAction = action;
		projectStatusDialog.open();
	}

	// svelte-ignore non_reactive_update
	let invList: ProjectInvoiceListCard;

	// svelte-ignore non_reactive_update
	let completedTasks: ProjectTasksCompletedChartCard;

	// svelte-ignore non_reactive_update
	let cntrList: ProjectContractsList;
</script>

<svelte:head>
	<title>View Project</title>
</svelte:head>

<div class="mx-auto grid grid-cols-4 gap-2 lg:container">
	<div class="col-span-4 rounded bg-neutral-50 p-2 text-right dark:bg-neutral-950">
		{#await projectPromise}
			<span>...</span>
		{:then res}
			{#if res}
				{#if auth.role == 'administrator' || auth.role == 'staff'}
					<DropdownMenu.Root>
						<DropdownMenu.Trigger
							class="inline-flex cursor-pointer items-center gap-1 rounded bg-neutral-100 px-4 py-2
						text-xs transition hover:bg-neutral-300 dark:bg-neutral-900 
						dark:hover:bg-neutral-800"
						>
							New <CaretDown />
						</DropdownMenu.Trigger>
						<DropdownMenu.Content class="mr-4 *:text-xs">
							<DropdownMenu.Item onclick={newInvoiceDialog.open}>Invoice / Quote</DropdownMenu.Item>
							<DropdownMenu.Item onclick={newTaskDialog.open}>Task</DropdownMenu.Item>
							<DropdownMenu.Item onclick={newContractDialog.open}>Contract</DropdownMenu.Item>
						</DropdownMenu.Content>
					</DropdownMenu.Root>

					{#if auth.role == 'administrator'}
						<button
							onclick={manageMembersDialog.open}
							class="cursor-pointer rounded bg-neutral-100 px-4 py-2
						text-xs transition hover:bg-neutral-300 dark:bg-neutral-900
						dark:hover:bg-neutral-800"
						>
							Manage Members
						</button>
					{/if}
					<DropdownMenu.Root>
						<DropdownMenu.Trigger
							class="inline-flex cursor-pointer items-center gap-1 rounded bg-neutral-100 px-4 py-2
						text-xs transition hover:bg-neutral-300 dark:bg-neutral-900 
						dark:hover:bg-neutral-800"
						>
							Mark As <CaretDown />
						</DropdownMenu.Trigger>
						<DropdownMenu.Content class="mr-4 *:text-xs">
							{#if res.status != 'started'}
								<DropdownMenu.Item onclick={() => openStatusDialog('started')}>
									Started
								</DropdownMenu.Item>
							{/if}
							{#if res.status != 'paused'}
								<DropdownMenu.Item onclick={() => openStatusDialog('paused')}>
									Paused
								</DropdownMenu.Item>
							{/if}
							{#if res.status != 'cancelled'}
								<DropdownMenu.Item onclick={() => openStatusDialog('cancelled')}>
									Cancelled
								</DropdownMenu.Item>
							{/if}
							{#if res.status != 'completed'}
								<DropdownMenu.Item onclick={() => openStatusDialog('completed')}>
									Completed
								</DropdownMenu.Item>
							{/if}
						</DropdownMenu.Content>
					</DropdownMenu.Root>

					{#if auth.role == 'administrator'}
						<button
							onclick={projectDeleteDialog.open}
							class="cursor-pointer rounded bg-red-700 px-4 py-2 text-xs text-red-50 transition
						hover:bg-red-600 dark:bg-red-400 dark:text-red-950
						dark:hover:bg-red-500"
						>
							Delete
						</button>
					{/if}
				{/if}
			{/if}
		{/await}
	</div>
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
		<ProjectMembersCard projectId={id} bind:this={projectMembersCard} />
	</div>

	<div class="col-span-2 min-h-80 rounded bg-neutral-50 dark:bg-neutral-950">
		<ProjectInvoicePaidChartCard projectId={id} />
	</div>

	<div class="col-span-4 grid h-84 grid-cols-4 gap-2 overflow-hidden">
		<div class="col-span-2 rounded bg-neutral-50 dark:bg-neutral-950">
			<ProjectTasksCompletedChartCard bind:this={completedTasks} projectId={id} />
		</div>

		<div class="col-span-2 rounded bg-neutral-50 dark:bg-neutral-950">
			<ProjectFilesListCard projectId={id} />
		</div>
	</div>

	<div class="col-span-4 max-h-100 min-h-60 rounded bg-neutral-50 dark:bg-neutral-950">
		<ProjectInvoiceListCard bind:this={invList} projectId={id} />
	</div>

	<div class="col-span-4 max-h-100 min-h-60 rounded bg-neutral-50 dark:bg-neutral-950">
		<ProjectContractsList bind:this={cntrList} projectId={id} />
	</div>

	<div class="col-span-4 max-h-100 rounded bg-neutral-50 dark:bg-neutral-950">
		<ProjectKanbanCard projectId={id} />
	</div>
</div>

{#if selectedAction}
	<Dialog.ProjectStatusConfirm
		projectId={id}
		bind:open={projectStatusDialog.isOpen}
		status={selectedAction}
		onSuccess={loadProject}
	/>
{/if}

<Dialog.ProjectDeleteDialog
	projectId={id}
	bind:open={projectDeleteDialog.isOpen}
	onSuccess={() => goto(resolve('/projects'))}
/>

{#if projectPromise}
	<Dialog.ManageProjectMembers
		projectId={id}
		bind:open={manageMembersDialog.isOpen}
		onSuccess={refreshMembers}
	/>
{/if}

<Dialog.NewInvoice
	projectId={id}
	bind:open={newInvoiceDialog.isOpen}
	onSuccess={invList && invList.refresh}
/>

<Dialog.NewTask
	bind:open={newTaskDialog.isOpen}
	projectId={id}
	onSuccess={completedTasks && completedTasks.refresh}
/>

<Dialog.NewContract
	projectId={id}
	bind:open={newContractDialog.isOpen}
	onSuccess={cntrList && cntrList.refresh}
/>
