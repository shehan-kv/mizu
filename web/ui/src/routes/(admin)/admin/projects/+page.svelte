<script lang="ts">
	import { page } from '$app/state';
	import * as Table from '$lib/components/ui/table';
	import * as Dialog from '$lib/components/dialogs';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import SearchBar from '$lib/components/SearchBar.svelte';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import Trash from 'phosphor-svelte/lib/Trash';
	import DotsThree from 'phosphor-svelte/lib/DotsThree';
	import Pagination from '$lib/components/Pagination.svelte';
	import { getProjects, type Project, type ProjectStatus } from '$lib/api/projects';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { onMount } from 'svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import FilterSelect from '$lib/components/FilterSelect.svelte';
	import FilterInput from '$lib/components/FilterInput.svelte';
	import { formatDate } from '$lib/utils/formatDate';
	import { createDialogState } from '$lib/components/dialogs/createDialogState.svelte';
	import ListChecks from 'phosphor-svelte/lib/ListChecks';

	const MAX_LIMIT = 100;
	const MIN_LIMIT = 1;
	const DEFAULT_LIMIT = 30;
	const DEFAULT_PAGE_NUMBER = 1;

	const params = new URLSearchParams(page.url.searchParams.toString());

	let q = $state(params.get('q') || '');
	let status = $state(params.get('status') || '');
	let pageNum = $state(Number(params.get('page')) || DEFAULT_PAGE_NUMBER);
	let limit = $state(Math.min(Number(params.get('limit')) || DEFAULT_LIMIT, MAX_LIMIT));

	let projectsPromise: Promise<PaginatedResponse<Project>> | null = $state(null);

	let abortController: AbortController | null = null;
	function loadProjects() {
		if (abortController) {
			abortController.abort();
		}

		abortController = new AbortController();

		projectsPromise = getProjects(q, pageNum, limit, status, abortController.signal);
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

	const newProjectDialog = createDialogState();
	const projectStatusDialog = createDialogState();
	const projectDeleteDialog = createDialogState();

	type SelectedProject = Project & { action?: ProjectStatus };

	let selectedProject: SelectedProject | null = $state(null);

	function openStatusDialog(project: Project, action: ProjectStatus) {
		selectedProject = { ...project, action };
		projectStatusDialog.open();
	}

	function openDeleteDialog(project: Project) {
		selectedProject = { ...project };
		projectDeleteDialog.open();
	}
</script>

<svelte:head>
	<title>Projects</title>
</svelte:head>

<div class="grid h-full auto-rows-[min-content_1fr_min-content] gap-6">
	<div class="space-y-2">
		<div>
			<button
				onclick={newProjectDialog.open}
				class="cursor-pointer rounded bg-neutral-800 px-4 py-2 text-xs text-neutral-50 transition
				hover:bg-neutral-950 dark:bg-neutral-200 dark:text-neutral-950 dark:hover:bg-neutral-50"
			>
				New Project
			</button>
		</div>
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
	</div>

	{#await projectsPromise}
		<Spinner />
	{:then res}
		{#if res && res.data}
			<div class="mx-auto gap-4 overflow-y-auto lg:container">
				{#if res.data.length == 0}
					<ErrorMessage variant="info" text="Projects Not Found" />
				{/if}
				<div class="overflow-y-auto">
					{#if res.data.length > 0}
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
								{#each res.data as project}
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
												<a href={`/admin/projects/${project.id}`} class="inline-block" title="View">
													<ArrowRight size={18} />
												</a>

												<DropdownMenu.Root>
													<DropdownMenu.Trigger
														class="cursor-pointer p-1 hover:text-neutral-950 dark:hover:text-neutral-50"
													>
														<DotsThree size={18} />
													</DropdownMenu.Trigger>
													<DropdownMenu.Content class="mr-4 *:text-xs">
														<DropdownMenu.Item class="py-2">
															<a
																href={`/admin/projects/${project.id}/tasks`}
																class="flex items-center gap-3"
															>
																<ListChecks size={18} />View Tasks
															</a>
														</DropdownMenu.Item>
														<DropdownMenu.Separator />
														<DropdownMenu.Group class="text-xs">
															<DropdownMenu.Label class="text-xs">Mark As</DropdownMenu.Label>
															{#if project.status != 'started'}
																<DropdownMenu.Item
																	class="pl-4 text-xs"
																	onclick={() => openStatusDialog(project, 'started')}
																>
																	Started
																</DropdownMenu.Item>
															{/if}
															{#if project.status != 'paused'}
																<DropdownMenu.Item
																	class="pl-4 text-xs"
																	onclick={() => openStatusDialog(project, 'paused')}
																>
																	Paused
																</DropdownMenu.Item>
															{/if}
															{#if project.status != 'cancelled'}
																<DropdownMenu.Item
																	class="pl-4 text-xs"
																	onclick={() => openStatusDialog(project, 'cancelled')}
																>
																	Cancelled
																</DropdownMenu.Item>
															{/if}
															{#if project.status != 'completed'}
																<DropdownMenu.Item
																	class="pl-4 text-xs"
																	onclick={() => openStatusDialog(project, 'completed')}
																>
																	Completed
																</DropdownMenu.Item>
															{/if}
														</DropdownMenu.Group>
														<DropdownMenu.Separator />
														<DropdownMenu.Item
															class="py-2"
															onclick={() => openDeleteDialog(project)}
														>
															<Trash />Delete
														</DropdownMenu.Item>
													</DropdownMenu.Content>
												</DropdownMenu.Root>
											</div>
										</Table.Cell>
									</Table.Row>
								{/each}
							</Table.Body>
						</Table.Root>
					{/if}
				</div>
			</div>
			{#if res.data.length > 0}
				<div class="container mx-auto flex justify-end">
					<Pagination bind:page={pageNum} count={res.count} perPage={res.limit} />
				</div>
			{/if}
		{/if}
	{:catch err}
		{#if err instanceof APIBadRequestError}
			<ErrorMessage variant="warn" text="Invalid Request" retry={loadProjects} />
		{:else if err instanceof APIForbiddenError}
			<ErrorMessage
				variant="warn"
				text="You Don't Have Permission To View These Projects"
				retry={loadProjects}
			/>
		{:else if err instanceof APINotFoundError}
			<ErrorMessage variant="info" text="Not Found" retry={loadProjects} />
		{:else if err instanceof APIServerError}
			<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadProjects} />
		{:else}
			<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadProjects} />
		{/if}
	{/await}
</div>

<Dialog.NewProject bind:open={newProjectDialog.isOpen} onSuccess={loadProjects} />

{#if selectedProject}
	<Dialog.ProjectDeleteDialog
		bind:open={projectDeleteDialog.isOpen}
		projectId={selectedProject.id}
		onSuccess={loadProjects}
	/>

	{#if selectedProject.action}
		<Dialog.ProjectStatusConfirm
			bind:open={projectStatusDialog.isOpen}
			status={selectedProject.action}
			projectId={selectedProject.id}
			onSuccess={loadProjects}
		/>
	{/if}
{/if}
