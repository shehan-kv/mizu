<script lang="ts">
	import { Dialog } from 'bits-ui';
	import WarningCircle from 'phosphor-svelte/lib/WarningCircle';
	import X from 'phosphor-svelte/lib/X';

	interface Props {
		open: boolean;
		onDiscard: () => any;
	}
	let { open = $bindable(), onDiscard }: Props = $props();
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
					<WarningCircle size={18} class="text-red-500" />
					Discard Data
				</p>
				<p class="mt-2 text-sm">
					Are you sure you want to discard any entered data?. This action cannot be undone and all
					unsaved changes will be lost.
				</p>
				<div class="mt-6 space-x-1 text-right text-xs *:cursor-pointer *:rounded *:px-6 *:py-3">
					<Dialog.Close class="hover:bg-neutral-100 dark:hover:bg-neutral-900">Cancel</Dialog.Close>
					<button
						onclick={onDiscard}
						class="cursor-pointer rounded bg-red-300 text-xs text-red-950 transition
						hover:bg-red-400 dark:bg-red-900 dark:text-red-50
						dark:hover:bg-red-800"
					>
						Discard
					</button>
				</div>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
