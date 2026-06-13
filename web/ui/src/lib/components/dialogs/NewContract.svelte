<script lang="ts">
	import { Dialog } from 'bits-ui';
	import X from 'phosphor-svelte/lib/X';
	import InputLabel from '../InputLabel.svelte';
	import TextEditor from '../TextEditor.svelte';
	import AiSuggestionsButton from '../AiSuggestionsButton.svelte';
	import { toast } from 'svelte-sonner';
	import { createContract } from '$lib/api/contracts';
	import { createDialogState } from './createDialogState.svelte';
	import ConfirmDiscardData from './ConfirmDiscardData.svelte';
	import { ApiError } from '$lib/api/client';
	import UserCard from '../UserCard.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import type { ProjectMember } from '$lib/api/projects';
	import SearchProjectMember from '../SearchProjectMember.svelte';
	import Info from 'phosphor-svelte/lib/Info';

	interface Props {
		open: boolean;
		projectId: string;
		onSuccess?: () => unknown;
	}
	let { open = $bindable(), projectId, onSuccess }: Props = $props();

	let req = $state<{
		name: string;
		terms: string;
		signatories: ProjectMember[];
	}>({
		name: '',
		terms: '',
		signatories: []
	});

	function validateReq() {
		return (
			req.name.trim().length != 0 && req.terms.trim().length != 0 && req.signatories.length != 0
		);
	}
	function resetReq() {
		req.name = '';
		req.terms = '';
		req.signatories = [];
	}

	function addSignatory(member: ProjectMember) {
		const exists = req.signatories.find((m) => m.id == member.id);
		if (!exists) {
			req.signatories.push(member);
			req.signatories = [...req.signatories];
		} else {
			toast.info('Already Added');
		}
	}

	function removeSignatory(member: ProjectMember) {
		req.signatories = req.signatories.filter((m) => m.id != member.id);
	}

	let isAiEnabled = $state(false);

	let createAbort: AbortController | null = $state(null);
	async function handleCreate() {
		if (!validateReq()) {
			toast.error('Missing Required Fields');
			return;
		}

		if (createAbort) {
			createAbort.abort();
		}
		createAbort = new AbortController();

		try {
			await createContract(
				projectId,
				{
					name: req.name,
					terms: req.terms,
					signatoryIds: req.signatories.map((s) => s.id)
				},
				createAbort.signal
			);
			toast.success('Successfully Created');
			resetReq();
			onSuccess?.();
			open = false;
		} catch (error) {
			if (error instanceof ApiError) {
				toast.error(error.message);
			} else {
				toast.error('An Error Occurred');
			}
		}
	}

	let discardDialog = createDialogState();

	function handleDiscard() {
		resetReq();
		discardDialog.close();
		open = false;
	}
</script>

<Dialog.Root
	bind:open
	onOpenChange={(state) => {
		if (!state && (req.name.length > 0 || req.terms.length > 0 || req.signatories.length > 0)) {
			open = true;
			discardDialog.open();
		} else {
			resetReq();
			open = false;
		}
	}}
>
	<Dialog.Portal>
		<Dialog.Overlay
			class="data-[state=open]:animate-in data-[state=closed]:animate-out 
			data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed 
			inset-0 z-50 bg-neutral-100/80 dark:bg-black/80"
		/>
		<Dialog.Content
			class="bg-background data-[state=open]:animate-in data-[state=closed]:animate-out 
			data-[state=closed]:slide-out-to-bottom-8 data-[state=closed]:fade-out
			data-[state=open]:slide-in-from-bottom-8 data-[state=open]:fade-in 
			fixed top-1/2 left-1/2 z-50 grid w-full max-w-4xl -translate-x-1/2 -translate-y-1/2 auto-rows-[min-content_1fr] gap-4 rounded outline-hidden 
			duration-250"
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

			<div class="px-6 pb-6">
				<p class="font-bold">Create New Contract</p>

				<div class="mt-4">
					<div class="grid grid-cols-2 gap-4 text-sm">
						<div>
							<InputLabel htmlFor="name" text="Name" required />
							<input
								type="text"
								id="name"
								bind:value={req.name}
								class="mt-1 block w-full rounded border border-neutral-200 bg-neutral-100 p-2
							outline-hidden dark:border-neutral-800 dark:bg-neutral-900"
							/>
						</div>
					</div>

					<div class="mt-4 space-y-1 text-sm">
						<p>
							Contract Terms <span class="text-xs text-red-600 dark:text-red-500">(Required)</span>
						</p>
						<div
							class="grid h-50 grid-rows-[1fr_min-content] rounded bg-neutral-100 px-3 pt-3 pb-2 dark:bg-neutral-900"
						>
							<TextEditor
								bind:value={req.terms}
								autoSuggest={isAiEnabled}
								placeholder="Write your contract terms here"
							/>
							<div class="pt-2">
								<AiSuggestionsButton
									{isAiEnabled}
									class="cursor-pointer rounded bg-neutral-200 p-2 
									text-xs hover:bg-neutral-300 dark:bg-neutral-950 
									dark:hover:bg-neutral-800"
									onclick={() => (isAiEnabled = !isAiEnabled)}
								/>
							</div>
						</div>
					</div>

					<div class="mt-4 space-y-1 text-sm">
						<p>
							Signatories
							<span class="text-xs text-red-600 dark:text-red-500"> (Required) </span>
						</p>

						<div class="my-4 h-25 space-y-2 overflow-scroll">
							{#if req.signatories.length > 0}
								{#each req.signatories as signatory (signatory.id)}
									<div class="flex items-center justify-between gap-2">
										<UserCard
											image={signatory.image}
											name={`${signatory.firstName} ${signatory.lastName}`}
											role={signatory.role}
											title={signatory.title}
										/>

										<button
											class="cursor-pointer rounded p-2 transition hover:bg-neutral-100
											dark:hover:bg-neutral-900"
											onclick={() => removeSignatory(signatory)}
										>
											<X size={18} />
										</button>
									</div>
								{/each}
							{:else}
								<ErrorMessage text="No Signatories Yet" variant="info" />
							{/if}
						</div>

						<div class="space-y-2">
							<SearchProjectMember {projectId} onSelect={addSignatory} />
							<div class="flex items-start gap-1 text-xs text-neutral-500">
								<Info size={16} />
								<p>At Least 1 Client and 1 Administrator/Staff Member Are Required</p>
							</div>
						</div>
					</div>
				</div>

				<div class="mt-6 space-x-1 text-right text-xs *:cursor-pointer *:rounded *:px-6 *:py-3">
					<Dialog.Close class="hover:bg-neutral-100 dark:hover:bg-neutral-900">Cancel</Dialog.Close>
					<button
						onclick={handleCreate}
						class="bg-neutral-800 text-neutral-50 transition hover:bg-neutral-950
                    dark:bg-neutral-200 dark:text-neutral-950 dark:hover:bg-neutral-50"
					>
						Create
					</button>
				</div>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<ConfirmDiscardData open={discardDialog.isOpen} onDiscard={handleDiscard} />
