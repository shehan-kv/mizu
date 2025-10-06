import { goto } from '$app/navigation';
import {
	APIBadRequestError,
	APIError,
	APIForbiddenError,
	APINotFoundError,
	APIServerError,
	NetworkError
} from './errors';

export interface Channel {
	id: number;
	projectId?: number;
	name: string;
	createdAt: Date;
}

export async function getChannels(signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch('/api/v1/messages/channels', {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch channels: ${err}`);
	}

	if (!res.ok) {
		switch (res.status) {
			case 400:
				throw new APIBadRequestError('Bad request');
			case 401:
				goto('/sign-in');
			case 403:
				throw new APIForbiddenError('Forbidden');
			case 404:
				throw new APINotFoundError(`Channels not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as Channel[];
	return payload;
}

export interface ChannelMember {
	id: number;
	firstName: string;
	lastName: string;
	image?: string;
	title?: string;
	role?: string;
}

export async function getMembersByChannel(channelId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/messages/members/${channelId}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch channels: ${err}`);
	}

	if (!res.ok) {
		switch (res.status) {
			case 400:
				throw new APIBadRequestError('Bad request');
			case 401:
				goto('/sign-in');
			case 403:
				throw new APIForbiddenError('Forbidden');
			case 404:
				throw new APINotFoundError(`Channels not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as ChannelMember[];
	return payload;
}
