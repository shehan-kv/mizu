<script lang="ts">
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import KanbanTaskList from './KanbanTaskList.svelte';
	import type { UserRole } from '$lib/api/users';
	import { resolve } from '$app/paths';

	interface Props {
		projectId: string;
		role?: UserRole;
	}
	let { projectId, role = 'client' }: Props = $props();

	// svelte-ignore non_reactive_update
	let linksPrefix = '';
	if (role == 'administrator') linksPrefix = '/admin';
	if (role == 'staff') linksPrefix = '/staff';
</script>

<div class="grid h-full w-full grid-rows-[min-content_1fr] overflow-hidden rounded border">
	<div class="flex items-center justify-between bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
		<p class="text-sm">Kanban Board</p>
		<a
			href={resolve(`${linksPrefix}/projects/${projectId}/kanban`)}
			class="flex items-center gap-1 text-sm"
		>
			<span>View</span>
			<ArrowRight />
		</a>
	</div>
	<div class="grid grid-cols-3 gap-2 overflow-scroll px-6 py-2">
		<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
			<p class="py-3 text-center text-sm">Backlog</p>
			<div class="overflow-scroll">
				<KanbanTaskList {projectId} status="backlog" role="admin" />
			</div>
		</div>
		<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
			<p class="py-3 text-center text-sm">In-Progress</p>
			<div class="overflow-scroll">
				<KanbanTaskList {projectId} status="in-progress" role="admin" />
			</div>
		</div>
		<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
			<p class="py-3 text-center text-sm">Completed</p>
			<div class="overflow-scroll">
				<KanbanTaskList {projectId} status="completed" role="admin" />
			</div>
		</div>
	</div>
</div>
