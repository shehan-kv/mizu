<script lang="ts">
	import { debounce } from '$lib/utils/debounce';
	import MagnifyingGlass from 'phosphor-svelte/lib/MagnifyingGlass';
	import Spinner from './Spinner.svelte';
	import UserCard from './UserCard.svelte';
	import ErrorMessage from './ErrorMessage.svelte';
	import { getProjectMembers, type ProjectMember } from '$lib/api/projects';

	interface Props {
		projectId: string;
		onSelect: (member: ProjectMember) => unknown;
	}

	let { projectId, onSelect }: Props = $props();

	let searchTerm = $state('');

	let membersPromise: Promise<ProjectMember[]> | null = $state(null);
	let abort: AbortController | null = null;
	function searchMembers() {
		if (!searchTerm) {
			if (abort) {
				abort.abort();
				abort = null;
			}

			membersPromise = null;
			return;
		}

		if (abort) {
			abort.abort();
		}

		abort = new AbortController();

		membersPromise = getProjectMembers(projectId, searchTerm, abort.signal);
	}

	const memberSearchDebounced = debounce(() => {
		searchMembers();
	}, 300);
</script>

<div
	class="relative flex items-center gap-1 rounded border
    border-neutral-200 bg-neutral-100
	dark:border-neutral-800 dark:bg-neutral-900 focus-within:[&>div.absolute]:block"
>
	<input
		bind:value={searchTerm}
		oninput={memberSearchDebounced}
		type="text"
		class="peer grow p-2 text-sm outline-hidden placeholder:text-xs placeholder:italic"
		placeholder="Search For Project Members..."
	/>
	<MagnifyingGlass size={16} class="mx-2" />

	<div
		class="absolute top-10 left-0 hidden max-h-50 min-h-10 w-full overflow-scroll
									rounded bg-neutral-900 px-2 py-3 ring-0 transition"
	>
		{#if !searchTerm && !membersPromise}
			<p class="text-xs text-neutral-300">Start Typing To Search</p>
		{/if}

		{#await membersPromise}
			<Spinner size={16} />
		{:then res}
			{#if res && res.length > 0}
				<div>
					{#each res as member (member.id)}
						<div
							class="cursor-pointer rounded p-2 hover:bg-neutral-950"
							onmousedown={() => {
								onSelect(member);
							}}
							role="button"
							tabindex="0"
							onkeydown={(e) => {
								if (e.key === 'Enter' || e.key === ' ') {
									e.preventDefault();
									onSelect(member);
								}
							}}
						>
							<UserCard
								id={member.id}
								hasImage={member.hasImage}
								role={member.role}
								title={member.title}
								name={`${member.firstName} ${member.lastName}`}
							/>
						</div>
					{/each}
				</div>
			{:else if res && searchTerm}
				<ErrorMessage variant="info" text="Members Not Found" />
			{/if}
		{/await}
	</div>
</div>
