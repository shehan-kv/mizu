import { goto } from '$app/navigation';
import type { USER_ROLES } from '$lib/constants/user';
import {
	APIBadRequestError,
	APIConflictError,
	APIError,
	APIForbiddenError,
	APINotFoundError,
	APIServerError,
	NetworkError
} from './errors';

export type UserRole = (typeof USER_ROLES)[number];

export interface User {
	id: number;
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
	page: number;
	limit: number;
}
export async function getUsers(query: UserQuery, signal?: AbortSignal) {
	const url = new URLSearchParams();

	// set "q" param if q is truthy
	if (query.q) url.set('q', query.q);

	// set "role" param if status is truthy
	if (query.role) url.set('role', query.role);

	url.set('page', query.page.toString());
	url.set('limit', query.limit.toString());

	let res: Response;
	try {
		res = await fetch(`/api/v1/users/?${url.toString()}`, {
			method: 'GET',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to fetch users: ${err}`);
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
				throw new APINotFoundError(`Not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}

	const payload = (await res.json()) as PaginatedResponse<User>;
	return payload;
}

export interface UserCreateParams {
	firstName: string;
	lastName: string;
	email: string;
	title?: string;
	role: string;
	isActive: boolean;
	projects: number[];
}
export async function createUser(req: UserCreateParams, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/users/ `, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(req),
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to create user: ${err}`);
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
				throw new APINotFoundError(`Not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}

export async function activateUser(userId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/users/${userId}/activate`, {
			method: 'PUT',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to activate user: ${err}`);
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
				throw new APINotFoundError(`Not found`);
			case 409:
				throw new APIConflictError(`Already activated`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}

export async function deactivateUser(userId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/users/${userId}/deactivate`, {
			method: 'PUT',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to deactivate user: ${err}`);
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
				throw new APINotFoundError(`Not found`);
			case 409:
				throw new APIConflictError(`Already deactivated`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}

export async function deleteUser(userId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/users/${userId}`, {
			method: 'DELETE',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to delete user: ${err}`);
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
				throw new APINotFoundError(`Not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}

export async function sendUserVerificationEmail(userId: number, signal?: AbortSignal) {
	let res: Response;
	try {
		res = await fetch(`/api/v1/users/${userId}/resend-verification`, {
			method: 'POST',
			signal
		});
	} catch (err) {
		throw new NetworkError(`Failed to resend user verification email: ${err}`);
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
				throw new APINotFoundError(`Not found`);
			case 500:
				throw new APIServerError('Internal server error');
			default:
				throw new APIError(`Unexpected error: ${res.status} ${res.statusText}`, res.status);
		}
	}
}
