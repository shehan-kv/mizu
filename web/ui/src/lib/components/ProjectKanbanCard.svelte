<script lang="ts">
	import ArrowRight from 'phosphor-svelte/lib/ArrowRight';
	import KanbanTaskList from './KanbanTaskList.svelte';
	import { resolve } from '$app/paths';

	interface Props {
		projectId: string;
	}
	let { projectId }: Props = $props();
</script>

<div class="grid h-full w-full grid-rows-[min-content_1fr] overflow-hidden">
	<div class="flex items-center justify-between border-b px-6 py-2">
		<p class="text-sm">Kanban Board</p>
		<a href={resolve(`/projects/${projectId}/kanban`)} class="flex items-center gap-1 text-sm">
			<span>View</span>
			<ArrowRight />
		</a>
	</div>
	<div class="grid grid-cols-3 gap-2 overflow-scroll px-6 py-2">
		<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
			<p class="py-3 text-center text-sm">Backlog</p>
			<div class="overflow-scroll">
				<KanbanTaskList {projectId} status="backlog" />
			</div>
		</div>
		<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
			<p class="py-3 text-center text-sm">In-Progress</p>
			<div class="overflow-scroll">
				<KanbanTaskList {projectId} status="in-progress" />
			</div>
		</div>
		<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
			<p class="py-3 text-center text-sm">Completed</p>
			<div class="overflow-scroll">
				<KanbanTaskList {projectId} status="completed" />
			</div>
		</div>
	</div>
</div>
