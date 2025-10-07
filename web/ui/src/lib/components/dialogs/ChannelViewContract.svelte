<script lang="ts">
	import Envelope from 'phosphor-svelte/lib/Envelope';
	import CheckCircle from 'phosphor-svelte/lib/CheckCircle';
	import WarningCircle from 'phosphor-svelte/lib/WarningCircle';
	import ArrowCounterClockwise from 'phosphor-svelte/lib/ArrowCounterClockwise';
	import DownloadSimple from 'phosphor-svelte/lib/DownloadSimple';

	import ContractRevisions from '../ContractRevisions.svelte';
	import Spinner from '../Spinner.svelte';
	import ConfirmRejectContract from './ConfirmRejectContract.svelte';
	import ConfirmSignContract from './ConfirmSignContract.svelte';
	import { createDialogState } from './createDialogState.svelte';
	import FullScreenDialog from './FullScreenDialog.svelte';
	import RequestRevision from './RequestRevision.svelte';
	import { getContractVersions, type Contract, type ContractVersion } from '$lib/api/contracts';
	import { formatDate } from '$lib/utils/formatDate';

	interface Props {
		open: Boolean;
		contract: Contract;
	}
	let { open = $bindable() }: Props = $props();

	let activeSidebar: 'ABOUT' | 'REVISION' = $state('ABOUT');

	let versionsPromise: Promise<ContractVersion[]> | null = $state(null);
	let versionAbortController: AbortController | null = null;

	let selectedVersion: ContractVersion | null = $state(null);

	async function loadVersions() {
		if (versionAbortController) {
			versionAbortController.abort();
		}

		versionAbortController = new AbortController();

		versionsPromise = getContractVersions(contract.id, versionAbortController.signal).then(
			(versions) => {
				if (versions.length > 0) selectedVersion = versions[0];
				return versions;
			}
		);
	}

	let contract = {
		id: 1,
		name: 'Website Redesign',
		userSignedStatus: 'UNSIGNED',
		versions: [
			{ id: 1, version: 'V1.0.0', createdDate: new Date().toUTCString() },
			{ id: 2, version: 'V1.0.1', createdDate: new Date().toUTCString() },
			{ id: 3, version: 'V1.0.2', createdDate: new Date().toUTCString() },
			{ id: 4, version: 'V1.0.3', createdDate: new Date().toUTCString() }
		],

		parties: [
			{
				name: 'Alice',
				signed: true,
				image: null
			},
			{
				name: 'Bob',
				signed: true,
				image: null
			},
			{
				name: 'Yvonne',
				signed: true,
				image: null
			},
			{
				name: 'Lucas Barrett',
				signed: false,
				image: null
			}
		]
	};

	function openContractVersion(version: ContractVersion) {
		selectedVersion = version;
	}

	$effect(() => {
		if (!open) return;
		loadVersions();
	});

	let confirmSignDialog = createDialogState();
	let confirmRejectDialog = createDialogState();
	let requestRevisionDialog = createDialogState();
</script>

{#snippet cardTitle(text: string)}
	<p class="mb-4 text-xs text-neutral-500">{text}</p>
{/snippet}

<FullScreenDialog bind:open>
	{#await versionsPromise}
		<Spinner />
	{:then versions}
		<div class="container mx-auto grid auto-rows-[min-content_1fr] gap-6 overflow-y-auto">
			<div class="space-y-0.5">
				<p class="text-xs text-neutral-500">Contract</p>
				<p class="font-bold">{contract.name}</p>
			</div>

			{#if versions && selectedVersion}
				<div class="grid grid-cols-4 gap-2 overflow-hidden">
					<div
						class="col-span-3 grid h-full auto-rows-[min-content_1fr] gap-2 overflow-y-auto rounded border"
					>
						<div
							class="sticky top-0 flex items-center justify-between bg-neutral-100 p-2 px-6 dark:bg-neutral-900"
						>
							<p class="text-xs">
								{selectedVersion.version} - Created At {formatDate(selectedVersion.createdAt)}
							</p>

							<div
								class="space-x-1 text-neutral-700 *:cursor-pointer *:px-2 *:py-1.5 *:hover:text-neutral-950 dark:text-neutral-400
							*:dark:hover:text-neutral-50"
							>
								<button title="Download Contract"><DownloadSimple size={18} /> </button>
								<button title="Email Me This Version"><Envelope size={18} /> </button>
							</div>
						</div>
						<div class="overflow-y-auto p-6">
							<p class="max-w-2xl whitespace-pre-line">
								{selectedVersion?.contract}
							</p>
						</div>
					</div>
					<div class="grid auto-rows-[min-content_1fr] space-y-2 overflow-y-auto overflow-y-auto">
						<div
							class="grid grid-cols-2 gap-2 text-neutral-500 transition
					*:cursor-pointer *:border-neutral-900 *:px-8 *:py-3 *:text-left
					*:text-xs *:hover:underline"
						>
							<button
								class:dark:text-neutral-50={activeSidebar == 'ABOUT'}
								class:text-neutral-950={activeSidebar == 'ABOUT'}
								onclick={() => (activeSidebar = 'ABOUT')}>ABOUT CONTRACT</button
							>
							<button
								class:dark:text-neutral-50={activeSidebar == 'REVISION'}
								class:text-neutral-950={activeSidebar == 'REVISION'}
								onclick={() => (activeSidebar = 'REVISION')}>REVISION HISTORY</button
							>
						</div>
						{#if activeSidebar == 'ABOUT'}
							<div class="rounded bg-neutral-50 p-8 dark:bg-neutral-900/30">
								{@render cardTitle('PARTIES')}

								<div class="space-y-3">
									{#each contract.parties as party}
										<div class="flex gap-1.5">
											<div class="size-10 rounded-full bg-neutral-200/80 dark:bg-neutral-800">
												{#if party.image}
													<img
														src={party.image}
														alt={`${party.name} profile picture`}
														class="size-full object-cover"
													/>
												{:else if party.name}
													<div
														class="flex size-full items-center justify-center text-xs text-neutral-500"
													>
														{party.name[0]}
													</div>
												{/if}
											</div>
											<div>
												<p class="text-sm">{party.name}</p>
												<p class="text-xs text-neutral-500 dark:text-neutral-400">
													{party.signed ? 'Signed' : 'Pending Signature'}
												</p>
											</div>
										</div>
									{/each}
								</div>
							</div>
							<div class="rounded bg-neutral-50 p-8 dark:bg-neutral-900/30">
								{@render cardTitle('SIGN CONTRACT')}

								{#if contract.userSignedStatus == 'UNSIGNED'}
									<div class="space-y-6">
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
										<button
											onclick={requestRevisionDialog.open}
											class="inline-flex cursor-pointer items-center gap-1.5 text-sm hover:underline"
										>
											<ArrowCounterClockwise size={18} />
											Request Revision</button
										>
									</div>
								{:else if contract.userSignedStatus == 'SIGNED'}
									<p class="inline-flex items-center gap-1 text-sm">
										<CheckCircle size={20} class="text-emerald-500" /> You've Already Signed This Contract
									</p>
								{:else if contract.userSignedStatus == 'REJECTED'}
									<p class="inline-flex items-center gap-1 text-sm">
										<CheckCircle size={18} class="text-rose-500" /> You've Already Rejected This Contract
									</p>
								{/if}
							</div>
							<div class="rounded bg-neutral-50 p-8 dark:bg-neutral-900/30">
								{@render cardTitle('VERSIONS')}
								<div
									class="text-sm text-neutral-700 *:block *:w-full *:cursor-pointer *:py-0.5
								*:text-left *:hover:text-neutral-950 *:hover:underline dark:text-neutral-300 *:dark:hover:text-neutral-50"
								>
									{#each versions as version}
										<button onclick={() => openContractVersion(version)}>
											{version.version}
										</button>
									{/each}
								</div>
							</div>
						{:else}
							<div
								class="grid auto-rows-[min-content_1fr] overflow-y-auto rounded bg-neutral-50
							p-8 dark:bg-neutral-900/30"
							>
								{@render cardTitle('REVISIONS')}

								<div class="overflow-y-auto">
									<ContractRevisions contractId={contract.id} />
								</div>
							</div>
						{/if}
					</div>
				</div>
			{/if}
		</div>
	{/await}
</FullScreenDialog>

<ConfirmSignContract bind:open={confirmSignDialog.isOpen} close={confirmSignDialog.close} />
<ConfirmRejectContract bind:open={confirmRejectDialog.isOpen} close={confirmRejectDialog.close} />
<RequestRevision bind:open={requestRevisionDialog.isOpen} close={requestRevisionDialog.close} />
