<script lang="ts">
	import { onMount } from 'svelte';
	import Spinner from './Spinner.svelte';
	import { getContractRevisions, type Contract, type ContractRevision } from '$lib/api/contracts';
	import ErrorMessage from './ErrorMessage.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import { formatDate } from '$lib/utils/formatDate';
	import { toTitleCase } from '$lib/utils/toTitleCase';

	interface Props {
		contractId: number;
	}
	let { contractId }: Props = $props();

	let revisionsPromise: Promise<PaginatedResponse<ContractRevision>> | null = $state(null);
	let abortController = new AbortController();

	function loadRevisions() {
		if (abortController) {
			abortController.abort();
		}

		abortController = new AbortController();

		revisionsPromise = getContractRevisions(
			contractId,
			{ page: 1, limit: 30 },
			abortController.signal
		);
	}

	onMount(() => {
		loadRevisions();
	});
</script>

{#await revisionsPromise}
	<Spinner />
{:then res}
	{#if res && res.data.length > 0}
		<div
			class="before:content-[' '] animate-in fade-in slide-in-from-bottom-2
			dark:before:-z-1 relative space-y-8 duration-500 before:absolute
            before:left-[6px] before:top-1 before:z-auto before:min-h-full
            before:w-1 before:border-l-[1px] before:border-dashed
			before:border-neutral-600"
		>
			{#each res.data as revision, idx (revision)}
				<div data-index={idx} class="flex items-start gap-2">
					<div
						class="z-1 mt-1 size-3 shrink-0 rounded-full border border-2 bg-neutral-50
					dark:z-auto dark:bg-neutral-950"
						class:border-emerald-500={revision.status == 'accepted'}
						class:dark:border-emerald-700={revision.status == 'accepted'}
						class:border-rose-500={revision.status == 'rejected'}
						class:dark:border-rose-700={revision.status == 'rejected'}
					></div>
					<div>
						<p class="text-sm">{revision.title}</p>
						<p class="text-xs text-neutral-500">
							{#if revision.status == 'accepted' || revision.status == 'rejected'}
								{toTitleCase(revision.status)}
								{#if revision.updatedAt}
									On {formatDate(revision.updatedAt)}
								{/if}
							{:else}
								{toTitleCase(revision.status)} - Created On {formatDate(revision.createdAt)}
							{/if}
						</p>
						<p class="mt-1.5 text-xs">
							{revision.description}
						</p>
					</div>
				</div>
			{/each}
		</div>
	{:else}
		<ErrorMessage variant="info" text="Revisions Not Found" />
	{/if}
{:catch err}
	{#if err instanceof APIBadRequestError}
		<ErrorMessage variant="warn" text="Invalid Request" retry={loadRevisions} />
	{:else if err instanceof APIForbiddenError}
		<ErrorMessage
			variant="warn"
			text="You Don't Have Permission To View These Revisions"
			retry={loadRevisions}
		/>
	{:else if err instanceof APINotFoundError}
		<ErrorMessage variant="info" text="Not Found" retry={loadRevisions} />
	{:else if err instanceof APIServerError}
		<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadRevisions} />
	{:else}
		<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadRevisions} />
	{/if}
{/await}
