<script lang="ts">
	import { getContractRevision, type ContractRevision } from '$lib/api/contracts';
	import type { UserRole } from '$lib/api/users';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import { Dialog } from 'bits-ui';
	import Info from 'phosphor-svelte/lib/Info';
	import X from 'phosphor-svelte/lib/X';
	import ArrowSquareOut from 'phosphor-svelte/lib/ArrowSquareOut';
	import Spinner from '../Spinner.svelte';
	import { formatDate } from '$lib/utils/formatDate';
	import Checks from 'phosphor-svelte/lib/Checks';
	import { createDialogState } from './createDialogState.svelte';
	import ContractRevisionStatusConfirm from './ContractRevisionStatusConfirm.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';
	import ErrorMessage from '../ErrorMessage.svelte';

	interface Props {
		open: boolean;
		contractId: number;
		revisionId: number;
		role?: UserRole;
		onUpdate?: () => any;
	}
	let { open = $bindable(), contractId, revisionId, role = 'client', onUpdate }: Props = $props();

	let revision: Promise<ContractRevision> | null = $state(null);
	let abort: AbortController | null = null;
	function loadRevision() {
		if (abort) {
			abort.abort();
		}
		abort = new AbortController();

		revision = getContractRevision(contractId, revisionId, abort.signal);
	}

	$effect(() => {
		if (open) {
			loadRevision();
		}
	});

	let confirmDialog = createDialogState();
	let confirmStatus: null | 'accepted' | 'rejected' = $state(null);
	function updateStatus(status: 'accepted' | 'rejected') {
		confirmStatus = status;
		confirmDialog.open();
	}

	// svelte-ignore non_reactive_update
	let linksPrefix = '';
	if (role == 'admin') linksPrefix = '/admin';
	if (role == 'staff') linksPrefix = '/staff';
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay
			class="data-[state=open]:animate-in data-[state=closed]:animate-out 
			data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed 
			inset-0 z-50 bg-neutral-100/80 dark:bg-black/80"
		/>
		<Dialog.Content
			class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:slide-out-to-bottom-8 data-[state=closed]:fade-out 
			data-[state=open]:slide-in-from-bottom-8 data-[state=open]:fade-in
			outline-hidden duration-250 
			fixed left-1/2 top-1/2 z-50 grid w-full max-w-xl -translate-x-1/2 -translate-y-1/2 auto-rows-[min-content_1fr] gap-4 overflow-hidden rounded 
			bg-white dark:bg-neutral-950"
		>
			<div class="text-right">
				<Dialog.Close
					class="cursor-pointer rounded-bl bg-neutral-50 px-4 py-2
					transition duration-150
					hover:bg-neutral-950 hover:text-neutral-50 dark:bg-neutral-900
					hover:dark:bg-neutral-50 hover:dark:text-neutral-950"
				>
					<X class="size-3" />
				</Dialog.Close>
			</div>

			{#await revision}
				<Spinner />
			{:then res}
				{#if res}
					<div class="px-6 pb-6">
						<div class="space-y-1">
							<p class="font-bold">{res.title}</p>

							<div
								class="flex w-fit items-center gap-1 rounded bg-neutral-200 px-2 py-1 dark:bg-neutral-800"
							>
								{#if res.status == 'accepted'}
									<Checks class="text-emerald-500" />
								{:else if res.status == 'rejected'}
									<Info weight="fill" class="text-yellow-500" />
								{:else}
									<Info weight="fill" size={14} />
								{/if}
								<p class="text-xs">{toTitleCase(res.status)}</p>
							</div>
						</div>

						<hr class="my-4" />

						<div class="grid grid-cols-2 gap-3">
							<div>
								<p class="text-xs text-neutral-400">Contract</p>
								<p class="flex items-center gap-1 text-sm">
									{res.contractName}
									<a
										href={`${linksPrefix}/contracts/${res.contractId}`}
										class="text-neutral-400 transition hover:text-neutral-50"
									>
										<ArrowSquareOut size={16} />
									</a>
								</p>
							</div>
							<div>
								<p class="text-xs text-neutral-400">Project</p>
								<p class="flex items-center gap-1 text-sm">
									{res.project.name}
									<a
										href={`${linksPrefix}/projects/${res.project.id}`}
										class="text-neutral-400 transition hover:text-neutral-50"
									>
										<ArrowSquareOut size={16} />
									</a>
								</p>
							</div>
							<div>
								<p class="text-xs text-neutral-400">Created At</p>
								<p class="text-sm">{formatDate(res.createdAt)}</p>
							</div>
							<div>
								<p class="text-xs text-neutral-400">Updated At</p>
								<p class="text-sm">{res.updatedAt ? formatDate(res.updatedAt) : 'N/A'}</p>
							</div>
							<div>
								<p class="text-xs text-neutral-400">Started By</p>
								<p class="text-sm">{res.reqUser.firstName} {res.reqUser.lastName}</p>
							</div>
							<div>
								<p class="text-xs text-neutral-400">Responded By</p>
								<p class="text-sm">
									{res.resUser ? `${res.resUser.firstName} ${res.resUser.lastName}` : 'N/A'}
								</p>
							</div>
						</div>

						<hr class="my-4" />

						<p class="max-h-50 overflow-y-scroll text-sm">{res.description}</p>

						{#if res.status == 'pending' && (role == 'admin' || role == 'staff')}
							<div
								class="mt-6 space-x-1 text-right text-xs *:cursor-pointer *:rounded *:px-6 *:py-3"
							>
								<button
									onclick={() => updateStatus('rejected')}
									class="bg-red-700 text-red-50 transition hover:bg-red-800
                    			dark:bg-red-800 dark:text-red-50 dark:hover:bg-red-700"
								>
									Reject
								</button>
								<button
									onclick={() => updateStatus('accepted')}
									class="bg-neutral-800 text-neutral-50 transition hover:bg-neutral-950
                    			dark:bg-neutral-200 dark:text-neutral-950 dark:hover:bg-neutral-50"
								>
									Accept
								</button>
							</div>
						{/if}
					</div>
				{:else}
					<ErrorMessage variant="info" text="Not Found" retry={loadRevision} />
				{/if}
			{:catch err}
				{#if err instanceof APIBadRequestError}
					<ErrorMessage variant="warn" text="Invalid Request" retry={loadRevision} />
				{:else if err instanceof APIForbiddenError}
					<ErrorMessage
						variant="warn"
						text="You Don't Have Permission To View This Revision Request"
						retry={loadRevision}
					/>
				{:else if err instanceof APINotFoundError}
					<ErrorMessage variant="info" text="Not Found" retry={loadRevision} />
				{:else if err instanceof APIServerError}
					<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadRevision} />
				{:else}
					<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadRevision} />
				{/if}
			{/await}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

{#if confirmStatus}
	<ContractRevisionStatusConfirm
		bind:open={confirmDialog.isOpen}
		{contractId}
		{revisionId}
		status={confirmStatus}
		onSuccess={() => {
			loadRevision();
			onUpdate && onUpdate();
		}}
	/>
{/if}
