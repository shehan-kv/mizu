import type { Channel } from '$lib/api/messages';

const channels = $state<Channel[]>([]);

function setChannels(items: Channel[]) {
	channels.splice(0, channels.length, ...items.sort(sortChannels));
}

function updateActivity(channelId: string, updatedAt: Date) {
	const channel = channels.find((c) => c.id === channelId);

	if (!channel) return;

	channel.updatedAt = updatedAt;

	channels.sort(sortChannels);
}

function sortChannels(a: Channel, b: Channel) {
	return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime();
}

export const channelStore = {
	channels,
	setChannels,
	updateActivity
};
