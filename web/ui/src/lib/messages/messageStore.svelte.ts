import type { Message } from '$lib/api/messages';

type MessageState = {
	activeChannelId: string | null;

	messages: Message[];

	unreadCounts: Record<string, number>;

	oldestLoadedPage: number;

	pageSize: number;
};

const state = $state<MessageState>({
	activeChannelId: null,

	messages: [],

	unreadCounts: {},

	oldestLoadedPage: 1,

	pageSize: 100
});

function setActiveChannel(channelId: string) {
	state.activeChannelId = channelId;

	clearUnread(channelId);
}

function replaceMessages(messages: Message[]) {
	state.messages = messages;
}

function prependMessages(messages: Message[]) {
	state.messages = [...messages, ...state.messages];
}

function appendMessage(message: Message) {
	const exists = state.messages.some((m) => m.id === message.id);

	if (exists) return;

	state.messages.push(message);
}

function incrementUnread(channelId: string) {
	state.unreadCounts[channelId] = (state.unreadCounts[channelId] ?? 0) + 1;
}

function clearUnread(channelId: string) {
	state.unreadCounts[channelId] = 0;
}

function resetPagination(pageSize = 100) {
	state.oldestLoadedPage = 1;
	state.pageSize = pageSize;
}

function incrementLoadedPage() {
	state.oldestLoadedPage++;
}

export const messageStore = {
	state,

	setActiveChannel,
	replaceMessages,
	prependMessages,
	appendMessage,
	incrementUnread,
	clearUnread,
	resetPagination,
	incrementLoadedPage
};
