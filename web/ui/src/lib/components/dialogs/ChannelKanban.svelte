<script lang="ts">
	import FullScreenDialog from './FullScreenDialog.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import KanbanTaskList from '../KanbanTaskList.svelte';
	import type { Channel } from '$lib/api/messages';

	interface Props {
		open: Boolean;
		channel: Channel;
	}
	let { open = $bindable(), channel }: Props = $props();
</script>

<FullScreenDialog bind:open>
	{#if !channel}
		<ErrorMessage variant="info" text="Channel Not Selected" />
	{:else if !channel.projectId}
		<ErrorMessage variant="info" text="Project Not Selected" />
	{:else}
		<div class="grid auto-rows-[min-content_1fr] gap-6 overflow-y-auto px-5">
			<div class="flex-none">
				<div class="container mx-auto">
					<p class="font-bold">Kanban Board - {channel.name}</p>
				</div>
			</div>

			<div
				class="container mx-auto grid h-full auto-rows-[min-content_1fr] grid-cols-3 gap-2 overflow-y-auto"
			>
				<p class="border-b py-3.5 text-center text-sm">Backlog</p>
				<p class="border-b py-3.5 text-center text-sm">In-Progress</p>
				<p class="border-b py-3.5 text-center text-sm">Completed</p>

				<div class="h-full space-y-2 overflow-y-auto">
					<KanbanTaskList projectId={channel.projectId} status="backlog" />
				</div>
				<div class="h-full space-y-2 overflow-y-auto">
					<KanbanTaskList projectId={channel.projectId} status="in-progress" />
				</div>
				<div class="h-full space-y-2 overflow-y-auto">
					<KanbanTaskList projectId={channel.projectId} status="completed" />
				</div>
			</div>
		</div>
	{/if}
</FullScreenDialog>
