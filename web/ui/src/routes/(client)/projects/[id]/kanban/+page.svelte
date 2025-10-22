<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { getProjectDetails, type ProjectDetails } from '$lib/api/projects';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import KanbanTaskList from '$lib/components/KanbanTaskList.svelte';

	let id = Number(page.params.id);

	let projectPromise: Promise<ProjectDetails> | null = $state(null);
	let projectAbort: AbortController | null = null;
	function loadProject() {
		if (projectAbort) {
			projectAbort.abort();
		}

		projectAbort = new AbortController();

		projectPromise = getProjectDetails(id, projectAbort.signal).then((d) => {
			document.title = 'Kanban Board - ' + d.name;
			return d;
		});
	}

	onMount(() => {
		loadProject();
	});
</script>

<svelte:head>
	<title>Kanban Board</title>
</svelte:head>

<div class="grid h-full auto-rows-[min-content_1fr] gap-6">
	<div class="mx-auto lg:container">
		<div class="flex w-fit items-center gap-3 text-sm text-neutral-700 dark:text-neutral-400">
			{#await projectPromise}
				<p class="">...</p>
			{:then res}
				<a href={`/projects/${res?.id}`} class="underline">{res?.name}</a>
			{/await}

			<ChevronRight size={18} />
			<p>Kanban Board</p>
		</div>
	</div>

	<div class="mx-auto grid h-full auto-rows-[min-content_1fr] gap-6 lg:container">
		<div class="grid grid-cols-3 gap-2 overflow-scroll py-2">
			<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
				<p class="py-3 text-center text-sm">Backlog</p>
				<div class="overflow-scroll">
					<KanbanTaskList projectId={id} status="backlog" />
				</div>
			</div>
			<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
				<p class="py-3 text-center text-sm">In-Progress</p>
				<div class="overflow-scroll">
					<KanbanTaskList projectId={id} status="in-progress" />
				</div>
			</div>
			<div class="grid grid-rows-[min-content_1fr] overflow-hidden">
				<p class="py-3 text-center text-sm">Completed</p>
				<div class="overflow-scroll">
					<KanbanTaskList projectId={id} status="completed" />
				</div>
			</div>
		</div>
	</div>
</div>
