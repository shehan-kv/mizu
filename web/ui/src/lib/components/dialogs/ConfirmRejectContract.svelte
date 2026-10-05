<script lang="ts">
	import WarningCircle from 'phosphor-svelte/lib/WarningCircle';
	import X from 'phosphor-svelte/lib/X';
	import { Dialog } from 'bits-ui';
	import { rejectContract } from '$lib/api/contracts';
	import { toast } from 'svelte-sonner';
	import { ApiError } from '$lib/api/client';

	interface Props {
		open: boolean;
		contractId: string;
		onSuccess?: () => unknown;
	}
	let { open = $bindable(), contractId, onSuccess }: Props = $props();

	let abortController: AbortController | null = null;

	async function handleReject() {
		if (abortController) {
			abortController.abort();
		}

		abortController = new AbortController();
		try {
			rejectContract(contractId, abortController.signal);
			toast.success('Successfully Rejected');
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
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay
			class="data-[state=open]:animate-in data-[state=closed]:animate-out 
			data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed 
			inset-0 z-50 bg-neutral-100/80 dark:bg-black/80"
		/>
		<Dialog.Content
			class="data-[state=open]:animate-in data-[state=closed]:animate-out 
			data-[state=open]:fade-in-0 data-[state=closed]:fade-out-0 
			data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95
			fixed top-1/2 
			left-1/2 z-50 grid w-full max-w-xl -translate-x-1/2 -translate-y-1/2 auto-rows-[min-content_1fr] gap-4 overflow-hidden rounded bg-white outline-hidden 
			duration-250 dark:bg-neutral-950"
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
				<p class="inline-flex items-center gap-1 font-bold">
					<WarningCircle weight="fill" size={18} class="text-red-400 dark:text-red-500" />
					Reject Contract – This Action is Irreversible
				</p>
				<p class="mt-2 text-sm">
					This action cannot be undone or reversed. Please ensure you have reviewed the contract
					carefully before proceeding.
				</p>
				<div class="mt-6 space-x-1 text-right text-xs *:cursor-pointer *:rounded *:px-6 *:py-3">
					<Dialog.Close class="hover:bg-neutral-100 dark:hover:bg-neutral-900">Cancel</Dialog.Close>
					<button
						onclick={handleReject}
						class="bg-neutral-800 text-neutral-50 transition hover:bg-neutral-950
                    dark:bg-neutral-200 dark:text-neutral-950 dark:hover:bg-neutral-50"
					>
						Reject
					</button>
				</div>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
