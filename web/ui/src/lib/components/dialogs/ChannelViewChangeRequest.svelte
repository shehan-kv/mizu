<script lang="ts">
	import FullScreenDialog from './FullScreenDialog.svelte';
	import * as Message from '$lib/components/message';
	import TextEditor from '../TextEditor.svelte';
	import SendButton from '../SendButton.svelte';
	import AiSuggestionsButton from '../AiSuggestionsButton.svelte';
	import {
		createChangeRequestEntry,
		getChangeRequestDetails,
		type ChangeRequest,
		type ChangeRequestDetails
	} from '$lib/api/changeRequest';
	import Spinner from '../Spinner.svelte';
	import ErrorMessage from '../ErrorMessage.svelte';
	import { formatDate } from '$lib/utils/formatDate';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import Checks from 'phosphor-svelte/lib/Checks';
	import Hourglass from 'phosphor-svelte/lib/Hourglass';
	import { toast } from 'svelte-sonner';
	import {
		APIBadRequestError,
		APIError,
		APIForbiddenError,
		APINotFoundError,
		APIServerError,
		NetworkError
	} from '$lib/api/errors';

	interface Props {
		open: Boolean;
		request: ChangeRequest;
	}
	let { open = $bindable(), request }: Props = $props();

	let isAiEnabled = $state(false);

	let requestPromise: Promise<ChangeRequestDetails> | null = $state(null);

	let loadAbortController: AbortController | null = null;
	let entryAbortController: AbortController | null = null;
	function loadRequest() {
		if (!request.id) {
			return;
		}

		if (loadAbortController) {
			loadAbortController.abort();
		}

		loadAbortController = new AbortController();

		requestPromise = getChangeRequestDetails(request.id, loadAbortController.signal);
	}

	// svelte-ignore non_reactive_update
	let editor: TextEditor | null = null;

	async function createEntry(content: string) {
		if (!content) {
			toast.error('Required Field Missing');
			return;
		}

		try {
			await createChangeRequestEntry(request.id, { content }, entryAbortController?.signal);
			toast.success('Entry Created Successfully');
			loadRequest();
		} catch (error) {
			if (error instanceof APIBadRequestError) toast.error('Invalid Request');
			if (error instanceof APIForbiddenError) toast.error('Not Authorized');
			if (error instanceof APINotFoundError) toast.error('Not Found');
			if (error instanceof APIServerError) toast.error('Server Error');
			if (error instanceof APIError) toast.error('Unexpected Error, Try Again');
			if (error instanceof NetworkError) toast.error('Request Failed, Try Again');
		}
	}

	$effect(() => {
		if (!open) return;
		loadRequest();
	});
</script>

<FullScreenDialog bind:open>
	<div class="grid auto-rows-[min-content_1fr] gap-6 overflow-y-auto">
		<div class="container mx-auto space-y-0.5">
			<p class="text-xs text-neutral-500">Change Request</p>
			<p class="font-bold">{request.title}</p>
		</div>

		{#if !request.id}
			<ErrorMessage variant="warn" text="Request ID Not Found" />
		{:else}
			{#await requestPromise}
				<Spinner />
			{:then res}
				{#if res}
					<div class="overflow-y-auto">
						{#if res.entries.length == 0}
							<ErrorMessage variant="info" text="Change Request Entries Not Found" />
						{/if}
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
									{#if res.status == 'in-progress'}
										<span class="relative flex size-3">
											<span
												class="absolute inline-flex h-full w-full animate-ping rounded-full
									bg-green-500 opacity-75 dark:bg-green-600"
											>
											</span>
											<span
												class="relative inline-flex size-3 rounded-full bg-green-500 dark:bg-green-600"
											>
											</span>
										</span>
									{:else if res.status == 'waiting'}
										<Hourglass size={18} class="text-yellow-500" />
									{:else if res.status == 'closed'}
										<Checks size={18} class="text-emerald-500" />
									{/if}
									<p class="text-sm">
										{#if res.status == 'waiting'}
											Waiting For Client Reply
										{:else}
											{toTitleCase(res.status)}
										{/if}
									</p>
								</div>
								{#if res.status != 'closed'}
									<div class="space-y-2 rounded bg-neutral-100 p-4 dark:bg-neutral-900">
										<div class="border-b pb-2">
											<TextEditor
												bind:this={editor}
												autoSuggest={isAiEnabled}
												placeholder="Write your reply here"
												onSubmit={createEntry}
											/>
										</div>
										<div class="flex items-end justify-end text-xs">
											<AiSuggestionsButton
												{isAiEnabled}
												class="cursor-pointer rounded p-2 hover:bg-neutral-200 dark:hover:bg-neutral-800"
												onclick={() => (isAiEnabled = !isAiEnabled)}
											/>
											<SendButton onclick={() => editor?.submit()} />
										</div>
									</div>
								{/if}

								{#if res.entries.length > 0}
									{#each res.entries as entry}
										<div class="bg-white p-4 dark:bg-neutral-950">
											<div class="space-y-2">
												<Message.Member
													name={`${entry.user.firstName} ${entry.user.lastName}`}
													title={entry.user.title}
													image={entry.user.image}
													role={entry.user.role}
												/>
												<div class="space-y-1">
													<p>{entry.content}</p>
													<p class="text-xs text-neutral-500">{formatDate(entry.createdAt)}</p>
												</div>
											</div>
										</div>
									{/each}
								{/if}
							</div>
						</div>
					</div>
				{/if}
			{/await}
		{/if}
	</div>
</FullScreenDialog>
