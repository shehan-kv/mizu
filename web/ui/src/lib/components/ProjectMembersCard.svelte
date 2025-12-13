<script lang="ts">
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { getProjectMembers, type ProjectMember } from '$lib/api/projects';
	import { onMount } from 'svelte';
	import ErrorMessage from './ErrorMessage.svelte';
	import Spinner from './Spinner.svelte';
	import UserCard from './UserCard.svelte';

	interface Props {
		projectId: number;
	}

	let { projectId }: Props = $props();

	let members: Promise<ProjectMember[]> | null = $state(null);
	let membersAbort: AbortController | null = null;
	function loadMembers() {
		if (membersAbort) {
			membersAbort.abort();
		}
		membersAbort = new AbortController();

		members = getProjectMembers(projectId, membersAbort.signal);
	}

	export function refresh() {
		loadMembers();
	}

	onMount(() => {
		loadMembers();
	});
</script>

<div class="grid h-full grid-rows-[min-content_1fr] overflow-hidden rounded border">
	<div class="bg-neutral-100 px-6 py-2 dark:bg-neutral-900">
		<p class="text-sm">Members</p>
	</div>

	<div class="space-y-2 overflow-scroll px-6 py-4">
		{#await members}
			<Spinner />
		{:then res}
			{#if res && res.length > 0}
				{#each res as member (member)}
					<UserCard
						image={member.image}
						role={member.role}
						title={member.title}
						name={`${member.firstName} ${member.lastName}`}
					/>
				{/each}
			{:else}
				<ErrorMessage variant="warn" text="Members Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof APIBadRequestError}
				<ErrorMessage variant="warn" text="Invalid Request" retry={loadMembers} />
			{:else if err instanceof APIForbiddenError}
				<ErrorMessage
					variant="warn"
					text="You Don't Have Permission To View Members"
					retry={loadMembers}
				/>
			{:else if err instanceof APINotFoundError}
				<ErrorMessage variant="info" text="Not Found" retry={loadMembers} />
			{:else if err instanceof APIServerError}
				<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadMembers} />
			{:else}
				<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadMembers} />
			{/if}
		{/await}
	</div>
</div>
