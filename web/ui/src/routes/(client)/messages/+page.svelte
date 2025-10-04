<script lang="ts">
	import Files from 'phosphor-svelte/lib/Files';
	import FileText from 'phosphor-svelte/lib/FileText';
	import Invoice from 'phosphor-svelte/lib/Invoice';
	import Kanban from 'phosphor-svelte/lib/Kanban';
	import UploadSimple from 'phosphor-svelte/lib/UploadSimple';

	import * as Message from '$lib/components/message';
	import * as Dialog from '$lib/components/dialogs';

	import type { Channel, ChannelMessage, Member, UserMessage } from '$lib/components/message/types';
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import { onMount } from 'svelte';
	import { channelData, memebersData, messageData } from '$lib/components/message/mockData';
	import TextEditor from '$lib/components/TextEditor.svelte';
	import Spinner from '$lib/components/Spinner.svelte';
	import { createDialogState } from '$lib/components/dialogs/createDialogState.svelte';
	import SendButton from '$lib/components/SendButton.svelte';
	import AiSuggestionsButton from '$lib/components/AiSuggestionsButton.svelte';
	import Swap from 'phosphor-svelte/lib/Swap';

	// svelte-ignore non_reactive_update
	let editor: TextEditor | null = null;
	// svelte-ignore non_reactive_update
	let chatWindow: HTMLDivElement | null = null;

	let isAiEnabled = $state(false);
	let loading = $state({
		channels: true,
		members: true,
		messages: true
	});

	let channels = $state<Channel[]>([]);
	let selectedChannel: Channel | null = $state(null);
	let channelMembers: Member[] = $state([]);
	let channelMessages: ChannelMessage[] = $state([]);

	$effect(() => {
		if (!selectedChannel) return;

		loading.members = true;
		channelMembers = [];
		setTimeout(() => {
			channelMembers = memebersData;
			loading.members = false;
		}, 1000);
	});

	$effect(() => {
		if (!selectedChannel) return;

		loading.messages = true;
		channelMessages = [];
		setTimeout(() => {
			channelMessages = messageData;
			loading.messages = false;
		}, 1000);
	});

	$effect(() => {
		if (channelMessages.length > 0) {
			scrollToBottom();
		}
	});

	onMount(() => {
		fetchChannels();
	});

	function fetchChannels() {
		loading.channels = true;
		setTimeout(() => {
			channels = channelData;
			loading.channels = false;
			selectedChannel = channels[0];
		}, 1000);
	}
	function switchChannel(channel: Channel) {
		selectedChannel = channel;
	}

	let fileDialog = createDialogState();
	let contractDialog = createDialogState();
	let invoiceDialog = createDialogState();
	let kanbanDialog = createDialogState();
	let allTicketDialog = createDialogState();
	let newTicketDialog = createDialogState();

	let messageToSend = $state('');

	function appendMessage() {
		let id = channelMessages.slice(-1)[0].id || 0;
		id++;
		let newMessage: UserMessage = {
			id: id,
			date: new Date().toLocaleString(),
			name: 'Shehan',
			message: messageToSend,
			title: 'Founder - Mizu',
			image: '',
			type: 'USER'
		};

		channelMessages.push(newMessage);
		messageToSend = '';
	}

	function scrollToBottom() {
		if (chatWindow) {
			const threshold = 150;
			const position = chatWindow.scrollTop + chatWindow.clientHeight;
			const height = chatWindow.scrollHeight;

			let isUserNearBottom = height - position <= threshold;

			if (isUserNearBottom) {
				chatWindow.lastElementChild?.scrollIntoView({ behavior: 'smooth' });
			}
		}
	}
	function sendMessage(message: string) {
		if (!message) return;
		messageToSend = message;
		appendMessage();
	}
</script>

<svelte:head>
	<title>Messages</title>
</svelte:head>

<div class="grid h-full grid-cols-[20rem_1fr] overflow-y-auto rounded border">
	<div
		class="grid h-full auto-rows-[1fr_1fr_min-content] overflow-hidden bg-neutral-100 dark:bg-neutral-900"
	>
		<div class="grid h-full auto-rows-[min-content_1fr] gap-2 overflow-y-auto p-4">
			<p class="bg-neutral-100 text-xs text-neutral-400 dark:bg-neutral-900 dark:text-neutral-500">
				CHANNELS
			</p>
			{#if loading.channels}
				<Spinner />
			{:else if !loading.channels && selectedChannel}
				<div
					class="text-sm text-neutral-700 *:block *:cursor-pointer *:border-l *:px-2 *:py-0.5 *:transition
				*:hover:text-neutral-950 *:hover:underline dark:text-neutral-300 *:dark:hover:text-neutral-50"
				>
					{#each channels as channel}
						<button
							class:border-sky-500={selectedChannel.id == channel.id}
							onclick={() => switchChannel(channel)}>#{channel.name}</button
						>
					{/each}
				</div>
			{:else}
				<ErrorMessage variant="channel" text="No Channels Found" retry={fetchChannels} />
			{/if}
		</div>

		<div class="grid h-full auto-rows-[min-content_1fr] gap-2 overflow-y-auto border-t p-4">
			<p class="bg-neutral-100 text-xs text-neutral-400 dark:bg-neutral-900 dark:text-neutral-500">
				MEMBERS ({channelMembers.length})
			</p>

			<div class=" space-y-2.5">
				{#if loading.members}
					<Spinner />
				{:else if !loading.members && channelMembers.length > 0}
					{#each channelMembers as member (member)}
						<Message.Member image={member.image} name={member.name} title={member.title} />
					{/each}
				{:else}
					<ErrorMessage variant="user" text="No Members Found" />
				{/if}
			</div>
		</div>

		{#if selectedChannel}
			<div class="border-t p-4 text-sm text-neutral-700 dark:text-neutral-300">
				<div
					class="*:block *:flex *:w-full *:cursor-pointer *:items-center *:gap-2 *:py-1.5 *:text-left
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
				<p class="mt-3 flex items-center gap-2 text-neutral-500">
					<Swap class="size-5" />Change Requests
				</p>
				<div
					class="mt-1.5 border-l pl-4 *:block *:w-full *:cursor-pointer *:py-1
				*:text-left *:hover:text-neutral-950 *:hover:underline *:disabled:text-neutral-400 *:disabled:hover:no-underline
				*:dark:hover:text-neutral-50 *:dark:disabled:text-neutral-600"
				>
					<button onclick={allTicketDialog.open} disabled={!selectedChannel.projectId}>
						All Change Requests
					</button>
					<button onclick={newTicketDialog.open} disabled={!selectedChannel.projectId}>
						New Request
					</button>
				</div>
			</div>
		{:else}
			<div class="border-t p-4">
				<ErrorMessage variant="warn" text="Select Channel To See Options" />
			</div>
		{/if}
	</div>

	{#if loading.messages}
		<Spinner />
	{:else if !loading.messages && selectedChannel}
		<div class="grid auto-rows-[1fr_min-content] overflow-y-auto p-4">
			{#if channelMessages.length > 0}
				<div class="grow overflow-y-auto whitespace-pre-line pb-8" bind:this={chatWindow}>
					{#each channelMessages as message (message.id)}
						{#if message.type == 'USER'}
							<Message.User
								name={message.name}
								title={message.title}
								date={message.date}
								image={message.image}
								message={message.message}
							/>
						{:else if message.type == 'QUOTE'}
							<Message.System variant="invoice" message={message.message} date={message.date} />
						{:else if message.type == 'INVOICE'}
							<Message.System variant="invoice" message={message.message} date={message.date} />
						{:else if message.type == 'FILE_UPLOAD'}
							<Message.System
								variant="file"
								message={message.message}
								download={message.link}
								date={message.date}
							/>
						{/if}
					{/each}
				</div>
			{:else}
				<ErrorMessage variant="message" text="No Messages Yet" />
			{/if}

			<div class="space-y-1">
				<div class="space-y-2 rounded-sm bg-neutral-100 p-3 text-sm dark:bg-neutral-900">
					<div
						class="flex justify-end space-x-1 text-right text-xs *:cursor-pointer *:rounded
					*:p-2 *:hover:bg-neutral-200 *:dark:hover:bg-neutral-800"
					>
						<AiSuggestionsButton onclick={() => (isAiEnabled = !isAiEnabled)} {isAiEnabled} />
						<button class="inline-flex items-center gap-1.5">
							<UploadSimple />Upload File
						</button>
					</div>
					<div class="grid grid-cols-[1fr_min-content]">
						<div class="max-h-15 overflow-y-auto border-b">
							<TextEditor bind:this={editor} autoSuggest={isAiEnabled} onSubmit={sendMessage} />
						</div>
						<SendButton onclick={() => editor?.submit()} />
					</div>
				</div>
			</div>
		</div>
	{:else}
		<div><ErrorMessage variant="warn" text="Select Channel To Send Messages" /></div>
	{/if}
</div>

{#if selectedChannel}
	<Dialog.ChannelFiles
		bind:open={fileDialog.isOpen}
		close={fileDialog.close}
		channel={selectedChannel}
	/>
{/if}

{#if selectedChannel}
	<Dialog.ChannelContracts bind:open={contractDialog.isOpen} channel={selectedChannel} />
{/if}

{#if selectedChannel}
	<Dialog.ChannelInvoices bind:open={invoiceDialog.isOpen} channel={selectedChannel} />
{/if}

{#if selectedChannel}
	<Dialog.ChannelKanban bind:open={kanbanDialog.isOpen} channel={selectedChannel} />
{/if}

{#if selectedChannel?.projectId}
	<Dialog.NewChangeRequest
		bind:open={newTicketDialog.isOpen}
		projectId={selectedChannel.projectId}
	/>
{/if}

<Dialog.ChannelTickets
	bind:open={allTicketDialog.isOpen}
	close={allTicketDialog.close}
	channel={selectedChannel}
/>
