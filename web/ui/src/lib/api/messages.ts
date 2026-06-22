import { apiFetch, BASE_URL } from './client';
import type { PaginatedResponse } from './page';

export interface Channel {
	id: string;
	projectId?: string;
	name: string;
	createdAt: Date;
	updatedAt: Date;
}

export async function getChannelsByMember(memberId: string, signal?: AbortSignal) {
	return apiFetch<Channel[]>(`messages/members/${memberId}/channels`, {
		method: 'GET',
		signal
	});
}

export async function getChannels(signal?: AbortSignal) {
	return apiFetch<Channel[]>(`messages/channels`, {
		method: 'GET',
		signal
	});
}

export interface ChannelMember {
	id: string;
	firstName: string;
	lastName: string;
	hasImage: boolean;
	title?: string;
	role: string;
}

export async function getChannelMembers(channelId: string, signal?: AbortSignal) {
	return apiFetch<ChannelMember[]>(`messages/channels/${channelId}/members`, {
		method: 'GET',
		signal
	});
}

export interface MessageSender {
	id: string;
	firstName: string;
	lastName: string;
	hasImage: boolean;
	title?: string;
	role: string;
}

export interface Message {
	id: string;
	channelId: string;
	sender: {
		id: string;
		firstName: string;
		lastName: string;
		hasImage: boolean;
		title?: string;
		role?: string;
	};
	isSystem: string;
	content: string;
	createdAt: Date;
}

export async function getChannelMessages(
	channelId: string,
	page: number,
	limit: number,
	signal?: AbortSignal
) {
	const params = new URLSearchParams();

	params.set('page', page.toString());
	params.set('limit', limit.toString());

	return apiFetch<PaginatedResponse<Message>>(
		`messages/channels/${channelId}?${params.toString()}`,
		{
			method: 'GET',
			signal
		}
	);
}

export interface ChannelFile {
	id: string;
	channelId: string;
	user: MessageSender;
	name: string;
	mimeType: string;
	size: number;
	uploadedAt: Date;
}

export interface FileQuery {
	q?: string;
	page: number;
	limit: number;
}

export async function getChannelFiles(channelId: string, query: FileQuery, signal?: AbortSignal) {
	const params = new URLSearchParams();

	if (query.q) {
		params.set('q', query.q);
	}

	params.set('page', query.page.toString());
	params.set('limit', query.limit.toString());

	return apiFetch<PaginatedResponse<ChannelFile>>(
		`messages/channels/${channelId}/files?${params.toString()}`,
		{
			method: 'GET',
			signal
		}
	);
}

export async function getProjectFiles(projectId: string, query: FileQuery, signal?: AbortSignal) {
	const params = new URLSearchParams();

	if (query.q) {
		params.set('q', query.q);
	}

	params.set('page', query.page.toString());
	params.set('limit', query.limit.toString());

	return apiFetch<PaginatedResponse<ChannelFile>>(
		`messages/projects/${projectId}/files?${params.toString()}`,
		{
			method: 'GET',
			signal
		}
	);
}

export function downloadChannelFile(fileId: string) {
	window.location.href = `${BASE_URL}/messages/files/${fileId}`;
}

export async function openFileInNewTab(fileId: string) {
	const response = await fetch(`${BASE_URL}messages/files/${fileId}`);

	const blob = await response.blob();
	const url = URL.createObjectURL(blob);

	window.open(url, '_blank', 'noopener,noreferrer');
}

export interface CreateChannelParams {
	projectId?: string;
	name: string;
	memberIds: string[];
}

export function createChannel(req: CreateChannelParams, signal?: AbortSignal) {
	return apiFetch<void>(`messages/channels`, {
		method: 'POST',
		body: JSON.stringify(req),
		signal
	});
}

export interface CreateMessageParams {
	content: string;
}
export function createMessage(channelId: string, req: CreateMessageParams, signal?: AbortSignal) {
	return apiFetch<void>(`messages/${channelId}`, {
		method: 'POST',
		body: JSON.stringify(req),
		signal
	});
}

export function uploadMessageFile(
	channelId: string,
	file: File,
	options?: {
		signal?: AbortSignal;
		onProgress?: (percent: number) => void;
	}
): Promise<void> {
	return new Promise((resolve, reject) => {
		const formData = new FormData();
		formData.append('file', file);

		const xhr = new XMLHttpRequest();

		// Progress
		xhr.upload.onprogress = (e) => {
			if (!e.lengthComputable) return;

			options?.onProgress?.(Math.round((e.loaded / e.total) * 100));
		};

		// Success
		xhr.onload = () => {
			if (xhr.status >= 200 && xhr.status < 300) {
				resolve();
			} else {
				reject(new Error(`Upload failed (${xhr.status})`));
			}
		};

		// Failure
		xhr.onerror = () => {
			reject(new Error('Network error'));
		};

		// Abort
		xhr.onabort = () => {
			reject(new DOMException('Aborted', 'AbortError'));
		};

		// Connect AbortSignal -> xhr.abort()
		if (options?.signal) {
			if (options.signal.aborted) {
				xhr.abort();
				return;
			}

			options.signal.addEventListener('abort', () => xhr.abort(), { once: true });
		}

		xhr.open('POST', `${BASE_URL}/messages/channels/${channelId}/files`);

		xhr.send(formData);
	});
}

export interface ChannelMemberReplaceRequest {
	memberIds: string[];
}

export async function replaceChannelMembers(
	channelId: string,
	req: ChannelMemberReplaceRequest,
	signal: AbortSignal
) {
	return apiFetch<void>(`messages/channels/${channelId}/members`, {
		method: 'PUT',
		body: JSON.stringify(req),
		signal
	});
}
