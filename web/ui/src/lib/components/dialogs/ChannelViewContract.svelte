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
	import {
		getContractSignatures,
		getContractVersions,
		type Contract,
		type ContractSignature,
		type ContractVersion
	} from '$lib/api/contracts';
	import { formatDate } from '$lib/utils/formatDate';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import ErrorMessage from '../ErrorMessage.svelte';
	import {
		APIBadRequestError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError
	} from '$lib/api/errors';

	interface Props {
		open: Boolean;
		contract: Contract;
	}
	let { open = $bindable(), contract }: Props = $props();

	let activeSidebar: 'ABOUT' | 'REVISION' = $state('ABOUT');

	let signaturesPromise: Promise<ContractSignature[]> | null = $state(null);
	let signatureAbortController: AbortController | null = null;

	async function loadSignatures() {
		if (!selectedVersion) return;

		if (signatureAbortController) {
			signatureAbortController.abort();
		}

		signatureAbortController = new AbortController();

		signaturesPromise = getContractSignatures(selectedVersion.id, signatureAbortController.signal);
	}

	let versionsPromise: Promise<ContractVersion[]> | null = $state(null);
	let versionAbortController: AbortController | null = null;
	let selectedVersion: ContractVersion | null = $state(null);

	function loadVersions() {
		if (versionAbortController) {
			versionAbortController.abort();
		}

		versionAbortController = new AbortController();

		versionsPromise = getContractVersions(contract.id, versionAbortController.signal).then(
			(versions) => {
				if (versions.length > 0) selectedVersion = versions[0];
				loadSignatures();
				return versions;
			}
		);
	}

	function openContractVersion(version: ContractVersion) {
		selectedVersion = version;
	}

	let confirmSignDialog = createDialogState();
	let confirmRejectDialog = createDialogState();
	let requestRevisionDialog = createDialogState();

	$effect(() => {
		if (!open) return;
		loadVersions();
	});
</script>

{#snippet cardTitle(text: string)}
	<p class="mb-4 text-xs text-neutral-500">{text}</p>
{/snippet}

<FullScreenDialog bind:open>
	<div class="container mx-auto grid auto-rows-[min-content_1fr] gap-6 overflow-y-auto">
		<div class="space-y-0.5">
			<p class="text-xs text-neutral-500">Contract</p>
			<p class="font-bold">{contract.name}</p>
		</div>

		{#await versionsPromise}
			<Spinner />
		{:then versions}
			{#if versions && selectedVersion}
				<div class="grid grid-cols-4 gap-2 overflow-hidden">
					<div
						class="col-span-3 grid h-full auto-rows-[min-content_1fr] gap-2 overflow-y-auto rounded border"
					>
						<div
							class="sticky top-0 flex items-center justify-between bg-neutral-100 p-2 px-6 dark:bg-neutral-900"
						>
							{#if selectedVersion}
								<p class="text-xs">
									{selectedVersion.version} - Created At {formatDate(selectedVersion.createdAt)}
								</p>
							{/if}

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
					<div class="grid auto-rows-[min-content_1fr] space-y-2 overflow-y-auto">
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
									{#await signaturesPromise}
										<Spinner />
									{:then signatures}
										{#if signatures && signatures.length > 0}
											{#each signatures as signature}
												<div class="flex gap-1.5">
													<div class="size-10 rounded-full bg-neutral-200/80 dark:bg-neutral-800">
														{#if signature.image}
															<img
																src={signature.image}
																alt={`${signature.firstName} ${signature.lastName} profile picture`}
																class="size-full object-cover"
															/>
														{:else}
															<div
																class="flex size-full items-center justify-center text-xs text-neutral-500"
															>
																{signature.firstName[0]}
															</div>
														{/if}
													</div>
													<div>
														<p class="text-sm">{signature.firstName} {signature.lastName}</p>
														<p class="text-xs text-neutral-500 dark:text-neutral-400">
															{signature.status == 'pending'
																? 'Pending Signature'
																: toTitleCase(signature.status)}
														</p>
													</div>
												</div>
											{/each}
										{:else}
											<ErrorMessage variant="info" text="Signatures Not Found" />
										{/if}
									{:catch err}
										{#if err instanceof APIBadRequestError}
											<ErrorMessage variant="warn" text="Invalid Request" retry={loadSignatures} />
										{:else if err instanceof APIForbiddenError}
											<ErrorMessage
												variant="warn"
												text="You Don't Have Permission To View These Signatures"
												retry={loadSignatures}
											/>
										{:else if err instanceof APINotFoundError}
											<ErrorMessage variant="info" text="Not Found" retry={loadSignatures} />
										{:else if err instanceof APIServerError}
											<ErrorMessage
												variant="warn"
												text="Server Ran Into An Error"
												retry={loadSignatures}
											/>
										{:else}
											<ErrorMessage
												variant="warn"
												text="An Unexpected Error Occured"
												retry={loadSignatures}
											/>
										{/if}
									{/await}
								</div>
							</div>
							<div class="rounded bg-neutral-50 p-8 dark:bg-neutral-900/30">
								{@render cardTitle('SIGN CONTRACT')}

								{#if selectedVersion && selectedVersion.status == 'pending'}
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
								{:else if selectedVersion.status == 'signed'}
									<p class="inline-flex items-center gap-1 text-sm">
										<CheckCircle size={20} class="text-emerald-500" /> You've Already Signed This Contract
									</p>
								{:else if selectedVersion.status == 'rejected'}
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
									<ContractRevisions {contract} />
								</div>
							</div>
						{/if}
					</div>
				</div>
			{:else}
				<ErrorMessage variant="info" text="Versions Not Found" />
			{/if}
		{:catch err}
			{#if err instanceof APIBadRequestError}
				<ErrorMessage variant="warn" text="Invalid Request" retry={loadVersions} />
			{:else if err instanceof APIForbiddenError}
				<ErrorMessage
					variant="warn"
					text="You Don't Have Permission To View These Versions"
					retry={loadVersions}
				/>
			{:else if err instanceof APINotFoundError}
				<ErrorMessage variant="info" text="Not Found" retry={loadVersions} />
			{:else if err instanceof APIServerError}
				<ErrorMessage variant="warn" text="Server Ran Into An Error" retry={loadVersions} />
			{:else}
				<ErrorMessage variant="warn" text="An Unexpected Error Occured" retry={loadVersions} />
			{/if}
		{/await}
	</div>
</FullScreenDialog>

{#if selectedVersion}
	<ConfirmSignContract
		bind:open={confirmSignDialog.isOpen}
		version={selectedVersion}
		onSuccess={loadVersions}
	/>

	<ConfirmRejectContract
		bind:open={confirmRejectDialog.isOpen}
		version={selectedVersion}
		onSuccess={loadVersions}
	/>
{/if}

<RequestRevision bind:open={requestRevisionDialog.isOpen} {contract} onSuccess={loadVersions} />
