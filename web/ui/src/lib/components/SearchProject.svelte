<script lang="ts">
	import { getUsers } from '$lib/api/users';
	import { debounce } from '$lib/utils/debounce';
	import MagnifyingGlass from 'phosphor-svelte/lib/MagnifyingGlass';
	import Spinner from './Spinner.svelte';
	import UserCard from './UserCard.svelte';
	import ErrorMessage from './ErrorMessage.svelte';
	import { getProjects, type Project } from '$lib/api/projects';
	import { toTitleCase } from '$lib/utils/toTitleCase';

	interface Props {
		onSelect: (project: Project) => any;
	}

	let { onSelect }: Props = $props();

	let searchTerm = $state('');

	let projectsPromise: Promise<PaginatedResponse<Project>> | null = $state(null);
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

		projectsPromise = getProjects(searchTerm.trim(), 1, 50, undefined, abort.signal);
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
		class="outline-hidden peer grow p-2 text-sm placeholder:text-xs placeholder:italic"
		placeholder="Search For Projects..."
	/>
	<MagnifyingGlass size={16} class="mx-2" />

	<div
		class="max-h-50 absolute left-0 top-10 hidden min-h-10 w-full overflow-scroll
									rounded bg-neutral-900 px-2 py-3 ring-0 transition"
	>
		{#if !searchTerm && !projectsPromise}
			<p class="text-xs text-neutral-300">Start Typing To Search</p>
		{/if}

		{#await projectsPromise}
			<Spinner size={16} />
		{:then res}
			{#if res && res.data.length > 0}
				<div>
					{#each res.data as project (project)}
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
							<p class="text-xs text-neutral-400">{toTitleCase(project.status)}</p>
						</div>
					{/each}
				</div>
			{:else if res && searchTerm}
				<ErrorMessage variant="info" text="Projects Not Found" />
			{/if}
		{/await}
	</div>
</div>
