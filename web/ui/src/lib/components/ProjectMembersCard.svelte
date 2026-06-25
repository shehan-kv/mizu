<script lang="ts">
	import { getProjectMembers, type ProjectMember } from '$lib/api/projects';
	import { onMount } from 'svelte';
	import ErrorMessage from './ErrorMessage.svelte';
	import Spinner from './Spinner.svelte';
	import UserCard from './UserCard.svelte';
	import { ApiError } from '$lib/api/client';

	interface Props {
		projectId: string;
	}

	let { projectId }: Props = $props();

	let members: Promise<ProjectMember[]> | null = $state(null);
	let membersAbort: AbortController | null = null;
	function loadMembers() {
		if (membersAbort) {
			membersAbort.abort();
		}
		membersAbort = new AbortController();

		members = getProjectMembers(projectId, '', membersAbort.signal);
	}

	export function refresh() {
		loadMembers();
	}

	onMount(() => {
		loadMembers();
	});
</script>

<div class="grid h-full grid-rows-[min-content_1fr] overflow-hidden">
	<div class="border-b px-6 py-2">
		<p class="text-sm">Members</p>
	</div>

	<div class="space-y-2 overflow-scroll px-6 py-4">
		{#await members}
			<Spinner />
		{:then res}
			{#if res && res.length > 0}
				{#each res as member (member.id)}
					<UserCard
						id={member.id}
						hasImage={member.hasImage}
						role={member.role}
						title={member.title}
						name={`${member.firstName} ${member.lastName}`}
					/>
				{/each}
			{:else}
				<ErrorMessage variant="warn" text="Members Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof ApiError}
				<ErrorMessage variant="warn" text={err.message} retry={loadMembers} />
			{:else}
				<ErrorMessage variant="warn" text="An Error Occurred" retry={loadMembers} />
			{/if}
		{/await}
	</div>
</div>
