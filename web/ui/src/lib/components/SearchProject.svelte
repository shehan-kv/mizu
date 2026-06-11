<script lang="ts">
	import { debounce } from '$lib/utils/debounce';
	import MagnifyingGlass from 'phosphor-svelte/lib/MagnifyingGlass';
	import Spinner from './Spinner.svelte';
	import ErrorMessage from './ErrorMessage.svelte';
	import { getProjectStats, type ProjectStat } from '$lib/api/projects';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';
	import type { PaginatedResponse } from '$lib/api/page';

	interface Props {
		onSelect: (project: ProjectStat) => unknown;
	}

	let { onSelect }: Props = $props();

	let searchTerm = $state('');

	let projectsPromise: Promise<PaginatedResponse<ProjectStat>> | null = $state(null);
	let abort: AbortController | null = null;
	function search() {
		if (!searchTerm) {
			if (abort) {
				abort.abort();
				abort = null;
			}

			abort = null;
			return;
		}

		if (abort) {
			abort.abort();
		}

		abort = new AbortController();

		projectsPromise = getProjectStats({ q: searchTerm.trim(), page: 1, limit: 50 }, abort.signal);
	}

	const searchDebounced = debounce(() => {
		search();
	}, 300);
</script>

<div
	class="relative flex items-center gap-1 rounded border
    border-neutral-200 bg-neutral-100
	dark:border-neutral-800 dark:bg-neutral-900 focus-within:[&>div.absolute]:block"
>
	<input
		bind:value={searchTerm}
		oninput={searchDebounced}
		type="text"
		class="peer grow p-2 text-sm outline-hidden placeholder:text-xs placeholder:italic"
		placeholder="Search For Projects..."
	/>
	<MagnifyingGlass size={16} class="mx-2" />

	<div
		class="absolute top-10 left-0 hidden max-h-50 min-h-10 w-full overflow-scroll
									rounded bg-neutral-900 px-2 py-3 ring-0 transition"
	>
		{#if !searchTerm && !projectsPromise}
			<p class="text-xs text-neutral-300">Start Typing To Search</p>
		{/if}

		{#await projectsPromise}
			<Spinner size={16} />
		{:then res}
			{#if res && res.items.length > 0}
				<div>
					{#each res.items as project (project.id)}
						<div
							class="cursor-pointer rounded p-2 hover:bg-neutral-950"
							onmousedown={() => {
								onSelect(project);
							}}
							role="button"
							tabindex="0"
							onkeydown={(e) => {
								if (e.key === 'Enter' || e.key === ' ') {
									e.preventDefault();
									onSelect(project);
								}
							}}
						>
							<p class="text-sm">{project.name}</p>
							<p class="text-xs text-neutral-400">{toTitleCaseDashed(project.status)}</p>
						</div>
					{/each}
				</div>
			{:else if res && searchTerm}
				<ErrorMessage variant="info" text="Projects Not Found" />
			{/if}
		{/await}
	</div>
</div>
