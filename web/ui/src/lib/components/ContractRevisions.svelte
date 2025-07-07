<script lang="ts">
	import { onMount } from 'svelte';
	import Spinner from './Spinner.svelte';

	let { contractId } = $props();

	let isLoading = $state(false);
	let revisions = $state([
		{
			id: 1,
			createdDate: new Date().toUTCString(),
			title: 'Request to Update Payment Terms',
			version: 'V1.0.0',
			status: 'APPROVED',
			statusDate: new Date().toUTCString(),
			description:
				'Request to update payment terms from 15-day net to 30-day net payment schedule to align with the latest financial discussions. This affects sections 4.2 and 4.3.'
		},
		{
			id: 2,
			createdDate: new Date().toUTCString(),
			title: 'Request to Add Data Privacy Clause',
			version: 'V1.0.0',
			status: 'REJECTED',
			statusDate: new Date().toUTCString(),
			description:
				'Request clarification on intellectual property ownership created during the contract period, specifying that IP remains with the contracting party unless otherwise agreed.'
		},
		{
			id: 3,
			createdDate: new Date().toUTCString(),
			title: 'Request to Extend Termination Notice Period',
			version: 'V1.0.0',
			status: 'APPROVED',
			statusDate: new Date().toUTCString(),
			description:
				'Request to extend the termination notice period from 30 days to 60 days to allow both parties more time for transition and contract closure.'
		},
		{
			id: 4,
			createdDate: new Date().toUTCString(),
			title: 'Request to Revise Liability and Indemnification Clauses',
			version: 'V1.0.0',
			status: 'APPROVED',
			statusDate: new Date().toUTCString(),
			description:
				'Request to revise liability clauses to limit damages to direct losses only and update indemnification terms to better allocate risk between the parties.'
		}
	]);

	onMount(() => {
		let controller = new AbortController();

		// TODO: Added to simulate a network call. Remove
		isLoading = true;
		let timeout = setTimeout(() => {
			isLoading = false;
		}, 1000);

		// TODO: Fetch revisions here

		return () => {
			// TODO: Remove clearTimeout
			clearTimeout(timeout);
			controller.abort();
		};
	});
</script>

{#if isLoading}
	<Spinner />
{:else}
	<div
		class="before:content-[' '] animate-in fade-in slide-in-from-bottom-2
			dark:before:-z-1 relative space-y-8 duration-500 before:absolute
            before:left-[6px] before:top-1 before:z-auto before:min-h-full
            before:w-1 before:border-l-[1px] before:border-dashed
			before:border-neutral-600"
	>
		{#each revisions as revision, idx (revision)}
			<div data-index={idx} class="flex items-start gap-2">
				<div
					class="z-1 mt-1 size-3 shrink-0 rounded-full border border-2 bg-neutral-50
					dark:z-auto dark:bg-neutral-950"
					class:border-emerald-500={revision.status == 'APPROVED'}
					class:dark:border-emerald-700={revision.status == 'APPROVED'}
					class:border-rose-500={revision.status == 'REJECTED'}
					class:dark:border-rose-700={revision.status == 'REJECTED'}
				></div>
				<div>
					<p class="text-sm">{revision.title}</p>
					<p class="text-xs text-neutral-500">
						{#if revision.status == 'APPROVED'}
							Approved
						{:else if revision.status == 'REJECTED'}
							Rejected
						{/if}
						on {revision.statusDate}
					</p>
					<p class="mt-1.5 text-xs">
						{revision.description}
					</p>
				</div>
			</div>
		{/each}
	</div>
{/if}
