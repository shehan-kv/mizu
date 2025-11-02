<script lang="ts">
	import { getUsers, type User } from '$lib/api/users';
	import { debounce } from '$lib/utils/debounce';
	import MagnifyingGlass from 'phosphor-svelte/lib/MagnifyingGlass';
	import Spinner from './Spinner.svelte';
	import UserCard from './UserCard.svelte';
	import ErrorMessage from './ErrorMessage.svelte';

	interface Props {
		onSelect: (user: User) => any;
	}

	let { onSelect }: Props = $props();

	let searchTerm = $state('');

	let membersPromise: Promise<PaginatedResponse<User>> | null = $state(null);
	let membersAbort: AbortController | null = null;
	function searchMembers() {
		if (!searchTerm) {
			if (membersAbort) {
				membersAbort.abort();
				membersAbort = null;
			}

			membersPromise = null;
			return;
		}

		if (membersAbort) {
			membersAbort.abort();
		}

		membersAbort = new AbortController();

		membersPromise = getUsers({ q: searchTerm.trim(), page: 1, limit: 50 }, membersAbort.signal);
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
		class="outline-hidden peer grow p-2 text-sm placeholder:text-xs placeholder:italic"
		placeholder="Search For Members..."
	/>
	<MagnifyingGlass size={16} class="mx-2" />

	<div
		class="max-h-50 absolute left-0 top-10 hidden min-h-10 w-full overflow-scroll
									rounded bg-neutral-900 px-2 py-3 ring-0 transition"
	>
		{#if !searchTerm && !membersPromise}
			<p class="text-xs text-neutral-300">Start Typing To Search</p>
		{/if}

		{#await membersPromise}
			<Spinner size={16} />
		{:then res}
			{#if res && res.data.length > 0}
				<div>
					{#each res.data as member}
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
								image={member.image}
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
