<script lang="ts">
	import Envelope from 'phosphor-svelte/lib/Envelope';
	import CheckCircle from 'phosphor-svelte/lib/CheckCircle';
	import WarningCircle from 'phosphor-svelte/lib/WarningCircle';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';
	import { formatDate } from '$lib/utils/formatDate';
	import { createDialogState } from './dialogs/createDialogState.svelte';
	import ErrorMessage from './ErrorMessage.svelte';
	import ConfirmSignContract from './dialogs/ConfirmSignContract.svelte';
	import ConfirmRejectContract from './dialogs/ConfirmRejectContract.svelte';
	import { type Contract } from '$lib/api/contracts';
	import { toTitleCaseDashed } from '$lib/utils/toTitleCaseDashed';

	interface Props {
		contract: Contract;
		refresh?: () => unknown;
	}
	let { contract, refresh }: Props = $props();

	let confirmSignDialog = createDialogState();
	let confirmRejectDialog = createDialogState();
</script>

{#snippet cardTitle(text: string)}
	<p class="mb-4 text-xs text-neutral-700 dark:text-neutral-400">{text}</p>
{/snippet}

<div class="grid grid-cols-4 gap-2 overflow-hidden">
	<div
		class="col-span-3 grid h-full auto-rows-[min-content_1fr] gap-2 overflow-y-auto rounded border"
	>
		<div
			class="sticky top-0 flex items-center justify-between bg-neutral-100 p-2 px-6 dark:bg-neutral-900"
		>
			<p class="text-sm">
				{contract.name} - Created At {formatDate(contract.createdAt)}
			</p>

			<div
				class="space-x-1 text-neutral-700 *:cursor-pointer *:px-2 *:py-1.5 *:hover:text-neutral-950 dark:text-neutral-400
							*:dark:hover:text-neutral-50"
			>
				<button title="Download Contract As PDF"><DownloadSimple size={18} /> </button>
				<button title="Email Me"><Envelope size={18} /> </button>
			</div>
		</div>
		<div class="overflow-y-auto p-6">
			<p class="max-w-2xl whitespace-pre-line">
				{contract.terms}
			</p>
		</div>
	</div>
	<div class="grid auto-rows-[min-content_1fr] space-y-2 overflow-y-auto">
		<div class="rounded border p-4">
			{@render cardTitle('SIGN CONTRACT')}

			{#if contract.memberSignatoryStatus == 'pending'}
				<div
					class="grid grid-cols-2 gap-2 *:inline-flex
										*:cursor-pointer *:items-center *:justify-center
										*:gap-2 *:py-3 *:text-sm *:transition"
				>
					<button
						onclick={confirmSignDialog.open}
						class="rounded bg-neutral-800 text-neutral-50 hover:bg-neutral-950
											dark:bg-neutral-200 dark:text-neutral-950 dark:hover:bg-neutral-50"
					>
						<CheckCircle size={18} />
						Sign
					</button>
					<button
						onclick={confirmRejectDialog.open}
						class="rounded hover:bg-neutral-200 dark:text-neutral-50
											dark:hover:bg-neutral-900"
					>
						<WarningCircle size={18} />
						Reject
					</button>
				</div>
			{:else if contract.memberSignatoryStatus == 'signed'}
				<p class="inline-flex items-center gap-1 text-sm">
					<CheckCircle size={20} class="text-emerald-500" /> You've Already Signed This Contract
				</p>
			{:else if contract.memberSignatoryStatus == 'rejected'}
				<p class="inline-flex items-center gap-1 text-sm">
					<CheckCircle size={18} class="text-rose-500" /> You've Already Rejected This Contract
				</p>
			{/if}
		</div>

		<div class="rounded border p-4">
			{@render cardTitle('SIGNATORIES')}

			<div class="space-y-3">
				{#if contract.signatories.length > 0}
					{#each contract.signatories as signatory (signatory.id)}
						<div class="flex gap-1.5">
							<div class="size-10 rounded-full bg-neutral-200/80 dark:bg-neutral-800">
								{#if signatory.image}
									<img
										src={signatory.image}
										alt={`${signatory.firstName} ${signatory.lastName} profile picture`}
										class="size-full object-cover"
									/>
								{:else}
									<div class="flex size-full items-center justify-center text-xs text-neutral-500">
										{signatory.firstName[0]}
									</div>
								{/if}
							</div>
							<div>
								<p class="text-sm">{signatory.firstName} {signatory.lastName}</p>
								<p class="text-xs text-neutral-600 dark:text-neutral-400">
									{signatory.status == 'pending'
										? 'Pending Signature'
										: toTitleCaseDashed(signatory.status)}
								</p>
							</div>
						</div>
					{/each}
				{:else}
					<ErrorMessage variant="info" text="Signatures Not Found" />
				{/if}
			</div>
		</div>
	</div>
</div>

<ConfirmSignContract
	bind:open={confirmSignDialog.isOpen}
	contractId={contract.id}
	onSuccess={refresh}
/>

<ConfirmRejectContract
	bind:open={confirmRejectDialog.isOpen}
	contractId={contract.id}
	onSuccess={refresh}
/>
