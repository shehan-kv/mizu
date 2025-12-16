<script lang="ts">
	import {
		APIBadRequestError,
		APIConflictError,
		APIError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError,
		NetworkError
	} from '$lib/api/errors';
	import {
		markProjectCancelled,
		markProjectCompleted,
		markProjectPaused,
		markProjectStarted,
		type ProjectTaskStatus
	} from '$lib/api/projects';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import { Dialog } from 'bits-ui';
	import WarningCircle from 'phosphor-svelte/lib/WarningCircle';
	import X from 'phosphor-svelte/lib/X';
	import { toast } from 'svelte-sonner';

	interface Props {
		open: boolean;
		taskId: number;
		status: ProjectTaskStatus;
		onSuccess?: () => any;
	}

	let { open = $bindable(), taskId, status, onSuccess }: Props = $props();

	let changeAbort: AbortController | null = null;
	async function handleStatusChange() {
		if (changeAbort) {
			changeAbort.abort();
		}

		changeAbort = new AbortController();

		try {
			switch (status) {
				case 'backlog':
					await markProjectStarted(taskId, changeAbort.signal);
					break;
				case 'in-progress':
					await markProjectPaused(taskId, changeAbort.signal);
					break;
				case 'completed':
					await markProjectCompleted(taskId, changeAbort.signal);
					break;
				default:
					break;
			}

			toast.success(`Successfully ${toTitleCase(status)}`);
			onSuccess && onSuccess();
			open = false;
		} catch (error) {
			if (error instanceof APIBadRequestError) {
				toast.error('Invalid Request');
			} else if (error instanceof APIForbiddenError) {
				toast.error('Not Authorized');
			} else if (error instanceof APINotFoundError) {
				toast.error('Not Found');
			} else if (error instanceof APIServerError) {
				toast.error('Server Error');
			} else if (error instanceof APIConflictError) {
				toast.error(`Already In ${toTitleCase(status)}`);
			} else if (error instanceof APIError) {
				toast.error('Unexpected Error, Try Again');
			} else if (error instanceof NetworkError) {
				toast.error('Request Failed, Try Again');
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
			class="bg-background data-[state=open]:animate-in data-[state=closed]:animate-out 
			data-[state=closed]:slide-out-to-bottom-8 data-[state=closed]:fade-out
			data-[state=open]:slide-in-from-bottom-8 data-[state=open]:fade-in 
			outline-hidden duration-250 fixed left-1/2 top-1/2 z-50 grid w-full max-w-xl -translate-x-1/2 -translate-y-1/2 auto-rows-[min-content_1fr] gap-4 
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
				<p class="inline-flex items-center gap-1 font-bold">
					<WarningCircle weight="fill" size={18} class="text-red-400 dark:text-red-500" />
					Move To {toTitleCase(status)}
				</p>
				<p class="mt-2 text-sm">
					Please ensure you have reviewed the task carefully before proceeding.
				</p>
				<div class="mt-6 space-x-1 text-right text-xs *:cursor-pointer *:rounded *:px-6 *:py-3">
					<Dialog.Close class="hover:bg-neutral-100 dark:hover:bg-neutral-900">Cancel</Dialog.Close>
					<button
						onclick={handleStatusChange}
						class="bg-neutral-800 text-neutral-50 transition hover:bg-neutral-950
                    dark:bg-neutral-200 dark:text-neutral-950 dark:hover:bg-neutral-50"
					>
						Move
					</button>
				</div>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
