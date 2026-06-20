import type { USER_ROLES } from '$lib/constants/user';
import { apiFetch } from './client';
import type { PaginatedResponse } from './page';

export type UserRole = (typeof USER_ROLES)[number];

export interface UserCreateParams {
	firstName: string;
	lastName: string;
	email: string;
	title?: string;
	role: string;
	isActive: boolean;
	projectIds: string[];
}

export function createUser(req: UserCreateParams, signal?: AbortSignal) {
	return apiFetch<void>('users', {
		method: 'POST',
		body: JSON.stringify(req),
		signal
	});
}

export interface User {
	id: string;
	firstName: string;
	lastName: string;
	title?: string;
	email: string;
	role: UserRole;
	image?: string;
	createdAt: Date;
	lastLogin?: Date;
	isActive: boolean;
	isVerified: boolean;
}

export interface UserQuery {
	q?: string;
	role?: string;
	active?: string;
	verified?: string;
	page: number;
	limit: number;
}

export async function getUsers(query: UserQuery, signal?: AbortSignal) {
	const params = new URLSearchParams();

	if (query.q) params.set('q', query.q);

	if (query.role) {
		params.set('role', query.role);
	}

	if (query.active == 'active') {
		params.set('isActive', 'true');
	} else if (query.active == 'disabled') {
		params.set('isActive', 'false');
	}

	if (query.verified == 'verified') {
		params.set('isVerified', 'true');
	} else if (query.verified == 'pending') {
		params.set('isVerified', 'false');
	}

	params.set('page', query.page.toString());
	params.set('limit', query.limit.toString());

	return apiFetch<PaginatedResponse<User>>(`users?${params.toString()}`, {
		method: 'GET',
		signal
	});
}

export async function getMe(signal?: AbortSignal) {
	return apiFetch<User>('users/me', {
		method: 'GET',
		signal
	});
}

export async function activateUser(userId: string, signal?: AbortSignal) {
	return apiFetch<void>(`users/${userId}/activate`, {
		method: 'PUT',
		signal
	});
}

export async function deactivateUser(userId: string, signal?: AbortSignal) {
	return apiFetch<void>(`users/${userId}/deactivate`, {
		method: 'PUT',
		signal
	});
}

export async function getUser(userId: string, signal?: AbortSignal) {
	return apiFetch<User>(`users/${userId}`, {
		method: 'GET',
		signal
	});
}

export async function deleteUser(userId: string, signal?: AbortSignal) {
	return apiFetch<void>(`users/${userId}`, {
		method: 'DELETE',
		signal
	});
}

export async function regenerateUserVerification(userId: string, signal?: AbortSignal) {
	return apiFetch<void>(`users/verifications/${userId}/regenerate-verification`, {
		method: 'POST',
		signal
	});
}

export interface VerificationConfirmParams {
	password: string;
	confirmPassword: string;
}

export async function verifyUser(
	verificationId: string,
	req: VerificationConfirmParams,
	signal?: AbortSignal
) {
	return apiFetch<void>(`users/verifications/${verificationId}/confirm`, {
		method: 'POST',
		body: JSON.stringify(req),
		signal
	});
}

export async function validateUserVerification(verificationId: string, signal?: AbortSignal) {
	return apiFetch<void>(`users/verifications/${verificationId}`, {
		method: 'GET',
		signal
	});
}

export interface SignInParams {
	email: string;
	password: string;
	rememberMe: boolean;
}

export interface SignInResponse {
	status: string;
	role: string;
}
export function signIn(req: SignInParams) {
	return apiFetch<SignInResponse>('users/auth/sign-in', {
		method: 'POST',
		body: JSON.stringify(req)
	});
}

export function signOut() {
	return apiFetch<void>('users/auth/sign-out', {
		method: 'POST'
	});
}
