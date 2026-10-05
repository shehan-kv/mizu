import type { Message } from '$lib/api/messages';

type MessageState = {
	activeChannelId: string | null;

	messages: Message[];

	unreadCounts: Record<string, number>;

	oldestMessageId: string | null;

	pageSize: number;

	hasMoreOlderMessages: boolean;
};

const state = $state<MessageState>({
	activeChannelId: null,

	messages: [],

	unreadCounts: {},

	oldestMessageId: null,

	pageSize: 100,

	hasMoreOlderMessages: true
});

function setActiveChannel(channelId: string) {
	state.activeChannelId = channelId;

	clearUnread(channelId);
}

function replaceMessages(messages: Message[]) {
	state.messages = messages;

	state.hasMoreOlderMessages = true;

	if (messages.length > 0) {
		state.oldestMessageId = messages[0].id;
	} else {
		state.oldestMessageId = null;
	}
}

function prependMessages(messages: Message[]) {
	state.messages = [...messages, ...state.messages];

	if (state.messages.length > 0) {
		state.oldestMessageId = state.messages[0].id;
	} else {
		state.oldestMessageId = null;
	}
}

function appendMessage(message: Message) {
	const last = state.messages[state.messages.length - 1];

	// prevent duplicates
	if (state.messages.some((m) => m.id === message.id)) return;

	if (!last || new Date(message.createdAt) >= new Date(last.createdAt)) {
		state.messages.push(message);
	} else {
		state.messages = [...state.messages, message].sort(
			(a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
		);
	}
}

function incrementUnread(channelId: string) {
	state.unreadCounts[channelId] = (state.unreadCounts[channelId] ?? 0) + 1;
}

function clearUnread(channelId: string) {
	state.unreadCounts[channelId] = 0;
}

function resetMessages() {
	state.activeChannelId = null;
	state.messages = [];
	state.oldestMessageId = null;
	state.hasMoreOlderMessages = true;
}

export const messageStore = {
	state,

	setActiveChannel,
	replaceMessages,
	prependMessages,
	appendMessage,
	incrementUnread,
	clearUnread,
	resetMessages
};
