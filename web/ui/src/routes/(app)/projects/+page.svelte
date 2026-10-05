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
	import { getProjectStats, type ProjectStat, type ProjectStatus } from '$lib/api/projects';

	import { onDestroy, onMount } from 'svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import FilterSelect from '$lib/components/FilterSelect.svelte';
	import { formatDate } from '$lib/utils/formatDate';
	import { createDialogState } from '$lib/components/dialogs/createDialogState.svelte';
	import ListChecks from 'phosphor-svelte/lib/ListChecks';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import type { PaginatedResponse } from '$lib/api/page';
	import { ApiError } from '$lib/api/client';
	import { PROJECT_STATUS } from '$lib/constants/project';
	import { auth } from '$lib/auth/auth.svelte';

	const MAX_LIMIT = 100;
	const DEFAULT_LIMIT = 25;
	const DEFAULT_PAGE_NUMBER = 1;

	const params = new SvelteURLSearchParams(page.url.searchParams.toString());

	let q = $state(params.get('q') || '');
	let status = $state(PROJECT_STATUS.find((s) => s === params.get('status')) ?? '');
	let pageNum = $state(Number(params.get('page')) || DEFAULT_PAGE_NUMBER);

	const limitParam = Number(params.get('limit'));
	let limit = $state(
		Number.isFinite(limitParam)
			? Math.min(Math.max(limitParam, DEFAULT_LIMIT), MAX_LIMIT).toString()
			: DEFAULT_LIMIT.toString()
	);

	let promise: Promise<PaginatedResponse<ProjectStat>> | null = $state(null);

	let abort: AbortController | null = null;
	function loadProjects() {
		if (abort) {
			abort.abort();
		}

		abort = new AbortController();

		promise = getProjectStats(
			{ q, page: pageNum, limit: Number(limit), status: status },
			abort.signal
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
		updateUrlParam();
		loadProjects();
	}

	onMount(() => {
		loadProjects();
	});

	onDestroy(() => {
		abort?.abort();
	});

	const newProjectDialog = createDialogState();
	const projectStatusDialog = createDialogState();
	const projectDeleteDialog = createDialogState();

	type SelectedProject = ProjectStat & { action?: ProjectStatus };

	let selectedProject: SelectedProject | null = $state(null);

	function openStatusDialog(project: ProjectStat, action: ProjectStatus) {
		selectedProject = { ...project, action };
		projectStatusDialog.open();
	}

	function openDeleteDialog(project: ProjectStat) {
		selectedProject = { ...project };
		projectDeleteDialog.open();
	}
</script>

<svelte:head>
	<title>Projects</title>
</svelte:head>

<div
	class="grid h-full auto-rows-[min-content_1fr_min-content] gap-6 rounded bg-neutral-50 p-4 dark:bg-neutral-950"
>
	<div class="mx-auto flex justify-between lg:container">
		<div class="flex gap-2">
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
				<FilterSelect
					bind:value={limit}
					onchange={handleFilter}
					name="Limit"
					options={[
						{ value: '25', label: '25' },
						{ value: '50', label: '50' },
						{ value: '75', label: '75' },
						{ value: '100', label: '100' }
					]}
				/>
			</div>
		</div>
		{#if auth.role == 'administrator'}
			<button
				onclick={newProjectDialog.open}
				class="cursor-pointer rounded bg-neutral-800 px-4 py-2 text-xs text-neutral-50 transition
				hover:bg-neutral-950 dark:bg-neutral-200 dark:text-neutral-950 dark:hover:bg-neutral-50"
			>
				New Project
			</button>
		{/if}
	</div>

	{#await promise}
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
								{#each res.items as project (project.id)}
									<Table.Row>
										<Table.Cell>{project.name}</Table.Cell>
										<Table.Cell>
											<div class="flex items-center gap-1.5">
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
											</div>
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

												{#if auth.role == 'administrator' || auth.role == 'staff'}
													<DropdownMenu.Root>
														<DropdownMenu.Trigger
															class="cursor-pointer p-1 hover:text-neutral-950 dark:hover:text-neutral-50"
														>
															<DotsThree size={18} />
														</DropdownMenu.Trigger>
														<DropdownMenu.Content class="mr-4 *:text-xs">
															<DropdownMenu.Item class="py-2">
																<a
																	href={resolve(`/projects/${project.id}/tasks`)}
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

															{#if auth.role == 'administrator'}
																<DropdownMenu.Separator />
																<DropdownMenu.Item
																	class="py-2"
																	onclick={() => openDeleteDialog(project)}
																>
																	<Trash />Delete
																</DropdownMenu.Item>
															{/if}
														</DropdownMenu.Content>
													</DropdownMenu.Root>
												{/if}
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
