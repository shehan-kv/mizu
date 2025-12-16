<script lang="ts">
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Dialog from '$lib/components/dialogs';
	import { getProjectTasks, type ProjectTask, type ProjectTaskStatus } from '$lib/api/projects';
	import Square from 'phosphor-svelte/lib/Square';
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
	import type { UserRole } from '$lib/api/users';
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import DotsThree from 'phosphor-svelte/lib/DotsThree';
	import { createDialogState } from './dialogs/createDialogState.svelte';

	interface Props {
		projectId: number;
		status: ProjectTaskStatus;
		role?: UserRole;
	}
	let { projectId, status, role = 'client' }: Props = $props();

	let isLoading = $state(false);

	let page = $state(1);
	let limit = $state(30);

	let tasks: ProjectTask[] = $state([]);

	let abort: AbortController | null = null;
	let loadError: unknown | null = $state(null);

	async function loadTasks() {
		isLoading = true;
		if (abort) {
			abort.abort();
		}

		abort = new AbortController();

		try {
			const res = await getProjectTasks(projectId, { status, page, limit }, abort.signal);
			tasks.push(...res.data);
			return res;
		} catch (error) {
			loadError = error;
		} finally {
			isLoading = false;
		}
	}

	function refresh() {
		tasks = [];
		loadTasks();
	}

	let container: HTMLDivElement | null = null;

	async function handleScroll() {
		if (!container || isLoading) return;
		const { scrollTop, clientHeight, scrollHeight } = container;

		if (scrollTop + clientHeight >= scrollHeight) {
			page++;
			const res = await loadTasks();

			// Reset the page counter so each new scroll request
			// fetches the next sequential page relative to the
			// pages already loaded; for example, if the current
			// page is 1, repeated scrolls will request page 2
			// when the next fetch is performed.
			if (!res || res.data.length == 0) page--;
		}
	}

	onMount(() => {
		loadTasks();
	});

	let setStatusDialog = createDialogState();

	type SelectedTask = ProjectTask & { action?: ProjectTaskStatus };
	let selectedTask: SelectedTask | null = $state(null);
	function openStatusDialog(task: ProjectTask, action: ProjectTaskStatus) {
		selectedTask = { ...task, action };
		setStatusDialog.open();
	}

	let assigneeDialog = createDialogState();
	function openAssigneeDialog(task: ProjectTask) {
		selectedTask = task;
		assigneeDialog.open();
	}

	// svelte-ignore non_reactive_update
	let linksPrefix = '';
	if (role == 'admin') linksPrefix = '/admin';
	if (role == 'staff') linksPrefix = '/staff';
</script>

<div class="h-full space-y-2 overflow-y-auto" bind:this={container} onscroll={handleScroll}>
	{#if loadError}
		<div>
			{#if loadError instanceof APIBadRequestError}
				<ErrorMessage variant="warn" text="Invalid Request" retry={loadTasks} />
			{:else if loadError instanceof APIForbiddenError}
				<ErrorMessage
					variant="warn"
					text="You Don't Have Permission To View These Project Tasks"
					retry={loadTasks}
				/>
			{:else if loadError instanceof APINotFoundError}
				<ErrorMessage variant="info" text="Not Found" retry={loadTasks} />
			{:else if loadError instanceof APIServerError}
				<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadTasks} />
			{/if}
		</div>
	{/if}

	{#if !loadError}
		{#if tasks.length > 0}
			{#each tasks as task (task)}
				<div class="rounded bg-neutral-50 p-6 dark:bg-neutral-900">
					<div class="flex items-end justify-between text-xs">
						<div class="space-x-2">
							{#if task.priority == 'high'}
								<p class="inline-flex items-center gap-1 text-red-800 dark:text-red-200">
									<Square weight="fill" size={14} class="text-rose-600 dark:text-rose-400" />
									High Priority
								</p>
							{:else if task.priority == 'medium'}
								<p class="inline-flex items-center gap-2 text-amber-800 dark:text-amber-200">
									<Square weight="fill" size={14} class="text-yellow-600 dark:text-amber-500" />
									Medium Priority
								</p>
							{:else if task.priority == 'low'}
								<p class="inline-flex items-center gap-1 text-emerald-800 dark:text-emerald-200">
									<Square weight="fill" size={14} class="text-emerald-600 dark:text-emerald-500" />
									Low Priority
								</p>
							{:else}
								<p>{toTitleCase(task.priority)}</p>
							{/if}
						</div>

						{#if role == 'admin'}
							<div class="flex items-center gap-1 text-right">
								<DropdownMenu.Root>
									<DropdownMenu.Trigger
										class="inline-flex cursor-pointer items-center gap-1 rounded bg-neutral-200 px-2 py-1
									text-xs transition hover:bg-neutral-300 dark:bg-neutral-800 
									dark:hover:bg-neutral-700"
									>
										<DotsThree size={14} />
									</DropdownMenu.Trigger>
									<DropdownMenu.Content class="mr-4 *:text-xs">
										{#if status != 'backlog'}
											<DropdownMenu.Item onclick={() => openStatusDialog(task, 'backlog')}>
												Move To Backlog
											</DropdownMenu.Item>
										{/if}
										{#if status != 'in-progress'}
											<DropdownMenu.Item onclick={() => openStatusDialog(task, 'in-progress')}>
												Move To In-Progress
											</DropdownMenu.Item>
										{/if}
										{#if status != 'completed'}
											<DropdownMenu.Item onclick={() => openStatusDialog(task, 'completed')}>
												Move To Completed
											</DropdownMenu.Item>
										{/if}
										<DropdownMenu.Item onclick={() => openAssigneeDialog(task)}>
											Manage Assignees
										</DropdownMenu.Item>
									</DropdownMenu.Content>
								</DropdownMenu.Root>

								<a
									href={`${linksPrefix}/projects/${projectId}/kanban/${task.id}`}
									class="inline-block cursor-pointer rounded bg-neutral-200 px-2 py-1
									transition hover:bg-neutral-300 dark:bg-neutral-800
									dark:hover:bg-neutral-700"
									title="View"
								>
									<ArrowRight size={14} />
								</a>
							</div>
						{/if}
					</div>

					<div class="mt-4 space-y-0.5 text-xs text-neutral-400">
						<p>Added - {formatDate(task.createdAt)}</p>
						<p class="text-neutral-400">
							{formatMinutes(task.estTimeMinutes)} Estimated
						</p>
					</div>

					<p class="mt-4 text-sm font-bold">{task.name}</p>
					<p class="mt-0.5 text-xs text-neutral-700 dark:text-neutral-400">
						{task.description}
					</p>
					<div class="mt-6">
						<div>
							<p class="text-xs text-neutral-700 dark:text-neutral-400">Assigned To</p>

							<p class="text-sm">
								{task.assignees.map((a) => `${a.firstName} ${a.lastName}`).join(', ')}
							</p>
						</div>
					</div>
				</div>
			{/each}
		{:else}
			<div>
				<ErrorMessage variant="info" text="Tasks Not Found" retry={loadTasks} />
			</div>
		{/if}
	{/if}

	<div>
		{#if isLoading}
			<Spinner />
		{/if}
	</div>
</div>

{#if selectedTask}
	{#if selectedTask.action}
		<Dialog.TaskStatusConfirm
			bind:open={setStatusDialog.isOpen}
			taskId={selectedTask.id}
			status={selectedTask.action}
			onSuccess={refresh}
		/>
	{/if}

	<Dialog.ManageTaskAssignees
		bind:open={assigneeDialog.isOpen}
		{projectId}
		taskId={selectedTask.id}
		onSuccess={refresh}
	/>
{/if}
