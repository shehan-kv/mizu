<script>
	import { Dialog } from 'bits-ui';
	import XIcon from '../icons/XIcon.svelte';

	let { open = $bindable(), close, children } = $props();
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
			data-[state=closed]:slide-out-to-bottom data-[state=closed]:fade-out
			data-[state=open]:slide-in-from-bottom data-[state=open]:fade-in 
			outline-hidden fixed inset-8 z-50 grid auto-rows-[min-content_1fr] gap-4 overflow-hidden 
			rounded pb-4 duration-500 "
		>
			<div class="text-right">
				<Dialog.Close
					class="cursor-pointer rounded-bl bg-neutral-50 px-4 py-2
					transition duration-150
					hover:bg-neutral-950 hover:text-neutral-50 dark:bg-neutral-900
					hover:dark:bg-neutral-50 hover:dark:text-neutral-950"
				>
					<XIcon class="size-3" />
				</Dialog.Close>
			</div>
			{@render children()}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
