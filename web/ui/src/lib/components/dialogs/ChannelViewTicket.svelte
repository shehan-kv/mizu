<script lang="ts">
	import FullScreenDialog from './FullScreenDialog.svelte';
	import * as Message from '$lib/components/message';
	import TextEditor from '../TextEditor.svelte';
	import SendButton from '../SendButton.svelte';
	import AiSuggestionsButton from '../AiSuggestionsButton.svelte';

	interface Props {
		open: Boolean;
		close: () => void;
		ticket?: any | null;
	}
	let { open = $bindable(), close, ticket }: Props = $props();

	let isAiEnabled = $state(false);

	$effect(() => console.log(isAiEnabled));
</script>

<FullScreenDialog bind:open {close}>
	<div class="grid auto-rows-[min-content_1fr] gap-6 overflow-y-auto">
		<div class="container mx-auto space-y-0.5">
			<p class="text-xs text-neutral-500">Support Ticket</p>
			<p class="font-bold">{ticket.subject}</p>
		</div>
		<div class="overflow-y-auto">
			<div class="relative">
				<div
					class="before:content-[' '] mx-auto max-w-3xl space-y-20
				before:absolute before:left-1/2 before:top-0 before:-z-10 before:min-h-full
				before:-translate-x-1/2 before:border-l before:border-dashed
				before:border-neutral-400 dark:before:border-neutral-600"
				>
					<div
						class="sticky top-0 flex items-center justify-center gap-2 bg-white px-6 py-4 dark:bg-neutral-950"
					>
						<span class="relative flex size-3">
							<span
								class="absolute inline-flex h-full w-full animate-ping rounded-full
								bg-green-500 opacity-75 dark:bg-green-600"
							>
							</span>
							<span class="relative inline-flex size-3 rounded-full bg-green-500 dark:bg-green-600"
							></span>
						</span>
						<p class="text-sm">In-Progress</p>
					</div>

					<div class="space-y-2 rounded bg-neutral-100 p-4 dark:bg-neutral-900">
						<div class="border-b pb-2">
							<TextEditor
								autoSuggest={isAiEnabled}
								placeholder="Write your reply here"
								onSubmit={() => console.log('submit called!')}
							/>
						</div>
						<div class="flex items-end justify-end text-xs">
							<AiSuggestionsButton
								{isAiEnabled}
								class="cursor-pointer rounded p-2 hover:bg-neutral-200 dark:hover:bg-neutral-800"
								onclick={() => (isAiEnabled = !isAiEnabled)}
							/>
							<SendButton onclick={() => console.log('submit called!')} />
						</div>
					</div>

					<div class="bg-white p-4 dark:bg-neutral-950">
						<div class="space-y-2">
							<Message.Member name="Carlos Hernandez" title="Client" />
							<div class="space-y-1">
								<p class="">
									Lorem ipsum dolor sit, amet consectetur adipisicing elit. Temporibus nesciunt
									necessitatibus officia ipsum harum quos esse! A nesciunt sunt voluptates amet quas
									consequuntur, aut reiciendis impedit dicta error iste suscipit!
								</p>
								<p class="text-xs text-neutral-500">{new Date().toLocaleString()}</p>
							</div>
						</div>
					</div>
					<div class="bg-white p-4 dark:bg-neutral-950">
						<div class="space-y-2">
							<Message.Member name="Carlos Hernandez" title="Client" />
							<div class="space-y-1">
								<p class="">
									Lorem ipsum dolor sit, amet consectetur adipisicing elit. Temporibus nesciunt
									necessitatibus officia ipsum harum quos esse! A nesciunt sunt voluptates amet quas
									consequuntur, aut reiciendis impedit dicta error iste suscipit!
								</p>
								<p class="text-xs text-neutral-500">{new Date().toLocaleString()}</p>
							</div>
						</div>
					</div>
					<div class="bg-white p-4 dark:bg-neutral-950">
						<div class="space-y-2">
							<Message.Member name="Carlos Hernandez" title="Client" />
							<div class="space-y-1">
								<p class="">
									Lorem ipsum dolor sit, amet consectetur adipisicing elit. Temporibus nesciunt
									necessitatibus officia ipsum harum quos esse! A nesciunt sunt voluptates amet quas
									consequuntur, aut reiciendis impedit dicta error iste suscipit!
								</p>
								<p class="text-xs text-neutral-500">{new Date().toLocaleString()}</p>
							</div>
						</div>
					</div>
					<div class="bg-white p-4 dark:bg-neutral-950">
						<div class="space-y-2">
							<Message.Member name="Carlos Hernandez" title="Client" />
							<div class="space-y-1">
								<p class="">
									Lorem ipsum dolor sit, amet consectetur adipisicing elit. Temporibus nesciunt
									necessitatibus officia ipsum harum quos esse! A nesciunt sunt voluptates amet quas
									consequuntur, aut reiciendis impedit dicta error iste suscipit!
								</p>
								<p class="text-xs text-neutral-500">{new Date().toLocaleString()}</p>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</FullScreenDialog>
