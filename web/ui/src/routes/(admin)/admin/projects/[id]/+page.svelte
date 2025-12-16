<script lang="ts">
	import { page } from '$app/state';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Dialog from '$lib/components/dialogs';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { getProjectDetails, type ProjectDetails, type ProjectStatus } from '$lib/api/projects';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import { formatDate } from '$lib/utils/formatDate';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import { onMount } from 'svelte';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import KanbanTaskList from '$lib/components/KanbanTaskList.svelte';
	import CaretDown from 'phosphor-svelte/lib/CaretDown';

	import { createDialogState } from '$lib/components/dialogs/createDialogState.svelte';
	import { goto } from '$app/navigation';
	import ProjectMembersCard from '$lib/components/ProjectMembersCard.svelte';
	import InvoiceListCard from '$lib/components/InvoiceListCard.svelte';
	import ProjectInvoicePaidChart from '$lib/components/ProjectInvoicePaidChart.svelte';
	import ProjectTasksCompletedChart from '$lib/components/ProjectTasksCompletedChart.svelte';
	import ProjectFilesList from '$lib/components/ProjectFilesList.svelte';
	import ProjectContractsList from '$lib/components/ProjectContractsList.svelte';
	import ProjectChangeRequestList from '$lib/components/ProjectChangeRequestList.svelte';

	let id = Number(page.params.id);

	let projectPromise: Promise<ProjectDetails> | null = $state(null);
	let projectAbort: AbortController | null = null;

	function loadProject() {
		if (projectAbort) {
			projectAbort.abort();
		}
		projectAbort = new AbortController();

		projectPromise = getProjectDetails(id, projectAbort.signal);
	}

	onMount(() => {
		loadProject();
	});

	let projectMembersCard: ProjectMembersCard | null = $state(null);
	function refreshMembers() {
		projectMembersCard && projectMembersCard.refresh();
	}

	const projectStatusDialog = createDialogState();
	const projectDeleteDialog = createDialogState();
	const manageMembersDialog = createDialogState();
	const newInvoiceDialog = createDialogState();
	const newChReqDialog = createDialogState();
	const newTaskDialog = createDialogState();
	const newContractDialog = createDialogState();

	let selectedAction: ProjectStatus | null = $state(null);

	function openStatusDialog(action: ProjectStatus) {
		selectedAction = action;
		projectStatusDialog.open();
	}

	// svelte-ignore non_reactive_update
	let invList: InvoiceListCard;

	// svelte-ignore non_reactive_update
	let completedTasks: ProjectTasksCompletedChart;

	// svelte-ignore non_reactive_update
	let cntrList: ProjectContractsList;

	// svelte-ignore non_reactive_update
	let chReqList: ProjectChangeRequestList;
</script>

<svelte:head>
	<title>View Project</title>
</svelte:head>

<div class="mx-auto grid grid-cols-4 gap-2 lg:container">
	<div class="col-span-4 text-right">
		{#await projectPromise}
			<span>...</span>
		{:then res}
			{#if res}
				<DropdownMenu.Root>
					<DropdownMenu.Trigger
						class="inline-flex cursor-pointer items-center gap-1 rounded bg-neutral-200 px-4 py-2
						text-xs transition hover:bg-neutral-300 dark:bg-neutral-900 
						dark:hover:bg-neutral-800"
					>
						New <CaretDown />
					</DropdownMenu.Trigger>
					<DropdownMenu.Content class="mr-4 *:text-xs">
						<DropdownMenu.Item onclick={newInvoiceDialog.open}>Invoice / Quote</DropdownMenu.Item>
						<DropdownMenu.Item onclick={newTaskDialog.open}>Task</DropdownMenu.Item>
						<DropdownMenu.Item onclick={newContractDialog.open}>Contract</DropdownMenu.Item>
						<DropdownMenu.Item onclick={newChReqDialog.open}>Change Request</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
				<button
					onclick={manageMembersDialog.open}
					class="cursor-pointer rounded bg-neutral-200 px-4 py-2
						text-xs transition hover:bg-neutral-300 dark:bg-neutral-900
						dark:hover:bg-neutral-800"
				>
					Manage Members
				</button>
				<DropdownMenu.Root>
					<DropdownMenu.Trigger
						class="inline-flex cursor-pointer items-center gap-1 rounded bg-neutral-200 px-4 py-2
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
				<button
					onclick={projectDeleteDialog.open}
					class="cursor-pointer rounded bg-red-300 px-4 py-2 text-xs text-red-950 transition
						hover:bg-red-400 dark:bg-red-900 dark:text-red-50
						dark:hover:bg-red-800"
				>
					Delete
				</button>
			{/if}
		{/await}
	</div>
	<div class="space-y-2 rounded border p-6">
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
						{toTitleCase(res.status)}
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
						<p class="text-xs text-neutral-500">Change Requests</p>
						<p class="text-sm">
							{res.changeReqClosedCount} / {res.changeReqCount} Closed
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
			{#if err instanceof APIBadRequestError}
				<ErrorMessage variant="warn" text="Invalid Request" retry={loadProject} />
			{:else if err instanceof APIForbiddenError}
				<ErrorMessage
					variant="warn"
					text="You Don't Have Permission To View This Project"
					retry={loadProject}
				/>
			{:else if err instanceof APINotFoundError}
				<ErrorMessage variant="info" text="Not Found" retry={loadProject} />
			{:else if err instanceof APIServerError}
				<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadProject} />
			{:else}
				<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadProject} />
			{/if}
		{/await}
	</div>

	<div class="min-h-50 max-h-100">
		<ProjectMembersCard projectId={id} bind:this={projectMembersCard} />
	</div>

	<div class="col-span-2 min-h-80">
		<ProjectInvoicePaidChart projectId={id} />
	</div>

	<div class="h-84 col-span-4 grid grid-cols-4 gap-2 overflow-hidden">
		<div class="col-span-2">
			<ProjectTasksCompletedChart bind:this={completedTasks} projectId={id} />
		</div>

		<div class="col-span-2">
			<ProjectFilesList projectId={id} />
		</div>
	</div>

	<InvoiceListCard bind:this={invList} projectId={id} role="admin" />

	<div class="min-h-50 max-h-100 col-span-4">
		<ProjectContractsList bind:this={cntrList} projectId={id} role="admin" />
	</div>

	<div class="min-h-50 max-h-100 col-span-4">
		<ProjectChangeRequestList bind:this={chReqList} projectId={id} role="admin" />
	</div>

	<div class="max-h-100 col-span-4 grid grid-rows-[min-content_1fr] overflow-hidden rounded border">
		<div class="flex items-center justify-between bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
			<p class="text-sm">Kanban Board</p>
			<a href={`/projects/${id}/kanban`} class="flex items-center gap-1 text-sm">
				<span>View</span>
				<ArrowRight />
			</a>
		</div>
		<div class="grid grid-cols-3 gap-2 overflow-scroll px-6 py-2">
			<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
				<p class="py-3 text-center text-sm">Backlog</p>
				<div class="overflow-scroll">
					<KanbanTaskList projectId={id} status="backlog" role="admin" />
				</div>
			</div>
			<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
				<p class="py-3 text-center text-sm">In-Progress</p>
				<div class="overflow-scroll">
					<KanbanTaskList projectId={id} status="in-progress" role="admin" />
				</div>
			</div>
			<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
				<p class="py-3 text-center text-sm">Completed</p>
				<div class="overflow-scroll">
					<KanbanTaskList projectId={id} status="completed" role="admin" />
				</div>
			</div>
		</div>
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
	onSuccess={() => goto('/admin/projects')}
/>

{#if projectPromise}
	<Dialog.ManageMembers
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

<Dialog.NewChangeRequest
	bind:open={newChReqDialog.isOpen}
	projectId={id}
	onSuccess={chReqList && chReqList.refresh}
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
