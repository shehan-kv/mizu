import { apiFetch } from './client';
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
	image?: string;
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
	image?: string;
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
		image?: string;
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
	return apiFetch<PaginatedResponse<ChannelFile>>(`messages/channels/${channelId}/files`, {
		method: 'GET',
		signal
	});
}

export function downloadChannelFile(fileId: string) {
	window.location.href = `/messages/files/${fileId}`;
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

export async function uploadMessageFile(channelId: string, file: File, signal?: AbortSignal) {
	const formData = new FormData();

	formData.append('file', file);

	return apiFetch<void>(`/messages/channels/${channelId}/files`, {
		method: 'POST',
		body: formData,
		signal
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
