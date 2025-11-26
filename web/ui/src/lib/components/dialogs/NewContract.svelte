<script lang="ts">
	import { Dialog } from 'bits-ui';
	import X from 'phosphor-svelte/lib/X';
	import InputLabel from '../InputLabel.svelte';
	import TextEditor from '../TextEditor.svelte';
	import AiSuggestionsButton from '../AiSuggestionsButton.svelte';
	import { toast } from 'svelte-sonner';
	import { createContract } from '$lib/api/contracts';
	import {
		APIBadRequestError,
		APIConflictError,
		APIError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError,
		NetworkError
	} from '$lib/api/errors';
	import { createDialogState } from './createDialogState.svelte';
	import ConfirmDiscardData from './ConfirmDiscardData.svelte';

	interface Props {
		open: boolean;
		projectId: number;
		onSuccess?: () => any;
	}
	let { open = $bindable(), projectId, onSuccess }: Props = $props();

	let req = $state({
		name: '',
		version: '',
		contract: ''
	});

	function validateReq() {
		return (
			req.name.trim().length != 0 &&
			req.version.trim().length != 0 &&
			req.contract.trim().length != 0
		);
	}
	function resetReq() {
		req.name = '';
		req.version = '';
		req.contract = '';
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
			await createContract(projectId, req, createAbort.signal);
			toast.success('Successfully Created');
			resetReq();
			onSuccess && onSuccess();
			open = false;
		} catch (error) {
			if (error instanceof APIBadRequestError) {
				toast.error('Invalid Request');
			} else if (error instanceof APIForbiddenError) {
				toast.error('Not Authorized');
			} else if (error instanceof APINotFoundError) {
				toast.error('Not Found');
			} else if (error instanceof APIConflictError) {
				toast.error('Already Exists');
			} else if (error instanceof APIServerError) {
				toast.error('Server Error');
			} else if (error instanceof APIError) {
				toast.error('Unexpected Error, Try Again');
			} else if (error instanceof NetworkError) {
				toast.error('Request Failed, Try Again');
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
		if (!state && (req.name.length > 0 || req.contract.length > 0 || req.version)) {
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
			outline-hidden duration-250 fixed left-1/2 top-1/2 z-50 grid w-full max-w-4xl -translate-x-1/2 -translate-y-1/2 auto-rows-[min-content_1fr] gap-4 
			rounded"
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
								class="outline-hidden mt-1 block w-full rounded border border-neutral-200 bg-neutral-100
							p-2 dark:border-neutral-800 dark:bg-neutral-900"
							/>
						</div>

						<div>
							<InputLabel htmlFor="version" text="Version" required />
							<input
								type="text"
								id="version"
								placeholder="Example: v1.0.0"
								bind:value={req.version}
								class="outline-hidden mt-1 block w-full rounded border border-neutral-200 bg-neutral-100
							p-2 placeholder:text-xs placeholder:italic dark:border-neutral-800 dark:bg-neutral-900"
							/>
						</div>
					</div>

					<div class="mt-4 space-y-1 text-sm">
						<p>Contract <span class="text-xs text-red-600 dark:text-red-500">(Required)</span></p>
						<div
							class="grid h-60 grid-rows-[1fr_min-content] rounded bg-neutral-100 px-3 pb-2 pt-3 dark:bg-neutral-900"
						>
							<TextEditor
								bind:value={req.contract}
								autoSuggest={isAiEnabled}
								placeholder="Write your contract here"
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
