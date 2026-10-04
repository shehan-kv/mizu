<script lang="ts">
	import Files from 'phosphor-svelte/lib/Files';
	import FileText from 'phosphor-svelte/lib/FileText';
	import Invoice from 'phosphor-svelte/lib/Invoice';
	import Kanban from 'phosphor-svelte/lib/Kanban';

	import * as Message from '$lib/components/message';
	import * as Dialog from '$lib/components/dialogs';

	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import { onDestroy, onMount, tick } from 'svelte';
	import TextEditor from '$lib/components/TextEditor.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import { createDialogState } from '$lib/components/dialogs/createDialogState.svelte';
	import SendButton from '$lib/components/SendButton.svelte';
	import AiSuggestionsButton from '$lib/components/AiSuggestionsButton.svelte';
	import {
		createMessage,
		getChannelMembers,
		getChannelMessages,
		getChannels,
		type Channel,
		type ChannelMember
	} from '$lib/api/messages';
	import { messageStore } from '$lib/messages/messageStore.svelte';
	import { channelStore } from '$lib/messages/channelStore.svelte';
	import { toast } from 'svelte-sonner';
	import { ApiError } from '$lib/api/client';
	import { toTitleCase } from '$lib/utils/toTitleCase';
	import FileUploadButton from '$lib/components/FileUploadButton.svelte';
	import { auth } from '$lib/auth/auth.svelte';

	let chatWindow: HTMLDivElement | null = $state(null);

	let loading = $state({
		channels: false,
		members: false,
		messages: false
	});

	let errors = $state<{
		channels: string | null;
		members: string | null;
		messages: string | null;
	}>({
		channels: null,
		members: null,
		messages: null
	});

	let selectedChannel: Channel | null = $state(null);
	let channelMembers: ChannelMember[] = $state([]);

	let membersAbort: AbortController | null = null;
	async function loadMembers() {
		if (!messageStore.state.activeChannelId) return;

		errors.members = null;

		if (membersAbort) {
			membersAbort.abort();
		}
		membersAbort = new AbortController();

		try {
			loading.members = true;
			channelMembers = await getChannelMembers(
				messageStore.state.activeChannelId,
				membersAbort.signal
			);
		} catch (err) {
			if (err instanceof ApiError) {
				errors.members = err.message;
			} else {
				errors.members = 'Failed to load members';
			}
		} finally {
			loading.members = false;
		}
	}

	let channelAbort: AbortController | null = null;
	async function loadChannels() {
		errors.channels = null;

		if (channelAbort) {
			channelAbort.abort();
		}
		channelAbort = new AbortController();

		try {
			loading.channels = true;
			const channels = await getChannels(channelAbort.signal);

			channelStore.setChannels(channels);
		} catch (err) {
			if (err instanceof ApiError) {
				errors.channels = err.message;
			} else {
				errors.channels = 'Failed to load channels';
			}
		} finally {
			loading.channels = false;
		}
	}

	let messageAbort: AbortController | null = null;
	async function loadMessages(channelId: string) {
		errors.messages = null;

		if (!channelId) return;

		if (messageAbort) {
			messageAbort.abort();
		}
		messageAbort = new AbortController();

		try {
			loading.messages = true;

			const messages = await getChannelMessages(channelId, 100, undefined, messageAbort.signal);
			messageStore.replaceMessages(messages);
		} catch (err) {
			if (err instanceof ApiError) {
				errors.messages = err.message;
			} else {
				errors.messages = 'Failed to load messages';
			}
		} finally {
			loading.messages = false;
			await tick();
			scrollToBottom(true);
		}
	}

	let loadingOlder = $state(false);
	let olderAbort: AbortController | null = null;
	async function loadOlderMessages() {
		if (!messageStore.state.activeChannelId) return;
		if (!messageStore.state.oldestMessageId) return;
		if (!messageStore.state.hasMoreOlderMessages) return;
		if (!chatWindow) return;
		if (loadingOlder) return;

		olderAbort?.abort();
		olderAbort = new AbortController();

		loadingOlder = true;

		const previousHeight = chatWindow.scrollHeight;

		try {
			const messages = await getChannelMessages(
				messageStore.state.activeChannelId,
				100,
				messageStore.state.oldestMessageId,
				olderAbort.signal
			);

			if (messages.length < messageStore.state.pageSize) {
				messageStore.state.hasMoreOlderMessages = false;
			}

			if (messages.length > 0) {
				messageStore.prependMessages(messages);

				await tick();

				const newHeight = chatWindow.scrollHeight;
				chatWindow.scrollTop += newHeight - previousHeight;
			}
		} catch (err) {
			if (err instanceof ApiError) {
				toast.error(toTitleCase(err.message));
			} else {
				toast.error('Failed To Load Older Messages');
			}
		} finally {
			loadingOlder = false;
		}
	}

	function handleMessageScroll() {
		if (!chatWindow) return;

		const threshold = 10;

		if (chatWindow.scrollTop <= threshold) {
			loadOlderMessages();
		}
	}

	async function switchChannel(channel: Channel) {
		selectedChannel = channel;

		messageStore.setActiveChannel(channel.id);

		await Promise.all([loadMembers(), loadMessages(channel.id)]);
	}

	onMount(() => {
		loadChannels();
	});

	onDestroy(() => {
		messageStore.resetMessages();
	});

	let fileDialog = createDialogState();
	let contractDialog = createDialogState();
	let invoiceDialog = createDialogState();
	let kanbanDialog = createDialogState();
	let newChannelDialog = createDialogState();
	let manageMembersDialog = createDialogState();

	function scrollToBottom(force = false) {
		const messages = messageStore.state.messages;

		if (!chatWindow || messages.length === 0) return;

		const threshold = 150;

		const position = chatWindow.scrollTop + chatWindow.clientHeight;
		const height = chatWindow.scrollHeight;
		const isNearBottom = height - position <= threshold;

		if (force || isNearBottom) {
			chatWindow.scrollTo({
				top: chatWindow.scrollHeight,
				behavior: 'smooth'
			});
		}
	}

	$effect(() => {
		const count = messageStore.state.messages.length;
		if (count > 0 && chatWindow) {
			tick().then(() => {
				scrollToBottom();
			});
		}
	});

	let messageToSend = $state('');
	async function sendMessage() {
		if (!messageToSend || !messageStore.state.activeChannelId) return;

		try {
			await createMessage(messageStore.state.activeChannelId, { content: messageToSend });
		} catch (error) {
			if (error instanceof ApiError) {
				toast(toTitleCase(error.message));
			} else {
				toast('An Error Occured');
			}
		}
	}
</script>

<svelte:head>
	<title>Messages</title>
</svelte:head>

<div
	class="grid h-full grid-cols-[20rem_1fr] overflow-y-auto rounded bg-neutral-50 dark:bg-neutral-950"
>
	<div class="grid h-full auto-rows-[1fr_1fr_min-content] overflow-hidden border-r">
		<div class="grid h-full auto-rows-[min-content_1fr] gap-2 overflow-y-auto p-4">
			<p class="bg-neutral-50 text-xs text-neutral-500 dark:bg-neutral-950 dark:text-neutral-500">
				CHANNELS
			</p>
			{#if loading.channels}
				<Spinner />
			{:else if channelStore.channels.length == 0}
				<ErrorMessage variant="channel" text="No Channels Found" retry={loadChannels} />
			{:else}
				<div>
					{#each channelStore.channels as channel (channel.id)}
						<button
							class="block w-full cursor-pointer border-l px-2 py-0.5
									text-left text-sm text-neutral-700 transition
									hover:text-neutral-950 hover:underline dark:text-neutral-300 dark:hover:text-neutral-50"
							class:border-sky-500={selectedChannel?.id == channel.id}
							class:border-red-500={messageStore.state.unreadCounts[channel.id] > 0}
							onclick={() => switchChannel(channel)}
						>
							#{channel.name}
						</button>
					{/each}
				</div>
			{/if}

			{#if !loading.channels && errors.channels}
				<ErrorMessage variant="warn" text={errors.channels} retry={loadChannels} />
			{/if}
		</div>

		<div class="grid h-full auto-rows-[min-content_1fr] gap-2 overflow-y-auto border-t p-4">
			<p class="bg-neutral-50 text-xs text-neutral-500 dark:bg-neutral-950 dark:text-neutral-500">
				MEMBERS
			</p>
			{#if loading.members}
				<Spinner />
			{:else if !selectedChannel}
				<ErrorMessage variant="warn" text="Select A Channel To See Members" />
			{:else if channelMembers.length == 0}
				<ErrorMessage variant="user" text="No Members Found" />
			{:else}
				<div class=" space-y-2.5">
					{#each channelMembers as member (member.id)}
						<Message.Member
							id={member.id}
							hasImage={member.hasImage}
							name={`${member.firstName} ${member.lastName}`}
							title={member.title}
							role={member.role}
						/>
					{/each}
				</div>
			{/if}

			{#if !loading.members && errors.members}
				<ErrorMessage variant="warn" text={errors.members} retry={loadChannels} />
			{/if}
		</div>

		{#if selectedChannel}
			<div class="border-t p-4 text-sm text-neutral-700 dark:text-neutral-300">
				<div
					class="*:flex *:w-full *:cursor-pointer *:items-center *:gap-2 *:py-1.5 *:text-left
			*:hover:text-neutral-950 *:hover:underline *:disabled:text-neutral-400 *:disabled:hover:no-underline
			*:dark:hover:text-neutral-50 *:dark:disabled:text-neutral-600"
				>
					<button onclick={fileDialog.open}><Files size={18} />Files</button>
					<button onclick={contractDialog.open} disabled={!selectedChannel.projectId}>
						<FileText size={18} />Contracts
					</button>
					<button onclick={invoiceDialog.open} disabled={!selectedChannel.projectId}>
						<Invoice size={18} />Invoices & Quotes
					</button>
					<button onclick={kanbanDialog.open} disabled={!selectedChannel.projectId}>
						<Kanban size={18} />Kanban Board
					</button>
				</div>
			</div>
		{:else}
			<div class="border-t p-4">
				<ErrorMessage variant="warn" text="Select A Channel To See Options" />
			</div>
		{/if}
	</div>

	<div class="grid auto-rows-[min-content_1fr] overflow-hidden">
		{#if auth.role == 'administrator'}
			<div class="space-x-1 border-b p-2 text-right text-xs">
				<button
					onclick={newChannelDialog.open}
					class="cursor-pointer rounded px-4 py-2 text-neutral-700
				transition hover:bg-neutral-200 dark:text-neutral-300
				hover:dark:bg-neutral-800"
				>
					New Channel
				</button>
				<button
					disabled={!selectedChannel}
					onclick={manageMembersDialog.open}
					class="hover:bg-netural-200 cursor-pointer rounded px-4
				py-2 text-neutral-700 transition hover:bg-neutral-200 disabled:cursor-not-allowed
				dark:text-neutral-300 hover:dark:bg-neutral-800"
				>
					Manage Members
				</button>
			</div>
		{:else}
			<div></div>
		{/if}

		{#if loading.messages}
			<Spinner />
		{:else if !loading.messages && selectedChannel}
			<div class="grid auto-rows-[1fr_min-content] overflow-y-auto px-4 pb-4">
				{#if messageStore.state.messages.length > 0}
					<div
						class="grow overflow-y-auto pb-8 whitespace-pre-line"
						bind:this={chatWindow}
						onscroll={handleMessageScroll}
					>
						{#if loadingOlder}
							<div class="py-2">
								<Spinner />
							</div>
						{/if}

						{#each messageStore.state.messages as message (message.id)}
							{#if !message.isSystem}
								<Message.User
									name={message.sender.firstName + ' ' + message.sender.lastName}
									title={message.sender.title}
									date={message.createdAt}
									hasImage={message.sender.hasImage}
									id={message.sender.id}
									message={message.content}
								/>
							{:else}
								<Message.System date={message.createdAt} message={message.content} />
							{/if}
						{/each}
					</div>
				{:else}
					<ErrorMessage variant="message" text="No Messages Yet" />
				{/if}

				<div class="space-y-1">
					<div class="space-y-2 rounded-sm bg-neutral-100 p-3 text-sm dark:bg-neutral-900">
						<div class="flex justify-end space-x-1 text-right text-xs">
							<AiSuggestionsButton
								class="cursor-pointer rounded p-2 hover:bg-neutral-200 dark:hover:bg-neutral-800"
							/>
							<FileUploadButton channelId={selectedChannel.id} />
						</div>
						<div class="grid grid-cols-[1fr_min-content]">
							<div class="max-h-15 overflow-y-auto border-b">
								<TextEditor bind:value={messageToSend} onSubmit={sendMessage} />
							</div>
							<SendButton onclick={sendMessage} />
						</div>
					</div>
				</div>
			</div>
		{:else}
			<div><ErrorMessage variant="warn" text="Select A Channel To Send Messages" /></div>
		{/if}
	</div>
</div>

{#if selectedChannel}
	<Dialog.ChannelFiles bind:open={fileDialog.isOpen} channel={selectedChannel} />
	<Dialog.ChannelContracts bind:open={contractDialog.isOpen} channel={selectedChannel} />
	<Dialog.ChannelInvoices bind:open={invoiceDialog.isOpen} channel={selectedChannel} />
	<Dialog.ChannelKanban bind:open={kanbanDialog.isOpen} channel={selectedChannel} />
	<Dialog.ManageChannelMembers
		bind:open={manageMembersDialog.isOpen}
		channelId={selectedChannel.id}
		onSuccess={loadMembers}
	/>
{/if}

<Dialog.NewChannel bind:open={newChannelDialog.isOpen} />
