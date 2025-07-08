<script>
	import X from 'phosphor-svelte/lib/X';
	import { Dialog } from 'bits-ui';

	let { open = $bindable(), close } = $props();

	// TODO: Implement request revision functionality
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
			data-[state=closed]:slide-out-to-bottom-8 data-[state=closed]:fade-out 
			data-[state=open]:slide-in-from-bottom-8 data-[state=open]:fade-in
			outline-hidden duration-250 
			fixed left-1/2 top-1/2 z-50 grid w-full max-w-xl -translate-x-1/2
			-translate-y-1/2 auto-rows-[min-content_1fr] gap-4 overflow-hidden rounded 
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

			<div class="px-6 pb-6">
				<p class="font-bold">Request a Revision</p>
				<div class="mt-6 space-y-4">
					<div class="space-y-1 text-sm *:block">
						<label for="title">Subject</label>
						<input
							type="text"
							id="title"
							class="w-full rounded border border-neutral-200 bg-neutral-50 p-2
							dark:border-neutral-800 dark:bg-neutral-900"
						/>
					</div>
					<div class="space-y-1 text-sm *:block">
						<label for="description">Description</label>
						<textarea
							id="description"
							class="h-24 w-full resize-none rounded border border-neutral-200
							bg-neutral-50 p-2 dark:border-neutral-800 dark:bg-neutral-900"
						></textarea>
					</div>
				</div>
				<div class="mt-6 space-x-1 text-right text-xs *:cursor-pointer *:rounded *:px-6 *:py-3">
					<button onclick={close} class="hover:bg-neutral-100 dark:hover:bg-neutral-900">
						Cancel
					</button>
					<button
						class="bg-neutral-800 text-neutral-50 transition hover:bg-neutral-950
                    dark:bg-neutral-200 dark:text-neutral-950 dark:hover:bg-neutral-50"
					>
						Request
					</button>
				</div>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
